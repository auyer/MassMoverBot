package bot

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/auyer/massmoverbot/mover"
	"github.com/auyer/massmoverbot/utils"
	"github.com/bwmarrin/discordgo"
)

// Summon command moves all users to specified channel
func (bot *Bot) Summon(m *discordgo.MessageCreate, params []string) (string, error) {
	workerschann := make(chan []*discordgo.Session, 1)
	go utils.DetectServants(m.GuildID, append(bot.PowerupSessions, bot.CommanderSession), workerschann)

	guild, _ := bot.CommanderSession.Guild(m.GuildID)

	destination := utils.GetUserCurrentChannel(bot.CommanderSession, m.Author.ID, guild)

	if destination == "" {
		_, _ = bot.CommanderSession.ChannelMessageSend(m.ChannelID, fmt.Sprintf(bot.Messages[utils.GetGuildLocale(bot.DB, m)]["CantFindUser"], m.Author.Username))
		return "", errors.New("user not connected to any voice channel")
	}

	if !utils.CheckPermissions(bot.CommanderSession, destination, m.Author.ID, discordgo.PermissionVoiceMoveMembers) {
		_, _ = bot.CommanderSession.ChannelMessageSend(m.ChannelID, bot.Messages[utils.GetGuildLocale(bot.DB, m)]["NoPermissionsDestination"])
		return "", errors.New("no permission destination")
	}
	numParams := len(params)
	guildChannels := guild.Channels
	if numParams == 2 || numParams == 3 {
		afk := false
		if numParams == 3 {
			log.Println("Received summon command with 3 parameters on " + guild.Name + " , ID: " + guild.ID + " , by :" + m.Author.ID)
			if strings.ToLower(params[2]) != "afk" {
				_, _ = bot.CommanderSession.ChannelMessageSend(m.ChannelID, fmt.Sprintf(bot.Messages[utils.GetGuildLocale(bot.DB, m)]["SummonHelp"], bot.Prefix, bot.Prefix, bot.Prefix, bot.Prefix, bot.Prefix, utils.ListChannelsForHelpMessage(guild.Channels)))

				return "", nil
			}
			afk = true
		} else {
			log.Println("Received summon command with 2 parameters on " + guild.Name + " , ID: " + guild.ID + " , by :" + m.Author.ID)
		}
		if params[1] == "all" {

			num, err := mover.MoveAllMembers(<-workerschann, m, guild, destination, afk)
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
		}
		destination, err := utils.GetChannel(guildChannels, params[1])
		if err != nil {
			_, _ = bot.CommanderSession.ChannelMessageSend(m.ChannelID, fmt.Sprintf(bot.Messages[utils.GetGuildLocale(bot.DB, m)]["CantFindChannel"], params[1]))
			return "", err
		}

		num, err := mover.MoveAllMembers(<-workerschann, m, guild, destination, afk)
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

	}
	_, _ = bot.CommanderSession.ChannelMessageSend(m.ChannelID, fmt.Sprintf(bot.Messages[utils.GetGuildLocale(bot.DB, m)]["SummonHelp"], bot.Prefix, bot.Prefix, bot.Prefix, bot.Prefix, bot.Prefix, utils.ListChannelsForHelpMessage(guild.Channels)))

	return "", nil
}
