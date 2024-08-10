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

func (g *Game) GetNightCount() int {
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
func (g *Game) GameMessenger() *Messenger {
	return g.messenger
}

type DeepCloneGame struct {
	GuildID      string                        `bson:"guild_id" json:"guild_id" yaml:"guild_id" db:"guild_id" xml:"guild_id"`
	PlayersCount int                           `bson:"players_count" json:"players_count" yaml:"players_count" db:"players_count" xml:"players_count"`
	RolesConfig  *configPack.RolesConfig       `bson:"roles_config" json:"roles_config" yaml:"roles_config" db:"roles_config" xml:"roles_config"`
	nightCounter int                           `bson:"night_counter" json:"night_count" yaml:"night_count" db:"night_count" xml:"night_count"`
	TimeStart    time.Time                     `bson:"time_start" json:"time_start" yaml:"time_start" db:"time_start" xml:"time_start"`
	EndTime      time.Time                     `bson:"end_time" json:"end_time" yaml:"end_time" db:"end_time" xml:"end_time"`
	StartPlayers *playerPack.NonPlayingPlayers `bson:"start_players" json:"start_players" yaml:"start_players" db:"start_players" xml:"start_players"`
	Active       *playerPack.Players           `bson:"active" json:"active" yaml:"active" db:"active" xml:"active"`
	Dead         *playerPack.DeadPlayers       `bson:"dead" json:"dead" yaml:"dead" db:"dead" xml:"dead"`
	Spectators   *playerPack.NonPlayingPlayers `bson:"spectators" json:"spectators" yaml:"spectators" db:"spectators" xml:"spectators"`

	RoleChannels map[*rolesPack.Role]channelPack.RoleChannel `bson:"role_channels" json:"role_channels" yaml:"role_channels" db:"role_channels" xml:"role_channels"`

	MainChannel     channelPack.MainChannel `bson:"main_channel" json:"main_channel" yaml:"main_channel" db:"main_channel" xml:"main_channel"`
	NightVoting     *rolesPack.Role         `bson:"night_voting" json:"night_voting" db:"night_voting" xml:"night_voting"`
	VoteForYourself bool                    `bson:"vote_for_yourself" bson:"vote_for_yourself"`
	VotePing        int                     `bson:"vote_ping" json:"vote_ping" yaml:"vote_ping" db:"vote_ping" xml:"vote_ping"`
	PreviousState   State                   `bson:"previous_state" json:"previous_state" yaml:"previous_state" db:"previous_state" xml:"previous_state"`
	State           State                   `bson:"state" json:"state" yaml:"state" db:"state" xml:"state"`
	RenameMode      RenameMode              `bson:"rename_mode" json:"rename_mode" db:"rename_mode" xml:"rename_mode"`
	Logger          Logger
}

func (g *Game) GetDeepClone() DeepCloneGame {
	return DeepCloneGame{
		GuildID:       g.guildID,
		PlayersCount:  g.playersCount,
		RolesConfig:   g.rolesConfig,
		nightCounter:  g.nightCounter,
		TimeStart:     g.timeStart,
		EndTime:       g.endTime,
		StartPlayers:  g.startPlayers,
		Active:        g.active,
		Dead:          g.dead,
		Spectators:    g.spectators,
		RoleChannels:  make(map[*rolesPack.Role]channelPack.RoleChannel),
		MainChannel:   g.mainChannel,
		NightVoting:   g.nightVoting,
		VotePing:      g.votePing,
		PreviousState: g.previousState,
		State:         g.state,
		RenameMode:    g.renameMode,
	}
}
