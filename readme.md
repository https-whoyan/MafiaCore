# <h1 align="center"> MafiaCore</h1> [![Go Reference](https://pkg.go.dev/badge/github.com/https-whoyan/MafiaCore.svg)](https://pkg.go.dev/github.com/https-whoyan/MafiaCore) [![Go Report](https://goreportcard.com/badge/github.com/https-whoyan/MafiaCore)](https://goreportcard.com/report/github.com/https-whoyan/MafiaCore)
<hr>

**Open Source code to integrate the game “Mafia” into your application.**

[Install](#install) <br>
[Project Struct](#Architecture) <br>
[Game Rules](#Rules) <br>
[How it works?](#Usage) <br>

<hr>
## Install

```
go get -u github.com/https-whoyan/MafiaCore
```

<hr>

# Architecture
<pre>
<code style="display: block">
├── src/app
|     └── main.go
|            ├── Initialization of all packages with empty assignments
|            └── for no errors checking
|
├── channel
|     ├── Here is the interface channel on which the game will be played.
|     └── Also functions to add players, spectators, and remove users from the channel.
|
├── config
|     └── Here you will find all information regarding the role configurations of the game.
|
├── converter
|     ├── Useful functions for working with internal go types,
|     └── but which are absent in the standard go language package
|
├── fmt
|     └── FMTInterface. Look code. Used to formatting messages
|
├── game 
|     ├── game.go
|     |       ├── The structure of the game and its methods of
|     |       └── initialization, start, action, and ending.
|     ├── interaction.go
|     |       └── Logic on the interaction of roles on players or on the game.
|     ├── loaders.go
|     |       └── Methods of game struct to load channels and players
|     ├── logger.go
|     |       └── Interface to log all logs about games and logs definition
|     ├── message.go
|     |       ├── File used to send messages to channels (game channels) 
|     ├── reincarnation.go
|     |       └── Changing a player's role and verifying this in certain cases
|     ├── signal.go
|     |       └── An interface that informs your interpreter of new game states or runtime errors
|     ├── state.go
|     |       └── Micro state machine for game
|     ├── vote.go
|     |       └── A file containing all logic and vote processing.
|     ├── day.go
|     ├── night.go
|     └── timer.go
|
├── internal/tests
|
├── message
|     ├── Utils messages, not called in code 
|     └── but may be useful for your interpretation
|
├── player
|     ├── The structure of players, non-players, dead players and his collections.
|     └── Also, code for renaming users during and starting the game
|
├── roles
|     ├── All information about roles.
|     └── NOTE: Each role is a variable, not a separate struct. 
└── time
      └── consts.go
              └── Time constants for the game. 

</code>
</pre>

<hr>

## Rules

<h2 align="center"> This is not the classic mafia! </h2>

There are many roles presented in the game, you can find all of them along with a description in the roles folder.

**Please note that**
* Don may or may not know the mafia. It all depends on how you put the channel in the game.
* The detective does not check one player. Instead, he checks two players to see if they belong to the same team.
* The fool in the game plays for the peaceful. However, he wins by one vote when killed, and is considered the loser when the civilians are eliminated.
* Mistress blocks only night actions of a player, but in no way prevents him from voting in daytime voting.

<hr>

## Usage
### Game start
