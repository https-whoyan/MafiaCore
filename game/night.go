package game

import (
	"sort"
	"time"

	channelPack "github.com/https-whoyan/MafiaCore/channel"
	playerPack "github.com/https-whoyan/MafiaCore/player"
	rolesPack "github.com/https-whoyan/MafiaCore/roles"
	myTime "github.com/https-whoyan/MafiaCore/time"
)

// Night
// Actions with the game related to the night.
// Send will send signals to the channels about which role is currently voting. Comes from the g.run function
func (g *Game) Night() NightLog {
	select {
	case <-g.ctx.Done():
		return NightLog{}
	default:
		g.SetState(NightState)
		g.infoSender <- g.newSwitchStateSignal()

		err := g.messenger.Night.SendInitialNightMessage(g.mainChannel)
		safeSendErrSignal(g.errSender, err)

		// I'm getting the voting order
		g.RLock()
		orderToVote := g.rolesConfig.GetOrderToVote()
		g.RUnlock()

		// For each of the votes
		for _, votedRole := range orderToVote {
			// To avoid for shorting
			votedRoleClone := votedRole
			g.RoleNightAction(votedRoleClone)
		}
		// On this line, all votes are accepted.
		// I hereby signify that the voting is over.
		g.Lock()
		g.nightVoting = nil
		g.Unlock()
		g.RLock()

		// I do the rest of the interactions that come after the vote.
		var needToProcessPlayers []*playerPack.Player
		for _, p := range *g.active {
			if p.Role.CalculationOrder > 0 && !p.Role.UrgentCalculation {
				needToProcessPlayers = append(needToProcessPlayers, p)
			}
		}
		g.RUnlock()
		sort.Slice(needToProcessPlayers, func(i, j int) bool {
			return needToProcessPlayers[i].Role.CalculationOrder < needToProcessPlayers[j].Role.CalculationOrder
		})
		for _, p := range needToProcessPlayers {
			g.nightInteraction(p)
		}
		return g.NewNightLog()
	}

}

/*
RoleNightAction

	Counting variables, sending messages,
	adding to spectators, and like that.

	A follow-up call to the methods themselves is voice acceptance.
*/
func (g *Game) RoleNightAction(votedRole *rolesPack.Role) {

	select {
	case <-g.ctx.Done():
		return
	default:
		var err error

		g.Lock()
		g.nightVoting = votedRole
		g.Unlock()
		g.infoSender <- g.newSwitchVotingRoleSignal()
		// Finding all the players with that role.
		// And finding nightInteraction channel
		g.RLock()
		interactionChannel := g.roleChannels[votedRole]
		allPlayersWithRole := g.active.SearchAllPlayersWithRole(votedRole)
		g.RUnlock()

		var (
			nonEmptyVote1 playerPack.IDType = EmptyVoteInt
			nonEmptyVote2 playerPack.IDType = EmptyVoteInt
		)

		sendToOtherEmptyVotes := func(nonEmptyVoter *playerPack.Player) {
			voterLen := len(nonEmptyVoter.Votes)
			if votedRole.IsTwoVotes {
				nonEmptyVote1 = nonEmptyVoter.Votes[voterLen-2]
				nonEmptyVote2 = nonEmptyVoter.Votes[voterLen-1]
			} else {
				nonEmptyVote1 = nonEmptyVoter.Votes[voterLen-1]
			}

			// Set to other empty votes
			for _, playerWithRole := range *allPlayersWithRole {
				if playerWithRole == nonEmptyVoter {
					continue
				}
				if votedRole.IsTwoVotes {
					playerWithRole.Votes = append(playerWithRole.Votes, nonEmptyVote1, nonEmptyVote2)
				} else {
					playerWithRole.Votes = append(playerWithRole.Votes, nonEmptyVote1)
				}
			}
			return
		}
		findOrStandNotEmptyVoter := func() (nonEmptyVoter *playerPack.Player) {
			// Need to find a not empty Vote.
			for _, voter := range *allPlayersWithRole {
				voterVotesLen := len(voter.Votes)
				if voterVotesLen == 0 {
					continue
				}
				if voter.Votes[voterVotesLen-1] == EmptyVoteInt {
					continue
				}
				nonEmptyVoter = voter
				break
			}

			// If deadline pass, or player set EmptyVote, stand nonEmptyVoter
			// on somebody, and we'll put votes on the player.
			if nonEmptyVoter == nil {
				for _, p := range *allPlayersWithRole {
					nonEmptyVoter = p

					if votedRole.IsTwoVotes {
						p.Votes = append(p.Votes, nonEmptyVote1, nonEmptyVote2)
					} else {
						p.Votes = append(p.Votes, nonEmptyVote1)
					}

					break
				}
			}
			return
		}

		voteDeadlineInt := myTime.VotingDeadline
		voteDeadline := time.Second * time.Duration(voteDeadlineInt)

		containsNotMutedPlayers := false

		// I go through each player and, with a mention, invite them to Vote.
		// And if a player is locked, I tell him about it and add him to spectators for the duration of the Vote.
		for _, voter := range *allPlayersWithRole {
			if voter.InteractionStatus == playerPack.Muted {
				err = g.messenger.Night.SendToPlayerThatIsMutedMessage(voter, interactionChannel)
				safeSendErrSignal(g.errSender, err)

				// Add to spectator
				err = channelPack.FromUserToSpectator(interactionChannel, voter.Tag)
				safeSendErrSignal(g.errSender, err)

			} else {
				containsNotMutedPlayers = true
				err = g.messenger.Night.SendInvitingToVoteMessage(voter, voteDeadlineInt, interactionChannel)
				safeSendErrSignal(g.errSender, err)
			}
		}

		// From this differs in which channel the game will wait for the voice,
		//as well as the difference in the voice itself.
		var isTimerStop bool
		switch !votedRole.IsTwoVotes {
		case true:
			isTimerStop = g.oneVoteRoleNightVoting(containsNotMutedPlayers, voteDeadline)
		case false:
			isTimerStop = g.twoVoterRoleNightVoting(containsNotMutedPlayers, voteDeadline)
		}

		if isTimerStop && containsNotMutedPlayers {
			safeSendErrSignal(g.errSender, g.messenger.Night.InfoThatTimerIsDone(interactionChannel))
		}

		// Putting it back in the channel.
		for _, voter := range *allPlayersWithRole {
			if voter.InteractionStatus == playerPack.Muted {
				err = channelPack.FromSpectatorToUser(interactionChannel, voter.Tag)
				safeSendErrSignal(g.errSender, err)

				err = g.messenger.Night.SendThanksToMutedPlayerMessage(voter, interactionChannel)
				safeSendErrSignal(g.errSender, err)
			}
		}

		nonEmptyVoter := findOrStandNotEmptyVoter()
		sendToOtherEmptyVotes(nonEmptyVoter)

		// Case when roles need to urgent calculation
		g.infoLogger.Println(votedRole.Name, votedRole.UrgentCalculation)
		if votedRole.UrgentCalculation {
			message := g.nightInteraction(nonEmptyVoter)
			if message != nil {
				_, err = interactionChannel.Write([]byte(*message))
				safeSendErrSignal(g.errSender, err)
			}
		}
	}
}

