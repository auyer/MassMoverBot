package config

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/auyer/massmoverbot/db"
	"github.com/auyer/massmoverbot/utils"
	"github.com/dgraph-io/badger"
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
	version = "1.0"
	website = "github.com/auyer/massmoverbot/"
)

// Init runs all steps of configuration including printing some messages to the terminal
func Init() (ConfigurationParameters, *utils.Message, *badger.DB, error) {
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

	conn, err := initDB(configs.DatabasePath)
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

func initDB(DatabasePath string) (*badger.DB, error) {
	err := os.Mkdir(DatabasePath, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		log.Println("Error creating Databases folder: ", err)
		return nil, err
	}

	conn, err := db.ConnectDB(DatabasePath + "/db")
	if err != nil {
		log.Println("Error creating guildDB " + err.Error())
		return nil, err
	}

	bytesStats, err := db.GetDataTupleBytes(conn, "statistics")
	stats := map[string]int{}
	if err != nil {
		if err.Error() != "Key not found" {
			log.Println("Error reading guildDB " + err.Error())
			return conn, err
		}

		log.Println("Failed to get Statistics")
		stats["usrs"] = 0
		stats["movs"] = 0
		bytesStats, _ = json.Marshal(stats)
		_ = db.UpdateDataTupleBytes(conn, "statistics", bytesStats)
	}

	// stats := map[string]string{}
	err = json.Unmarshal(bytesStats, &stats)
	if err != nil {
		log.Println("Failed to decode Statistics")
		return conn, err
	}

	log.Println(fmt.Sprintf("Moved %d players in %d actions", stats["usrs"], stats["movs"]))

	return conn, nil
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
