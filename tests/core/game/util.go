package game

import (
	"strconv"
	"sync"

	"github.com/https-whoyan/MafiaBot/core/config"
	"github.com/https-whoyan/MafiaBot/core/game"
	"github.com/https-whoyan/MafiaBot/core/player"
	"github.com/https-whoyan/MafiaBot/core/roles"
	"github.com/https-whoyan/MafiaBot/tests/core/models"
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

func playersHelper(players []*player.Player) map[*roles.Role][]*player.Player {
	mp := make(map[*roles.Role][]*player.Player)
	for _, p := range players {
		mp[p.Role] = append(mp[p.Role], p)
	}
	return mp
}

type voteCfg struct {
	role  *roles.Role
	votes []int
}

func takeANight(g *game.Game, c votesCfg) {
	ch := make(chan game.Signal)
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		g.Night(ch)
		close(ch)
	}()
	go func() {
		defer wg.Done()
		select {
		case <-ch:
			return
		default:
			for s := range ch {
				votedRole := signalHandler(s)
				if votedRole == nil {
					continue
				}
				if votedRole.IsTwoVotes {
					vP := c[votedRole].toTwoVotePr(g.Active)
					g.TwoVoteChan <- vP
					continue
				}
				vP := c[votedRole].toVotePr(g.Active)
				g.VoteChan <- vP
			}
		}
	}()
	wg.Wait()
}

func (v voteCfg) toTwoVotePr(players []*player.Player) *game.TwoVotesProvider {
	return &game.TwoVotesProvider{
		VotedPlayerID:  strconv.Itoa(player.SearchAllPlayersWithRole(players, v.role)[0].ID),
		Vote1:          strconv.Itoa(v.votes[0]),
		Vote2:          strconv.Itoa(v.votes[1]),
		IsServerUserID: false,
	}
}

func (v voteCfg) toVotePr(players []*player.Player) *game.VoteProvider {
	return &game.VoteProvider{
		VotedPlayerID:  strconv.Itoa(player.SearchAllPlayersWithRole(players, v.role)[0].ID),
		Vote:           strconv.Itoa(v.votes[0]),
		IsServerUserID: false,
	}
}

type votesCfg map[*roles.Role]voteCfg