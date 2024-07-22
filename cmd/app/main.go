package main

import (
	"github.com/https-whoyan/MafiaCore/channel"
	"github.com/https-whoyan/MafiaCore/config"
	"github.com/https-whoyan/MafiaCore/converter"
	"github.com/https-whoyan/MafiaCore/fmt"
	"github.com/https-whoyan/MafiaCore/game"
	"github.com/https-whoyan/MafiaCore/message"
	"github.com/https-whoyan/MafiaCore/player"
	"github.com/https-whoyan/MafiaCore/roles"
	"github.com/https-whoyan/MafiaCore/time"

	"github.com/https-whoyan/MafiaCore/internal/tests/models"
)

func main() {
	// Import all packages to check for no errors.
	var (
		_ = &game.Game{}
		_ = &player.Player{}
		_ = channel.Channel(nil)
		_ = &config.Configs
		_ = fmt.FmtInterface(nil)
		_ = converter.GetMapKeys(map[int]int{})
		_ = message.GetStartPlayerDefinition(&player.Player{Role: roles.Mafia}, models.TestFMTInstance)
		_ = &roles.Role{}
		_ = time.FakeVotingMaxSeconds
	)
}
