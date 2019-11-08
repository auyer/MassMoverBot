package bot

import (
	"bytes"
	"encoding/base64"
	"image/png"
	"log"

	"github.com/auyer/massmoverbot/web/service"
	"github.com/bwmarrin/discordgo"
)

// GetGuildIDs will return all guild IDS acessible by the Bot
func (bot *Bot) GetGuildIDs() []string {
	var ids []string
	for _, guild := range bot.MoverSession.State.Guilds {
		ids = append(ids, guild.ID)
	}
	return ids
}

// GetGuilds will return a simple version of the guilds with the provided IDs
func (bot *Bot) GetGuilds(guildIDs ...string) []*service.SimpleGuild {
	var guilds []*service.SimpleGuild
	for _, guildID := range guildIDs {
		guild, err := bot.MoverSession.Guild(guildID) // TODO: Filter information
		if err != nil {
			log.Println("Failed to get server")
		}
		imgBase64Str := ""
		img, err := bot.MoverSession.GuildIcon(guildID)
		if err != nil {
			log.Println("Failed to get image. ", err)
		} else {
			buf := new(bytes.Buffer)
			err = png.Encode(buf, img)
			if err != nil {
				log.Println("Icon error: ", err)
			} else {
				imgBase64Str = base64.StdEncoding.EncodeToString(buf.Bytes())
			}
		}
		guilds = append(guilds, &service.SimpleGuild{ID: guild.ID, Name: guild.Name, Icon: imgBase64Str, MemberCount: guild.MemberCount})
	}
	return guilds
}

// GetGuild will return a full Guild object for a given ID
func (bot *Bot) GetGuild(guildID string) *discordgo.Guild {

	guild, err := bot.MoverSession.Guild(guildID)
	if err != nil {
		log.Println("Failed to get server")
	}

	return guild //&SimpleGuild{guild.ID, guild.Name, "", guild.MemberCount}
	// return guild
}
