package game

import (
	"errors"
	"github.com/https-whoyan/MafiaCore/player"
)

// This file contains everything about the voting mechanics.

// ___________________________________
// NightVoteProviderInterface
// ___________________________________

const (
	EmptyVoteStr = "-1"
	EmptyVoteInt = -1
)

// oneVoteProviderInterface Interface for voice reception
//
// Interface implementations allows you to implement the Vote command in your interpretation.
type oneVoteProviderInterface interface {
	// GetVotedPlayerID Provides 2 fields: information about the voting player.
	//
	// The 1st field provides the ID of the player who voted,
	// the second field is whether this ID is your server ID or in-game IDType.
	GetVotedPlayerID() (votedUserID string, isUserServerID bool)
	// GetVote provide one field: ID of the player being voted for.
	// If you need empty Vote, use the EmptyVoteStr constant.
	GetVote() (ID string, isServerUserID bool)
}

// NightVoteProviderInterface Interface for night voice reception
//
// Same as oneVoteProviderInterface, but using only at night.
type NightVoteProviderInterface interface {
	oneVoteProviderInterface
}

// OneVoteProvider default implementation of NightVoteProviderInterface
type OneVoteProvider struct {
	VotedPlayerID                                string
	Vote                                         string
	IsServerUserIDByPlayer, IsServerUserIDByVote bool
}

func NewVoteProvider(votedPlayerID string, vote string,
	isServerUserIDByPlayer, isServerUserIDByVote bool) NightVoteProviderInterface {
	return &OneVoteProvider{
		VotedPlayerID:          votedPlayerID,
		Vote:                   vote,
		IsServerUserIDByPlayer: isServerUserIDByPlayer,
		IsServerUserIDByVote:   isServerUserIDByVote,
	}
}

func (p *OneVoteProvider) GetVotedPlayerID() (votedUserID string, isUserServerID bool) {
	return p.VotedPlayerID, p.IsServerUserIDByPlayer
}
func (p *OneVoteProvider) GetVote() (ID string, isUserServerID bool) {
	return p.Vote, p.IsServerUserIDByVote
}

// DayVoteProviderInterface Interface for day voice reception
//
// Same as oneVoteProviderInterface, but using only at day.
type DayVoteProviderInterface interface {
	oneVoteProviderInterface
}

// NightTwoVoteProviderInterface A special channel used only  for roles that specify 2 voices rather
// than one (example: detective)
//
// Its peculiarity is that instead of one voice it uses
// 2 voices - IDs of 2 players it wants to check, so I decided to make a separate interface for it
type NightTwoVoteProviderInterface interface {
	GetVotedPlayerID() (votedUserID string, isUserServerID bool)
	GetVotes() (ID1 string, ID2 string, isServerUserID bool)
}

// NightTwoVotesProvider default implementation of NightTwoVoteProviderInterface
type NightTwoVotesProvider struct {
	VotedPlayerID                                string
	Vote1, Vote2                                 string
	IsServerUserIDByPlayer, IsServerUserIDByVote bool
}

func NewTwoVoteProvider(votedPlayerID string, vote1, vote2 string,
	isServerUserIDByPlayer, isServerUserIDByVote bool) NightTwoVoteProviderInterface {
	return &NightTwoVotesProvider{
		VotedPlayerID:          votedPlayerID,
		Vote1:                  vote1,
		Vote2:                  vote2,
		IsServerUserIDByPlayer: isServerUserIDByPlayer,
		IsServerUserIDByVote:   isServerUserIDByVote,
	}
}

func (p *NightTwoVotesProvider) GetVotedPlayerID() (votedUserID string, isUserServerID bool) {
	return p.VotedPlayerID, p.IsServerUserIDByPlayer
}
func (p *NightTwoVotesProvider) GetVotes() (ID1, ID2 string, isServerUserID bool) {
	return p.Vote1, p.Vote2, p.IsServerUserIDByVote
}

// _______________________________
// Vote Validators
// _______________________________

var (
	IsNotGameStartedErr = errors.New("game is not started")
	NilValidatorErr     = errors.New("nil Validator")

	InVotePlayerNotFound = errors.New("voted player not found")
	IncorrectVoteType    = errors.New("incorrect Vote type")

	VotePlayerNotFoundErr      = errors.New("vote player not found")
	PlayerIsMutedErr           = errors.New("player is muted")
	VotePlayerIsNotAliveErr    = errors.New("vote player is not alive")
	VotePingErr                = errors.New("player get same Vote before")
	IncorrectVoteTimeErr       = errors.New("incorrect Vote time")
	ToVotedPlayerIsNotAliveErr = errors.New("toVoted player is not alive")
	CannotVoteToYourselfErr    = errors.New("cannot vote to yourself")

	TwoVotesOneOfEmptyErr   = errors.New("both votes must be either blank or not blank")
	TwoVotesSimilarVotesErr = errors.New("votes are similar")
)

