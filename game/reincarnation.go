package game

import (
	"github.com/https-whoyan/MafiaCore/player"
	"github.com/https-whoyan/MafiaCore/roles"
	"github.com/samber/lo"
)

// This is where all the code regarding reincarnation and role reversal is contained.
// This may be necessary if a role is specified to become a different role in certain scenarios.

func (g *Game) reincarnation(p *player.Player) {
	switch p.Role {
	case roles.Don:
		g.donReincarnation(p)
	}
	return
}

func (g *Game) donReincarnation(p *player.Player) {
	// I find out he's the only one left on the mafia team.
	g.RLock()

	mafiaTeamCounter := lo.CountValues(lo.Map(
		lo.Values(*g.active),
		func(p *player.Player, _ int) roles.Team {
			return p.Role.Team
		}),
	)[roles.MafiaTeam]
	if mafiaTeamCounter > 1 {
		g.RUnlock()
		return
	}
	p.Role = roles.Mafia
	safeSendErrSignal(g.infoSender, g.roleChannels[roles.Don].RemoveUser(p.Tag))
	safeSendErrSignal(g.infoSender, g.roleChannels[roles.Mafia].AddPlayer(p.Tag))

	f := g.messenger.f
	g.RUnlock()
	var message string
	message = f.Bold("Hello, dear ") + f.Mention(p.ServerNick) + "." + f.LineSplitter()
	message += "You are the last player left alive from the mafia team, so you become mafia." + f.LineSplitter()
	message += f.Underline("Don't reveal yourself.")
	_, err := g.roleChannels[roles.Mafia].Write([]byte(message))
	safeSendErrSignal(g.infoSender, err)
}
