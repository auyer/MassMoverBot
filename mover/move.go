package mover

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/auyer/commanderBot/utils"
	"github.com/bwmarrin/discordgo"
)

// MoveDestination function moves discord users
func MoveDestination(s *discordgo.Session, workers []*discordgo.Session, m *discordgo.MessageCreate, guild *discordgo.Guild, prefix string, destination string) {
	for _, member := range guild.VoiceStates {
		if member.UserID == m.Author.ID {
			if !utils.CheckPermissions(s, member.ChannelID, m.Author.ID, discordgo.PermissionVoiceMoveMembers) {
				s.ChannelMessageSend(m.ChannelID, "Sorry, but you dont have permissions to move from your current channel")
				return
			}
			num, err := MoveMembers(workers, guild, member.ChannelID, destination)
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
}

// MoveOriginDestination function moves discord users
func MoveOriginDestination(s *discordgo.Session, workers []*discordgo.Session, m *discordgo.MessageCreate, guild *discordgo.Guild, prefix string, origin string, destination string) {
	num, err := MoveMembers(workers, guild, origin, destination)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Sorry, but: "+err.Error())
		log.Println(err.Error())
		return
	}
	s.ChannelMessageSend(m.ChannelID, "I Just moved "+num+" users for you.")
	return
}

// MoveMembers wraps MoveAndRetry with councurrent calls and error reporting.
/*
Inputs:
	s *discordgo.Session : the session that called this handler
	guildID string : the ID of the server (guild) where the request was originated
	userID string : the ID of the user that is going to be moved
	dest string : the ID of the Voice Channel the user will be moved to
*/
func MoveMembers(servants []*discordgo.Session, guild *discordgo.Guild, origin string, dest string) (string, error) {
	if origin == dest {
		return "", errors.New("destination and origin are the same")
	}
	num := 0
	var wg sync.WaitGroup
	for index, member := range guild.VoiceStates {
		if member.ChannelID == origin {
			wg.Add(1)
			go func(guildID, userID, dest string, servants []*discordgo.Session, index int) {
				defer wg.Done()
				err := MoveAndRetry(servants[index%len(servants)], guildID, userID, dest, 3)
				if err != nil {
					log.Println("Failed to move user with ID: "+userID, err)
				}
			}(guild.ID, member.UserID, dest, servants, index)
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
		time.Sleep(time.Millisecond * 20)
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
