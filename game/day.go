package game

import (
	"log"
	"math"
	"strconv"
	"time"

	"github.com/https-whoyan/MafiaCore/player"
)

const (
	DayPersentageToNextStage = 50
)

func (g *Game) Day() DayLog {
	select {
	case <-g.ctx.Done():
		return DayLog{}
	default:
		g.SetState(DayState)
		g.infoSender <- g.newSwitchStateSignal()

		g.RLock()
		deadline := CalculateDayDeadline(
			g.nightCounter, g.dead.Len(), g.rolesConfig.PlayersCount)
		g.RUnlock()
		safeSendErrSignal(g.infoSender, g.messenger.Day.SendMessageAboutNewDay(g.mainChannel, deadline))

		return g.StartDayVoting(deadline)
	}
}

func (g *Game) StartDayVoting(deadline time.Duration) DayLog {
	votesMp := make(map[int]int)
	occurrencesMp := make(map[int]int)

	g.timer(deadline)

	var kickedPlayerID = -1
	var breakDownDayPlayersCount = int(math.Ceil(float64(DayPersentageToNextStage*g.active.Len()) / 100.0))

	acceptTheVote := func(voteP NightVoteProviderInterface) (kickedID *int, isEndVoting bool) {
		var votedPlayerID = int(g.active.SearchPlayerByID(voteP.GetVotedPlayerID()).ID)
		var vote, _ = strconv.Atoi(voteP.GetVote())

		if prevVote, isContains := votesMp[votedPlayerID]; isContains {
			occurrencesMp[prevVote]--
		}
		occurrencesMp[vote]++
		votesMp[votedPlayerID] = vote

		// If occurrencesMp[vote] >= breakDownDayPlayersCount
		if occurrencesMp[vote] >= breakDownDayPlayersCount {
			kickedID = &vote
			return kickedID, true
		}
		// Case, when all players leave his vote
		if len(votesMp) == g.active.Len() {
			// Calculate pVote, which have maximum occurrences
			var (
				mxOccurrence = 0
				mxVote       = 0
			)

			for pVote, occurrence := range occurrencesMp {
				if occurrence > mxOccurrence {
					mxOccurrence = occurrence
					mxVote = pVote
				} else if occurrence == mxOccurrence {
					mxVote = -1
				}
			}

			if mxVote == -1 {
				return nil, true
			}
			kickedID = &mxVote
			return kickedID, true
		}

		return
	}

	dayLog := DayLog{
		DayNumber: g.nightCounter,
		IsSkip:    false,
	}

	standDayLog := func(kickedID *int) {
		dayLog.Kicked = kickedID
		dayLog.DayVotes = votesMp
		dayLog.IsSkip = false
		if kickedID == nil || *kickedID == -1 {
			dayLog.Kicked = nil
			dayLog.IsSkip = true
		}
	}

	for {
		isNeedToContinue := true
		select {
		case <-g.ctx.Done():
			break
		case <-g.timerDone:
			isNeedToContinue = false
			break
		case voteP := <-g.dayVoteChan:
			err := g.dayVote(voteP, nil)
			if err != nil {
				g.infoSender <- newErrSignal(err)
				break
			}
			maybeKickedID, isEnd := acceptTheVote(voteP)
			log.Println(maybeKickedID, isEnd)
			if isEnd {
				if maybeKickedID != nil {
					kickedPlayerID = *maybeKickedID
				} else {
					kickedPlayerID = -1
				}
				isNeedToContinue = false
				g.timerStop <- struct{}{}
				break
			}
		}
		if !isNeedToContinue {
			break
		}
	}

	standDayLog(&kickedPlayerID)
	return dayLog
}

// CalculateDayDeadline calculate the day max time.
func CalculateDayDeadline(nighCounter int, deadCount int, totalPlayers int) time.Duration {
	// Weights of aspects
	const (
		currNightCounterWeight  = 0.61
		deadCountWeight         = 0.68
		totalPlayersCountWeight = 0.27
	)

	var basicMinutes = 0.0 // TODO 2.2
	nightCounterAddMinutes := currNightCounterWeight * float64(nighCounter)
	deadCountAddMinutes := deadCountWeight * float64(deadCount)
	totalPlayersCountAddMinutes := totalPlayersCountWeight * float64(totalPlayers)

	totalTime := basicMinutes + nightCounterAddMinutes + deadCountAddMinutes + totalPlayersCountAddMinutes
	totalTimeMinutes := math.Ceil(totalTime)
	return time.Minute * time.Duration(totalTimeMinutes)
}

func (g *Game) AffectDay(l DayLog) (isFool bool) {
	if l.IsSkip {
		safeSendErrSignal(g.infoSender, g.messenger.Day.SendMessageThatDayIsSkipped(g.mainChannel))
		return
	}
	kickedPlayer := (*g.active)[player.IDType(*l.Kicked)]
	safeSendErrSignal(g.infoSender, g.messenger.Day.SendMessageAboutKickedPlayer(g.mainChannel, kickedPlayer))

	g.active.ToDead(kickedPlayer.ID, player.KilledByDayVoting, g.nightCounter, g.dead)
	return
}
