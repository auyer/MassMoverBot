package bot

import (
	"encoding/json"
	"fmt"
	"github.com/auyer/massmoverbot/commands"
	"github.com/auyer/massmoverbot/config"
	"github.com/auyer/massmoverbot/locale"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/auyer/massmoverbot/db"
	"github.com/auyer/massmoverbot/mover"
	"github.com/auyer/massmoverbot/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/dgraph-io/badger"
)

var commander *discordgo.Session
var botPrefix string

// RegEx used to split all command parameters, considering anything between quotes as a single parameter.
// Ex: `> move ThisChannel "That Channel"` will be processed as [">", "move", "ThisChannel", "That Channel"]
var commandRegEx, _ = regexp.Compile(`(".*?"|\S+)`)

// RegEx used to remove starting and ending quotes from the parameters
var parameterQuotesRegEx, _ = regexp.Compile(`(^"|"$)`)

// Close function ends the bot connection and closes its database
func Close() {
	log.Println("Closing")
	_ = commander.Close()
	for _, servant := range config.ServantList {
		_ = servant.Close()
	}
}

func setupBot(bot *discordgo.Session) error {
	bot.AddHandler(ready)

	_, err := commander.User("@me")
	if err != nil {
		return err
	}

	return nil
}

// Start function connects and ads the necessary handlers
func Start() error {
	servantTokens := config.Config.ServantTokens
	botPrefix = config.Config.BotPrefix

	var err error
	commander, err = discordgo.New("Bot " + config.Config.CommanderToken)
	if err != nil {
		log.Println("Error creating main session: ", err)
		return err
	}

	err = setupBot(commander)
	if err != nil {
		log.Println("Error setting up main session: ", err)
		return err
	}

	commander.AddHandler(guildCreate)
	commander.AddHandler(guildDelete)
	commander.AddHandler(messageHandler)

	err = commander.Open()
	if err != nil {
		log.Println("Error opening main Discord session: ", err)
		return err
	}

	for _, servantToken := range servantTokens {
		servant, err := discordgo.New("Bot " + servantToken)
		if err != nil {
			log.Println("Error creating PowerUp session: ", err)
			continue
		}

		err = setupBot(servant)
		if err != nil {
			log.Println("Error setting powerup session: ", err)
			continue
		}

		err = servant.Open()
		if err != nil {
			fmt.Println("Error Opening powerup session: ", err)
			continue
		}

		config.ServantList = append(config.ServantList, servant)
	}

	log.Println("Bot is running!")
	return nil
}

func ready(s *discordgo.Session, event *discordgo.Ready) {
	// Set the playing status.
	_ = s.UpdateStatus(0, botPrefix+" help")
}

// This function will be called (due to AddHandler above) every time a new
// guild is joined.
func guildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
	if event.Guild.Unavailable {
		return
	}
	log.Println("Joined " + event.Guild.Name + " (" + event.Guild.ID + ")" + " in " + event.Guild.Region)

	val, err := db.GetDataTuple(config.Conn, "M:"+event.Guild.ID)
	if err != nil {
		if err == badger.ErrKeyNotFound || val == "" {
			err = askMember(s, event.Guild.OwnerID, fmt.Sprintf(locale.Messages["LANG"]["WelcomeAndLang"], botPrefix, botPrefix))
			if err != nil {
				log.Println("Failed to send message to owner.")
				return
			}
			_ = db.UpdateDataTuple(config.Conn, "M:"+event.Guild.ID, "1")
		}
	}
}

