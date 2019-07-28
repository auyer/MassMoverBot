package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/auyer/massmoverbot/bot"

	"github.com/auyer/massmoverbot/config"
	_ "github.com/auyer/massmoverbot/statik"
)

func main() {
	config, messages, conn, err := config.Init()
	if err != nil {
		return
	}

	moverBot := bot.Init(config, messages, conn)
	err = moverBot.Start()
	if err != nil {
		log.Fatal(err)
	}
	defer moverBot.Close()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-quit

}
