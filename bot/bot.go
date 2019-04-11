package bot

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/auyer/massmoverbot/db"
	"github.com/auyer/massmoverbot/mover"
	"github.com/auyer/massmoverbot/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/dgraph-io/badger"
)

var commander *discordgo.Session
var servantList []*discordgo.Session
var conn *badger.DB
var botPrefix string
var messages map[string]map[string]string

// RegEx used to split all command parameters, considering anything between quotes as a single parameter.
// Ex: `> move ThisChannel "That Channel"` will be processed as [">", "move", "ThisChannel", "That Channel"]
var commandRegEx, _ = regexp.Compile(`(".*?"|\S+)`)

// Close function ends the bot connection and closes its database
func Close() {
	log.Println("Closing")
	commander.Close()
	for _, servant := range servantList {
		servant.Close()
	}
}

func setupBot(bot *discordgo.Session) (string, error) {

	bot.AddHandler(ready)
	u, err := commander.User("@me")
	if err != nil {
		log.Println("Error creating Discord session: ", err)
		return "", err
	}
	return u.ID, nil
}

// Start function connects and ads the necessary handlers
func Start(commanderToken string, servantTokens []string, prefix string, DBConnection *badger.DB, botMessages map[string]map[string]string) {
	messages = botMessages
	botPrefix = prefix
	conn = DBConnection
	var err error
	commander, err = discordgo.New("Bot " + commanderToken)
	for _, servantToken := range servantTokens {
		servant, err := discordgo.New("Bot " + servantToken)
		if err != nil {
			log.Println("Error creating Discord session: ", err)
			return
		}
		setupBot(servant)
		servantList = append(servantList, servant)

	}

	_, err = setupBot(commander)
	commander.AddHandler(guildCreate)
	commander.AddHandler(guildDelete)
	if err != nil {
		log.Println("Error creating Discord session: ", err)
		return
	}

	commander.AddHandler(messageHandler)

	for _, s := range servantList {
		err = s.Open()
		if err != nil {
			fmt.Println("Error Opening Bot: ", err)
		}
	}

	err = commander.Open()
	if err != nil {
		log.Println("Error creating Discord session: ", err)
		return
	}
	log.Println("Bot is running!")

}

func ready(s *discordgo.Session, event *discordgo.Ready) {
	// Set the playing status.
	s.UpdateStatus(0, botPrefix+" help")
}

// This function will be called (due to AddHandler above) every time a new
// guild is joined.
func guildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
	if event.Guild.Unavailable {
		return
	}
	log.Println("Joined " + event.Guild.Name + " (" + event.Guild.ID + ")" + " in " + event.Guild.Region)

	val, err := db.GetDataTuple(conn, "M:"+event.Guild.ID)
	if err != nil {
		if err == badger.ErrKeyNotFound || val == "" {
			err = askMember(s, event.Guild.OwnerID, fmt.Sprintf(messages["LANG"]["WelcomeAndLang"], botPrefix, botPrefix))
			if err != nil {
				log.Println("Failed to send mesage to owner.")
				return
			}
			db.UpdateDataTuple(conn, "M:"+event.Guild.ID, "1")
		}
	}
}

// guildDelete function will be called every time the bot leaves a guild.
func guildDelete(s *discordgo.Session, event *discordgo.GuildDelete) {
	if event.Guild.Unavailable {
		return
	}
	log.Println("Left " + event.Guild.Name + " (" + event.Guild.ID + ")")
	_, err := db.GetDataTuple(conn, event.Guild.ID)
	if err == nil {
		db.DeleteDataTuple(conn, event.Guild.ID)
	}
}

// askMember function is used to send a private message to a guild member
func askMember(s *discordgo.Session, owner string, message string) error {
	c, err := s.UserChannelCreate(owner)
	if err != nil {
		fmt.Println(err)
		return err
	}
	_, err = s.ChannelMessageSend(c.ID, message) // event.Guild.OwnerID
	return err
}

