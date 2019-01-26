package bot

import (
	"log"
	"os"
	"strings"

	"github.com/auyer/commanderBot/config"
	"github.com/auyer/commanderBot/db"
	"github.com/bwmarrin/discordgo"
)

var BotID string
var Bot *discordgo.Session

func Close() {
	log.Println("Closing")
	db.CloseDatabases()
	Bot.Close()
}

func Start() {
	Bot, err := discordgo.New("Bot " + config.Token)
	if err != nil {
		log.Println("Error creating Discord session: ", err)
		return
	}

	u, err := Bot.User("@me")
	if err != nil {
		log.Println("Error creating Discord session: ", err)
	}

	if _, err := os.Stat(config.DatabasesPath); os.IsNotExist(err) {
		err = os.Mkdir(config.DatabasesPath, os.ModePerm)
		if err != nil && err.Error() != "file exists" {
			log.Println("Error creating Databases folder: ", err)
			return
		}
	}

	BotID = u.ID

	Bot.AddHandler(ready)
	Bot.AddHandler(messageHandler)
	Bot.AddHandler(guildCreate)
	Bot.AddHandler(guildDelete)

	err = Bot.Open()

	if err != nil {
		log.Println("Error creating Discord session: ", err)
		return
	}
	log.Println("Bot is running!")

}

func ready(s *discordgo.Session, event *discordgo.Ready) {
	// Set the playing status.
	s.UpdateStatus(0, config.BotPrefix+" help")
}

// This function will be called (due to AddHandler above) every time a new
// guild is joined.
func guildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
	log.Println("Joined " + event.Guild.ID)

	if event.Guild.Unavailable {
		return
	}
	dbpointer, err := db.ConnectDB(config.DatabasesPath + event.Guild.ID)
	if err != nil {
		log.Println("Error creating guildDB " + err.Error())
	}
	db.PointerDict.Lock()
	db.PointerDict.Dict[event.Guild.ID] = dbpointer
	db.PointerDict.Unlock()

	for _, channel := range event.Guild.Channels {
		if channel.ID == event.Guild.ID {
			_, _ = s.ChannelMessageSend(channel.ID, "Bot Joined!")
			return
		}
	}
}

// This function will be called (due to AddHandler above) every time the bot
// leaves a guild.
func guildDelete(s *discordgo.Session, event *discordgo.GuildDelete) {

	if event.Guild.Unavailable {
		return
	}
	db.PointerDict.Lock()
	db.PointerDict.Dict[event.Guild.ID].Lock()
	db.PointerDict.Dict[event.Guild.ID].Close()
	db.PointerDict.Dict[event.Guild.ID].Unlock()
	delete(db.PointerDict.Dict, event.Guild.ID)
	db.PointerDict.Unlock()

	err := db.RemoveDatabase(config.DatabasesPath, event.Guild.ID)
	if err != nil {
		log.Println(err)
	}

}

func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if strings.HasPrefix(m.Content, config.BotPrefix) {
		if m.Author.ID == BotID {
			return
		}
		if m.Content == config.BotPrefix+" help" {
			s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+`, use `+config.BotPrefix+` to use all my commands !
			`+config.BotPrefix+`  move : Use this command to move users from one Voice Channel to another ! Type  `+config.BotPrefix+` move for help`)
		} else if strings.HasPrefix(m.Content, config.BotPrefix+" move") {
			move(s, m)
		} else {
			s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+", you said "+m.Content+" ... ehh ?")
		}
	}
}
