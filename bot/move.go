package bot

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/auyer/commanderBot/config"
	"github.com/bwmarrin/discordgo"
)

// move function moves discord users
func move(s *discordgo.Session, m *discordgo.MessageCreate) {
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
	if l == 2 { // IF 2 parameters: Get Help message
		message, err := s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+MoveHelper(channs))
		log.Println(message)
		if err != nil {
			log.Println(err.Error())
		}
	} else if l == 3 { // IF 3 parameters : Detect Authors Location
		param2, err := strconv.Atoi(params[2])
		var destination string
		if err != nil {
			destination, err = ChanByName(channs, params[2])
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Sorry, I can't find channel "+params[2]+".")
				return
			}
		} else {
			destination, err = ChanByPosNum(channs, param2)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Sorry, I can't find channel "+params[2]+".")
				return
			}
		}
		for _, member := range guild.VoiceStates {
			if member.UserID == m.Author.ID {
				num, err := MoveMembers(s, guild, c.GuildID, member.ChannelID, destination)
				if err != nil {
					log.Println(err.Error())
				}
				log.Print(s.ChannelMessageSend(m.ChannelID, "I Just moved "+num+" users for you."))
				return
			}
		}
		log.Print(s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+", I you need to be connected to a channel for me to find you. Type '"+config.BotPrefix+" move' to get help."))
	} else if l == 4 { // IF 4 parameters: Move from Origin to Destination
		var origin string
		param2, err := strconv.Atoi(params[2])
		if err != nil {
			origin, err = ChanByName(channs, params[2])
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Sorry, I can't find channel "+params[2]+".")
				return
			}
		} else {
			origin, err = ChanByPosNum(channs, param2)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Sorry, I can't find channel "+params[2]+".")
				return
			}
		}
		param3, err := strconv.Atoi(params[3])
		var destination string
		if err != nil {
			destination, err = ChanByName(channs, params[3])
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Sorry, I can't find channel "+params[2]+".")
				return
			}
		} else {
			destination, err = ChanByPosNum(channs, param3)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Sorry, I can't find channel "+params[2]+".")
				return
			}
		}

		num, err := MoveMembers(s, guild, c.GuildID, origin, destination)
		if err != nil {
			log.Print(s.ChannelMessageSend(m.ChannelID, "Sorry, but: "+err.Error()))
			log.Println(err.Error())
		}
		log.Print(s.ChannelMessageSend(m.ChannelID, "I Just moved "+num+" users for you."))
		return
	} else {

	}
}

// MoveMembers wraps MoveAndRetry with councurrent calls and error reporting.
func MoveMembers(s *discordgo.Session, guild *discordgo.Guild, id string, origin string, dest string) (string, error) {
	if origin == dest {
		return "", errors.New("destination and origin are the same")
	}
	num := 0
	var wg sync.WaitGroup
	for _, member := range guild.VoiceStates {
		if member.ChannelID == origin {
			wg.Add(1)
			go func(id, UserID, dest string) {
				num++
				defer wg.Done()
				MoveAndRetry(s, id, UserID, dest, 3)
			}(id, member.UserID, dest)
		}
	}
	wg.Wait()
	return fmt.Sprintf("Moved %d users", num), nil
}

// MoveAndRetry is a wrapper on top of discordgo.Session.GuildMemberMove with a retry function
func MoveAndRetry(s *discordgo.Session, id, userID, dest string, retry int) {
	err := s.GuildMemberMove(id, userID, dest)
	if err != nil {
		time.Sleep(time.Millisecond * 10)
		if retry >= 0 {
			log.Println("Failed to move user with ID: " + userID)
			return
		}
		MoveAndRetry(s, id, userID, dest, retry-1)
	}
}

// MoveHelper prints the help text. We will now list all channels available and the user must type both in chatMessage
func MoveHelper(channs []*discordgo.Channel) string {
	message := " You may use the bot with the following commands:\n\nView this help and the list of available channels\n\t" + config.BotPrefix + " move\n\nMove all users from your channel to the <integer:destination channel>\n\t" + config.BotPrefix + " move <number:destination channel>\n\nMove all users from <integer:origin channel> to the <integer:destination channel>\n\t" + config.BotPrefix + " move <number:origin channel> <number:destination channel>\n\nList of available channels:\n"
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

// ChanByPosNum retrieves channel id by the possition as displayed in the channel.
func ChanByPosNum(channs []*discordgo.Channel, posNum int) (string, error) {
	i := 0
	for _, chann := range channs {
		if chann.Type == 2 {
			i++
			if i == posNum {
				return chann.ID, nil
			}
		}
	}
	return "", errors.New("Not Found")
}

// ChanByName retrieves channel id by name. The comparison is case insensitive.
func ChanByName(channs []*discordgo.Channel, name string) (string, error) {
	for _, chann := range channs {
		if chann.Type == 2 {
			if strings.ToUpper(chann.Name) == strings.ToUpper(name) {
				return chann.ID, nil
			}
		}
	}
	return "", errors.New("Not Found")
}
