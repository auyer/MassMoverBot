package bot

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/auyer/massmoverbot/mover"
	"github.com/auyer/massmoverbot/utils"
	"github.com/bwmarrin/discordgo"
)

func (bot *Bot) Move(m *discordgo.MessageCreate, params []string) (string, error) {

	workerschann := make(chan []*discordgo.Session, 1)
	go utils.DetectServants(m.GuildID, append(bot.PowerupSessions, bot.CommanderSession), workerschann)

	guild, err := bot.CommanderSession.Guild(m.GuildID) // retrieving the server (guild) the message was originated from
	if err != nil {
		log.Println(err)
		_, _ = bot.CommanderSession.ChannelMessageSend(m.ChannelID, fmt.Sprintf(bot.Messages[utils.GetGuildLocale(bot.DB, m)]["NotInGuild"], m.Author.Mention()))
		return "", errors.New("notinguild")
	}
	numParams := len(params)

	guildChannels := guild.Channels // retrieving the list of guildChannels
	if numParams == 2 {
		log.Println("Received move command with 2 parameters on " + guild.Name + " , ID: " + guild.ID + " , by :" + m.Author.ID)
		destination, err := utils.GetChannel(guildChannels, params[1])
		if err != nil {
			_, _ = bot.CommanderSession.ChannelMessageSend(m.ChannelID, fmt.Sprintf(bot.Messages[utils.GetGuildLocale(bot.DB, m)]["CantFindChannel"], params[1]))
			return "", err
		}

		if !utils.CheckPermissions(bot.CommanderSession, destination, m.Author.ID, discordgo.PermissionVoiceMoveMembers) {
			_, _ = bot.CommanderSession.ChannelMessageSend(m.ChannelID, bot.Messages[utils.GetGuildLocale(bot.DB, m)]["NoPermissionsDestination"])
			return "", err
		}

		num, err := mover.MoveDestination(bot.CommanderSession, <-workerschann, m, guild, bot.Prefix, destination)
		if err != nil && num != "0" {
			_, _ = bot.CommanderSession.ChannelMessageSend(m.ChannelID, bot.Messages[utils.GetGuildLocale(bot.DB, m)]["CantMoveSomeUsers"])
		} else if err != nil {
			if err.Error() == "no permission origin" {
				_, _ = bot.CommanderSession.ChannelMessageSend(m.ChannelID, bot.Messages[utils.GetGuildLocale(bot.DB, m)]["NoPermissionsOrigin"])
			} else if err.Error() == "bot permission" {
				_, _ = bot.CommanderSession.ChannelMessageSend(m.ChannelID, bot.Messages[utils.GetGuildLocale(bot.DB, m)]["BotNoPermission"])
			} else if err.Error() == "cant find user" {
				_, _ = bot.CommanderSession.ChannelMessageSend(m.ChannelID, fmt.Sprintf(bot.Messages[utils.GetGuildLocale(bot.DB, m)]["CantFindUser"], m.Author.Mention(), bot.Prefix))
			} else {
				_, _ = bot.CommanderSession.ChannelMessageSend(m.ChannelID, fmt.Sprintf(bot.Messages[utils.GetGuildLocale(bot.DB, m)]["SorryBut"], err.Error()))
			}
			return "", err
		}

		_, _ = bot.CommanderSession.ChannelMessageSend(m.ChannelID, fmt.Sprintf(bot.Messages[utils.GetGuildLocale(bot.DB, m)]["JustMoved"], num))
		return num, err
	} else if numParams == 3 {
		log.Println("Received move command with 3 parameters on " + guild.Name + " , ID: " + guild.ID + " , by :" + m.Author.ID)
		origin, err := utils.GetChannel(guildChannels, params[1])
		if err != nil {
			_, _ = bot.CommanderSession.ChannelMessageSend(m.ChannelID, fmt.Sprintf(bot.Messages[utils.GetGuildLocale(bot.DB, m)]["CantFindChannel"], params[1]))
			return "", err
		}

		destination, err := utils.GetChannel(guildChannels, params[2])
		if err != nil {
			_, _ = bot.CommanderSession.ChannelMessageSend(m.ChannelID, fmt.Sprintf(bot.Messages[utils.GetGuildLocale(bot.DB, m)]["CantFindChannel"], params[2]))
			return "", err
		}

		if !utils.CheckPermissions(bot.CommanderSession, origin, m.Author.ID, discordgo.PermissionVoiceMoveMembers) {
			_, _ = bot.CommanderSession.ChannelMessageSend(m.ChannelID, bot.Messages[utils.GetGuildLocale(bot.DB, m)]["NoPermissionsOrigin"])
			return "", err
		}

		if !utils.CheckPermissions(bot.CommanderSession, destination, m.Author.ID, discordgo.PermissionVoiceMoveMembers) {
			_, _ = bot.CommanderSession.ChannelMessageSend(m.ChannelID, bot.Messages[utils.GetGuildLocale(bot.DB, m)]["NoPermissionsDestination"])
			return "", err
		}

		num, err := mover.MoveOriginDestination(bot.CommanderSession, <-workerschann, m, guild, bot.Prefix, origin, destination)
		if err != nil && num != "0" {
			_, _ = bot.CommanderSession.ChannelMessageSend(m.ChannelID, bot.Messages[utils.GetGuildLocale(bot.DB, m)]["CantMoveSomeUsers"])
		} else if err != nil {
			if err.Error() == "no permission origin" {
				_, _ = bot.CommanderSession.ChannelMessageSend(m.ChannelID, bot.Messages[utils.GetGuildLocale(bot.DB, m)]["NoPermissionsOrigin"])
			} else if err.Error() == "bot permission" {
				_, _ = bot.CommanderSession.ChannelMessageSend(m.ChannelID, bot.Messages[utils.GetGuildLocale(bot.DB, m)]["BotNoPermission"])
			} else if err.Error() == "cant find user" {
				_, _ = bot.CommanderSession.ChannelMessageSend(m.ChannelID, fmt.Sprintf(bot.Messages[utils.GetGuildLocale(bot.DB, m)]["CantFindUser"], m.Author.Mention(), bot.Prefix))
			} else {
				_, _ = bot.CommanderSession.ChannelMessageSend(m.ChannelID, fmt.Sprintf(bot.Messages[utils.GetGuildLocale(bot.DB, m)]["SorryBut"], err.Error()))
			}
			return "", err
		}

		_, _ = bot.CommanderSession.ChannelMessageSend(m.ChannelID, fmt.Sprintf(bot.Messages[utils.GetGuildLocale(bot.DB, m)]["JustMoved"], num))
		return num, err
	} else {
		log.Println("Received move command with " + strconv.Itoa(numParams) + " parameter(s) (help message) on " + guild.Name + " , ID: " + guild.ID)
		_, _ = bot.CommanderSession.ChannelMessageSend(m.ChannelID, fmt.Sprintf(bot.Messages[utils.GetGuildLocale(bot.DB, m)]["MoveHelper"], bot.Prefix, bot.Prefix, bot.Prefix, utils.ListChannelsForHelpMessage(guildChannels)))
		return "0", err
	}
}