// Helpers

func (g *Game) oneVoteHelper(vP oneVoteProviderInterface) (
	votedPlayer, toVotedPlayer *player.Player, isEmptyVote bool) {
	voterID, isServerVoterID := vP.GetVotedPlayerID()
	vote, isServerVote := vP.GetVote()
	if vote == EmptyVoteStr {
		isEmptyVote = true
	}
	g.RLock()
	defer g.RUnlock()
	votedPlayer = g.active.SearchPlayerByID(voterID, isServerVoterID)
	toVotedPlayer = g.active.SearchPlayerByID(vote, isServerVote)
	return
}

// _________________________
// Validators
// _________________________

// Common

func (g *Game) basicVoteValidator(vP interface{ GetVotedPlayerID() (string, bool) }) error {
	if g.ctx == nil {
		return IsNotGameStartedErr
	}
	if vP == nil {
		return NilValidatorErr
	}
	votedPlayerID, isServerID := vP.GetVotedPlayerID()
	votedPlayer := g.active.SearchPlayerByID(votedPlayerID, isServerID)
	if votedPlayer == nil {
		return InVotePlayerNotFound
	}
	if votedPlayer.LifeStatus != player.Alive {
		return VotePlayerIsNotAliveErr
	}
	if votedPlayer.InteractionStatus == player.Muted {
		return PlayerIsMutedErr
	}
	return nil
}

func (g *Game) oneVoteValidator(vP oneVoteProviderInterface) error {
	if err := g.basicVoteValidator(vP); err != nil {
		return err
	}
	vote, isServerID := vP.GetVote()
	if vote == EmptyVoteStr {
		return nil
	}
	toVoted := g.active.SearchPlayerByID(vote, isServerID)
	if toVoted == nil {
		return IncorrectVoteType
	}
	if toVoted.LifeStatus != player.Alive {
		return ToVotedPlayerIsNotAliveErr
	}
	return nil
}

func (g *Game) twoVoteValidatorOnlyVotes(gV interface{ GetVotes() (string, string, bool) }) error {
	vote1, vote2, isServerID := gV.GetVotes()
	if vote1 == EmptyVoteStr && vote2 == EmptyVoteStr {
		return nil
	}
	if vote1 == EmptyVoteStr || vote2 == EmptyVoteStr {
		return TwoVotesOneOfEmptyErr
	}
	if vote1 == vote2 {
		return TwoVotesSimilarVotesErr
	}
	toVoted1 := g.active.SearchPlayerByID(vote1, isServerID)
	toVoted2 := g.active.SearchPlayerByID(vote2, isServerID)
	if toVoted1 == nil || toVoted2 == nil {
		return IncorrectVoteType
	}
	if toVoted1.LifeStatus != player.Alive {
		return ToVotedPlayerIsNotAliveErr
	}
	if toVoted2.LifeStatus != player.Alive {
		return ToVotedPlayerIsNotAliveErr
	}
	return nil
}

// By Provider Type

func (g *Game) nightVoteValidator(vP NightVoteProviderInterface) error {
	err := g.oneVoteValidator(vP)
	if err != nil {
		return err
	}
	if g.state != NightState {
		return IncorrectVoteTimeErr
	}

	voter, toVoter, isEmpty := g.oneVoteHelper(vP)
	if isEmpty {
		return nil
	}
	if !g.voteForYourself && voter == toVoter {
		return CannotVoteToYourselfErr
	}
	if g.nightVoting != voter.Role {
		return IncorrectVoteTimeErr
	}
	// Check vote ping
	previousVotesMp := make(map[player.IDType]bool)
	startIndex := max(0, len(voter.Votes)-g.votePing)
	for i := startIndex; i <= len(voter.Votes)-1; i++ {
		previousVotesMp[voter.Votes[i]] = true
	}
	// ValidatedBefore
	if previousVotesMp[toVoter.ID] {
		return VotePingErr
	}
	return nil
}

