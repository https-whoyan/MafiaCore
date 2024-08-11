package game

import (
	"github.com/LastPossum/kamino"
	channelPack "github.com/https-whoyan/MafiaCore/channel"
	configPack "github.com/https-whoyan/MafiaCore/config"
	playerPack "github.com/https-whoyan/MafiaCore/player"
	rolesPack "github.com/https-whoyan/MafiaCore/roles"
	"time"
)

type DeepCloneGame struct {
	GuildID      string                        `bson:"guild_id" json:"guild_id" yaml:"guild_id" db:"guild_id" xml:"guild_id"`
	PlayersCount int                           `bson:"players_count" json:"players_count" yaml:"players_count" db:"players_count" xml:"players_count"`
	RolesConfig  *configPack.RolesConfig       `bson:"roles_config" json:"roles_config" yaml:"roles_config" db:"roles_config" xml:"roles_config"`
	NightCounter int                           `bson:"night_counter" json:"night_count" yaml:"night_count" db:"night_count" xml:"night_count"`
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

	NightLogs []NightLog `bson:"night_logs" json:"night_logs" yaml:"night_logs" db:"night_logs" xml:"night_logs"`
	DayLogs   []DayLog   `bson:"day_logs" json:"day_logs" yaml:"day_logs" db:"day_logs" xml:"day_logs"`
}

func (g *Game) GetDeepClone() (DeepCloneGame, error) {
	deepCloneGame, err := kamino.Clone(g)
	if err != nil {
		return DeepCloneGame{}, err
	}
	return DeepCloneGame{
		GuildID:       deepCloneGame.guildID,
		PlayersCount:  deepCloneGame.playersCount,
		RolesConfig:   deepCloneGame.rolesConfig,
		NightCounter:  deepCloneGame.nightCounter,
		TimeStart:     deepCloneGame.timeStart,
		EndTime:       deepCloneGame.endTime,
		StartPlayers:  deepCloneGame.startPlayers,
		Active:        deepCloneGame.active,
		Dead:          deepCloneGame.dead,
		Spectators:    deepCloneGame.spectators,
		RoleChannels:  deepCloneGame.roleChannels,
		MainChannel:   deepCloneGame.mainChannel,
		NightVoting:   deepCloneGame.nightVoting,
		VotePing:      deepCloneGame.votePing,
		PreviousState: deepCloneGame.previousState,
		State:         deepCloneGame.state,
		NightLogs:     deepCloneGame.nightLogs,
		DayLogs:       deepCloneGame.dayLogs,
		RenameMode:    deepCloneGame.renameMode,
	}, nil
}
