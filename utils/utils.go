package utils

import (
	"errors"
	"sort"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// GetUserCurrentChannel find the voice channel id for a connected user. Returns "" in case the user cant be found.
/*
Inputs:
	s *discordgo.Session : a bot session in the guild
	userID : the ID of the user
	guild *discordgo.Guild : the guild where to operate this function

Outputs: channel ID string, error
*/
func GetUserCurrentChannel(s *discordgo.Session, userID string, guild *discordgo.Guild) string {
	for _, member := range guild.VoiceStates {
		if member.UserID == userID {
			return member.ChannelID
		}
	}
	return ""
}

// GetChannel retrieves the channel ID by the name or position ID
/*
Inputs:
	channs []*discordgo.Channel : list of all channels in the server
	channelNameOrID string : the channel name or position integer

Outputs: channel ID string, error
*/
func GetChannel(channs []*discordgo.Channel, channelNameOrID string) (string, error) {
	var channel string
	var err error

	channelID, err := strconv.Atoi(channelNameOrID)
	if err != nil || channelID > 1000 {
		channel, err = chanByName(channs, channelNameOrID)
	} else {
		channel, err = chanByPosNum(channs, channelID-1)
	}
	if err != nil {
		return "", err
	}
	return channel, nil
}

// chanByPosNum retrieves channel id by the position as displayed in the channel.
/*
Inputs:
	chann []*discordgo.Channel : list of all channels in the server
	posNum integer: numer (position) of the channel

Outputs: id string, error
*/
func chanByPosNum(channs []*discordgo.Channel, posNum int) (string, error) {
	for _, chann := range channs {
		if chann.Type == 2 {
			if chann.Position == posNum {
				return chann.ID, nil
			}
		}
	}
	return "", errors.New("not nound")
}

// chanByName retrieves channel id by name. The comparison is case insensitive.
/*
Inputs:
	chann []*discordgo.Channel : list of all channels in the server
	name string: name of the desired channel

Outputs: id string, error
*/
func chanByName(channs []*discordgo.Channel, name string) (string, error) {
	name = strings.ToUpper(name)
	for _, chann := range channs {
		if chann.Type == 2 {
			if strings.ToUpper(chann.Name) == name {
				return chann.ID, nil
			}
		}
	}
	return "", errors.New("not found")
}

// CheckPermissions checks the permission for a User in a chennel
/*
Inputs:
	s *discordgo.Session : the Bot doing the check
	channelID string : the ID of the channel to check
	userID string : the ID of the user to check
	permission int : the permission Integer ( ex discordgo.PermissionVoiceMoveMembers)

Outputs: True/False
*/
func CheckPermissions(s *discordgo.Session, channelID string, userID string, permission int) bool {
	userPermission, err := s.State.UserChannelPermissions(userID, channelID)
	if err != nil || (userPermission&permission) != permission {
		return false
	}
	return true
}

// DetectPowerups will retrieve all bots logged in to the discord server with the provided ID.
/*
This function in ment to be used cuncurently.

Inputs:
	guildID string : the Discord Guild ID
	sseravnts []*discordgo.Session : all logged in Bot Sessions
	rchan chan []*discordgo.Session : the return channel

Output is sent in the rchan channel
*/
func DetectPowerups(guildID string, servants []*discordgo.Session, rchan chan []*discordgo.Session) {
	var workers []*discordgo.Session
	for _, powerup := range servants {
		_, err := powerup.State.Guild(guildID)
		if err == nil {
			workers = append(workers, powerup)
		}
	}
	rchan <- workers
}

// MoveHelper prints the help text for this command
/*
Inputs:
	voiceChannels []*discordgo.Channel : list of all channels in the server (used to list the numbers)
	prefix string: prefix used to call the bot (used to print in the message)

Outputs: message string
*/
func ListChannelsForHelpMessage(channels []*discordgo.Channel) string {
	sort.Slice(channels[:], func(i, j int) bool {
		return channels[i].Position < channels[j].Position
	})

	i := 0
	channelHelpList := ""
	for _, channel := range channels {
		if channel.Type == discordgo.ChannelTypeGuildVoice {
			i++
			channelHelpList = channelHelpList + strconv.Itoa(i) + " ) " + channel.Name + "\n"
		}
	}

	return channelHelpList
}

// AskMember function is used to send a private message to a guild member
func AskMember(s *discordgo.Session, member string, message string) error {
	c, err := s.UserChannelCreate(member)
	if err != nil {
		return err
	}
	_, err = s.ChannelMessageSend(c.ID, message) // event.Guild.OwnerID
	return err
}

// HaveIAskedMember function is used to send a private message to a guild member
/*
This function will return false in case of error and if the user recieved any message from a bot.
It checks the 10 last messages in the chat.
*/
func HaveIAskedMember(s *discordgo.Session, member string) bool {
	c, err := s.UserChannelCreate(member)
	if err != nil {
		return false
	}
	messages, err := s.ChannelMessages(c.ID, 10, "", "", "") // reading 10 messages to overcome possible user-sent messages
	if err != nil {
		return false
	}
	for _, message := range messages {
		if message.Author.Bot {
			return true
		}
	}
	return false
}
