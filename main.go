package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

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
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	bot.Bot.Close()
}
