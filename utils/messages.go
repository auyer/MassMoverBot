package utils

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// Message structure stores the messages and the amount of Formatting directives for each one of them ...
type Message struct {
	Messages           map[string]map[string]string
	FormaterDirectives map[string]map[string]int
}

// WelcomeAndLang function produces the Welcome message
func (m *Message) WelcomeAndLang(insertions string) *discordgo.MessageEmbed {
	a := make([]interface{}, m.FormaterDirectives["LANG"]["WelcomeAndLang"])
	memsetLoop(a, insertions)
	return &discordgo.MessageEmbed{
		Title:       ":globe_with_meridians:",
		Description: fmt.Sprintf(m.Messages["LANG"]["WelcomeAndLang"], a...),
		Color:       0x0099ff,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "massmover.github.io",
		},
	}
}

// LangSetupMessage produces the message for language setup
func (m *Message) LangSetupMessage(insertions string) *discordgo.MessageEmbed {
	a := make([]interface{}, m.FormaterDirectives["LANG"]["LangSetupMessage"])
	memsetLoop(a, insertions)
	return &discordgo.MessageEmbed{
		Title:       ":globe_with_meridians:",
		Description: fmt.Sprintf(m.Messages["LANG"]["LangSetupMessage"], a...),
		Color:       0x0099ff,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "massmover.github.io",
		},
	}
}

// LangSet produces the message reporting the new language set
func (m *Message) LangSet(lang string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       Flags[lang],
		Description: m.Messages[lang]["LangSet"],
		Color:       0x0099ff,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "massmover.github.io",
		},
	}
}

// CantFindChannel produces an error message for when a channel cant be found
func (m *Message) CantFindChannel(lang, insertions string) *discordgo.MessageEmbed {
	a := make([]interface{}, m.FormaterDirectives["MESSAGES"]["CantFindChannel"])
	memsetLoop(a, insertions)
	return &discordgo.MessageEmbed{
		Title:       ":x:",
		Description: fmt.Sprintf(m.Messages[lang]["CantFindChannel"], a...),
		Color:       0x0099ff,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "massmover.github.io",
		},
	}
}

// CantFindUser produces an error message for when a user cant be found
func (m *Message) CantFindUser(lang, prefix, mension string) *discordgo.MessageEmbed {
	// a := make([]interface{}, m.FormaterDirectives["MESSAGES"]["CantFindUser"])
	// memsetLoop(a, insertions)
	// variadicJoin([]interface{}{prefix}, []interface{}{mension})
	return &discordgo.MessageEmbed{
		Title:       ":x:",
		Description: fmt.Sprintf(m.Messages[lang]["CantFindUser"], prefix, mension),
		Color:       0x0099ff,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "massmover.github.io",
		},
	}
}

// CantMoveSomeUsers produces an error message for when some user cant be moved
func (m *Message) CantMoveSomeUsers(lang string) *discordgo.MessageEmbed {

	return &discordgo.MessageEmbed{
		Title:       ":x:",
		Description: m.Messages[lang]["CantMoveSomeUsers"],
		Color:       0x0099ff,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "massmover.github.io",
		},
	}
}

// BotNoPermission produces an error message for when the bot was denied permission on some action
func (m *Message) BotNoPermission(lang string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       ":x:",
		Description: m.Messages[lang]["BotNoPermission"],
		Color:       0x0099ff,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "massmover.github.io",
		},
	}
}

// GeneralHelp produces a message pointing to the help command
func (m *Message) GeneralHelp(lang, prefix, mension string) *discordgo.MessageEmbed {
	// a := make([]interface{}, m.FormaterDirectives["MESSAGES"]["GeneralHelp"])
	// memsetLoop(a, prefix)
	return &discordgo.MessageEmbed{
		Title:       ":question:",
		Description: fmt.Sprintf(m.Messages[lang]["GeneralHelp"], prefix, mension),
		Color:       0x0099ff,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "massmover.github.io",
		},
	}
}

