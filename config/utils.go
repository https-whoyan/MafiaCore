package config

import (
	"github.com/samber/lo"
	"math/rand"
	"sort"

	"github.com/https-whoyan/MafiaCore/roles"
)

func (cfg *RolesConfig) GetShuffledRolesConfig() []*roles.Role {
	var rolesArr []*roles.Role
	for _, roleConfig := range cfg.RolesMp {
		roleCount := roleConfig.Count
		role := roleConfig.Role
		for i := 1; i <= roleCount; i++ {
			rolesArr = append(rolesArr, role)
		}
	}

	rand.Shuffle(cfg.PlayersCount, func(i, j int) {
		rolesArr[i], rolesArr[j] = rolesArr[j], rolesArr[i]
	})

	return rolesArr
}

func (cfg *RolesConfig) GetTeamsByCfg() []roles.Team {
	mpTeams := lo.SliceToMap(
		lo.Values(cfg.RolesMp),
		func(config *RoleConfig) (roles.Team, bool) {
			return config.Role.Team, true
		},
	)

	teamsSlice := lo.Keys(mpTeams)
	sort.Slice(teamsSlice, func(i, j int) bool {
		return teamsSlice[i] < teamsSlice[j]
	})
	return teamsSlice
}

func (cfg *RolesConfig) GetMapKeyByTeamValuesRoleCfg() map[roles.Team][]*RoleConfig {
	teamsMp := make(map[roles.Team][]*RoleConfig)

	for _, roleCfg := range cfg.RolesMp {
		teamsMp[roleCfg.Role.Team] = append(teamsMp[roleCfg.Role.Team], roleCfg)
	}

	return teamsMp
}

func (cfg *RolesConfig) GetPlayersCountByTeam(team roles.Team) int {
	var count int
	for _, role := range cfg.RolesMp {
		if role.Role.Team == team {
			count += role.Count
		}
	}
	return count
}

func (cfg *RolesConfig) HasRole(role *roles.Role) bool {
	for _, roleCfg := range cfg.RolesMp {
		if roleCfg.Role == role {
			return true
		}
	}
	return false
}

func GetConfigByPlayersCountAndIndex(playersCount int, index int) *RolesConfig {
	return (*Configs[playersCount])[index]
}

func (cfg *RolesConfig) GetOrderToVote() []*roles.Role {
	rolesConfigs := lo.Values(cfg.RolesMp)
	var rolesArr []*roles.Role

	for _, roleConfig := range rolesConfigs {
		if roleConfig.Role.NightVoteOrder != -1 {
			rolesArr = append(rolesArr, roleConfig.Role)
		}
	}
	sort.Slice(rolesArr, func(i, j int) bool {
		return rolesArr[i].NightVoteOrder < rolesArr[j].NightVoteOrder
	})
	return rolesArr
}
