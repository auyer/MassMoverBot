package config

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/auyer/massmoverbot/db"
	"github.com/auyer/massmoverbot/locale"
	"github.com/bwmarrin/discordgo"
	"github.com/dgraph-io/badger"
	"github.com/rakyll/statik/fs"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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

const (
	version = "0.5-beta"
	website = "github.com/auyer/massmoverbot/"
)

var (
	red         = outer("31")
	cyan        = outer("36")
	Config      *ConfigurationStruct
	Conn        *badger.DB
	ServantList []*discordgo.Session
)

func Init() error {
	statikFS, err := fs.New()
	if err != nil {
		log.Fatal(err)
		return err
	}

	initLocales(statikFS)
	displayBanner(statikFS)

	err = initConfig()
	if err != nil {
		return err
	}

	err = initDB()
	if err != nil {
		return err
	}

	return nil
}

// ReadConfig function reads from the json file and stores the values
func initConfig() error {
	configFileLocation := flag.String("config", "./config.json", "Configuration File Location")
	flag.Parse()

	log.Print("Reading config file...")

	file, err := ioutil.ReadFile(*configFileLocation)

	if err != nil {
		log.Print(err.Error())
		return err
	}

	log.Print(string(file))

	err = json.Unmarshal(file, &Config)
	if err != nil {
		log.Print(err.Error())
		return err
	}

	return nil
}

func initDB() error {
	err := os.Mkdir(Config.DatabasesPath, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		log.Println("Error creating Databases folder: ", err)
		return err
	}

	Conn, err = db.ConnectDB(Config.DatabasesPath + "/db")
	if err != nil {
		log.Println("Error creating guildDB " + err.Error())
		return err
	}

	bytesStats, err := db.GetDataTupleBytes(Conn, "statistics")
	stats := map[string]int{}
	if err != nil {
		if err.Error() != "Key not found" {
			log.Println("Error reading guildDB " + err.Error())
			return err
		}

		log.Println("Failed to get Statistics")
		stats["usrs"] = 0
		stats["movs"] = 0
		bytesStats, _ = json.Marshal(stats)
		_ = db.UpdateDataTupleBytes(Conn, "statistics", bytesStats)
	}

	// stats := map[string]string{}
	err = json.Unmarshal(bytesStats, &stats)
	if err != nil {
		log.Println("Failed to decode Statistics")
		return err
	}

	log.Println(fmt.Sprintf("Moved %d players in %d actions", stats["usrs"], stats["movs"]))

	return nil
}

func initLocales(statikFS http.FileSystem) {
	mesagesFile, err := statikFS.Open("/messages.yaml")
	if err != nil {
		log.Print(err.Error())
		return
	}

	byteValue, _ := ioutil.ReadAll(mesagesFile)

	err = yaml.Unmarshal(byteValue, &locale.Messages)
	if err != nil {
		log.Fatal(err)
		return
	}
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
