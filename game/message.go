package game

import (
	"fmt"
	"io"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	myFMT "github.com/https-whoyan/MafiaCore/fmt"
	playerPack "github.com/https-whoyan/MafiaCore/player"
	rolesPack "github.com/https-whoyan/MafiaCore/roles"
	myTime "github.com/https-whoyan/MafiaCore/time"

	"github.com/samber/lo"
)

// !!!!!!!!!!!!!!!!!!
// The use of all functions here assumes that you g.Init() has already been
// called and successfully executed without errors.
// !!!!!!!!!!!!!!!!!!

var (
	playersCalling = []string{"poopsies", "players", "ladies and gentlemen", "citizens"}
	playerCalling  = []string{"poops", "ancient", "modern", "member"}
)

var (
	sInt = func(s int) string { return strconv.Itoa(s) }
	sCap = func(s string) string {
		if len(s) <= 1 {
			return s
		}
		return strings.ToUpper(string(s[0])) + strings.ToLower(s[1:])
	}
	getRandomPlayersCalling = func() string { return playersCalling[rand.Intn(len(playersCalling))] }
	getRandomPlayerCalling  = func() string { return playerCalling[rand.Intn(len(playerCalling))] }
)

type Messenger struct {
	f          myFMT.FmtInterface
	Init       *initMessenger
	Night      *nightMessenger
	Day        *dayMessenger
	AfterNight *afterNightMessenger
	Finish     *finishMessenger
	Public     *PublicMessanger
}

type primitiveMessenger struct {
	f myFMT.FmtInterface
	g *Game
}

func (m primitiveMessenger) sendMessage(msg string, writer io.Writer) error {
	_, err := writer.Write([]byte(msg))
	return err
}

func NewGameMessanger(f myFMT.FmtInterface, g *Game) *Messenger {
	base := &primitiveMessenger{f: f, g: g}
	return &Messenger{
		f:          f,
		Init:       &initMessenger{base},
		Night:      &nightMessenger{base},
		Day:        &dayMessenger{base},
		AfterNight: &afterNightMessenger{base},
		Finish:     &finishMessenger{base},
	}
}

// ____________
// Init
// ____________

type initMessenger struct {
	*primitiveMessenger
}

func (m initMessenger) SendStartMessage(writer io.Writer) error {
	var message string

	f := m.f

	nl := f.LineSplitter()
	dNl := nl + nl
	iL := f.InfoSplitter()

	message = f.Bold("Have a good day, " + getRandomPlayersCalling() + "!")
	message += dNl
	message += myFMT.BoldUnderline(f, "Today, our players:") + nl

	var aboutIDMessages []string
	activePlayers := lo.Values(*m.g.active)
	sort.Slice(activePlayers, func(i, j int) bool {
		return activePlayers[j].ID < activePlayers[j].ID
	})
	for _, player := range activePlayers {
		messageAboutPlayerID := f.Tab() + f.Bold(sCap(getRandomPlayerCalling())) + " " + f.Mention(player.ServerNick)
		messageAboutPlayerID += " with ID in game " + f.Block(sInt(int(player.ID)))

		aboutIDMessages = append(aboutIDMessages, messageAboutPlayerID)
	}
	message += strings.Join(aboutIDMessages, nl)

	if len(*m.g.spectators) != 0 {
		message += dNl
		message += "From behind the scenes to support us: "
		var spectatorMentions []string

		for _, spectator := range *m.g.spectators {
			spectatorMentions = append(spectatorMentions, f.Mention(spectator.ServerNick))
		}
		message += strings.Join(spectatorMentions, ", ")
	}

	message += nl + iL + nl
	message += myFMT.ItalicUnderline(f, "Selected game configuration:") + nl
	message += m.g.rolesConfig.GetMessageAboutConfig(f)
	message += nl + iL + nl

	// Redo it if it false!!!!
	message += "A private message has been sent to each of you, you can find your ID and role in it."
	message += nl
	message += "Also, " + f.Bold("if you have an active night role, you have been added to special channels, where "+
		"you can send commands to the bot anonymously")
	if len(*m.g.spectators) != 0 {
		message += f.Italic(" (but there's no hiding from observers))))")
	}
	message += "." + nl
	if m.g.renameMode != NotRenameMode {
		message += nl
		message += "Also, all participants have been prefixed with their IDs to make it more convenient for you."
	}
	message += dNl + iL + nl
	message += f.Bold("Welcome, welcome, welcome... Happy hunger games and the odds be ever in your favor! ") +
		f.Italic("(Or just have a good game!) 🍀")

	_, err := writer.Write([]byte(message))
	return err
}

