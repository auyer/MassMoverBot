package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

var (
	// Public variables
	Token         string
	BotPrefix     string
	DatabasesPath string

	// Private variables
	config *configStruct
)

type configStruct struct {
	Token         string `json:"Token"`
	BotPrefix     string `json:"BotPrefix"`
	DatabasesPath string `json:"DatabasesPath"`
}

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
