package roles

import (
	"strings"

	"github.com/https-whoyan/MafiaBot/core/message/fmt"
)

// For beauty messaging

var MappedEmoji = map[string]string{
	"Citizen":   "",
	"Detective": "",
	"Doctor":    "",
	"Don":       "",
	"Fool":      "",
	"Mafia":     "",
	"Maniac":    "",
	"Peaceful":  "",
	"Whose":     "",
}

var StringTeam = map[Team]string{
	PeacefulTeam: "❤️ Peaceful",
	MafiaTeam:    "🖤 Mafia Team",
	ManiacTeam:   "\U0001FA76 Maniac Team",
}

func GetEmojiByName(name string) string {
	return MappedEmoji[name]
}

// _____________________________________________________________________
// Beautiful presentations of roles to display information about them.
// _____________________________________________________________________

func GetDefinitionOfRole(f fmt.FmtInterface, roleName string) string {
	fixDescription := func(s string) string {
		words := strings.Split(s, " ")
		return strings.Join(words, " ")
	}

	role := MappedRoles[roleName]
	var message string

	name := f.Block(role.Name)
	team := f.Bold("Team: ") + StringTeam[role.Team]
	description := fixDescription(role.Description)
	message = name + f.LineSplitter() + f.LineSplitter() + team + f.LineSplitter() + description
	return message
}

func GetDefinitionsOfAllRoles(f fmt.FmtInterface, maxBytesLenInMessage int) (messages []string) {
	allRoles := GetAllSortedRoles()
	var allDescriptions []string

	bytesCounter := 0
	rolesCounter := 0

	infoSptr := f.LineSplitter() + f.InfoSplitter() + f.LineSplitter()

	for _, role := range allRoles {
		roleDescription := GetDefinitionOfRole(f, role.Name)
		// To avoid for looping
		nextTrueBytesMessage := len(roleDescription) + bytesCounter + len(infoSptr)*(rolesCounter-1)
		if nextTrueBytesMessage >= maxBytesLenInMessage {
			messages = append(messages, strings.Join(allDescriptions, infoSptr))
			allDescriptions = []string{}
			bytesCounter = 0
			rolesCounter = 0
		}
		bytesCounter += len(roleDescription)
		allDescriptions = append(allDescriptions, roleDescription)
	}
	if len(allDescriptions) > 0 {
		messages = append(messages, strings.Join(allDescriptions, infoSptr))
	}
	return
}