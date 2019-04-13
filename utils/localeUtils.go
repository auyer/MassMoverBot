package utils

import (
	"github.com/auyer/massmoverbot/config"
	"github.com/auyer/massmoverbot/db"
	"github.com/bwmarrin/discordgo"
	"strconv"
	"strings"
)

var langs = map[int]string{
	1: "EN",
	2: "PT",
}

// SelectLang selects a language code based on number or string code
/*
Input:
	choice string
Output:
	language string
*/func SelectLang(choice string) string {
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

func GetGuildLocale(m *discordgo.MessageCreate) string {
	lang, err := db.GetDataTuple(config.Conn, m.GuildID)
	if err != nil {
		lang = "EN"
	}

	return lang
}
