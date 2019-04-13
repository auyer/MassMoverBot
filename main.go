package main

import (
	"github.com/auyer/massmoverbot/bot"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/auyer/massmoverbot/config"
	_ "github.com/auyer/massmoverbot/statik"
)

func main() {
	if config.Init() != nil {
		return
	}

	err := bot.Start()
	if err != nil {
		log.Fatal(err)
	} else {
		defer bot.Close()
		defer config.Conn.Close()
	}

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-quit

}
