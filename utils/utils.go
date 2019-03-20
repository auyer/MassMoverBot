package utils

import (
	"errors"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var langs = map[int]string{
	1: "EN",
	2: "PT",
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
	var intparam int
	if intparam, err = strconv.Atoi(channelNameOrID); err != nil {
		channel, err = chanByName(channs, channelNameOrID)
	} else {
		channel, err = chanByPosNum(channs, intparam)
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
	return "", errors.New("Not Found")
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

// DetectServants will retrieve all bots logged in to the discord server with the provided ID.
/*
This function in ment to be used cuncurently.

Inputs:
	guildID string : the Discord Guild ID
	sseravnts []*discordgo.Session : all logged in Bot Sessions
	rchan chan []*discordgo.Session : the return channel

Output is sent in the rchan channel
*/
func DetectServants(guildID string, servants []*discordgo.Session, rchan chan []*discordgo.Session) {
	workers := []*discordgo.Session{}
	for _, servant := range servants {
		_, err := servant.State.Guild(guildID)
		if err == nil {
			workers = append(workers, servant)
		}
	}
	rchan <- workers
}

// SelectLang selects a language code based on number or string code
/*
Input:
	choice string
Output:
	language string
*/
func SelectLang(choice string) string {
	if intparam, err := strconv.Atoi(choice); err == nil {
		choice := langs[intparam]
		if choice != "" {
			return choice
		}
		return langs[1]
	}
	switch strings.ToUpper(choice) {
	case "EN":
		return "EN"
	case "PT":
		return "PT"
	case "BR":
		return "PT"
	default:
		return "EN"
	}
}
