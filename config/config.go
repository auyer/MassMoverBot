package config

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/auyer/massmoverbot/db"
	"github.com/auyer/massmoverbot/db/bdb"
	"github.com/auyer/massmoverbot/utils"
	"github.com/rakyll/statik/fs"
	"gopkg.in/yaml.v3"
)

// ConfigurationParameters stores the necessary info for a Multi Token bot
type ConfigurationParameters struct {
	MoverBotToken string   `json:"MoverBotToken"`
	PowerupTokens []string `json:"PowerupTokens"`
	BotPrefix     string   `json:"BotPrefix"`
	DatabasePath  string   `json:"DatabasePath"`
}

const (
	version = "1.0.1"
	website = "github.com/auyer/massmoverbot/"
)

// Init runs all steps of configuration including printing some messages to the terminal
func Init() (ConfigurationParameters, *utils.Message, db.DataStorage, error) {
	configFileLocation := flag.String("config", "./config.json", "Configuration File Location")
	flag.Parse()
	statikFS, err := fs.New()
	if err != nil {
		return ConfigurationParameters{}, nil, nil, err
	}

	messages, err := initLocales(statikFS)
	if err != nil {
		return ConfigurationParameters{}, nil, nil, err
	}
	messageFormaters, err := initLocaleFormatting(statikFS)
	if err != nil {
		return ConfigurationParameters{}, nil, nil, err
	}

	messagePack := &utils.Message{
		Messages:           messages,
		FormaterDirectives: messageFormaters,
	}

	displayBanner(statikFS)

	configs, err := readConfig(*configFileLocation)
	if err != nil {
		return configs, messagePack, nil, err
	}
	conn, err := bdb.NewBadgerDB(configs.DatabasePath)
	if err != nil {
		return configs, messagePack, conn, err
	}

	return configs, messagePack, conn, nil
}

// readConfig function reads from the json file and stores the values
func readConfig(configPath string) (ConfigurationParameters, error) {
	var config ConfigurationParameters

	log.Print("Reading config file...")

	file, err := ioutil.ReadFile(configPath)

	if err != nil {
		log.Print(err.Error())
		return config, err
	}

	log.Print(string(file))

	err = json.Unmarshal(file, &config)
	if err != nil {
		log.Print(err.Error())
		return config, err
	}

	return config, nil
}

func initLocales(statikFS http.FileSystem) (map[string]map[string]string, error) {
	mesagesFile, err := statikFS.Open("/messages.yaml")
	if err != nil {
		log.Print(err.Error())
		return nil, err
	}

	byteValue, _ := ioutil.ReadAll(mesagesFile)
	var messages map[string]map[string]string
	err = yaml.Unmarshal(byteValue, &messages)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return messages, nil
}

func initLocaleFormatting(statikFS http.FileSystem) (map[string]map[string]int, error) {
	mesagesFile, err := statikFS.Open("/messageFormaters.yaml")
	if err != nil {
		log.Print(err.Error())
		return nil, err
	}

	byteValue, _ := ioutil.ReadAll(mesagesFile)
	var messages map[string]map[string]int
	err = yaml.Unmarshal(byteValue, &messages)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return messages, nil
}

func displayBanner(statikFS http.FileSystem) {
	bannerFile, err := statikFS.Open("/banner.txt")
	if err != nil {
		log.Fatal(err)
		return
	}

	bannerBytes, _ := ioutil.ReadAll(bannerFile)
	banner := fmt.Sprintf("%s", bannerBytes)
	log.Printf(banner, red("v"+version), cyan(website))
}

// TEXT COLOUR FUNCTIONS
var (
	red  = outer("31")
	cyan = outer("36")
)

type (
	inner func(interface{}) string
)

func outer(n string) inner {
	return func(msg interface{}) string {
		b := new(bytes.Buffer)
		b.WriteString("\x1b[")
		b.WriteString(n)
		b.WriteString("m")
		return fmt.Sprintf("%s%v\x1b[0m", b.String(), msg)
	}
}
