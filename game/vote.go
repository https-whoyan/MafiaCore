package game

import (
	"errors"
	"strconv"

	"github.com/https-whoyan/MafiaCore/channel"
	"github.com/https-whoyan/MafiaCore/converter"
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

// NightVoteProvider default implementation of NightVoteProviderInterface
type NightVoteProvider struct {
	VotedPlayerID                                string
	Vote                                         string
	IsServerUserIDByPlayer, IsServerUserIDByVote bool
}

func NewVoteProvider(votedPlayerID string, vote string, isServerUserIDByPlayer, isServerUserIDByVote bool) NightVoteProviderInterface {
	return &NightVoteProvider{
		VotedPlayerID:          votedPlayerID,
		Vote:                   vote,
		IsServerUserIDByPlayer: isServerUserIDByPlayer,
		IsServerUserIDByVote:   isServerUserIDByVote,
	}
}

func (p *NightVoteProvider) GetVotedPlayerID() (votedUserID string, isUserServerID bool) {
	return p.VotedPlayerID, p.IsServerUserIDByPlayer
}
func (p *NightVoteProvider) GetVote() (ID string, isUserServerID bool) {
	return p.Vote, p.IsServerUserIDByVote
}

// DayVoteProviderInterface Interface for day voice reception
//
// Same as oneVoteProviderInterface, but using only at day.
type DayVoteProviderInterface interface {
	oneVoteProviderInterface
}

// TwoVoteProviderInterface A special channel used only  for roles that specify 2 voices rather
// than one (example: detective)
//
// Its peculiarity is that instead of one voice it uses
// 2 voices - IDs of 2 players it wants to check, so I decided to make a separate interface for it
type TwoVoteProviderInterface interface {
	GetVotedPlayerID() (votedUserID string, isUserServerID bool)
	GetVote() (ID1 string, ID2 string, isServerUserID bool)
}

// TwoVotesProvider default implementation of TwoVoteProviderInterface
type TwoVotesProvider struct {
	VotedPlayerID                                string
	Vote1, Vote2                                 string
	IsServerUserIDByPlayer, IsServerUserIDByVote bool
}

func NewTwoVoteProvider(votedPlayerID string, vote1, vote2 string,
	isServerUserIDByPlayer, isServerUserIDByVote bool) TwoVoteProviderInterface {
	return &TwoVotesProvider{
		VotedPlayerID:          votedPlayerID,
		Vote1:                  vote1,
		Vote2:                  vote2,
		IsServerUserIDByPlayer: isServerUserIDByPlayer,
		IsServerUserIDByVote:   isServerUserIDByVote,
	}
}

func (p *TwoVotesProvider) GetVotedPlayerID() (votedUserID string, isUserServerID bool) {
	return p.VotedPlayerID, p.IsServerUserIDByPlayer
}
func (p *TwoVotesProvider) GetVote() (ID1, ID2 string, isServerUserID bool) {
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

	IncorrectVoteChannel = errors.New("incorrect Vote channel")
	IncorrectVotedPlayer = errors.New("incorrect voted player")

	VotePlayerNotFound      = errors.New("vote player not found")
	PlayerIsMutedErr        = errors.New("player is muted")
	VotePlayerIsNotAlive    = errors.New("vote player is not alive")
	VotePingError           = errors.New("player get same Vote before")
	IncorrectVoteTime       = errors.New("incorrect Vote time")
	ToVotedPlayerIsNotAlive = errors.New("toVoted player is not alive")
	CannotVoteToYourself    = errors.New("cannot vote to yourself")
	OneVoteRequiredErr      = errors.New("one Vote required")

	TwoVoteRequiredErr      = errors.New("two Vote required")
	TwoVotesOneOfEmptyErr   = errors.New("both votes must be either blank or not blank")
	TwoVotesSimilarVotesErr = errors.New("votes are similar")
)

// _________________________
// Validators
// _________________________

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
		return VotePlayerIsNotAlive
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
	votedPlayer, isServerID := vP.GetVotedPlayerID()
	voter := g.active.SearchPlayerByID(votedPlayer, isServerID)
	vote, isServerID := vP.GetVote()
	if vote == EmptyVoteStr {
		return nil
	}
	toVoted := g.active.SearchPlayerByID(vote, isServerID)
	if toVoted == nil {
		return IncorrectVoteType
	}
	if toVoted.LifeStatus != player.Alive {
		return ToVotedPlayerIsNotAlive
	}
	if !g.voteForYourself && voter == toVoted {
		return CannotVoteToYourself
	}
	return nil
}

