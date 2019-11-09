package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/auyer/massmoverbot/bot"
	"github.com/auyer/massmoverbot/web"
	"github.com/auyer/massmoverbot/web/handler"

	"github.com/auyer/massmoverbot/config"
	_ "github.com/auyer/massmoverbot/statik"
)

func main() {
	botConfig, messages, conn, oauthConfig, err := config.Init()
	if err != nil {
		return
	}

	moverBot := bot.Init(botConfig, messages, conn)
	err = moverBot.Start()
	if err != nil {
		log.Fatal(err)
	}
	defer moverBot.Close()

	webHandler, err := handler.NewHandler(oauthConfig, moverBot)
	if err != nil {
		log.Fatal(err)
		return
	}

	go web.Run(webHandler, ":8080")

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-quit

}
