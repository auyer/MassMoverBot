package bot

import (
	"log"
	"os"
	"strings"

	"github.com/auyer/commanderBot/config"
	"github.com/auyer/commanderBot/mover"
	"github.com/bwmarrin/discordgo"
)

var botID string
var bot *discordgo.Session

// Close function ends the bot connection and closes its database
func Close() {
	log.Println("Closing")
	// db.CloseDatabases()
	bot.Close()
}

// Start function connects and ads the necessary handlers
func Start() {
	var err error
	bot, err = discordgo.New("Bot " + config.Token)
	if err != nil {
		log.Println("Error creating Discord session: ", err)
		return
	}

	u, err := bot.User("@me")
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

	botID = u.ID

	bot.AddHandler(ready)
	bot.AddHandler(messageHandler)
	bot.AddHandler(guildCreate)
	bot.AddHandler(guildDelete)

	err = bot.Open()

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
	log.Println("Joined " + event.Guild.ID + " (" + event.Guild.Name + ")")

	if event.Guild.Unavailable {
		return
	}
	// Database functionality not in use
	/* dbpointer, err := db.ConnectDB(config.DatabasesPath + event.Guild.ID)
	 if err != nil {
		log.Println("Error creating guildDB " + err.Error())
	}
	db.PointerDict.Lock()
	db.PointerDict.Dict[event.Guild.ID] = dbpointer
	db.PointerDict.Unlock()
	*/

	for _, channel := range event.Guild.Channels {
		if channel.ID == event.Guild.ID {
			_, _ = s.ChannelMessageSend(channel.ID, "Bot Joined!")
			return
		}
	}
}

// guildDelete function will be called every time the bot leaves a guild.
func guildDelete(s *discordgo.Session, event *discordgo.GuildDelete) {

	if event.Guild.Unavailable {
		return
	}
	// db.PointerDict.Lock()
	// db.PointerDict.Dict[event.Guild.ID].Lock()
	// db.PointerDict.Dict[event.Guild.ID].Close()
	// db.PointerDict.Dict[event.Guild.ID].Unlock()
	// delete(db.PointerDict.Dict, event.Guild.ID)
	// db.PointerDict.Unlock()

	// err := db.RemoveDatabase(config.DatabasesPath, event.Guild.ID)
	// if err != nil {
	// 	log.Println(err)
	// }

}

// messageHandler function will be called when the bot reads a message
func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if strings.HasPrefix(m.Content, config.BotPrefix) {
		if m.Author.ID == botID {
			return
		}
		if m.Content == config.BotPrefix || m.Content == config.BotPrefix+" help" {
			s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+`, use `+config.BotPrefix+` to use all my commands !
			`+config.BotPrefix+`  move : Use this command to move users from one Voice Channel to another ! Type  `+config.BotPrefix+` move for help`)
		} else if strings.HasPrefix(m.Content, config.BotPrefix+" move") {
			mover.Move(s, m, config.BotPrefix)
		} else {
			s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+", you said "+m.Content+" ... ehh ?")
		}
	}
}
