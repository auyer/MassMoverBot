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
	Prefix          string
	MoverBotToken   string
	MoverSession    *discordgo.Session
	PowerupTokens   []string
	DB              *badger.DB
	PowerupSessions []*discordgo.Session
	Messages        *utils.Message
}

// Close function ends the bot connection and closes its database
func (bot *Bot) Close() {
	log.Println("Shutting Down bot")
	err := bot.MoverSession.Close()
	if err != nil {
		log.Println("Failed closing main connection")
	}
	for _, powerupBot := range bot.PowerupSessions {
		err = powerupBot.Close()
		if err != nil {
			log.Println("Failed closing powerup connection")
		}
	}
	log.Println("Closing Database")
	bot.DB.Close()

}

// RegEx used to split all command parameters, considering anything between quotes as a single parameter.
// Ex: `> move ThisChannel "That Channel"` will be processed as [">", "move", "ThisChannel", "That Channel"]
var commandRegEx, _ = regexp.Compile(`(".*?"|\S+)`)

// RegEx used to remove starting and ending quotes from the parameters
var parameterQuotesRegEx, _ = regexp.Compile(`(^"|"$)`)

func (bot *Bot) setupBot(s *discordgo.Session) error {
	s.AddHandler(bot.ready)

	_, err := s.User("@me")
	if err != nil {
		return err
	}

	return nil
}

// Init creates the first bot object
func Init(configs config.ConfigurationParameters, messages *utils.Message, conn *badger.DB) *Bot {
	return &Bot{Prefix: configs.BotPrefix, MoverBotToken: configs.MoverBotToken, PowerupTokens: configs.PowerupTokens, Messages: messages, DB: conn}
}

// Start function connects and ads the necessary handlers
func (bot *Bot) Start() error {

	var err error
	commander, err := discordgo.New("Bot " + bot.MoverBotToken)
	if err != nil {
		log.Println("Error creating main session: ", err)
		return err
	}
	bot.MoverSession = commander

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
	var powerupSessions []*discordgo.Session

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

		powerupSessions = append(powerupSessions, powerup)
		bot.PowerupSessions = powerupSessions
	}

	log.Println("Bot is fully running!")

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
				err = utils.AskMember(s, event.Guild.OwnerID, bot.Messages.WelcomeAndLang(bot.Prefix))
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
		lang := utils.GetGuildLocale(bot.DB, m.GuildID)

		// Split params using regex
		params := commandRegEx.FindAllString(m.Content[1:], -1)
		numParams := len(params)

		// If no parameter was passed, show the help message
		if numParams == 0 {
			_, err := s.ChannelMessageSendEmbed(m.ChannelID, bot.Messages.GeneralHelp(lang, m.Author.Mention(), bot.Prefix))
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
			moved, err := bot.Summon(m, params)
			if err != nil {

				return
			}
			bot.bumpStatistics(moved)

		case "lang":
			_, err := bot.MoverSession.Guild(m.GuildID) // retrieving the server (guild) the message was originated from
			if err != nil {
				log.Println(err)
				_, _ = bot.MoverSession.ChannelMessageSendEmbed(m.ChannelID, bot.Messages.NotInGuild(utils.GetGuildLocale(bot.DB, m.GuildID), m.Author.Mention()))
				return
			}
			if numParams == 2 {
				chosenLang := utils.SelectLang(params[1])
				_ = db.UpdateDataTuple(bot.DB, m.GuildID, chosenLang)
				lang = chosenLang
				_, _ = s.ChannelMessageSendEmbed(m.ChannelID, bot.Messages.LangSet(lang))
			} else {
				_, _ = s.ChannelMessageSendEmbed(m.ChannelID, bot.Messages.LangSetupMessage(bot.Prefix))
			}

		default:
			_, _ = s.ChannelMessageSendEmbed(m.ChannelID, bot.Messages.HelpMessage(lang, bot.Prefix))
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
	_ = bot.MoverSession.UpdateStatus(0, fmt.Sprintf("Moved %d players \n ! %s help", stats["usrs"], bot.Prefix))
	for _, powerupSession := range bot.PowerupSessions {
		_ = powerupSession.UpdateStatus(0, fmt.Sprintf("Moved %d players \n ! %s help", stats["usrs"], bot.Prefix))
	}
	stats["movs"]++
	bytesStats, _ = json.Marshal(stats)
	err = db.UpdateDataTupleBytes(bot.DB, "statistics", bytesStats)
	if err != nil {
		log.Println(err)
		log.Println(stats)
	}
	return
}
