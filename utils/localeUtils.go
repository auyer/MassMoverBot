package utils

import (
	"strconv"
	"strings"
)

var langs = map[int]string{
	1: "EN",
	2: "PT",
	3: "ES",
	4: "FR",
}

// Flags maps language codes to a emoji strings representing the flags of the countries
var Flags = map[string]string{
	"EN": ":flag_us:",
	"PT": ":flag_br:",
	"ES": ":flag_es:",
	"FR": ":flag_fr:",
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
