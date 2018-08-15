package bot

import (
	"errors"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/auyer/commanderBot/config"
	"github.com/auyer/commanderBot/db"
	"github.com/bwmarrin/discordgo"
)

var BotID string
var Bot *discordgo.Session

// var dbDict map[string]*badger.DB

func Close() {
	log.Println("Closing")
	db.CloseDatabases()
	// Bot.Close()
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

	os.Mkdir(config.DatabasesPath, os.ModePerm)
	log.Print("Bot is running!")

}

func ready(s *discordgo.Session, event *discordgo.Ready) {
	// Set the playing status.
	s.UpdateStatus(0, config.BotPrefix+" help")
}

// This function will be called (due to AddHandler above) every time a new
// guild is joined.
func guildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
	log.Println("joined ! " + event.Guild.ID)

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
	dbp := db.PointerDict.Dict[event.Guild.ID]
	// dbp2 := &dbp
	dbp.Close()
	delete(db.PointerDict.Dict, event.Guild.ID)
	db.PointerDict.Unlock()

	err := db.RemoveDatabase(config.DatabasesPath, event.Guild.ID)
	if err != nil {
		log.Println(err)
	}
	db.PointerDict.Lock()
	delete(db.PointerDict.Dict, event.Guild.ID)
	db.PointerDict.Unlock()

}

func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if strings.HasPrefix(m.Content, config.BotPrefix) {
		if m.Author.ID == BotID {
			return
		}
		if m.Content == config.BotPrefix+" help" {
			log.Print(s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+`, use `+config.BotPrefix+` to use all my commands !
			`+config.BotPrefix+` move...
			`+config.BotPrefix+` bla
			`+config.BotPrefix+` blu`))
		} else if strings.HasPrefix(m.Content, config.BotPrefix+" move") {
			c, err := s.State.Channel(m.ChannelID)
			if err != nil {
				log.Print(err.Error())
				return
			}
			guild, err := s.Guild(c.GuildID)
			channs := guild.Channels
			sort.Slice(channs[:], func(i, j int) bool {
				return channs[i].Position < channs[j].Position
			})
			params := strings.Split(m.Content, " ")
			l := len(params)
			if l == 2 {
				message, err := s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+moveHelper(channs))
				log.Println(message)
				if err != nil {
					log.Println(err.Error())
				}
				// } else if l == 3 { // l==3 TODO DETECT AUTHOR LOCATION
				// 	err = moveMembers(s, c.GuildID, "USER CHANNEL", chanByPosId(channs, params[2]))
				// 	if err != nil {
				// 		log.Println(err.Error())
				// 	}
			} else if l == 4 {
				err = moveMembers(s, c.GuildID, chanByPosId(channs, params[2]), chanByPosId(channs, params[3]))
				if err != nil {
					log.Println(err.Error())
				}
			} else {

			}
		} else {
			log.Print(s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+", you said "+m.Content+" ... ehh ?"))
		}
	}
}

func moveMembers(s *discordgo.Session, id string, origin string, dest string) error {
	if origin == dest {
		return errors.New("Destination and origin are the same")
	}
	oriChan, err := s.Channel(origin)
	if err != nil {
		return err
	}
	// var moveErr []error
	for _, member := range oriChan.Recipients {
		log.Println(member.Username)
		log.Println(oriChan.ID)
		// log.Print(s.GuildMemberMove(id, member.ID, dest))
		// if err != nil {
		// 	moveErr = append(moveErr, err)
		// }
	}
	return nil
}

// User has requested a move but has not yet specified any channels. We will now list all channels available and the user must type both in chatMessage
func moveHelper(channs []*discordgo.Channel) string {
	message := " You may use the bot with the following commands:\n\nView this help and the list of available channels\n\t" + config.BotPrefix + " move\n\nMove all users from your channel to the <integer:destination channel>\n\t" + config.BotPrefix + " move <integer:destination channel>\n\nMove all users from <integer:origin channel> to the <integer:destination channel>\n\t" + config.BotPrefix + " move <integer:origin channel> <integer:destination channel>\n\nList of available channels:\n"
	sort.Slice(channs[:], func(i, j int) bool {
		return channs[i].Position < channs[j].Position
	})
	i := 0
	for _, chann := range channs {
		if chann.Type == 2 {
			i++
			message = message + strconv.Itoa(i) + " ) " + chann.Name + "\n"
		}
	}
	return message
}

func chanByPosId(channs []*discordgo.Channel, posId string) string {
	i := 0
	for _, chann := range channs {
		if chann.Type == 2 {
			i++
			if strconv.Itoa(i) == posId {
				return chann.ID
			}
		}
	}
	return "error"
}
