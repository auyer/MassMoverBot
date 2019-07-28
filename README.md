# MassMover Discord bot

[![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](http://godoc.org/github.com/auyer/massmoverbot)
[![Go Report Card](https://goreportcard.com/badge/github.com/auyer/massmoverbot)](https://goreportcard.com/report/github.com/auyer/massmoverbot)
[![LICENSE MIT](https://img.shields.io/badge/license-MIT-brightgreen.svg)](https://img.shields.io/badge/license-MIT-brightgreen.svg) <!--  [![Release](https://img.shields.io/github/release/auyer/massmoverbot.svg)](https://github.com/auyer/massmoverbot/releases/latest) -->

The MassMover bot is a Multi-Token Discord Bot.
This can be used to split intensive operations, or API limited opperations like the "User Voice Channel Move" opperations.

It is capable of using N "PowerUp" Bot connections to perform fast mass "User Move" operations, and the request will be executed by the ammount of bots connected to the server.

![GIF: moving 28 user with 2 "powerups"](https://raw.githubusercontent.com/auyer/MassMoverHugoPage/master/static/img/half.gif)

## Usage

You can invite the public version of the bot in its page: [massmover.github.io](http://massmover.github.io/).

## - > [Invite the Bot](http://massmover.github.io/)

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
The rest will be the powerup Tokens, and should be added in a List.

The Permission integer for the commander must be 16780288 (Move, Read messages, and Write messages)
While the Servants can be 16777216 since they dont send any messages

## Installation
You can get the lastest binary [here](https://github.com/auyer/massmoverbot/releases/latest).

Unzip it, and create a configuration file in the same directory, or point to it when executing with the  `-config` flag.
## Building Yourself
```go
go get -u github.com/auyer/massmoverbot
```
Build using 
```go
go build .
```

## Changing and building the messages

All messages are stored in [public/messages.yaml](public/messages.yaml) file, and loaded by the `Statik` pre compilation.
To build the messages, it is necessary to get the statik package, and run the command in the root if the commanderBot repository.

````
go get github.com/rakyll/statik
statik
```
