# MassMover bot: a Multi-Token Bot

[![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](http://godoc.org/github.com/auyer/massmoverbot) [![Go Report Card](https://goreportcard.com/badge/github.com/auyer/massmoverbot)](https://goreportcard.com/report/github.com/auyer/massmoverbot) [![LICENSE MIT](https://img.shields.io/badge/license-MIT-brightgreen.svg)](https://img.shields.io/badge/license-MIT-brightgreen.svg)

The MassMover bot is a Multi-Token Discord Bot.
This can be used to split intensive operations, or API limited opperations like the "User Voice Channel Move" opperation (limited by the API). 

It is capable of using N "PowerUp" Bot connections to perform fast mass "User Move" operations, and the request will be executed by the ammount of bots connected to the server.

## Usage

You can invite the public version of the bot in the [Bot Page](http://massmover.github.io/).

The current prefix for calling the bot is `>`.
The possible commands are:

 - `> help` -> prints general help message
 - `> move` -> prints move command help and all channels visible to the bot
 - `> move destination` -> moves users from your current channel to the destination channel.

    Exemple : `> move chat_y` , or `> move 2`

 - `> move origin destination` -> moves users from the origin channel to the destination channel

    Exemple : `> move chat_x chat_y` , or `> move 1 2`
 - `> lang` -> prints language configuration help message
 - `> lang option` -> changes the bot language to a specific language.

    Exemple: `> lang EN` or `> lang 1` will set the English language.
     
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
go get -u github.com/auyer/massmoverbot
```
Build using 
```go
go build .
```
If you want to cross compile it to run in a different OS or architecturte (like a rapberry pi), do:
```go
CGO_ENABLED=0 CC=arm-linux-gnueabi-cc GOOS=linux GOARCH=arm GOARM=6 go build .
```

## Changing/Building the messages

All messages are stored in [public/messages.yaml](public/messages.yaml) file, and loaded by the `Statik` pre compilation.
To build the messages, it is necessary to get the statik package, and run the command in the root if the commanderBot repository.

````
go get github.com/rakyll/statik
statik
```