// NotInGuild produces an error message was sent outside a guild
func (m *Message) NotInGuild(lang, insertions string) *discordgo.MessageEmbed {
	a := make([]interface{}, m.FormaterDirectives["MESSAGES"]["NotInGuild"])
	memsetLoop(a, insertions)
	return &discordgo.MessageEmbed{
		Title:       ":x:",
		Description: fmt.Sprintf(m.Messages[lang]["NotInGuild"], a...),
		Color:       0x0099ff,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "massmover.github.io",
		},
	}
}

// HelpMessage produces a message with the existing commands for the bot
func (m *Message) HelpMessage(lang, insertions string) *discordgo.MessageEmbed {
	a := make([]interface{}, m.FormaterDirectives["MESSAGES"]["HelpMessage"])
	memsetLoop(a, insertions)

	return &discordgo.MessageEmbed{
		Title:       ":question:",
		Description: fmt.Sprintf(m.Messages[lang]["HelpMessage"], a...),
		Color:       0x0099ff,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "massmover.github.io",
		},
	}
}

// JustMoved produces a message with the results of the move action
func (m *Message) JustMoved(lang, insertions string) *discordgo.MessageEmbed {
	a := make([]interface{}, m.FormaterDirectives["MESSAGES"]["JustMoved"])
	memsetLoop(a, insertions)
	return &discordgo.MessageEmbed{
		Title:       ":arrow_up_down:",
		Description: fmt.Sprintf(m.Messages[lang]["JustMoved"], a...),
		Color:       0x0099ff,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "massmover.github.io",
		},
	}
}

// MoveHelper produces a help message for the move command
func (m *Message) MoveHelper(lang, prefix, rooms string) *discordgo.MessageEmbed {
	a := make([]interface{}, m.FormaterDirectives["MESSAGES"]["MoveHelper"]-1)
	memsetLoop(a, prefix)
	a = append(a, rooms)

	return &discordgo.MessageEmbed{
		Title:       ":arrow_up_down: :question:",
		Description: fmt.Sprintf(m.Messages[lang]["MoveHelper"], a...),
		Color:       0x0099ff,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "massmover.github.io",
		},
	}

}

// SummonHelp produces a help message dor the summon command
func (m *Message) SummonHelp(lang, prefix, rooms string) *discordgo.MessageEmbed {
	a := make([]interface{}, m.FormaterDirectives["MESSAGES"]["SummonHelp"]-1)
	memsetLoop(a, prefix)
	a = append(a, rooms)
	return &discordgo.MessageEmbed{
		Title:       ":inbox_tray: :question:",
		Description: fmt.Sprintf(m.Messages[lang]["SummonHelp"], a...),
		Color:       0x0099ff,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "massmover.github.io",
		},
	}
}

// NoPermissionsDestination produces an error message for when the user has no permission on the destination channel
func (m *Message) NoPermissionsDestination(lang string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       ":x:",
		Description: m.Messages[lang]["NoPermissionsDestination"],
		Color:       0x0099ff,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "massmover.github.io",
		},
	}
}

// NoPermissionsOrigin produces an error message for when the user has no permission on the origin channel
func (m *Message) NoPermissionsOrigin(lang string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       ":x:",
		Description: m.Messages[lang]["NoPermissionsOrigin"],
		Color:       0x0099ff,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "massmover.github.io",
		},
	}
}

// SorryBut produces a error message for unnatural situations
func (m *Message) SorryBut(lang, insertions string) *discordgo.MessageEmbed {
	a := make([]interface{}, m.FormaterDirectives["MESSAGES"]["SorryBut"])
	memsetLoop(a, insertions)
	return &discordgo.MessageEmbed{
		Title:       ":x:",
		Description: fmt.Sprintf(m.Messages[lang]["SorryBut"], a...),
		Color:       0x0099ff,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "massmover.github.io",
		},
	}
}

func memsetLoop(a []interface{}, v interface{}) {
	for i := range a {
		a[i] = v
	}
}
