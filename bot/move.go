package bot

import (
	"errors"
	"log"
	"strconv"

	"github.com/auyer/massmoverbot/mover"
	"github.com/auyer/massmoverbot/utils"
	"github.com/bwmarrin/discordgo"
)

// Move function deals with the possible parameters for a move command
/*
Inputs:
	m *discordgo.MessageCreate : the message received by the bot
	 params []string : all the parameters used in the message

Outputs:
	string : number of users moved by this command
 	error : a error if something wrong happened
*/
func (bot *Bot) Move(m *discordgo.MessageCreate, params []string) (string, error) {

	workerschann := make(chan []*discordgo.Session, 1)
	go utils.DetectPowerups(m.GuildID, append(bot.PowerupSessions, bot.MoverSession), workerschann)

	guild, err := bot.MoverSession.State.Guild(m.GuildID) // retrieving the server (guild) the message was originated from
	if err != nil {
		log.Println(err)
		_, _ = bot.MoverSession.ChannelMessageSendEmbed(m.ChannelID, bot.Messages.NotInGuild(bot.GetGuildLocale(m.GuildID), m.Author.Mention()))
		return "", errors.New("notinguild")
	}
	numParams := len(params)

	guildChannels := guild.Channels // retrieving the list of guildChannels
	if numParams == 2 {
		log.Println("Received move command with 2 parameters on " + guild.Name + " , ID: " + guild.ID + " , by :" + m.Author.ID)
		destination, err := utils.GetChannel(guildChannels, params[1])
		if err != nil {
			_, _ = bot.MoverSession.ChannelMessageSendEmbed(m.ChannelID, bot.Messages.CantFindChannel(bot.GetGuildLocale(m.GuildID), params[1]))
			return "", err
		}

		if !utils.CheckPermissions(bot.MoverSession, destination, m.Author.ID, discordgo.PermissionVoiceMoveMembers) {
			_, _ = bot.MoverSession.ChannelMessageSendEmbed(m.ChannelID, bot.Messages.NoPermissionsDestination(bot.GetGuildLocale(m.GuildID)))
			return "", err
		}

		num, err := mover.MoveDestination(bot.MoverSession, <-workerschann, m, guild, bot.Prefix, destination)
		if err != nil && num != "0" {
			_, _ = bot.MoverSession.ChannelMessageSendEmbed(m.ChannelID, bot.Messages.CantMoveSomeUsers(bot.GetGuildLocale(m.GuildID)))
		} else if err != nil {
			if err.Error() == "no permission origin" {
				_, _ = bot.MoverSession.ChannelMessageSendEmbed(m.ChannelID, bot.Messages.NoPermissionsOrigin(bot.GetGuildLocale(m.GuildID)))
			} else if err.Error() == "bot permission" {
				_, _ = bot.MoverSession.ChannelMessageSendEmbed(m.ChannelID, bot.Messages.BotNoPermission(bot.GetGuildLocale(m.GuildID)))
			} else if err.Error() == "cant find user" {
				_, _ = bot.MoverSession.ChannelMessageSendEmbed(m.ChannelID, bot.Messages.CantFindUser(bot.GetGuildLocale(m.GuildID), m.Author.Mention(), bot.Prefix))
			} else {
				_, _ = bot.MoverSession.ChannelMessageSendEmbed(m.ChannelID, bot.Messages.SorryBut(bot.GetGuildLocale(m.GuildID), err.Error()))
			}
			return "", err
		}

		_, _ = bot.MoverSession.ChannelMessageSendEmbed(m.ChannelID, bot.Messages.JustMoved(bot.GetGuildLocale(m.GuildID), num))
		return num, err
	} else if numParams == 3 {
		log.Println("Received move command with 3 parameters on " + guild.Name + " , ID: " + guild.ID + " , by :" + m.Author.ID)
		origin, err := utils.GetChannel(guildChannels, params[1])
		if err != nil {
			_, _ = bot.MoverSession.ChannelMessageSendEmbed(m.ChannelID, bot.Messages.CantFindChannel(bot.GetGuildLocale(m.GuildID), params[1]))
			return "", err
		}

		destination, err := utils.GetChannel(guildChannels, params[2])
		if err != nil {
			_, _ = bot.MoverSession.ChannelMessageSendEmbed(m.ChannelID, bot.Messages.CantFindChannel(bot.GetGuildLocale(m.GuildID), params[2]))
			return "", err
		}

		if !utils.CheckPermissions(bot.MoverSession, origin, m.Author.ID, discordgo.PermissionVoiceMoveMembers) {
			_, _ = bot.MoverSession.ChannelMessageSendEmbed(m.ChannelID, bot.Messages.NoPermissionsOrigin(bot.GetGuildLocale(m.GuildID)))
			return "", err
		}

		if !utils.CheckPermissions(bot.MoverSession, destination, m.Author.ID, discordgo.PermissionVoiceMoveMembers) {
			_, _ = bot.MoverSession.ChannelMessageSendEmbed(m.ChannelID, bot.Messages.NoPermissionsDestination(bot.GetGuildLocale(m.GuildID)))
			return "", err
		}

		num, err := mover.MoveOriginDestination(bot.MoverSession, <-workerschann, m, guild, bot.Prefix, origin, destination)
		if err != nil && num != "0" {
			_, _ = bot.MoverSession.ChannelMessageSendEmbed(m.ChannelID, bot.Messages.CantMoveSomeUsers(bot.GetGuildLocale(m.GuildID)))
		} else if err != nil {
			if err.Error() == "no permission origin" {
				_, _ = bot.MoverSession.ChannelMessageSendEmbed(m.ChannelID, bot.Messages.NoPermissionsOrigin(bot.GetGuildLocale(m.GuildID)))
			} else if err.Error() == "bot permission" {
				_, _ = bot.MoverSession.ChannelMessageSendEmbed(m.ChannelID, bot.Messages.BotNoPermission(bot.GetGuildLocale(m.GuildID)))
			} else if err.Error() == "cant find user" {
				_, _ = bot.MoverSession.ChannelMessageSendEmbed(m.ChannelID, bot.Messages.CantFindUser(bot.GetGuildLocale(m.GuildID), m.Author.Mention(), bot.Prefix))
			} else {
				_, _ = bot.MoverSession.ChannelMessageSendEmbed(m.ChannelID, bot.Messages.SorryBut(bot.GetGuildLocale(m.GuildID), err.Error()))
			}
			return "", err
		}

		_, _ = bot.MoverSession.ChannelMessageSendEmbed(m.ChannelID, bot.Messages.JustMoved(bot.GetGuildLocale(m.GuildID), num))
		return num, err
	} else {
		log.Println("Received move command with " + strconv.Itoa(numParams) + " parameter(s) (help message) on " + guild.Name + " , ID: " + guild.ID)
		_, _ = bot.MoverSession.ChannelMessageSendEmbed(m.ChannelID, bot.Messages.MoveHelper(bot.GetGuildLocale(m.GuildID), bot.Prefix, utils.ListChannelsForHelpMessage(guildChannels)))
		return "0", err
	}
}
