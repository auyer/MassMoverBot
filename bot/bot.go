package bot

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/auyer/commanderBot/mover"
	"github.com/auyer/commanderBot/utils"
	"github.com/bwmarrin/discordgo"
)

var commanderID string
var commander *discordgo.Session
var servantList []*discordgo.Session
var botPrefix string

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
func Start(commanderToken string, servantTokens []string, prefix string) {
	botPrefix = prefix
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
	s.UpdateStatus(0, botPrefix+" help")
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
	if strings.HasPrefix(m.Content, botPrefix) {
		if m.Author.ID == commanderID {
			return
		}
		if m.Content == botPrefix || m.Content == botPrefix+" help" {
			s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+`, use `+botPrefix+` to use all my commands !
			`+botPrefix+`  move : Use this command to move users from one Voice Channel to another ! Type  `+botPrefix+` move for help`)
		} else if strings.HasPrefix(m.Content, botPrefix+" move") {
			workerschann := make(chan []*discordgo.Session, 1)
			go utils.DetectServants(m.GuildID, append(servantList, s), workerschann)
			guild, err := s.Guild(m.GuildID) // retrieving the server (guild) the message was originated from
			if err != nil {
				log.Panic(err)
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
					s.ChannelMessageSend(m.ChannelID, "Sorry, I can't find channel "+params[2]+".")
				}
				if !utils.CheckPermissions(s, destination, m.Author.ID, discordgo.PermissionVoiceMoveMembers) {
					s.ChannelMessageSend(m.ChannelID, "Sorry, but you dont have permissions to move to the Destination channel")
					return
				}
				mover.MoveDestination(s, <-workerschann, m, guild, botPrefix, destination)
				return
			} else if length == 4 {
				log.Println("Received 4 parameter move command on " + guild.Name + " , ID: " + guild.ID + " , by :" + m.Author.ID)
				origin, err := utils.GetChannel(channs, params[2])
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, "Sorry, I can't find channel "+params[2]+".")
				}
				destination, err := utils.GetChannel(channs, params[3])
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, "Sorry, I can't find channel "+params[3]+".")
				}
				if !utils.CheckPermissions(s, origin, m.Author.ID, discordgo.PermissionVoiceMoveMembers) {
					s.ChannelMessageSend(m.ChannelID, "Sorry, but you dont have permissions from the Origin channel")
					return
				}
				if !utils.CheckPermissions(s, destination, m.Author.ID, discordgo.PermissionVoiceMoveMembers) {
					s.ChannelMessageSend(m.ChannelID, "Sorry, but you dont have permissions to move to the Destination channel")
					return
				}
				mover.MoveOriginDestination(s, <-workerschann, m, guild, botPrefix, origin, destination)
				return
			}
			log.Println("Sending help message on " + guild.Name + " , ID: " + guild.ID)
			s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+mover.MoveHelper(channs, botPrefix))
		} else {
			s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+", you said "+m.Content+" ... ehh ?")
		}
	}
}

// servantTest function will be called when the bot reads a message
// func servantTest(s *discordgo.Session, m *discordgo.MessageCreate) {
// 	if strings.HasPrefix(m.Content, botPrefix+"s") {
// 		if m.Author.ID == commanderID {
// 			return
// 		}
// 		s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+" Sir yes sir !")
// 	}
// }
