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

	"gopkg.in/yaml.v2"

	_ "github.com/auyer/commanderBot/statik"
	"github.com/rakyll/statik/fs"

	"github.com/auyer/commanderBot/config"

	"github.com/auyer/commanderBot/bot"
)

const (
	version = "0.3-beta"
	website = "github.com/auyer/commanderbot/"
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

	mesagesFile, err := statikFS.Open("/messages.yaml")
	if err != nil {
		log.Print(err.Error())
		return
	}

	byteValue, _ := ioutil.ReadAll(mesagesFile)

	var messages map[string]string

	err = yaml.Unmarshal(byteValue, &messages)
	if err != nil {
		log.Fatal(err)
		return
	}

	bot.Start(config.CommanderToken, config.ServantTokens, config.BotPrefix, messages)
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
