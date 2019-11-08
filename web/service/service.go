package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/oauth2"
)

// // DiscordService interface ...
// type DiscordService interface {
// 	GetGuilds(IDs ...string) []*SimpleGuild
// 	GetGuild(ID string) *discordgo.Guild
// 	GetGuildIDs() []string
// }

// Bot interface ...
type Bot interface {
	GetGuild(guildID string) *discordgo.Guild
	GetGuilds(guildIDs ...string) []*SimpleGuild
	GetGuildIDs() []string
}

// type Service struct {
// 	bot Bot
// }

// func NewService(bot Bot) *Service {
// 	return &Service{
// 		bot,
// 	}
// }

// func (s Service) GetGuild(id string) *discordgo.Guild {
// 	return s.bot.GetGuild(id)
// }
// func (s Service) GetGuilds(guildIDs ...string) []*SimpleGuild {
// 	return s.bot.GetGuilds(guildIDs...)
// }
// func (s Service) GetGuildIDs() []string {
// 	return (s.bot.GetGuildIDs())
// }

// UserRequest function will reach discord API with the user Token and retrieve its basic informations
func UserRequest(ctx context.Context, token *oauth2.Token) (*discordgo.User, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://discordapp.com/api/users/@me", nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to create Request. %s", err)
	}
	req.Header.Add("authorization", fmt.Sprintf("%s %s", token.TokenType, token.AccessToken))
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	var user discordgo.User
	err = json.NewDecoder(res.Body).Decode(&user)
	return &user, err
}

// GuildsRequest function will reach discord API with the user Token and retrieve its guilds
func GuildsRequest(ctx context.Context, token *oauth2.Token) ([]*discordgo.Guild, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://discordapp.com/api/users/@me/guilds", nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to create Request. %s", err)
	}
	req.Header.Add("authorization", fmt.Sprintf("%s %s", token.TokenType, token.AccessToken))
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	var guilds []*discordgo.Guild
	err = json.NewDecoder(res.Body).Decode(&guilds)
	return guilds, err
}