func (g *Game) nightTwoVoteProviderValidator(vP NightTwoVoteProviderInterface) error {
	if basicErr := g.basicVoteValidator(vP); basicErr != nil {
		return basicErr
	}
	err := g.twoVoteValidatorOnlyVotes(vP)
	if err != nil {
		return err
	}
	if g.state != NightState {
		return IncorrectVoteTimeErr
	}

	votedPlayerID, isServerIDByPlayer := vP.GetVotedPlayerID()
	vote1, vote2, isServerIDByVote := vP.GetVotes()

	g.RLock()
	defer g.RUnlock()
	votedPlayer := g.active.SearchPlayerByID(votedPlayerID, isServerIDByPlayer)
	toVotePlayer1 := g.active.SearchPlayerByID(vote1, isServerIDByVote)
	toVotePlayer2 := g.active.SearchPlayerByID(vote2, isServerIDByVote)
	if g.nightVoting != votedPlayer.Role {
		return IncorrectVoteTimeErr
	}
	if vote1 == EmptyVoteStr && vote2 == EmptyVoteStr {
		return nil
	}
	if toVotePlayer1 == nil || toVotePlayer2 == nil {
		return VotePlayerNotFoundErr
	}
	if !g.voteForYourself && (votedPlayer == toVotePlayer1 || votedPlayer == toVotePlayer2) {
		return CannotVoteToYourselfErr
	}
	return nil
}

func (g *Game) dayVoteValidator(vP DayVoteProviderInterface) error {
	err := g.oneVoteValidator(vP)
	if err != nil {
		return err
	}
	if g.state != DayState {
		return IncorrectVoteTimeErr
	}
	voter, toVoter, isEmpty := g.oneVoteHelper(vP)
	if isEmpty {
		return nil
	}
	if !g.voteForYourself && voter == toVoter {
		return CannotVoteToYourselfErr
	}
	return nil
}

// _______________________________
// Vote Functions
// _______________________________

// nightOneVote used after validation, and stand votes
func (g *Game) nightOneVote(vP NightVoteProviderInterface) {
	voter, toVote, isEmpty := g.oneVoteHelper(vP)
	g.Lock()
	defer g.Unlock()
	if isEmpty {
		voter.Votes = append(voter.Votes, EmptyVoteInt)
		return
	}
	voter.Votes = append(voter.Votes, toVote.ID)
}

// nightTwoVote used after validation, and stand votes
func (g *Game) nightTwoVote(vP NightTwoVoteProviderInterface) {
	voterID, isServerVoterID := vP.GetVotedPlayerID()
	vote1, vote2, isServerVoteID := vP.GetVotes()

	g.RLock()
	voter := g.active.SearchPlayerByID(voterID, isServerVoterID)
	g.RUnlock()
	if vote1 == EmptyVoteStr && vote2 == EmptyVoteStr {
		g.Lock()
		voter.Votes = append(voter.Votes, EmptyVoteInt, EmptyVoteInt)
		g.Unlock()
		return
	}
	g.RLock()
	voter1ID := g.active.SearchPlayerByID(vote1, isServerVoteID).ID
	voter2ID := g.active.SearchPlayerByID(vote2, isServerVoteID).ID
	g.Lock()
	defer g.Unlock()
	voter.Votes = append(voter.Votes, voter1ID, voter2ID)
}

// dayVote used after validation, and stand votes
func (g *Game) dayVote(vP NightVoteProviderInterface) {
	voter, toVote, isEmpty := g.oneVoteHelper(vP)
	g.Lock()
	defer g.Unlock()
	if isEmpty {
		voter.DayVote = EmptyVoteInt
		return
	}
	voter.DayVote = toVote.ID
}

// __________________________________________

// Functions of game to set voting.
// You can use only this functions to interact for vote system

// SetNightVote Checks the voice for errors, and if it's ok, puts it in right away.
func (g *Game) SetNightVote(nightVote NightVoteProviderInterface) error {
	var err error
	err = g.nightVoteValidator(nightVote)
	if err != nil {
		return err
	}
	g.nightOneVote(nightVote)
	g.voteAccepted <- struct{}{}
	return nil
}

// SetNightTwoVote Checks the voice for errors, and if it's ok, puts it in right away.
func (g *Game) SetNightTwoVote(nightVote NightTwoVoteProviderInterface) error {
	var err error
	err = g.nightTwoVoteProviderValidator(nightVote)
	if err != nil {
		return err
	}
	g.nightTwoVote(nightVote)
	g.voteAccepted <- struct{}{}
	return nil
}

func (g *Game) SetDayVote(dayVote DayVoteProviderInterface) error {
	var err error
	err = g.dayVoteValidator(dayVote)
	if err != nil {
		return err
	}
	g.dayVote(dayVote)
	return nil
}