func (g *Game) twoVoteValidator(gV interface{ GetVotes() (string, string, bool) }) error {
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
		return ToVotedPlayerIsNotAlive
	}
	if toVoted2.LifeStatus != player.Alive {
		return ToVotedPlayerIsNotAlive
	}
	return nil
}

// nightVoteValidator also check roleChannel.Role and vP.VotedPlayer role.
// Use nil if you don't need for this checking
func (g *Game) nightVoteValidator(vP NightVoteProviderInterface, roleChannel channel.RoleChannel) error {
	if err := g.oneVoteValidator(vP); err != nil {
		return err
	}

	if g.state != NightState {
		return IncorrectVoteTime
	}

	votedPlayerID, isServerIDByVoter := vP.GetVotedPlayerID()
	votedPlayer := g.active.SearchPlayerByID(votedPlayerID, isServerIDByVoter)
	if g.nightVoting != votedPlayer.Role {
		return IncorrectVotedPlayer
	}
	if roleChannel != nil && g.nightVoting != roleChannel.GetRole() {
		return IncorrectVoteChannel
	}
	vote, isServerIDByVote := vP.GetVote()
	toVotePlayer := g.active.SearchPlayerByID(vote, isServerIDByVote)
	toVotePlayerID := toVotePlayer.ID
	previousVotesMp := make(map[int]bool)
	startIndex := max(0, len(votedPlayer.Votes)-g.votePing)
	for i := startIndex; i <= len(votedPlayer.Votes)-1; i++ {
		previousVotesMp[votedPlayer.Votes[i]] = true
	}
	// ValidatedBefore
	if previousVotesMp[toVotePlayerID] {
		return VotePingError
	}
	return nil
}

// dayVoteValidatorByChannelIID performs the same validation as dayVoteValidator
func (g *Game) dayVoteValidatorByChannelIID(vP NightVoteProviderInterface, channelIID string) error {
	var allChannels []channel.Channel
	allRoleChannels := converter.GetMapValues(g.roleChannels)
	allChannels = append(allChannels, channel.RoleSliceToChannelSlice(allRoleChannels)...)

	allChannels = append(allChannels, g.mainChannel)

	channelVotedFrom := channel.SearchChannelByGameID(allChannels, channelIID)
	if channelVotedFrom == nil {
		return IncorrectVoteChannel
	}
	return g.dayVoteValidator(vP)
}

func (g *Game) dayVoteValidator(vP NightVoteProviderInterface) error {
	if g.state != DayState {
		return IncorrectVoteTime
	}
	return g.nightVoteProviderValidator(vP)
}

func (g *Game) twoVoteProviderValidator(vP TwoVoteProviderInterface) error {
	if vP == nil {
		return NilValidatorErr
	}
	votedPlayerID, isServerID := vP.GetVotedPlayerID()
	votedPlayer := g.active.SearchPlayerByID(votedPlayerID, isServerID)
	if votedPlayer == nil {
		return InVotePlayerNotFound
	}
	if votedPlayer.InteractionStatus == player.Muted {
		return PlayerIsMutedErr
	}
	if g.nightVoting != votedPlayer.Role {
		return IncorrectVotedPlayer
	}
	if !votedPlayer.Role.IsTwoVotes {
		return OneVoteRequiredErr
	}
	vote1, vote2 := vP.GetVote()
	if vote1 == EmptyVoteStr && vote2 == EmptyVoteStr {
		return nil
	}
	if vote1 == EmptyVoteStr || vote2 == EmptyVoteStr {
		return IncorrectVoteType
	}
	if !isServerID {
		_, err1 := strconv.Atoi(vote1)
		_, err2 := strconv.Atoi(vote2)
		if err1 != nil || err2 != nil {
			return IncorrectVoteType
		}
	}
	if votedPlayer.LifeStatus != player.Alive {
		return VotePlayerIsNotAlive
	}
	toVotePlayer1 := g.active.SearchPlayerByID(votedPlayerID, isServerID)
	toVotePlayer2 := g.active.SearchPlayerByID(votedPlayerID, isServerID)
	if toVotePlayer1 == nil || toVotePlayer2 == nil {
		return VotePlayerNotFound
	}
	return nil
}

// nightTwoVoteValidatorByChannelIID performs the same validation as nightVoteValidator.
//
// Use it, if you want, that day Vote should be in a particular channel.
func (g *Game) nightTwoVoteValidatorByChannelIID(vP TwoVoteProviderInterface, channelIID string) error {
	sliceChannels := converter.GetMapValues(g.roleChannels)
	foundedChannel := channel.SearchRoleChannelByID(sliceChannels, channelIID)
	if foundedChannel == nil {
		return IncorrectVoteChannel
	}
	return g.nightTwoVoteValidator(vP, foundedChannel)
}

