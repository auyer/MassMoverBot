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
	requesterChannel := utils.GetUserCurrentChannel(s, m.Author.ID, guild)
	if requesterChannel == "" {
		return "", errors.New("cant find user")
	}
	if !utils.CheckPermissions(s, requesterChannel, m.Author.ID, discordgo.PermissionVoiceMoveMembers) {
		return "", errors.New("no permission origin")
	}
	return MoveOriginDestination(s, workers, m, guild, prefix, requesterChannel, destination)
}

// MoveOriginDestination function moves discord users
func MoveOriginDestination(s *discordgo.Session, workers []*discordgo.Session, m *discordgo.MessageCreate, guild *discordgo.Guild, prefix string, origin string, destination string) (string, error) {
	return MoveMembers(workers, guild, origin, destination)
}

// MoveMembers wraps MoveAndRetry with councurrent calls and error reporting.
/*
Inputs:
	s *discordgo.Session : the session that called this handler
	guildID string : the ID of the server (guild) where the request was originated
	origin string : the ID of the Voice Channel the user will be moved from
	dest string : the ID of the Voice Channel the user will be moved to
*/
func MoveMembers(servants []*discordgo.Session, guild *discordgo.Guild, origin string, dest string) (string, error) {
	if origin == dest {
		return "", errors.New("destination and origin are the same")
	}
	num := 0
	var wg sync.WaitGroup
	errchan := make(chan error, 10)
	for index, member := range guild.VoiceStates {
		if member.ChannelID == origin {
			wg.Add(1)
			go func(guildID, userID, dest string, servants []*discordgo.Session, index int, errchan chan error) {
				defer wg.Done()
				err := MoveAndRetry(servants[index%len(servants)], guildID, userID, dest, 3)
				if err != nil {
					log.Println("Failed to move user with ID: "+userID, err)
					errchan <- errors.New("bot permission")
				}
			}(guild.ID, member.UserID, dest, servants, index, errchan)
			num++
		}
	}
	go func() {
		wg.Wait()
		close(errchan)
	}()
	var err error
	for errFromChan := range errchan {
		num--
		err = errFromChan
	}

	return strconv.Itoa(num), err
}

// MoveAllMembers wraps MoveAndRetry with councurrent calls and error reporting.
/*
Inputs:
	s *discordgo.Session : the session that called this handler
	m *discordgo.MessageCreate : the message event used to check for permissions to move
	guildID string : the ID of the server (guild) where the request was originated
	dest string : the ID of the Voice Channel the user will be moved to
	afk bool : move users from afk channel
*/
func MoveAllMembers(servants []*discordgo.Session, m *discordgo.MessageCreate, guild *discordgo.Guild, dest string, afk bool) (string, error) {
	num := 0
	var wg sync.WaitGroup
	errchan := make(chan error, 10)
	for index, member := range guild.VoiceStates {
		if member.ChannelID != dest {

			if !afk && guild.AfkChannelID == member.ChannelID {
				continue
			}

			wg.Add(1)
			go func(guildID, userID, dest string, servants []*discordgo.Session, index int, errchan chan error) {
				defer wg.Done()
				if !utils.CheckPermissions(servants[index%len(servants)], member.ChannelID, m.Author.ID, discordgo.PermissionVoiceMoveMembers) {
					return
				}
				err := MoveAndRetry(servants[index%len(servants)], guildID, userID, dest, 3)
				if err != nil {
					log.Println("Failed to move user with ID: "+userID, err)
					errchan <- errors.New("bot permission")
				}
			}(guild.ID, member.UserID, dest, servants, index, errchan)
			num++

		}
	}

	go func() {
		wg.Wait()
		close(errchan)
	}()
	var err error
	for errFromChan := range errchan {
		num--
		err = errFromChan
	}

	return strconv.Itoa(num), err
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
