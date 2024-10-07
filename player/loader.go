package player

import (
	"errors"
	"github.com/https-whoyan/MafiaCore/config"
)

// ___________________________________
// Use to start a game starting.
// Role reversal, to put it simply.
// ___________________________________

func generateListToN(n int) []int {
	var IDs []int
	for i := 1; i <= n; i++ {
		IDs = append(IDs, i)
	}

	return IDs
}

func GeneratePlayers(tags []string, oldUsernames []string,
	serverUsernames []string, cfg *config.RolesConfig) (Players, error) {
	if len(tags) != cfg.PlayersCount {
		return nil, errors.New("unexpected mismatch of playing participants and configs")
	}
	if len(tags) != len(oldUsernames) {
		return nil, errors.New("unexpected mismatch of playing participants and nicknames")
	}

	n := len(tags)
	IDs := generateListToN(n)
	rolesArr := cfg.GetShuffledRolesConfig()

	players := make(Players)

	for i := 0; i <= n-1; i++ {
		newPlayer := NewPlayer(IDType(IDs[i]), tags[i], oldUsernames[i], serverUsernames[i], rolesArr[i])
		players[IDType(IDs[i])] = newPlayer
	}

	return players, nil
}

// _____________________________________________________________
// Load NonPlayingPlayers
// 2 different player loading options for your convenience
// _____________________________________________________________

// First

func GenerateNonPlayingPLayers(tags []string, usernames []string, serverUsernames []string) (*NonPlayingPlayers, error) {
	if len(tags) != len(usernames) {
		message := "unexpected mismatch of playing participants and nicknames"
		return nil, errors.New(message)
	}
	var players = &NonPlayingPlayers{}
	for i, tag := range tags {
		newPlayer := NewNonPlayingPlayer(tag, usernames[i], serverUsernames[i])
		players.Append(newPlayer)
	}
	return players, nil
}

// Second

func GenerateEmptyPlayersByFunc(
	x any,
	getTagUsernameAndServerUsername func(x any, index int) (string, string, string),
	countOfNewPlayers int,
) (*NonPlayingPlayers, error) {
	isRecovered := false
	defer func() {
		if r := recover(); r != nil {
			isRecovered = true
		}
	}()

	var players NonPlayingPlayers

	for i := 0; i <= countOfNewPlayers-1; i++ {
		tag, username, serverUsername := getTagUsernameAndServerUsername(x, i)
		var newPlayer *NonPlayingPlayer
		NewNonPlayingPlayer(tag, username, serverUsername)
		players[i] = newPlayer
	}
	if isRecovered {
		returnedErrorText := "panic recovered, invalid usage of GenerateEmptyPlayersByFunc function"
		return nil, errors.New(returnedErrorText)
	}
	return &players, nil
}
