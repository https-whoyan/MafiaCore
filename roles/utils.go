package roles

import (
	"github.com/samber/lo"
	"slices"
	"sort"
)

// Utils.

// __________________
// This contains all the functions that link the role name to the role, and things like that.
// __________________

func GetAllNightInteractionRolesNames() []string {
	return lo.Filter(
		lo.MapToSlice(
			MappedRoles,
			func(name string, role *Role) string {
				return name
			},
		),
		func(name string, _ int) bool {
			return MappedRoles[name].NightVoteOrder != -1
		},
	)
}

func GetInteractionRoleNamesWhoHasOwnChat() []string {
	roles := GetAllNightInteractionRolesNames()
	donIndex := slices.Index(roles, Don.Name)
	if donIndex == -1 {
		return roles
	}
	return slices.Delete(roles, donIndex, donIndex+1)
}

func GetAllRolesNames() []string {
	allRoles := GetAllSortedRoles()
	var roleNames []string
	for _, role := range allRoles {
		roleNames = append(roleNames, role.Name)
	}

	return roleNames
}

func GetRoleByName(roleName string) (*Role, bool) {
	role, ok := MappedRoles[roleName]
	return role, ok
}

func GetAllSortedRoles() []*Role {
	allRoles := lo.Values(MappedRoles)

	sort.Slice(allRoles, func(i, j int) bool {
		return allRoles[i].Team < allRoles[j].Team
	})

	return allRoles
}

func GetAllTeams() []Team {
	mpTeams := make(map[Team]bool)
	for _, role := range MappedRoles {
		mpTeams[role.Team] = true
	}

	teamsSlice := lo.Keys(mpTeams)
	sort.Slice(teamsSlice, func(i, j int) bool {
		return teamsSlice[i] < teamsSlice[j]
	})
	return teamsSlice
}
