package game

import (
	"context"
	"github.com/https-whoyan/MafiaCore/player"
	"github.com/https-whoyan/MafiaCore/roles"
)

// Logger allows you to save game information.
//
// The implementation is thrown when the game is initialized Init,
// logs are automatically loaded and saved to the implementation for
// saving in the run and finish methods.
type Logger interface {
	InitNewGame(ctx context.Context, g DeepCloneGame) error
	SaveNightLog(ctx context.Context, g DeepCloneGame, log NightLog) error
	SaveDayLog(ctx context.Context, g DeepCloneGame, log DayLog) error
	SaveFinishLog(ctx context.Context, g DeepCloneGame, log FinishLog) error
	NameAGame(ctx context.Context, g DeepCloneGame, gameName string) error
}

// ____________
// NightLog
// ____________

// NightLog saves all votes, as well as the IDs of those
// players who turned out to be dead based on the results of the night.
type NightLog struct {
	NightNumber int `json:"number"`
	// Key - ID of the voted player
	// Value - usually a vote, but in case the role uses 2 votes - 2 votes at once.
	NightVotes map[player.IDType][]player.IDType `json:"votes"`
	Dead       []player.IDType                   `json:"dead"`
}

// NewNightLog Gives the log after nightfall.
// Panics if not called after night or during voting.
func (g *Game) NewNightLog() NightLog {
	if g.ctx == nil {
		panic("Game is not initialized")
	}
	select {
	case <-g.ctx.Done():
		return NightLog{}
	default:
		if g.state != NightState {
			panic("Inappropriate use not after overnight")
		}
		if g.nightVoting != nil {
			panic("the function is called during the night, not after it!")
		}

		g.RLock()
		defer g.RUnlock()

		nightNumber := g.nightCounter
		nightVotes := make(map[player.IDType][]player.IDType)
		for _, p := range *g.active {
			if p.Role.NightVoteOrder == -1 {
				continue
			}

			var votes []player.IDType
			n := len(p.Votes)
			if p.Role.IsTwoVotes {
				votes = []player.IDType{p.Votes[n-2], p.Votes[n-1]}
			} else {
				votes = []player.IDType{p.Votes[n-1]}
			}
			nightVotes[p.ID] = votes
		}
		var dead []player.IDType
		for _, p := range *g.active {
			if p.LifeStatus == player.Dead {
				dead = append(dead, p.ID)
			}
		}
		return NightLog{
			NightNumber: nightNumber,
			NightVotes:  nightVotes,
			Dead:        dead,
		}
	}
}

// _______________
// DayLog
// _______________

type DayLog struct {
	DayNumber int `json:"number"`
	// Key - ID of the player who was voted for during the day to be excluded from the game
	// Value - number of votes.
	DayVotes map[player.IDType]player.IDType `json:"votes"`
	Kicked   *player.IDType                  `json:"kicked"`
	IsSkip   bool                            `json:"isSkip"`
}

type FinishLog struct {
	WinnerTeam  *roles.Team `json:"winnerTeam"`
	IsFool      bool        `json:"isFool"`
	TotalNights int         `json:"totalNights"`
}

func (g *Game) NewFinishLog(winnerTeam *roles.Team, isFool bool) FinishLog {
	if isFool {
		// Validation

		g.RLock()
		defer g.RUnlock()

		containsFool := false
		for _, players := range *g.dead {
			for _, p := range players {
				if p.Role == roles.Fool {
					containsFool = true
				}
			}
		}

		if !containsFool {
			panic("Fool is not killed!")
		}

		return FinishLog{
			WinnerTeam:  nil,
			IsFool:      isFool,
			TotalNights: g.nightCounter,
		}
	}

	trueWinnerTeam := g.UnderstandWinnerTeam()
	if *trueWinnerTeam != *winnerTeam {
		panic("WinnerTeam is not determined! The game can still turn around!!")
	}

	return FinishLog{
		WinnerTeam:  winnerTeam,
		IsFool:      false,
		TotalNights: g.nightCounter,
	}
}
