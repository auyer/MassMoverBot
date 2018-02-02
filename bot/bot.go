package bot

import (
	"log"
	"strings"

	"github.com/auyer/commanderBot/config"
	"github.com/bwmarrin/discordgo"
)

var BotID string
var bot *discordgo.Session

func Start() {
	bot, err := discordgo.New("Bot " + config.Token)

	if err != nil {
		log.Println("Error creating Discord session: ", err)
		return
	}

	u, err := bot.User("@me")

	if err != nil {
		log.Println("Error creating Discord session: ", err)
	}

	BotID = u.ID

	bot.AddHandler(messageHandler)

	err = bot.Open()

	if err != nil {
		log.Println("Error creating Discord session: ", err)
		return
	}

	log.Print("Bot is running!")
}

func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if strings.HasPrefix(m.Content, config.BotPrefix) {
		if m.Author.ID == BotID {
			return
		}
		if m.Content == "~c Help" {
			log.Print(s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+`, use ~c to use all my commands !
			~c
			~c
			~c`))
		} else {
			log.Print(s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+", you said "+m.Content+" ... ehh ?"))
		}
	}
}
