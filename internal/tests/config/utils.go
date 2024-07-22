package config

import (
	"math/rand"

	"github.com/https-whoyan/MafiaCore/config"
)

func GetRandomConfig() *config.RolesConfig {
	var cfgPlayersCount int
	for playersCount, available := range config.Configs {
		if len(*available) == 0 {
			continue
		}
		cfgPlayersCount = playersCount
		break
	}
	available := *(config.Configs[cfgPlayersCount])
	availableLen := len(available)
	return available[rand.Intn(availableLen-1)]
}
