# CommanderBot: a Multi-Token Bot

[![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](http://godoc.org/github.com/auyer/commanderBot) [![Go Report Card](https://goreportcard.com/badge/github.com/auyer/commanderBot)](https://goreportcard.com/report/github.com/auyer/commanderBot) [![LICENSE MIT](https://img.shields.io/badge/license-MIT-brightgreen.svg)](https://img.shields.io/badge/license-MIT-brightgreen.svg)

The CommanderBot is a Multi-Token Discord Bot.
This can be used to split intensive operations, or API limited opperations like the "User Voice Channel Move" opperation (limited by the API). 

It is capable of using N "Servant" Bot connections to perform fast mass "User Move" operations, and the request will be executed by the ammount of bots connected to the server.

## Usage

You can invite the public version of the bot in the [Bot Page](http://commandermultibot.github.io/).

The current prefix for calling the bot is `-c`.
The possible commands are:

 - `-c help` -> prints general help message
 - `-c move` -> prints move command help and all channels visible to the bot
 - `-c move destination` -> moves users from your current channel to the destination channel.

    Exemple : `-c move chat_y` , or `-c move 2`

 - `-c move origin destination` -> moves users from the origin channel to the destination channel

    Exemple : `-c move chat_x chat_y` , or `-c move 1 2`
 - `-c lang` -> prints language configuration help message
 - `-c lang option` -> changes the bot language to a specific language.

    Exemple: `-c lang EN` or `-c lang 1` will set the English language.
     
     The current options are: 
     - **1** or **EN** for English
     - **2** or **PT** or **BR** for Portuguese


## Configuration

Get your Discord Bot tokens (at least one) in the [official developers portal](https://discordapp.com/developers)

The first one should be in the "CommanderToken" slot in the config file.
The rest will be the servant Tokens, and should be added in a List.

The Permission integer for the commander must be 16780288 (Move, Read messages, and Write messages)
While the Servants can be 16777216 since they dont send any messages

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