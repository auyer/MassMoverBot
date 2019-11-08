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
	config, messages, conn, oauthConfig, err := config.Init()
	if err != nil {
		return
	}

	moverBot := bot.Init(config, messages, conn)
	err = moverBot.Start()
	if err != nil {
		log.Fatal(err)
	}
	defer moverBot.Close()

	webHandler := handler.NewHandler(oauthConfig, moverBot)

	go web.Run(webHandler, ":80")

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-quit

}
