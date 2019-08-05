package utils

import (
	"strconv"
	"strings"

	"github.com/auyer/massmoverbot/db"
	"github.com/dgraph-io/badger"
)

var langs = map[int]string{
	1: "EN",
	2: "PT",
	3: "ES",
	4: "FR",
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
	case "ES":
		return "ES"
	case "FR":
		return "FR"
	default:
		return "EN"
	}
}

// GetGuildLocale function will return the language for a guild, returning EN by default.
/*
Input:
	conn *badger.DB : a connection with the badger db
	GuildID string : the ID of the guild
Output:
	language string
*/
func GetGuildLocale(conn *badger.DB, GuildID string) string {
	lang, err := db.GetDataTuple(conn, GuildID)
	if err != nil {
		lang = "EN"
	}

	return lang
}
