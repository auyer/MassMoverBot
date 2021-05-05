package config

import (
	"testing"
)

var (
	testConfigPath = "../config.model.json"
	testConfig     = ConfigurationParameters{
		MoverBotToken: "MoverBotToken",
		PowerupTokens: []string{"powerupToken1", "powerupToken2"},
		BotPrefix:     "-",
		DatabasePath:  "./databases/",
	}
)

func TestConfRead(t *testing.T) {
	//TESTING with Config File
	conf, err := readConfig(testConfigPath)
	if err != nil {
		t.Errorf("Unable to Read Configuration: " + err.Error())
	}
	for i := range conf.PowerupTokens {
		if conf.PowerupTokens[i] != testConfig.PowerupTokens[i] {
			t.Errorf("Powerup Tokens do not match")
		}
	}
	if conf.BotPrefix == testConfig.BotPrefix && conf.MoverBotToken == testConfig.MoverBotToken && testConfig.DatabasePath == conf.DatabasePath {
		return
	}
	t.Errorf("String parameters not matching")
}

func TestFileMissing(t *testing.T) {
	//TESTING with Config File
	_, err := readConfig(testConfigPath + testConfigPath)
	if err == nil {
		t.Errorf("Unable to catch missing file exeption")
	}
}

func TestWrongFile(t *testing.T) {
	//TESTING with Config File
	_, err := readConfig("./config_test.go")
	if err == nil {
		t.Errorf("Unable to cach wrong format exeption")
	}
}
