package utils

import (
	"fmt"
)

// Message ...
type Message struct {
	Messages           map[string]map[string]string
	FormaterDirectives map[string]map[string]int
}

// WelcomeAndLang ...
func (m *Message) WelcomeAndLang(insertions string) string {
	a := make([]interface{}, m.FormaterDirectives["LANG"]["WelcomeAndLang"])
	memsetLoop(a, insertions)
	return fmt.Sprintf(m.Messages["LANG"]["WelcomeAndLang"], a...)
}

// LangSetupMessage ...
func (m *Message) LangSetupMessage(insertions string) string {
	a := make([]interface{}, m.FormaterDirectives["LANG"]["LangSetupMessage"])
	memsetLoop(a, insertions)
	return fmt.Sprintf(m.Messages["LANG"]["LangSetupMessage"], a...)
}

// LangSet ...
func (m *Message) LangSet(lang string) string {
	return m.Messages[lang]["LangSet"]
}

// CantFindChannel ...
func (m *Message) CantFindChannel(lang, insertions string) string {
	a := make([]interface{}, m.FormaterDirectives["MESSAGES"]["CantFindChannel"])
	memsetLoop(a, insertions)
	return fmt.Sprintf(m.Messages[lang]["CantFindChannel"], a...)
}

// CantFindUser ...
func (m *Message) CantFindUser(lang, prefix, mension string) string {
	// a := make([]interface{}, m.FormaterDirectives["MESSAGES"]["CantFindUser"])
	// memsetLoop(a, insertions)
	return fmt.Sprintf(m.Messages[lang]["CantFindUser"], prefix, mension)
}

// CantMoveSomeUsers ...
func (m *Message) CantMoveSomeUsers(lang string) string {
	return m.Messages[lang]["CantMoveSomeUsers"]
}

// BotNoPermission ...
func (m *Message) BotNoPermission(lang string) string {
	return m.Messages[lang]["BotNoPermission"]
}

// GeneralHelp ...
func (m *Message) GeneralHelp(lang, prefix, mension string) string {
	// a := make([]interface{}, m.FormaterDirectives["MESSAGES"]["GeneralHelp"])
	// memsetLoop(a, prefix)
	return fmt.Sprintf(m.Messages[lang]["GeneralHelp"], prefix, mension)
}

// NotInGuild ...
func (m *Message) NotInGuild(lang, insertions string) string {
	a := make([]interface{}, m.FormaterDirectives["MESSAGES"]["NotInGuild"])
	memsetLoop(a, insertions)
	return fmt.Sprintf(m.Messages[lang]["NotInGuild"], a...)
}

// HelpMessage ...
func (m *Message) HelpMessage(lang, insertions string) string {
	a := make([]interface{}, m.FormaterDirectives["MESSAGES"]["HelpMessage"])
	memsetLoop(a, insertions)
	return fmt.Sprintf(m.Messages[lang]["HelpMessage"], a...)
}

// JustMoved ...
func (m *Message) JustMoved(lang, insertions string) string {
	a := make([]interface{}, m.FormaterDirectives["MESSAGES"]["JustMoved"])
	memsetLoop(a, insertions)
	return fmt.Sprintf(m.Messages[lang]["JustMoved"], a...)
}

// MoveHelper ...
func (m *Message) MoveHelper(lang, prefix, rooms string) string {
	a := make([]interface{}, m.FormaterDirectives["MESSAGES"]["MoveHelper"]-1)
	memsetLoop(a, prefix)
	a = append(a, rooms)
	return fmt.Sprintf(m.Messages[lang]["MoveHelper"], a...)
}

// SummonHelp ...
func (m *Message) SummonHelp(lang, prefix, rooms string) string {
	a := make([]interface{}, m.FormaterDirectives["MESSAGES"]["SummonHelp"]-1)
	memsetLoop(a, prefix)
	a = append(a, rooms)
	return fmt.Sprintf(m.Messages[lang]["SummonHelp"], a...)
}

// NoPermissionsDestination ...
func (m *Message) NoPermissionsDestination(lang string) string {
	return m.Messages[lang]["NoPermissionsDestination"]
}

// NoPermissionsOrigin ...
func (m *Message) NoPermissionsOrigin(lang string) string {
	return m.Messages[lang]["NoPermissionsOrigin"]
}

// SorryBut ...
func (m *Message) SorryBut(lang, insertions string) string {
	a := make([]interface{}, m.FormaterDirectives["MESSAGES"]["SorryBut"])
	memsetLoop(a, insertions)
	return fmt.Sprintf(m.Messages[lang]["SorryBut"], a...)
}

func memsetLoop(a []interface{}, v interface{}) {
	for i := range a {
		a[i] = v
	}
}

func variadicJoin(interfaceLists ...[]interface{}) []interface{} {
	var result []interface{}
	for _, list := range interfaceLists {
		for _, item := range list {
			result = append(result, item)
		}
	}
	return result
}
