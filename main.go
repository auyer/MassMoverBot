package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"

	"gopkg.in/yaml.v3"

	_ "github.com/auyer/massmoverbot/statik"
	"github.com/rakyll/statik/fs"

	"github.com/auyer/massmoverbot/config"
	"github.com/auyer/massmoverbot/db"

	"github.com/auyer/massmoverbot/bot"
)

const (
	version = "0.5-beta"
	website = "github.com/auyer/massmoverbot/"
)

func main() {

	statikFS, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}

	bannerFile, err := statikFS.Open("/banner.txt")
	if err != nil {
		log.Fatal(err)
	}
	bannerBytes, _ := ioutil.ReadAll(bannerFile)

	banner := fmt.Sprintf("%s", bannerBytes)

	log.Printf(banner, red("v"+version), cyan(website))
	configFile := flag.String("config", "./config.json", "Configuration File Location")
	flag.Parse()

	config, err := config.ReadConfig(*configFile)
	if err != nil {
		log.Print(err.Error())
		return
	}

	if _, err := os.Stat(config.DatabasesPath); os.IsNotExist(err) {
		err = os.Mkdir(config.DatabasesPath, os.ModePerm)
		if err != nil && err.Error() != "file exists" {
			log.Println("Error creating Databases folder: ", err)
			return
		}
	}
	conn, err := db.ConnectDB(config.DatabasesPath + "/db")
	if err != nil {
		log.Println("Error creating guildDB " + err.Error())
		return
	}
	bot.GetAndInitStats(conn)

	mesagesFile, err := statikFS.Open("/messages.yaml")
	if err != nil {
		log.Print(err.Error())
		return
	}

	byteValue, _ := ioutil.ReadAll(mesagesFile)

	var messages map[string]map[string]string

	err = yaml.Unmarshal(byteValue, &messages)
	if err != nil {
		log.Fatal(err)
		return
	}

	bot.Start(config.CommanderToken, config.ServantTokens, config.BotPrefix, conn, messages)
	defer bot.Close()
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-quit

}

// TEXT COLOUR FUNCTIONS
type (
	inner func(interface{}) string
)

var (
	red  = outer("31")
	cyan = outer("36")
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
