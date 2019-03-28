package bot

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/auyer/commanderBot/db"
	"github.com/auyer/commanderBot/mover"
	"github.com/auyer/commanderBot/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/dgraph-io/badger"
)

var commander *discordgo.Session
var servantList []*discordgo.Session
var conn *badger.DB
var botPrefix string
var messages map[string]map[string]string

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
		// servant.AddHandler(servantTest)
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
	if strings.HasPrefix(m.Content, botPrefix+" lang") {
		params := strings.Split(m.Content, " ") // spliting the user request
		if len(params) == 3 {
			chosenLang := utils.SelectLang(params[2])
			db.UpdateDataTuple(conn, m.GuildID, chosenLang)
			langg = chosenLang
			s.ChannelMessageSend(m.ChannelID, messages[langg]["LangSet"])
			return
		}
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(messages["LANG"]["LangSetupMessage"], botPrefix, botPrefix))
		return
	}
	if strings.HasPrefix(m.Content, botPrefix) {
		if m.Author.Bot {
			return
		}
		if m.Content == botPrefix {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(messages[langg]["GeneralHelp"], m.Author.Mention(), botPrefix))
		} else if m.Content == botPrefix+" help" {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(messages[langg]["HelpMessage"], botPrefix, botPrefix))
		} else if strings.HasPrefix(m.Content, botPrefix+" move") {
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
			params := strings.Split(m.Content, " ") // spliting the user request
			length := len(params)
			if length == 3 {
				log.Println("Received 3 parameter move command on " + guild.Name + " , ID: " + guild.ID + " , by :" + m.Author.ID)
				destination, err := utils.GetChannel(channs, params[2])
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(messages[langg]["CantFindChannel"], params[2]))
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
				go bumpStatistics(conn)
				return
			} else if length == 4 {
				log.Println("Received 4 parameter move command on " + guild.Name + " , ID: " + guild.ID + " , by :" + m.Author.ID)
				origin, err := utils.GetChannel(channs, params[2])
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(messages[langg]["CantFindChannel"], params[2]))
				}
				destination, err := utils.GetChannel(channs, params[3])
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(messages[langg]["CantFindChannel"], params[3]))
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
				go bumpStatistics(conn)
				return
			}
			log.Println("Sending help message on " + guild.Name + " , ID: " + guild.ID)
			s.ChannelMessageSend(m.ChannelID, mover.MoveHelper(channs, messages[langg]["MoveHelper"], botPrefix))
		} else {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(messages[langg]["EhhMessage"], m.Author.Mention(), m.Content, botPrefix))
		}
	}
}

// Use this in bot status ?
func bumpStatistics(conn *badger.DB) {
	sts, err := db.GetDataTuple(conn, "statistics")
	if err == nil {
		stsInt, err := strconv.Atoi(sts)
		if err != nil {
			db.UpdateDataTuple(conn, "statistics", "1")
			return
		}
		db.UpdateDataTuple(conn, "statistics", strconv.Itoa(stsInt+1))
		return
	}
	db.UpdateDataTuple(conn, "statistics", "1")
	return
}
