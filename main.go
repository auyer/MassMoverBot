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
	config, err := config.ReadConfig()
	if err != nil {
		log.Print(err.Error())
		return
	}

	bot.Start(config.CommanderToken, config.ServantTokens, config.BotPrefix)
	defer bot.Close()
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-quit

}