// guildDelete function will be called every time the bot leaves a guild.
func guildDelete(s *discordgo.Session, event *discordgo.GuildDelete) {
	if event.Guild.Unavailable {
		return
	}
	log.Println("Left " + event.Guild.Name + " (" + event.Guild.ID + ")")
	_, err := db.GetDataTuple(config.Conn, event.Guild.ID)
	if err == nil {
		_ = db.DeleteDataTuple(config.Conn, event.Guild.ID)
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
	lang := utils.GetGuildLocale(m)

	// Is this message from a human && Does the message have the bot prefix?
	if !m.Author.Bot && strings.HasPrefix(m.Content, botPrefix) {

		// Split params using regex
		params := commandRegEx.FindAllString(m.Content[1:], -1)
		numParams := len(params)

		// If no parameter was passed, show the help message
		if numParams == 0 {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(locale.Messages[lang]["GeneralHelp"], m.Author.Mention(), botPrefix))
			log.Println("", err)
			return
		}

		for i := 0; i < numParams; i++ {
			params[i] = parameterQuotesRegEx.ReplaceAllString(params[i], "")
		}

		switch strings.ToLower(params[0]) {
		case "move":
			workerschann := make(chan []*discordgo.Session, 1)
			go utils.DetectServants(m.GuildID, append(config.ServantList, s), workerschann)

			guild, err := s.Guild(m.GuildID) // retrieving the server (guild) the message was originated from
			if err != nil {
				log.Println(err)
				_, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(locale.Messages[lang]["NotInGuild"], m.Author.Mention()))
				return
			}

			guildChannels := guild.Channels // retrieving the list of guildChannels
			if numParams == 2 {
				log.Println("Received move command with 2 parameters on " + guild.Name + " , ID: " + guild.ID + " , by :" + m.Author.ID)
				destination, err := utils.GetChannel(guildChannels, params[1])
				if err != nil {
					_, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(locale.Messages[lang]["CantFindChannel"], params[1]))
					return
				}

				if !utils.CheckPermissions(s, destination, m.Author.ID, discordgo.PermissionVoiceMoveMembers) {
					_, _ = s.ChannelMessageSend(m.ChannelID, locale.Messages[lang]["NoPermissionsDestination"])
					return
				}

				num, err := mover.MoveDestination(s, <-workerschann, m, guild, botPrefix, destination)
				if err != nil {
					if err.Error() == "no permission origin" {
						_, _ = s.ChannelMessageSend(m.ChannelID, locale.Messages[lang]["NoPermissionsOrigin"])
					} else if err.Error() == "cant find user" {
						_, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(locale.Messages[lang]["CantFindUser"], m.Author.Mention(), botPrefix))
					} else {
						_, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(locale.Messages[lang]["SorryBut"], err.Error()))
					}
					return
				}

				_, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(locale.Messages[lang]["JustMoved"], num))
				bumpStatistics(num, s, config.Conn)
				return
			} else if numParams == 3 {
				log.Println("Received move command with 3 parameters on " + guild.Name + " , ID: " + guild.ID + " , by :" + m.Author.ID)
				origin, err := utils.GetChannel(guildChannels, params[1])
				if err != nil {
					_, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(locale.Messages[lang]["CantFindChannel"], params[1]))
					return
				}

				destination, err := utils.GetChannel(guildChannels, params[2])
				if err != nil {
					_, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(locale.Messages[lang]["CantFindChannel"], params[2]))
					return
				}

				if !utils.CheckPermissions(s, origin, m.Author.ID, discordgo.PermissionVoiceMoveMembers) {
					_, _ = s.ChannelMessageSend(m.ChannelID, locale.Messages[lang]["NoPermissionsOrigin"])
					return
				}

				if !utils.CheckPermissions(s, destination, m.Author.ID, discordgo.PermissionVoiceMoveMembers) {
					_, _ = s.ChannelMessageSend(m.ChannelID, locale.Messages[lang]["NoPermissionsDestination"])
					return
				}

				num, err := mover.MoveOriginDestination(s, <-workerschann, m, guild, botPrefix, origin, destination)
				if err != nil {
					_, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(locale.Messages[lang]["JustMoved"], err.Error()))
					log.Println(err.Error())
					return
				}

				_, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(locale.Messages[lang]["JustMoved"], num))
				go bumpStatistics(num, s, config.Conn)
			} else {
				log.Println("Received move command with " + strconv.Itoa(numParams) + " parameter(s) (help message) on " + guild.Name + " , ID: " + guild.ID)
				_, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(locale.Messages[lang]["MoveHelper"], botPrefix, botPrefix, botPrefix, utils.ListChannelsForHelpMessage(guildChannels)))
			}

		case "summon":
			_, _ = summon.Summon(s, m, params)

		case "help":
			_, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(locale.Messages[lang]["HelpMessage"], botPrefix, botPrefix))

		case "lang":
			if numParams == 2 {
				chosenLang := utils.SelectLang(params[1])
				_ = db.UpdateDataTuple(config.Conn, m.GuildID, chosenLang)
				lang = chosenLang
				_, _ = s.ChannelMessageSend(m.ChannelID, locale.Messages[lang]["LangSet"])
			} else {
				_, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(locale.Messages["LANG"]["LangSetupMessage"], botPrefix, botPrefix))
			}

		default:
			_, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(locale.Messages[lang]["EhhMessage"], m.Author.Mention(), m.Content, botPrefix))
		}
	}
}

// bumpStatistics adds 1 to the "movs" stats and 'moved' to the "movd"
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
	_ = s.UpdateStatus(0, fmt.Sprintf("Moved %d players \n ! %s help", stats["usrs"], botPrefix))
	stats["movs"]++
	bytesStats, _ = json.Marshal(stats)
	err = db.UpdateDataTupleBytes(conn, "statistics", bytesStats)
	if err != nil {
		log.Println(err)
		log.Println(stats)
	}
	return
}
