package bot

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/auyer/massmoverbot/config"
	"github.com/auyer/massmoverbot/db"
	"github.com/auyer/massmoverbot/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/dgraph-io/badger"
)

// Bot struct
type Bot struct {
	Prefix           string
	CommanderToken   string
	CommanderSession *discordgo.Session
	PowerupTokens    []string
	DB               *badger.DB
	PowerupSessions  []*discordgo.Session
	Messages         map[string]map[string]string
}

// Close finishes all bot connections
func (bot *Bot) Close() {
	log.Println("Closing")
	_ = bot.CommanderSession.Close()
	bot.DB.Close()
	for _, powerupBot := range bot.PowerupSessions {
		_ = powerupBot.Close()
	}
}

// RegEx used to split all command parameters, considering anything between quotes as a single parameter.
// Ex: `> move ThisChannel "That Channel"` will be processed as [">", "move", "ThisChannel", "That Channel"]
var commandRegEx, _ = regexp.Compile(`(".*?"|\S+)`)

// RegEx used to remove starting and ending quotes from the parameters
var parameterQuotesRegEx, _ = regexp.Compile(`(^"|"$)`)

// Close function ends the bot connection and closes its database

func (bot *Bot) setupBot(s *discordgo.Session) error {
	s.AddHandler(bot.ready)

	_, err := s.User("@me")
	if err != nil {
		return err
	}

	return nil
}

// Init creates the first bot object
func Init(configs config.ConfigurationParameters, messages map[string]map[string]string, conn *badger.DB) *Bot {
	return &Bot{Prefix: configs.BotPrefix, CommanderToken: configs.CommanderToken, PowerupTokens: configs.PowerupTokens, Messages: messages, DB: conn}
}

// Start function connects and ads the necessary handlers
func (bot *Bot) Start() error {

	var err error
	commander, err := discordgo.New("Bot " + bot.CommanderToken)
	if err != nil {
		log.Println("Error creating main session: ", err)
		return err
	}
	bot.CommanderSession = commander

	err = bot.setupBot(commander)
	if err != nil {
		log.Println("Error setting up main session: ", err)
		return err
	}

	commander.AddHandler(bot.guildCreate)
	commander.AddHandler(bot.guildDelete)
	commander.AddHandler(bot.messageHandler)

	err = commander.Open()
	if err != nil {
		log.Println("Error opening main Discord session: ", err)
		return err
	}
	var powerupList []*discordgo.Session

	for _, powerupToken := range bot.PowerupTokens {
		powerup, err := discordgo.New("Bot " + powerupToken)
		if err != nil {
			log.Println("Error creating PowerUp session: ", err)
			continue
		}

		err = bot.setupBot(powerup)
		if err != nil {
			log.Println("Error setting powerup session: ", err)
			continue
		}

		err = powerup.Open()
		if err != nil {
			log.Println("Error Opening powerup session: ", err)
			continue
		}

		powerupList = append(powerupList, powerup)
	}
	// ADD POWERUPLIST TO COMMAND STRUCT

	log.Println("Bot is running!")
	return nil
}

func (bot *Bot) ready(s *discordgo.Session, event *discordgo.Ready) {
	// Set the playing status.
	bytesStats, err := db.GetDataTupleBytes(bot.DB, "statistics")
	if err != nil {
		log.Println("Failed to get Statistics")
		s.UpdateStatus(0, bot.Prefix+" help")
		return
	}
	stats := map[string]int{}
	err = json.Unmarshal(bytesStats, &stats)
	if err != nil {
		log.Println("Failed to decode Statistics")
		s.UpdateStatus(0, bot.Prefix+" help")
		return
	}
	_ = s.UpdateStatus(0, fmt.Sprintf("Moved %d players \n ! %s help", stats["usrs"], bot.Prefix))
}

// This function will be called (due to AddHandler above) every time a new
// guild is joined.
func (bot *Bot) guildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
	if event.Guild.Unavailable {
		return
	}
	log.Println("Joined " + event.Guild.Name + " (" + event.Guild.ID + ")" + " in " + event.Guild.Region)

	val, err := db.GetDataTuple(bot.DB, "M:"+event.Guild.OwnerID)
	if err != nil {
		if err == badger.ErrKeyNotFound || val == "" {
			if !utils.HaveIAskedMember(s, event.Guild.OwnerID) {
				err = utils.AskMember(s, event.Guild.OwnerID, fmt.Sprintf(bot.Messages["LANG"]["WelcomeAndLang"], bot.Prefix, bot.Prefix))
				if err != nil {
					log.Println("Failed to send message to owner.")
					return
				}
			}
			_ = db.UpdateDataTuple(bot.DB, "M:"+event.Guild.OwnerID, "1")
		}
	}
}

// guildDelete function will be called every time the bot leaves a guild.
func (bot *Bot) guildDelete(s *discordgo.Session, event *discordgo.GuildDelete) {
	if event.Guild.Unavailable {
		return
	}
	log.Println("Left " + event.Guild.Name + " (" + event.Guild.ID + ")")
	_, err := db.GetDataTuple(bot.DB, event.Guild.ID)
	if err == nil {
		_ = db.DeleteDataTuple(bot.DB, event.Guild.ID)
	}
}

// messageHandler function will be called when the bot reads a message
func (bot *Bot) messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Is this message from a human && Does the message have the bot prefix?
	if !m.Author.Bot && strings.HasPrefix(m.Content, bot.Prefix) {
		lang := utils.GetGuildLocale(bot.DB, m)

		// Split params using regex
		params := commandRegEx.FindAllString(m.Content[1:], -1)
		numParams := len(params)

		// If no parameter was passed, show the help message
		if numParams == 0 {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(bot.Messages[lang]["GeneralHelp"], m.Author.Mention(), bot.Prefix))
			log.Println("", err)
			return
		}

		for i := 0; i < numParams; i++ {
			params[i] = parameterQuotesRegEx.ReplaceAllString(params[i], "")
		}

		switch strings.ToLower(params[0]) {
		case "move":
			moved, err := bot.Move(m, params)
			if err != nil {

				return
			}
			bot.bumpStatistics(moved)

		case "summon":
			_, _ = bot.Summon(m, params)

		case "help":
			_, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(bot.Messages[lang]["HelpMessage"], bot.Prefix, bot.Prefix, bot.Prefix))

		case "lang":
			if numParams == 2 {
				chosenLang := utils.SelectLang(params[1])
				_ = db.UpdateDataTuple(bot.DB, m.GuildID, chosenLang)
				lang = chosenLang
				_, _ = s.ChannelMessageSend(m.ChannelID, bot.Messages[lang]["LangSet"])
			} else {
				_, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(bot.Messages["LANG"]["LangSetupMessage"], bot.Prefix, bot.Prefix))
			}

		default:
			_, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(bot.Messages[lang]["EhhMessage"], m.Author.Mention(), m.Content, bot.Prefix))
		}
	}
}

// bumpStatistics adds 1 to the "movs" stats and 'moved' to the "movd"
func (bot *Bot) bumpStatistics(moved string) {
	bytesStats, err := db.GetDataTupleBytes(bot.DB, "statistics")
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
	_ = bot.CommanderSession.UpdateStatus(0, fmt.Sprintf("Moved %d players \n ! %s help", stats["usrs"], bot.Prefix))
	stats["movs"]++
	bytesStats, _ = json.Marshal(stats)
	err = db.UpdateDataTupleBytes(bot.DB, "statistics", bytesStats)
	if err != nil {
		log.Println(err)
		log.Println(stats)
	}
	return
}
