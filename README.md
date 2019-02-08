# CommanderBot: a Multi-Token Bot

[![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](http://godoc.org/github.com/auyer/commanderBot) [![Go Report Card](https://goreportcard.com/badge/github.com/auyer/commanderBot)](https://goreportcard.com/report/github.com/auyer/commanderBot) [![LICENSE MIT](https://img.shields.io/badge/license-MIT-brightgreen.svg)](https://img.shields.io/badge/license-MIT-brightgreen.svg) [![cover.run](https://cover.run/go/github.com/auyer/commanderBot.svg?style=flat&tag=golang-1.10)](https://cover.run/go?tag=golang-1.10&repo=github.com%2Fauyer%2FcommanderBot)

The CommanderBot is a Multi-Token Discord Bot.
This can be used to split intensive operations, or API limited opperations like the "User Voice Channel Move" opperation (limited by the API). 

Is is capable of using N "Servant" Bot connections to perform fast mass "User Move" operations, and the request will be executed by the ammount of bots connected to the server.

## Configuration

Get your Discord Bot tokens (at least one) in the [official developers portal](https://discordapp.com/developers)

The first one should be in the "CommanderToken" slot in the config file.
The rest will be the servant Tokens, and should be added in a List.

The Permission integer must be 16780288 (Move, Read messages, and Write messages).

## Installation
You can get the lastest binary here: (soon)

## Building Yourself
```go
go get -u github.com/auyer/commanderBot
```
Build using 
```go
go build .
```
If you want to cross compile it to run in a different OS or architecturte (like a rapberry pi), do:
```go
CGO_ENABLED=0 CC=arm-linux-gnueabi-cc GOOS=linux GOARCH=arm GOARM=6 go build .
```

