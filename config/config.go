package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

var (
	// Token is the Discord API token used to connect
	Token string
	// BotPrefix is the string that shoud initiate a conversation with the bot
	BotPrefix string
	// DatabasesPath indicates the path where database files will be created
	DatabasesPath string
	// Private variables
	config *configStruct
)

type configStruct struct {
	Token         string `json:"Token"`
	BotPrefix     string `json:"BotPrefix"`
	DatabasesPath string `json:"DatabasesPath"`
}

// ReadConfig function reads from the json file and stores the values
func ReadConfig() error {
	log.Print("Reading config file...")

	file, err := ioutil.ReadFile("./config.json")

	if err != nil {
		log.Print(err.Error())
		return err
	}

	log.Print(string(file))

	err = json.Unmarshal(file, &config)

	if err != nil {
		log.Print(err.Error())
		return err
	}

	Token = config.Token
	BotPrefix = config.BotPrefix
	DatabasesPath = config.DatabasesPath

	return nil
}
