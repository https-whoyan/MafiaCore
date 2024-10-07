package game

import (
	"time"

	channelPack "github.com/https-whoyan/MafiaCore/channel"
	configPack "github.com/https-whoyan/MafiaCore/config"
	playerPack "github.com/https-whoyan/MafiaCore/player"
	rolesPack "github.com/https-whoyan/MafiaCore/roles"
)

func (g *Game) GuildID() string {
	return g.guildID
}
func (g *Game) GetState() State {
	return g.state
}

func (g *Game) GetActivePlayers() playerPack.Players {
	return *g.active
}
func (g *Game) GetStartPlayers() playerPack.NonPlayingPlayers {
	return *g.startPlayers
}
func (g *Game) GetSpectatorsOfGame() playerPack.NonPlayingPlayers {
	return *g.spectators
}

func (g *Game) GetNightVoting() *rolesPack.Role {
	return g.nightVoting
}
func (g *Game) GetNightsCount() int {
	return g.nightCounter
}
func (g *Game) PlayersCount() int {
	return g.playersCount
}
func (g *Game) GetConfig() *configPack.RolesConfig {
	return g.rolesConfig
}
func (g *Game) GetStartTime() time.Time {
	return g.timeStart
}
func (g *Game) GetEndTime() time.Time {
	return g.endTime
}

func (g *Game) GetVotePing() int {
	return g.votePing
}
func (g *Game) GameMessenger() Messenger {
	return *g.messenger
}

func (g *Game) GetRoleChannels() map[*rolesPack.Role]channelPack.RoleChannel {
	return g.roleChannels
}
func (g *Game) GetMainChannel() channelPack.MainChannel {
	return g.mainChannel
}

func (g *Game) GetGameLogger() Logger { return g.gameLogger }

func (g *Game) GetErrorChan() <-chan ErrSignal { return g.errChanDest }
func (g *Game) GetInfoChan() <-chan InfoSignal { return g.infoChanDest }
