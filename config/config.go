package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

// var (
// 	DatabasesPath string
// )

// ConfigurationStruct stores the necessary info for a Multi Token bot
type ConfigurationStruct struct {
	CommanderToken string   `json:"CommanderToken"`
	ServantTokens  []string `json:"ServantTokens"`
	BotPrefix      string   `json:"BotPrefix"`
	DatabasesPath  string   `json:"DatabasesPath"`
}

// ReadConfig function reads from the json file and stores the values
func ReadConfig(configFileLocation string) (*ConfigurationStruct, error) {
	var config *ConfigurationStruct
	log.Print("Reading config file...")

	file, err := ioutil.ReadFile(configFileLocation)

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

	// DatabasesPath = config.DatabasesPath

	return config, nil
}