// messageHandler function will be called when the bot reads a message
func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	langg, err := db.GetDataTuple(conn, m.GuildID)
	if err != nil {
		langg = "EN"
	}

	// Is this message from a human && Does the message have the bot prefix?
	if !m.Author.Bot && strings.HasPrefix(m.Content, botPrefix) {

		// Split params using regex
		params := commandRegEx.FindAllString(m.Content[1:], -1)
		numParams := len(params)

		// If no parameter was passed, show the help message
		if numParams == 0 {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(messages[langg]["GeneralHelp"], m.Author.Mention(), botPrefix))
			return
		}

		switch params[0] {
		case "lang":
			if numParams == 2 {
				chosenLang := utils.SelectLang(params[2])
				db.UpdateDataTuple(conn, m.GuildID, chosenLang)
				langg = chosenLang
				s.ChannelMessageSend(m.ChannelID, messages[langg]["LangSet"])
			} else {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(messages["LANG"]["LangSetupMessage"], botPrefix, botPrefix))
			}

		case "help":
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(messages[langg]["HelpMessage"], botPrefix, botPrefix))

		case "move":
			workerschann := make(chan []*discordgo.Session, 1)
			go utils.DetectServants(m.GuildID, append(servantList, s), workerschann)

			guild, err := s.Guild(m.GuildID) // retrieving the server (guild) the message was originated from
			if err != nil {
				log.Println(err)
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(messages[langg]["NotInGuild"], m.Author.Mention()))
				return
			}

			channs := guild.Channels // retrieving the list of channels and sorting (next line) them by position (in the users interface)
			sort.Slice(channs[:], func(i, j int) bool {
				return channs[i].Position < channs[j].Position
			})

			if numParams == 2 {
				log.Println("Received move command with 2 parameters on " + guild.Name + " , ID: " + guild.ID + " , by :" + m.Author.ID)
				destination, err := utils.GetChannel(channs, params[1])
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(messages[langg]["CantFindChannel"], params[1]))
				}

				if !utils.CheckPermissions(s, destination, m.Author.ID, discordgo.PermissionVoiceMoveMembers) {
					s.ChannelMessageSend(m.ChannelID, messages[langg]["NoPermissionsDestination"])
					return
				}

				num, err := mover.MoveDestination(s, <-workerschann, m, guild, botPrefix, destination)
				if err != nil {
					if err.Error() == "no permission origin" {
						s.ChannelMessageSend(m.ChannelID, messages[langg]["NoPermissionsOrigin"])
					} else if err.Error() == "cant find user" {
						s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(messages[langg]["CantFindUser"], m.Author.Mention(), botPrefix))
					} else {
						s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(messages[langg]["SorryBut"], err.Error()))
					}
					return
				}

				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(messages[langg]["JustMoved"], num))
				bumpStatistics(num, s, conn)
				return
			} else if numParams == 3 {
				log.Println("Received move command with 3 parameters on " + guild.Name + " , ID: " + guild.ID + " , by :" + m.Author.ID)
				origin, err := utils.GetChannel(channs, params[2])
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(messages[langg]["CantFindChannel"], params[2]))
					return
				}

				destination, err := utils.GetChannel(channs, params[3])
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(messages[langg]["CantFindChannel"], params[3]))
					return
				}

				if !utils.CheckPermissions(s, origin, m.Author.ID, discordgo.PermissionVoiceMoveMembers) {
					s.ChannelMessageSend(m.ChannelID, messages[langg]["NoPermissionsOrigin"])
					return
				}

				if !utils.CheckPermissions(s, destination, m.Author.ID, discordgo.PermissionVoiceMoveMembers) {
					s.ChannelMessageSend(m.ChannelID, messages[langg]["NoPermissionsDestination"])
					return
				}

				num, err := mover.MoveOriginDestination(s, <-workerschann, m, guild, botPrefix, origin, destination)
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(messages[langg]["JustMoved"], err.Error()))
					log.Println(err.Error())
					return
				}

				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(messages[langg]["JustMoved"], num))
				go bumpStatistics(num, s, conn)
			} else {
				log.Println("Received move command with " + strconv.Itoa(numParams) + " parameter(s) (help message) on " + guild.Name + " , ID: " + guild.ID)
				s.ChannelMessageSend(m.ChannelID, mover.MoveHelper(channs, messages[langg]["MoveHelper"], botPrefix))
			}

		default:
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(messages[langg]["EhhMessage"], m.Author.Mention(), m.Content, botPrefix))
		}
	}
}

// GetAndInitStats prints the current statistics, and set them up if there are none.
func GetAndInitStats(conn *badger.DB) {
	bytesStats, err := db.GetDataTupleBytes(conn, "statistics")
	stats := map[string]int{}
	if err != nil {
		log.Println("Failed to get Statistics")
		stats["usrs"] = 0
		stats["movs"] = 0
		bytesStats, _ = json.Marshal(stats)
		_ = db.UpdateDataTupleBytes(conn, "statistics", bytesStats)
		return
	}
	// stats := map[string]string{}
	err = json.Unmarshal(bytesStats, &stats)
	if err != nil {
		log.Println("Failed to decode Statistics")
		return
	}
	log.Println(fmt.Sprintf("Moved %d players in %d actions", stats["usrs"], stats["movs"]))
}

// bumpStatistics ads 1 to the "movs" stats and 'moved' to the "movd"
func bumpStatistics(moved string, s *discordgo.Session, conn *badger.DB) {
	bytesStats, err := db.GetDataTupleBytes(conn, "statistics")
	if err != nil {
		log.Println("Failed to get Statistics")
		return
	}
	stats := map[string]int{}
	err = json.Unmarshal(bytesStats, &stats)
	if err != nil {
		log.Println("Failed to decode Statistics")
		return
	}
	movedInt, _ := strconv.Atoi(moved)
	stats["usrs"] += movedInt
	s.UpdateStatus(0, fmt.Sprintf("Moved %d players \n ! %s help", stats["usrs"], botPrefix))
	stats["movs"]++
	bytesStats, _ = json.Marshal(stats)
	err = db.UpdateDataTupleBytes(conn, "statistics", bytesStats)
	if err != nil {
		log.Println(err)
		log.Println(stats)
	}
	return
}