/*
	The logic of accepting a role's Vote, and timers,
	depending on whether the role votes with 2 votes or one.
*/

func (g *Game) waitOneVoteRoleFakeTimer() {
	g.randomTimer()

	select {
	case <-g.timerDone:
		break
	case <-g.ctx.Done():
		break
	}
}

func (g *Game) oneVoteRoleNightVoting(containsNotMutedPlayers bool, deadline time.Duration) (isTimerStop bool) {
	if !containsNotMutedPlayers {
		g.waitOneVoteRoleFakeTimer()
		return
	}

	g.timer(deadline)

	select {
	case <-g.voteAccepted:
		g.timerStop <- struct{}{}
		break
	case <-g.timerDone:
		isTimerStop = true
		break
	case <-g.ctx.Done():
		break
	}
	return
}

func (g *Game) waitTwoVoteRoleFakeTimer() {
	g.randomTimer()

	select {
	case <-g.timerDone:
		break
	case <-g.ctx.Done():
		break
	}
}

func (g *Game) twoVoterRoleNightVoting(containsNotMutedPlayers bool, deadline time.Duration) (isTimerStop bool) {
	if !containsNotMutedPlayers {
		g.waitTwoVoteRoleFakeTimer()
		return
	}

	g.timer(deadline)

	select {
	case <-g.voteAccepted:
		g.infoLogger.Println("two vote accepted")
		g.timerStop <- struct{}{}
		break
	case <-g.timerDone:
		isTimerStop = true
		break
	case <-g.ctx.Done():
		break
	}
	return
}

// AffectNight changes players according to the night's actions.
// Errors during execution are sent to the channel
func (g *Game) AffectNight(l NightLog) {
	if !g.IsRunning() {
		panic("Game is not running")
	}
	if g.ctx == nil {
		panic("Game context is nil, then, don't initialed")
	}
	select {
	case <-g.ctx.Done():
		return
	default:
		g.ResetAllInteractionsStatuses()
		g.Lock()

		// Splitting arrays.
		var newDeadPersons = &playerPack.Players{}

		for _, deadID := range l.Dead {
			g.active.ToDead(deadID, playerPack.KilledAtNight, g.nightCounter, g.dead)
			newDeadPersons.Append(g.active)
		}

		// I will add add add all killed players after a minute of players a minute of
		// players after a minute, so, using goroutine.
		go g.AppendToSpectators(newDeadPersons, myTime.LastWordDeadline*time.Second)

		// Sending a message about who died today.
		err := g.messenger.AfterNight.SendAfterNightMessage(l, g.mainChannel)
		safeSendErrSignal(g.errSender, err)
		// Then, for each person try to do his reincarnation
		g.Unlock()
		for _, p := range *g.active {
			g.reincarnation(p)
		}
		return
	}
}

func (g *Game) AppendToSpectators(newSpectators interface{ GetTags() []string }, after time.Duration) {
	ticker := time.NewTicker(after)

	g.RLock()
	mainChannel := g.mainChannel
	roleChannels := g.roleChannels
	g.RUnlock()

	defer ticker.Stop()
	select {
	case <-g.ctx.Done():
		return
	case <-ticker.C:
		// I'm adding new dead players to the spectators in the channels (so they won't be so bored)
		for _, tag := range newSpectators.GetTags() {
			for _, interactionChannel := range roleChannels {
				select {
				case <-g.ctx.Done():
					return
				default:
					err := channelPack.FromUserToSpectator(interactionChannel, tag)
					safeSendErrSignal(g.errSender, err)
					break
				}
			}
			select {
			case <-g.ctx.Done():
				return
			default:
				err := channelPack.FromUserToSpectator(mainChannel, tag)
				safeSendErrSignal(g.errSender, err)
				break
			}
		}
	}
}
