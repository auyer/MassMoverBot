package bot

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"time"

	"github.com/auyer/massmoverbot/config"
	"github.com/auyer/massmoverbot/db"
	"github.com/auyer/massmoverbot/utils"
	"github.com/bwmarrin/discordgo"
)

// Bot struct
type Bot struct {
	Prefix          string
	MoverBotToken   string
	MoverSession    *discordgo.Session
	PowerupTokens   []string
	DB              db.DataStorage
	PowerupSessions []*discordgo.Session
	Messages        *utils.Message
	Closing         chan int
}

// Close function ends the bot connection and closes its database
func (bot *Bot) Close() {
	bot.Closing <- 1
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
func Init(configs config.ConfigurationParameters, messages *utils.Message, conn db.DataStorage) *Bot {
	c := make(chan int)
	return &Bot{Prefix: configs.BotPrefix, MoverBotToken: configs.MoverBotToken, PowerupTokens: configs.PowerupTokens, Messages: messages, DB: conn, Closing: c}
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
	commander.Identify.Intents = discordgo.MakeIntent(
		discordgo.IntentsGuilds | discordgo.IntentsGuildMessages |
			discordgo.IntentsGuildMembers | discordgo.IntentsGuildVoiceStates)
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
		powerup.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuilds)

		err = powerup.Open()
		if err != nil {
			log.Println("Error Opening powerup session: ", err)
			continue
		}

		powerupSessions = append(powerupSessions, powerup)
		bot.PowerupSessions = powerupSessions
	}

	log.Println("Bot is fully running!")
	go func() {
		for {
			select {
			case <-bot.Closing:
				log.Println("Halted Status Update")
				break
			case <-time.After(1200 * time.Second):
				stats, err := bot.DB.GetStatistics()
				if err != nil {
					log.Println(err)
				}
				bot.MoverSession.UpdateStatusComplex(discordgo.UpdateStatusData{Activities: []*discordgo.Activity{{Name: fmt.Sprintf("%s help | Moved %s players !", bot.Prefix, utils.FormatNumberWithSeparators(int64(stats["usrs"])))}}})
				for _, powerupSession := range bot.PowerupSessions {
					powerupSession.UpdateStatusComplex(discordgo.UpdateStatusData{Activities: []*discordgo.Activity{{Name: fmt.Sprintf("Moved %s players !", utils.FormatNumberWithSeparators(int64(stats["usrs"])))}}})
				}
			}
		}
	}() // this loop will update the bot status every 1.200 seconds
	return nil
}

func (bot *Bot) ready(s *discordgo.Session, event *discordgo.Ready) {
	// Set the playing status.
	stats, err := bot.DB.GetStatistics()
	if err != nil {
		log.Println("Failed to get Statistics", err)
		s.UpdateStatusComplex(discordgo.UpdateStatusData{Activities: []*discordgo.Activity{{Name: bot.Prefix + " help"}}})
		return
	}
	s.UpdateStatusComplex(discordgo.UpdateStatusData{Activities: []*discordgo.Activity{{Name: fmt.Sprintf("%s help | Moved %s players !", bot.Prefix, utils.FormatNumberWithSeparators(int64(stats["usrs"])))}}})
}

// bumpStatistics adds 1 to the "movs" stats and 'moved' to the "movd"
func (bot *Bot) bumpStatistics(moved string) {
	stats, err := bot.DB.GetStatistics()
	if err != nil {
		log.Println(err)
	}
	movedInt, _ := strconv.Atoi(moved)
	stats["usrs"] += movedInt
	err = bot.DB.SetStatistics(stats)
	if err != nil {
		log.Println(err)
		log.Println(stats)
	}
	return
}

// GetGuildLocale function will return the language for a guild, returning EN by default.
/*
Input:
	GuildID string : the ID of the guild
Output:
	language string
*/
func (bot *Bot) GetGuildLocale(GuildID string) string {
	lang, err := bot.DB.GetGuildLang(GuildID)
	if err != nil {
		lang = "EN"
	}

	return lang
}