// nightTwoVoteValidator also check roleChannel.Role and vP.VotedPlayer role.
// Use nil if you don't need for this checking
func (g *Game) nightTwoVoteValidator(vP TwoVoteProviderInterface, roleChannel channel.RoleChannel) error {
	if err := g.twoVoteProviderValidator(vP); err != nil {
		return err
	}

	if roleChannel == nil || g.nightVoting != roleChannel.GetRole() {
		return IncorrectVoteChannel
	}
	return nil
}

// _______________________________
// Vote Functions
// _______________________________

// OptionalChannelIID Optional field Mechanism.
type OptionalChannelIID struct{ channelIID string }

func NewOptionalChannelIID(channelIID string) *OptionalChannelIID {
	return &OptionalChannelIID{channelIID}
}

// nightOneVote used after validation, and stand votes
func (g *Game) nightOneVote(vP NightVoteProviderInterface) {
	votedPlayerID, isServerID := vP.GetVotedPlayerID()
	g.RLock()
	votedPlayer := g.active.SearchPlayerByID(votedPlayerID, isServerID)
	g.RUnlock()
	vote := vP.GetVote()
	g.Lock()
	if vote == EmptyVoteStr {
		votedPlayer.Votes = append(votedPlayer.Votes, EmptyVoteInt)
	} else {
		// validated Before
		intVote, _ := strconv.Atoi(vote)
		votedPlayer.Votes = append(votedPlayer.Votes, intVote)
	}
	g.Unlock()
}

// nightTwoVote opt is OptionalChannelIID optional field Mechanism.
//
// If you not need it, pass nil to the field.
// If yes, use NewOptionalChannelIID
//
// Immediately puts all the right votes and changes the value of the fields if no error occurred.
func (g *Game) nightTwoVote(vP TwoVoteProviderInterface, opt *OptionalChannelIID) error {
	var err error
	if opt == nil {
		err = g.nightTwoVoteValidator(vP, nil)
	} else {
		err = g.nightTwoVoteValidatorByChannelIID(vP, opt.channelIID)
	}
	if err != nil {
		return err
	}

	votedPlayerID, isServerID := vP.GetVotedPlayerID()
	g.RLock()
	votedPlayer := g.active.SearchPlayerByID(votedPlayerID, isServerID)
	g.RUnlock()
	vote1, vote2 := vP.GetVote()
	if vote1 != vote2 && (vote2 == EmptyVoteStr || vote1 == EmptyVoteStr) {
		return TwoVotesOneOfEmptyErr
	}
	g.Lock()
	if vote1 == EmptyVoteStr {
		votedPlayer.Votes = append(votedPlayer.Votes, EmptyVoteInt, EmptyVoteInt)
	} else {
		// validated Before
		intVote1, _ := strconv.Atoi(vote1)
		intVote2, _ := strconv.Atoi(vote2)
		votedPlayer.Votes = append(votedPlayer.Votes, intVote1, intVote2)
	}
	g.Unlock()
	return nil
}

// dayVote opt is OptionalChannelIID optional field Mechanism.
//
// If you not need it, pass nil to the field.
// If yes, use NewOptionalChannelIID
//
// Immediately puts all the right votes and changes the value of the fields if no error occurred.
func (g *Game) dayVote(vP NightVoteProviderInterface, opt *OptionalChannelIID) error {
	var err error
	if opt == nil {
		err = g.dayVoteValidator(vP)
	} else {
		err = g.dayVoteValidatorByChannelIID(vP, opt.channelIID)
	}
	if err != nil {
		return err
	}

	votedPlayerID, isServerID := vP.GetVotedPlayerID()
	g.RLock()
	votedPlayer := g.active.SearchPlayerByID(votedPlayerID, isServerID)
	g.RUnlock()
	vote := vP.GetVote()
	if vote == EmptyVoteStr {
		votedPlayer.DayVote = EmptyVoteInt
	}
	// validated Before
	votedPlayer.DayVote, _ = strconv.Atoi(vote)
	return nil
}

// Functions of game to set voting.
// You can use only this functions

func (g *Game) SetNightVote(nightVote NightVoteProviderInterface, ch *OptionalChannelIID) error {
	if err := g.basicVoteValidator(nightVote); err != nil {
		return err
	}
	if g.state != NightState {
		return IncorrectVoteTime
	}
	var err error
	if ch != nil {
		err = g.nightVoteValidatorByChannelIID(nightVote, ch.channelIID)
	} else {
		err = g.nightVoteValidator(nightVote, nil)
	}
	if err != nil {
		return err
	}
	g.nightOneVote(nightVote)
	return nil
}
