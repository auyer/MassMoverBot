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
	go utils.DetectPowerups(m.GuildID, append(bot.PowerupSessions, bot.MoverSession), workerschann)

	guild, err := bot.MoverSession.Guild(m.GuildID) // retrieving the server (guild) the message was originated from
	if err != nil {
		log.Println(err)
		_, _ = bot.MoverSession.ChannelMessageSend(m.ChannelID, fmt.Sprintf(bot.Messages[utils.GetGuildLocale(bot.DB, m.GuildID)]["NotInGuild"], m.Author.Mention()))
		return "", errors.New("notinguild")
	}
	numParams := len(params)

	guildChannels := guild.Channels // retrieving the list of guildChannels
	if numParams == 2 {
		log.Println("Received move command with 2 parameters on " + guild.Name + " , ID: " + guild.ID + " , by :" + m.Author.ID)
		destination, err := utils.GetChannel(guildChannels, params[1])
		if err != nil {
			_, _ = bot.MoverSession.ChannelMessageSend(m.ChannelID, fmt.Sprintf(bot.Messages[utils.GetGuildLocale(bot.DB, m.GuildID)]["CantFindChannel"], params[1]))
			return "", err
		}

		if !utils.CheckPermissions(bot.MoverSession, destination, m.Author.ID, discordgo.PermissionVoiceMoveMembers) {
			_, _ = bot.MoverSession.ChannelMessageSend(m.ChannelID, bot.Messages[utils.GetGuildLocale(bot.DB, m.GuildID)]["NoPermissionsDestination"])
			return "", err
		}

		num, err := mover.MoveDestination(bot.MoverSession, <-workerschann, m, guild, bot.Prefix, destination)
		if err != nil && num != "0" {
			_, _ = bot.MoverSession.ChannelMessageSend(m.ChannelID, bot.Messages[utils.GetGuildLocale(bot.DB, m.GuildID)]["CantMoveSomeUsers"])
		} else if err != nil {
			if err.Error() == "no permission origin" {
				_, _ = bot.MoverSession.ChannelMessageSend(m.ChannelID, bot.Messages[utils.GetGuildLocale(bot.DB, m.GuildID)]["NoPermissionsOrigin"])
			} else if err.Error() == "bot permission" {
				_, _ = bot.MoverSession.ChannelMessageSend(m.ChannelID, bot.Messages[utils.GetGuildLocale(bot.DB, m.GuildID)]["BotNoPermission"])
			} else if err.Error() == "cant find user" {
				_, _ = bot.MoverSession.ChannelMessageSend(m.ChannelID, fmt.Sprintf(bot.Messages[utils.GetGuildLocale(bot.DB, m.GuildID)]["CantFindUser"], m.Author.Mention(), bot.Prefix))
			} else {
				_, _ = bot.MoverSession.ChannelMessageSend(m.ChannelID, fmt.Sprintf(bot.Messages[utils.GetGuildLocale(bot.DB, m.GuildID)]["SorryBut"], err.Error()))
			}
			return "", err
		}

		_, _ = bot.MoverSession.ChannelMessageSend(m.ChannelID, fmt.Sprintf(bot.Messages[utils.GetGuildLocale(bot.DB, m.GuildID)]["JustMoved"], num))
		return num, err
	} else if numParams == 3 {
		log.Println("Received move command with 3 parameters on " + guild.Name + " , ID: " + guild.ID + " , by :" + m.Author.ID)
		origin, err := utils.GetChannel(guildChannels, params[1])
		if err != nil {
			_, _ = bot.MoverSession.ChannelMessageSend(m.ChannelID, fmt.Sprintf(bot.Messages[utils.GetGuildLocale(bot.DB, m.GuildID)]["CantFindChannel"], params[1]))
			return "", err
		}

		destination, err := utils.GetChannel(guildChannels, params[2])
		if err != nil {
			_, _ = bot.MoverSession.ChannelMessageSend(m.ChannelID, fmt.Sprintf(bot.Messages[utils.GetGuildLocale(bot.DB, m.GuildID)]["CantFindChannel"], params[2]))
			return "", err
		}

		if !utils.CheckPermissions(bot.MoverSession, origin, m.Author.ID, discordgo.PermissionVoiceMoveMembers) {
			_, _ = bot.MoverSession.ChannelMessageSend(m.ChannelID, bot.Messages[utils.GetGuildLocale(bot.DB, m.GuildID)]["NoPermissionsOrigin"])
			return "", err
		}

		if !utils.CheckPermissions(bot.MoverSession, destination, m.Author.ID, discordgo.PermissionVoiceMoveMembers) {
			_, _ = bot.MoverSession.ChannelMessageSend(m.ChannelID, bot.Messages[utils.GetGuildLocale(bot.DB, m.GuildID)]["NoPermissionsDestination"])
			return "", err
		}

		num, err := mover.MoveOriginDestination(bot.MoverSession, <-workerschann, m, guild, bot.Prefix, origin, destination)
		if err != nil && num != "0" {
			_, _ = bot.MoverSession.ChannelMessageSend(m.ChannelID, bot.Messages[utils.GetGuildLocale(bot.DB, m.GuildID)]["CantMoveSomeUsers"])
		} else if err != nil {
			if err.Error() == "no permission origin" {
				_, _ = bot.MoverSession.ChannelMessageSend(m.ChannelID, bot.Messages[utils.GetGuildLocale(bot.DB, m.GuildID)]["NoPermissionsOrigin"])
			} else if err.Error() == "bot permission" {
				_, _ = bot.MoverSession.ChannelMessageSend(m.ChannelID, bot.Messages[utils.GetGuildLocale(bot.DB, m.GuildID)]["BotNoPermission"])
			} else if err.Error() == "cant find user" {
				_, _ = bot.MoverSession.ChannelMessageSend(m.ChannelID, fmt.Sprintf(bot.Messages[utils.GetGuildLocale(bot.DB, m.GuildID)]["CantFindUser"], m.Author.Mention(), bot.Prefix))
			} else {
				_, _ = bot.MoverSession.ChannelMessageSend(m.ChannelID, fmt.Sprintf(bot.Messages[utils.GetGuildLocale(bot.DB, m.GuildID)]["SorryBut"], err.Error()))
			}
			return "", err
		}

		_, _ = bot.MoverSession.ChannelMessageSend(m.ChannelID, fmt.Sprintf(bot.Messages[utils.GetGuildLocale(bot.DB, m.GuildID)]["JustMoved"], num))
		return num, err
	} else {
		log.Println("Received move command with " + strconv.Itoa(numParams) + " parameter(s) (help message) on " + guild.Name + " , ID: " + guild.ID)
		_, _ = bot.MoverSession.ChannelMessageSend(m.ChannelID, fmt.Sprintf(bot.Messages[utils.GetGuildLocale(bot.DB, m.GuildID)]["MoveHelper"], bot.Prefix, bot.Prefix, bot.Prefix, utils.ListChannelsForHelpMessage(guildChannels)))
		return "0", err
	}
}