// ____________
// Night
// ____________

type nightMessenger struct {
	*primitiveMessenger
}

func (m *nightMessenger) SendInitialNightMessage(w io.Writer) error {
	f := m.f
	message := f.Bold("Night №") + f.Block(strconv.Itoa(m.g.nightCounter)) + " is coming." + f.LineSplitter()
	message += fmt.Sprintf("On this night you are played by %v players.", len(*m.g.active))
	return m.sendMessage(message, w)
}

func (m *nightMessenger) SendInvitingToVoteMessage(p *playerPack.Player, deadlineInSeconds int, w io.Writer) error {
	m.g.RLock()
	defer m.g.RUnlock()
	f := m.f
	message := f.Bold("Hello, " + f.Mention(p.ServerNick) + ". It's your turn to Vote.")
	message += f.LineSplitter()
	message += myFMT.BoldUnderline(f, fmt.Sprintf("Deadline: %v seconds.", deadlineInSeconds))
	return m.sendMessage(message, w)
}

func (m *nightMessenger) SendToPlayerThatIsMutedMessage(p *playerPack.Player, w io.Writer) error {
	m.g.RLock()
	defer m.g.RUnlock()

	message := "Oops.... someone was muted today!" + m.f.Mention(p.ServerNick) +
		" , just chill, bro."
	return m.sendMessage(message, w)
}

func (m *nightMessenger) SendThanksToMutedPlayerMessage(p *playerPack.Player, writer io.Writer) error {
	m.g.RLock()
	defer m.g.RUnlock()
	message := m.f.Bold(m.f.Mention(p.ServerNick) + ", always thanks!")
	return m.sendMessage(message, writer)
}

func (m *nightMessenger) InfoThatTimerIsDone(writer io.Writer) error {
	m.g.RLock()
	defer m.g.RUnlock()
	message := myFMT.BoldUnderline(m.f, "The timer has run out!") + m.f.LineSplitter() + m.f.LineSplitter() +
		m.f.Italic("next time, be quicker.....")
	return m.sendMessage(message, writer)
}

// ____________
// AfterNight
// ____________

type afterNightMessenger struct {
	*primitiveMessenger
}

// SendAfterNightMessage provide a message to main chat after game.
func (m afterNightMessenger) SendAfterNightMessage(l NightLog, w io.Writer) error {
	f := m.f
	message := myFMT.BoldItalic(f, "Dear citizens!") + f.LineSplitter()
	message += f.Bold("Today, we're losing")
	if len(l.Dead) == 0 {
		message += "....  " + myFMT.BoldUnderline(f, "Just our nerve cells...") + f.LineSplitter()
		message += f.Bold("Everyone survived.")
		return m.sendMessage(message, w)
	}
	message += " " + f.Block(strconv.Itoa(len(l.Dead))) +
		f.Bold(" people")
	var mentions []string
	lo.ToAnySlice(l.Dead)

	idsSet := lo.SliceToMap(l.Dead, func(id playerPack.IDType) (playerPack.IDType, bool) { return id, true })
	for _, p := range *m.g.active {
		if idsSet[p.ID] {
			mentions = append(mentions, f.Mention(p.ServerNick))
		}
	}
	message += " which is to say: " + strings.Join(mentions, ", ")
	message += f.LineSplitter() + f.LineSplitter()
	message += f.Bold("Dear victims, you have " +
		strconv.Itoa(myTime.LastWordDeadlineMinutes) + " minute to say your angry.")
	return m.sendMessage(message, w)
}

// _____
// Day
// _____

type dayMessenger struct {
	*primitiveMessenger
}

