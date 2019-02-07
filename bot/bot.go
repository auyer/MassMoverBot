package bot

import (
	"fmt"
	"log"
	"strings"

	"github.com/auyer/commanderBot/mover"
	"github.com/bwmarrin/discordgo"
)

var commanderID string
var commander *discordgo.Session
var servantList []*discordgo.Session
var BotPrefix string

// Close function ends the bot connection and closes its database
func Close() {
	log.Println("Closing")
	// db.CloseDatabases()
	commander.Close()
	for _, servant := range servantList {
		servant.Close()
	}
}

func setupBot(bot *discordgo.Session) (string, error) {

	bot.AddHandler(ready)
	bot.AddHandler(guildCreate)
	bot.AddHandler(guildDelete)
	u, err := commander.User("@me")
	if err != nil {
		log.Println("Error creating Discord session: ", err)
		return "", err
	}
	return u.ID, nil
}

// Start function connects and ads the necessary handlers
func Start(commanderToken string, servantTokens []string, botPrefix string) {
	BotPrefix = botPrefix
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

	// if _, err := os.Stat(config.DatabasesPath); os.IsNotExist(err) {
	// 	err = os.Mkdir(config.DatabasesPath, os.ModePerm)
	// 	if err != nil && err.Error() != "file exists" {
	// 		log.Println("Error creating Databases folder: ", err)
	// 		return
	// 	}
	// }

	commanderID, err = setupBot(commander)
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
	s.UpdateStatus(0, BotPrefix+" help")
}

// This function will be called (due to AddHandler above) every time a new
// guild is joined.
func guildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
	log.Println("Joined " + event.Guild.Name + " (" + event.Guild.ID + ")")

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
	if strings.HasPrefix(m.Content, BotPrefix) {
		if m.Author.ID == commanderID {
			return
		}
		if m.Content == BotPrefix || m.Content == BotPrefix+" help" {
			s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+`, use `+BotPrefix+` to use all my commands !
			`+BotPrefix+`  move : Use this command to move users from one Voice Channel to another ! Type  `+BotPrefix+` move for help`)
		} else if strings.HasPrefix(m.Content, BotPrefix+" move") {
			// s.UserChannelPermissions(m.Author.ID,m.Author.)
			mover.Move(s, servantList, m, BotPrefix)
		} else {
			s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+", you said "+m.Content+" ... ehh ?")
		}
	}
}

// servantTest function will be called when the bot reads a message
// func servantTest(s *discordgo.Session, m *discordgo.MessageCreate) {
// 	if strings.HasPrefix(m.Content, BotPrefix+"s") {
// 		if m.Author.ID == commanderID {
// 			return
// 		}
// 		s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+" Sir yes sir !")
// 	}
// }
