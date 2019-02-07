package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

var (
	// ServantTokens is the Discord API token used to connect
	ServantTokens []string
	// BotPrefix is the string that shoud initiate a conversation with the bot
	BotPrefix string
	// DatabasesPath indicates the path where database files will be created
	DatabasesPath string
	// Private variables
	config *ConfigurationStruct
)

type ConfigurationStruct struct {
	CommanderToken string   `json:"CommanderToken"`
	ServantTokens  []string `json:"ServantTokens"`
	BotPrefix      string   `json:"BotPrefix"`
	DatabasesPath  string   `json:"DatabasesPath"`
}

// ReadConfig function reads from the json file and stores the values
func ReadConfig() (*ConfigurationStruct, error) {
	var config *ConfigurationStruct
	log.Print("Reading config file...")

	file, err := ioutil.ReadFile("./config.json")

	if err != nil {
		log.Print(err.Error())
		return config, err
	}

	log.Print(string(file))

	err = json.Unmarshal(file, &config)

	if err != nil {
		log.Print(err.Error())
		return nil, err
	}

	ServantTokens = config.ServantTokens
	BotPrefix = config.BotPrefix
	DatabasesPath = config.DatabasesPath

	return config, nil
}