func (m dayMessenger) SendMessageAboutNewDay(w io.Writer, deadline time.Duration) error {
	f := m.f

	var message string
	message += "Comes a " + f.Block(strconv.Itoa(m.g.nightCounter)) + " day. "
	strMinutes := strconv.Itoa(int(math.Ceil(deadline.Minutes())))
	message += f.Bold("You have a ") + f.Block(strMinutes) + " minutes to set your votes."
	message += f.LineSplitter()
	message += f.LineSplitter()

	message += f.Bold("Skip voting will be, if ") + f.Block(strconv.Itoa(DayPercentageToNextStage)+"%") +
		" of player leave vote to skip."
	return m.sendMessage(message, w)
}

func (m dayMessenger) SendMessageThatDayIsSkipped(w io.Writer) error {
	var message string
	message = m.f.Bold("Today's vote has been skipped!")
	return m.sendMessage(message, w)
}

func (m dayMessenger) SendMessageAboutKickedPlayer(w io.Writer, kickedPlayer *playerPack.Player) error {
	var message string
	message = m.f.Bold("As a result of today's vote, the ousted... " + m.f.Mention(kickedPlayer.ServerNick))
	return m.sendMessage(message, w)
}

// _____________________
// Team victory message
// _____________________

type finishMessenger struct {
	*primitiveMessenger
}

func (m finishMessenger) basicEndGameMessage() string {
	var message string
	message = m.f.Bold("Dear ladies and gentlemen!") + m.f.LineSplitter()
	message += m.f.Tab() + myFMT.BoldUnderline(m.f, "Game is over!")
	message += m.f.LineSplitter() + m.f.InfoSplitter() + m.f.LineSplitter()
	return message + m.SendParticipantAboutMessage()
}

func (m finishMessenger) SendParticipantAboutMessage() string {
	f := m.f
	var message string
	message = myFMT.BoldUnderline(f, "And the roles of the participants were:") + f.LineSplitter() + f.LineSplitter()

	allPartitionsMp := m.g.active
	allPartitionsMp.Append(m.g.dead.ConvertToPlayers())

	allPartitionsSlice := lo.Values(*allPartitionsMp)

	sort.Slice(allPartitionsSlice, func(i, j int) bool {
		return allPartitionsSlice[i].ID < allPartitionsSlice[j].ID
	})

	for _, p := range allPartitionsSlice {
		playerMessage := "With ID " + f.Block(strconv.Itoa(int(p.ID)))
		playerMessage += " played " + f.Mention(p.ServerNick)
		playerMessage += " and " + f.Bold("his role was ") + myFMT.BoldUnderline(f, p.Role.Name)
		message += f.Tab() + playerMessage + f.LineSplitter()
	}
	message += f.InfoSplitter() + f.LineSplitter()

	return message
}

func (m finishMessenger) SendMessagesAboutEndOfGame(l FinishLog, w io.Writer) error {
	var message string
	if l.IsFool {
		message = m.getFoolWinnerMessage()
	} else {
		message = m.getTeamWinnerMessage(*l.WinnerTeam)
	}
	return m.sendMessage(message, w)
}

func (m finishMessenger) getTeamWinnerMessage(team rolesPack.Team) string {
	var message = m.basicEndGameMessage()

	message += m.f.Bold("This game was won by the team ") + rolesPack.StringTeam[team]
	message += m.f.LineSplitter()

	message += "Nice try!"
	return message
}

func (m finishMessenger) getFoolWinnerMessage() string {
	var message = m.basicEndGameMessage()

	message += m.f.Bold("You've been fooled by a fool!") +
		"The fool's goal is to get ousted during the day's voting."
	message += m.f.LineSplitter()

	// Search fool
	var fool = &playerPack.Player{}
	for _, p := range *m.g.active {
		if p.Role == rolesPack.Fool {
			fool = p
		}
	}
	foolMentions := m.f.Mention(fool.ServerNick)
	message += "Fool in this game was: " + foolMentions
	message += "Nice try!"
	return message
}

type PublicMessanger struct {
	*primitiveMessenger
}

func (p *PublicMessanger) SendMessageToMainChat(message string) error {
	_, err := p.g.mainChannel.Write([]byte(message))
	return err
}
