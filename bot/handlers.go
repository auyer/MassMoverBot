package bot

import (
	"log"
	"strings"

	"github.com/auyer/massmoverbot/utils"
	"github.com/bwmarrin/discordgo"
)

// messageHandler function will be called when the bot reads a message
func (bot *Bot) messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Is this message from a human && Does the message have the bot prefix?
	if !m.Author.Bot && strings.HasPrefix(m.Content, bot.Prefix) {
		lang := bot.GetGuildLocale(m.GuildID)

		// Split params using regex
		params := commandRegEx.FindAllString(m.Content[len(bot.Prefix):], -1)
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
		case "lang":
			_, err := bot.MoverSession.Guild(m.GuildID) // retrieving the server (guild) the message was originated from
			if err != nil {
				log.Println(err)
				_, _ = bot.MoverSession.ChannelMessageSendEmbed(m.ChannelID, bot.Messages.NotInGuild(bot.GetGuildLocale(m.GuildID), m.Author.Mention()))
				return
			}
			if numParams == 2 {
				chosenLang := utils.SelectLang(params[1])
				err := bot.DB.SetGuildLang(m.GuildID, chosenLang)
				if err != nil {
					_, _ = s.ChannelMessageSendEmbed(m.ChannelID, bot.Messages.SorryBut(lang, err.Error()))
					log.Println(err)
					return
				}
				lang = chosenLang
				_, _ = s.ChannelMessageSendEmbed(m.ChannelID, bot.Messages.LangSet(lang))
			} else {
				_, _ = s.ChannelMessageSendEmbed(m.ChannelID, bot.Messages.LangSetupMessage(bot.Prefix))
			}
		case "summon":
			moved, err := bot.Summon(m, params)
			if err != nil {

				return
			}
			bot.bumpStatistics(moved)
		case "move":
			moved, err := bot.Move(m, params)
			if err != nil {

				return
			}
			bot.bumpStatistics(moved)

		default:
			_, _ = s.ChannelMessageSendEmbed(m.ChannelID, bot.Messages.HelpMessage(lang, bot.Prefix))
		}
	}
}

// guildDelete function will be called every time the bot leaves a guild.
func (bot *Bot) guildDelete(s *discordgo.Session, event *discordgo.GuildDelete) {
	if event.Guild.Unavailable {
		return
	}
	log.Println("Left " + event.Guild.Name + " (" + event.Guild.ID + ")")
	err := bot.DB.DeleteGuildLang(event.Guild.ID)
	if err != nil {
		log.Println("Error clearing GuildLang.", err)
	}
}

// This function will be called (due to AddHandler above) every time a new
// guild is joined.
func (bot *Bot) guildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
	if event.Guild.Unavailable {
		return
	}
	log.Println("Joined " + event.Guild.Name + " (" + event.Guild.ID + ")" + " in " + event.Guild.Region)

	val, err := bot.DB.WasWelcomeMessageSent(event.Guild.OwnerID)
	if err != nil {
		if !val {
			if !utils.HaveIAskedMember(s, event.Guild.OwnerID) {
				err = utils.AskMember(s, event.Guild.OwnerID, bot.Messages.WelcomeAndLang(bot.Prefix))
				if err != nil {
					log.Println("Failed to send message to owner.")
					return
				}
			}
			err = bot.DB.SetWelcomeMessageSent(event.Guild.OwnerID, true)
			if err != nil {
				log.Println(err)
				return
			}
		}
	}
}
