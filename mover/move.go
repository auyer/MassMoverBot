package mover

import (
	"errors"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/auyer/massmoverbot/utils"
	"github.com/bwmarrin/discordgo"
)

// MoveDestination function moves discord users
func MoveDestination(s *discordgo.Session, workers []*discordgo.Session, m *discordgo.MessageCreate, guild *discordgo.Guild, prefix string, destination string) (string, error) {
	for _, member := range guild.VoiceStates {
		if member.UserID == m.Author.ID {
			if !utils.CheckPermissions(s, member.ChannelID, m.Author.ID, discordgo.PermissionVoiceMoveMembers) {
				return "", errors.New("no permission origin")
			}
			return MoveOriginDestination(s, workers, m, guild, prefix, member.ChannelID, destination)
		}
	}
	return "", errors.New("cant find user")
}

// MoveOriginDestination function moves discord users
func MoveOriginDestination(s *discordgo.Session, workers []*discordgo.Session, m *discordgo.MessageCreate, guild *discordgo.Guild, prefix string, origin string, destination string) (string, error) {
	num, err := MoveMembers(workers, guild, origin, destination)
	return num, err
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
	return strconv.Itoa(num), nil
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
