package mover

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Move function moves discord users
func Move(s *discordgo.Session, servants []*discordgo.Session, m *discordgo.MessageCreate, prefix string) {
	workers := []*discordgo.Session{}
	for _, servant := range servants {
		_, err := servant.State.Guild(m.GuildID)
		if err == nil {
			workers = append(workers, servant)
		}
	}
	workers = append(workers, s)
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
	if length == 3 { // IF 3 parameters : Detect Author's Location
		log.Println("Received 3 parameter move command on " + guild.Name + " , ID: " + guild.ID + " , by :" + m.Author.ID)
		param2, err := strconv.Atoi(params[2])
		var destination string
		if err != nil {
			destination, err = ChanByName(channs, params[2])
		} else {
			destination, err = ChanByPosNum(channs, param2)
		}
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Sorry, I can't find channel "+params[2]+".")
			return
		}
		// CHECK PERMISSIONS
		if !checkPermissions(s, destination, m.Author.ID, discordgo.PermissionVoiceMoveMembers) {
			s.ChannelMessageSend(m.ChannelID, "Sorry, but you dont have permissions to move to the Destination channel")
			return
		}
		for _, member := range guild.VoiceStates {
			if member.UserID == m.Author.ID {
				if !checkPermissions(s, member.ChannelID, m.Author.ID, discordgo.PermissionVoiceMoveMembers) {
					s.ChannelMessageSend(m.ChannelID, "Sorry, but you dont have permissions to move to the Destination channel")
					return
				}
				num, err := MoveMembers(workers, guild, m.GuildID, member.ChannelID, destination)
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, "Sorry, but: "+err.Error())
					log.Println(err.Error())
					return
				}
				s.ChannelMessageSend(m.ChannelID, "I Just moved "+num+" users for you.")
				return
			}
		}
		s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+", you need to be connected to a channel for me to find you. Type '"+prefix+" move' to get help.")
	} else if length == 4 { // IF 4 parameters: Move from Origin to Destination
		log.Println("Received 4 parameter move command on " + guild.Name + " , ID: " + guild.ID + " , by :" + m.Author.ID)
		var origin string
		param2, err := strconv.Atoi(params[2])
		if err != nil {
			origin, err = ChanByName(channs, params[2])
		} else {
			origin, err = ChanByPosNum(channs, param2)
		}
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Sorry, I can't find channel "+params[2]+".")
			return
		}
		if !checkPermissions(s, origin, m.Author.ID, discordgo.PermissionVoiceMoveMembers) {
			s.ChannelMessageSend(m.ChannelID, "Sorry, but you dont have permissions to move to the Destination channel")
			return
		}
		param3, err := strconv.Atoi(params[3])
		var destination string
		if err != nil {
			destination, err = ChanByName(channs, params[3])
		} else {
			destination, err = ChanByPosNum(channs, param3)
		}
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Sorry, I can't find channel "+params[3]+".")
			return
		}
		if !checkPermissions(s, destination, m.Author.ID, discordgo.PermissionVoiceMoveMembers) {
			s.ChannelMessageSend(m.ChannelID, "Sorry, but you dont have permissions to move to the Destination channel")
			return
		}
		num, err := MoveMembers(workers, guild, m.GuildID, origin, destination)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Sorry, but: "+err.Error())
			log.Println(err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "I Just moved "+num+" users for you.")
		return
	}
	log.Println("Sending help message on " + guild.Name + " , ID: " + guild.ID)
	_, err = s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+MoveHelper(channs, prefix))
}

// MoveMembers wraps MoveAndRetry with councurrent calls and error reporting.
/*
Inputs:
	s *discordgo.Session : the session that called this handler
	guildID string : the ID of the server (guild) where the request was originated
	userID string : the ID of the user that is going to be moved
	dest string : the ID of the Voice Channel the user will be moved to
*/
func MoveMembers(servants []*discordgo.Session, guild *discordgo.Guild, id string, origin string, dest string) (string, error) {
	if origin == dest {
		return "", errors.New("destination and origin are the same")
	}
	num := 0
	var wg sync.WaitGroup
	for index, member := range guild.VoiceStates {
		if member.ChannelID == origin {
			wg.Add(1)
			go func(id, userID, dest string, worker *discordgo.Session) {
				defer wg.Done()
				err := MoveAndRetry(worker, id, userID, dest, 3)
				if err != nil {
					log.Println("Failed to move user with ID: "+userID, err)
				}
				// worker.ChannelMessageSend(m)
			}(id, member.UserID, dest, servants[index%len(servants)])
			num++
		}
	}
	wg.Wait()
	return fmt.Sprintf("Moved %d users", num), nil
}

// MoveAndRetry is a wrapper on top of discordgo.Session.GuildMemberMove with a retry function
/*
Inputs:
	s *discordgo.Session : the session that called this handler
	guildID string : the ID of the server (guild) where the request was originated
	userID string : the ID of the user that is going to be moved
	dest string : the ID of the Voice Channel the user will be moved to
	retry int: the amount of retrys this function will allows
*/
func MoveAndRetry(s *discordgo.Session, guildID, userID, dest string, retry int) error {
	err := s.GuildMemberMove(guildID, userID, dest)
	if err != nil {
		time.Sleep(time.Millisecond * 30)
		if retry >= 0 {
			return err
		}
		MoveAndRetry(s, guildID, userID, dest, retry-1)
	}
	return nil
}

// MoveHelper prints the help text for this command
/*
Inputs:
	chann []*discordgo.Channel : list of all channels in the server (used to list the numbers)
	prefix string: prefix used to call the bot (used to print in the message)

Outputs: message string
*/
func MoveHelper(channs []*discordgo.Channel, prefix string) string {
	message := `View this help and the list of available channels\n\t" + prefix + " move\n\nMove all users from your channel to the <integer:destination channel>\n\t" + prefix + " move <number:destination channel>\n\nMove all users from <integer:origin channel> to the <integer:destination channel>\n\t" + prefix + " move <number:origin channel> <number:destination channel>\n\nList of available channels:\n`
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

// ChanByPosNum retrieves channel id by the position as displayed in the channel.
/*
Inputs:
	chann []*discordgo.Channel : list of all channels in the server
	posNum integer: numer (position) of the channel

Outputs: id string, error
*/
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
/*
Inputs:
	chann []*discordgo.Channel : list of all channels in the server
	name string: name of the desired channel

Outputs: id string, error
*/
func ChanByName(channs []*discordgo.Channel, name string) (string, error) {
	name = strings.ToUpper(name)
	for _, chann := range channs {
		if chann.Type == 2 {
			if strings.ToUpper(chann.Name) == name {
				return chann.ID, nil
			}
		}
	}
	return "", errors.New("Not Found")
}

// checkPermissions checks the permission for a User in a chennel
/*
Inputs:
	s *discordgo.Session : the Bot doing the check
	channelID string : the ID of the channel to check
	userID string : the ID of the user to check
	permission int : the permission Integer ( ex discordgo.PermissionVoiceMoveMembers)

Outputs: True/False
*/
func checkPermissions(s *discordgo.Session, channelID string, userID string, permission int) bool {
	userPermission, err := s.State.UserChannelPermissions(userID, channelID)
	if err != nil || (userPermission&permission) != permission {
		return false
	}
	return true
}
