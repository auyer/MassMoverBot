package main

import (
	"log"

	"github.com/auyer/commanderBot/config"

	"github.com/auyer/commanderBot/bot"
)

func main() {
	err := config.ReadConfig()
	if err != nil {
		log.Print(err.Error())
		return
	}

	bot.Start()
	<-make(chan struct{})
	return
}
