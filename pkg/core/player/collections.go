package player

import (
	"github.com/https-whoyan/MafiaBot/core/roles"
	"strconv"
)

// ________________________
// Types of Collections.
// ________________________

type NonPlayingPlayers []*NonPlayingPlayer

// Players Type
// Key - player IDType
//
// Used only for g.Active
type Players map[IDType]*Player

// DeadPlayers
// Key - dead player's role
// Value - slice of DeadPlayer s
type DeadPlayers map[*roles.Role][]*DeadPlayer

// ______________________________
// Methods for NonPlayingPlayers
// ______________________________

func (s NonPlayingPlayers) GetTags() []string {
	var tags []string

	for _, p := range s {
		tags = append(tags, p.Tag)
	}
	return tags
}

func (s NonPlayingPlayers) GetUsernames() []string {
	var usernames []string
	for _, p := range s {
		usernames = append(usernames, p.OldNick)
	}
	return usernames
}

func (s NonPlayingPlayers) GetServerNicknames() []string {
	var serverNames []string
	for _, p := range s {
		serverNames = append(serverNames, p.ServerNick)
	}
	return serverNames
}

// ______________________________
// Methods for Players
// ______________________________

func (s Players) SearchPlayerByServerID(ID string) *Player {
	for _, p := range s {
		if p.Tag == ID {
			return p
		}
	}

	return nil
}

func (s Players) SearchPlayerByGameID(ID string) *Player {
	intID, err := strconv.Atoi(ID)
	if err != nil {
		return nil
	}
	return s[IDType(intID)]
}

func (s Players) SearchPlayerByID(ID string, isServerID bool) *Player {
	if isServerID {
		return s.SearchPlayerByServerID(ID)
	}
	return s.SearchPlayerByGameID(ID)
}

func (s Players) GetTags() []string {
	var tags []string
	for _, p := range s {
		tags = append(tags, p.Tag)
	}
	return tags
}

func (s Players) GetServerNicknames() []string {
	var usernames []string
	for _, p := range s {
		usernames = append(usernames, p.ServerNick)
	}
	return usernames
}

func (s Players) SearchAllPlayersWithRole(role *roles.Role) Players {
	allPlayers := make(Players)
	for _, player := range s {
		if player.Role == role {
			allPlayers[player.ID] = player
		}
	}

	return allPlayers
}

// ______________________
// DeadPlayers func s
// ______________________

func (s DeadPlayers) GetTags() []string {
	var tags []string
	for _, deadPlayers := range s {
		for _, p := range deadPlayers {
			tags = append(tags, p.Tag)
		}
	}
	return tags
}

func (s *DeadPlayers) Add(players ...*DeadPlayer) {
	for _, p := range players {
		(*s)[p.Role] = append((*s)[p.Role], p)
	}
}
