package game

import (
	"context"
	"strconv"
	"sync"

	"github.com/https-whoyan/MafiaCore/config"
	"github.com/https-whoyan/MafiaCore/game"
	"github.com/https-whoyan/MafiaCore/player"
	"github.com/https-whoyan/MafiaCore/roles"

	"github.com/https-whoyan/MafiaCore/internal/tests/models"
)

func initHelper(cfg *config.RolesConfig) (*game.Game, error) {
	var internalErr error

	opts := []game.GameOption{
		game.FMTerOpt(models.TestFMTInstance),
		game.RenamePrOpt(models.TestRenameUserProviderInstance),
	}
	g := game.GetNewGame(models.TestingGuildID, opts...)

	allRoleChannels := models.NewTestChannels()
	mainChannel := models.NewTestMainChannels()

	internalErr = g.SetMainChannel(mainChannel)
	if internalErr != nil {
		return nil, internalErr
	}
	for _, roleCh := range allRoleChannels {
		internalErr = g.SetNewRoleChannel(roleCh)
		if internalErr != nil {
			return nil, internalErr
		}
	}

	testPlayers := models.GetTestPlayers(cfg.PlayersCount)
	g.SetStartPlayers(testPlayers)
	internalErr = g.Init(cfg)
	if internalErr != nil {
		return nil, internalErr
	}
	return g, nil
}

func signalHandler(s game.Signal) *roles.Role {
	if sSS, ok := s.(game.SwitchStateSignal); ok {
		if v, ok := sSS.Value.(game.SwitchNightVoteRoleSwitchValue); ok {
			return v.CurrentVotedRole
		}
	}
	return nil
}

func playersHelper(players player.Players) map[*roles.Role][]*player.Player {
	mp := make(map[*roles.Role][]*player.Player)
	for _, p := range players {
		mp[p.Role] = append(mp[p.Role], p)
	}
	return mp
}

type voteCfg struct {
	role  *roles.Role
	votes []player.IDType
}

func takeANight(g *game.Game, c votesCfg) error {
	ch := make(<-chan game.Signal)
	wg := sync.WaitGroup{}
	wg.Add(2)
	var (
		err      error
		standErr = func(fnErr error) {
			if err == nil && fnErr != nil {
				err = fnErr
			}
		}
	)
	go func() {
		defer wg.Done()
		ch = g.Run(context.Background())
	}()
	go func() {
		defer wg.Done()
		active := g.GetActivePlayers()
		for s := range ch {
			if g.GetState() == game.DayState {
				return
			}
			votedRole := signalHandler(s)
			if votedRole == nil {
				continue
			}
			if votedRole.IsTwoVotes {
				vP := c[votedRole].toTwoVotePr(&active)

				standErr(g.SetNightTwoVote(vP))
				continue
			}
			vP := c[votedRole].toVotePr(&active)
			standErr(g.SetNightVote(vP))
		}
	}()
	wg.Wait()
	return err
}

func (v voteCfg) toTwoVotePr(players *player.Players) *game.NightTwoVotesProvider {
	votedPlayers := *(players.SearchAllPlayersWithRole(v.role))
	var votedPlayer = &player.Player{}
	for _, p := range votedPlayers {
		votedPlayer = p
	}
	return &game.NightTwoVotesProvider{
		VotedPlayerID:          strconv.Itoa(int(votedPlayer.ID)),
		Vote1:                  strconv.Itoa(int(v.votes[0])),
		Vote2:                  strconv.Itoa(int(v.votes[1])),
		IsServerUserIDByPlayer: false,
		IsServerUserIDByVote:   false,
	}
}

func (v voteCfg) toVotePr(players *player.Players) *game.OneVoteProvider {
	votedPlayers := *(players.SearchAllPlayersWithRole(v.role))
	var votedPlayer = &player.Player{}
	for _, p := range votedPlayers {
		votedPlayer = p
	}
	return &game.OneVoteProvider{
		VotedPlayerID:          strconv.Itoa(int(votedPlayer.ID)),
		Vote:                   strconv.Itoa(int(v.votes[0])),
		IsServerUserIDByPlayer: false,
		IsServerUserIDByVote:   false,
	}
}

type votesCfg map[*roles.Role]voteCfg
