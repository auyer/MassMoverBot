package config

import (
	"testing"
)

var (
	testConfigPath = "../config.model.json"
	testConfig     = ConfigurationStruct{
		CommanderToken: "commanderToken",
		ServantTokens:  []string{"servantToken1", "servantToken2"},
		BotPrefix:      "-c",
		DatabasesPath:  "./databases/",
	}
)

func TestConfRead(t *testing.T) {
	//TESTING with Config File
	conf, err := ReadConfig(testConfigPath)
	if err != nil {
		t.Errorf("Unable to Read Configuration: " + err.Error())
	}
	for i := range conf.ServantTokens {
		if conf.ServantTokens[i] != testConfig.ServantTokens[i] {
			t.Errorf("Servant Tokens do not match")
		}
	}
	if conf.BotPrefix == testConfig.BotPrefix && conf.CommanderToken == testConfig.CommanderToken && conf.DatabasesPath == conf.DatabasesPath {
		return
	}
	t.Errorf("String parameters not matching")
}

func TestFileMissing(t *testing.T) {
	//TESTING with Config File
	_, err := ReadConfig(testConfigPath + testConfigPath)
	if err == nil {
		t.Errorf("Unable to catch missing file exeption")
	}
}

func TestWrongFile(t *testing.T) {
	//TESTING with Config File
	_, err := ReadConfig("./config_test.go")
	if err == nil {
		t.Errorf("Unable to cach wrong format exeption")
	}
}
