package game

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/https-whoyan/MafiaCore/log"
	"os"
	"sync"
	"time"

	channelPack "github.com/https-whoyan/MafiaCore/channel"
	configPack "github.com/https-whoyan/MafiaCore/config"
	fmtPack "github.com/https-whoyan/MafiaCore/fmt"
	playerPack "github.com/https-whoyan/MafiaCore/player"
	rolesPack "github.com/https-whoyan/MafiaCore/roles"
	timePack "github.com/https-whoyan/MafiaCore/time"
)

// This file describes the structure of the game, as well as the start and end functions of the game.

// ____________________
// Types and constants
// ____________________

type RenameMode int8

const (
	// NotRenameMode used if you not want to rename users in your implementations
	NotRenameMode RenameMode = iota
	// RenameInGuildMode used if you want to rename user everything in your guild
	RenameInGuildMode
	// RenameOnlyInMainChannelMode used if you want to rename user only in mainChannel
	RenameOnlyInMainChannelMode
	// RenameInAllChannelsMode used if you want to rename user in every channel (Roles and Main)
	RenameInAllChannelsMode
)

// ____________________
// Options
// ____________________

type Option func(g *Game)

func FMTerOpt(fmtEr fmtPack.FmtInterface) Option {
	return func(g *Game) {
		messenger := NewGameMessanger(fmtEr, g)
		g.messenger = messenger
	}
}
func RenamePrOpt(rP playerPack.RenameUserProviderInterface) Option {
	return func(g *Game) { g.renameProvider = rP }
}
func RenameModeOpt(mode RenameMode) Option {
	return func(g *Game) { g.renameMode = mode }
}
func VotePingOpt(votePing int) Option {
	return func(g *Game) { g.votePing = votePing }
}
func StorageOpt(logger Storage) Option {
	return func(g *Game) { g.storage = logger }
}
func InfoLoggerOpt(logger log.Logger) Option {
	return func(g *Game) { g.infoLogger = logger }
}
func ErrLoggerOpt(logger log.Logger) Option {
	return func(g *Game) { g.errorLogger = logger }
}
func VoteForYourselfOpt(voteForYourself bool) Option {
	return func(g *Game) { g.voteForYourself = voteForYourself }
}

// __________________
// Game struct
// __________________

type Game struct {
	sync.RWMutex
	ctx context.Context
	// Presents the server where the game is running.
	// Or GameID.
	// Depends on the implementation.
	//
	// Possibly, may be empty.
	guildID      string
	playersCount int
	rolesConfig  *configPack.RolesConfig
	nightCounter int
	infoLogger   log.Logger
	errorLogger  log.Logger

	timeStart time.Time
	endTime   time.Time

	startPlayers *playerPack.NonPlayingPlayers
	active       *playerPack.Players
	dead         *playerPack.DeadPlayers
	spectators   *playerPack.NonPlayingPlayers

	// Presents to the application which chat is used for which role.
	// key: str - role name
	roleChannels map[*rolesPack.Role]channelPack.RoleChannel
	mainChannel  channelPack.MainChannel

	// Keeps what role is voting (in night) right now.
	nightVoting *rolesPack.Role

	nightLogs []NightLog
	dayLogs   []DayLog

	voteAccepted chan struct{}
	dayVoteChan  chan DayVoteProviderInterface
	// Can the player choose himself
	voteForYourself bool
	// votePing presents a delay number for voting for the same player again.
	//
	// Example: A player has voted for players with IDs 5, 4, 3, and votePing is 2.
	// So the player will not be able to Vote for players 4 and 3 the next night.
	//
	// Default value: 1.
	//
	// Adjustable by option.
	votePing int

	timerDone chan struct{}
	timerStop chan struct{}

	previousState State
	state         State
	messenger     *Messenger
	// Use to rename user in your interpretation
	renameProvider playerPack.RenameUserProviderInterface
	renameMode     RenameMode
	errSender      chan<- ErrSignal
	infoSender     chan<- InfoSignal
	errChanDest    <-chan ErrSignal
	infoChanDest   <-chan InfoSignal
	finishFuncOnce *sync.Once
	finishOnce     *sync.Once
	storage        Storage
}

func GetNewGame(ctx context.Context, guildID string, opts ...Option) *Game {
	start := playerPack.NonPlayingPlayers{}
	active := make(playerPack.Players)
	dead := make(playerPack.DeadPlayers)
	spectators := playerPack.NonPlayingPlayers{}
	errChan := make(chan ErrSignal)
	infoChan := make(chan InfoSignal)
	logger := log.New(os.Stderr, fmt.Sprintf("Game, ID: %v\t", guildID), log.Ldate|log.Ltime|log.Lshortfile)
	newGame := &Game{
		guildID: guildID,
		state:   NonDefinedState,
		// Chan s create.
		voteAccepted: make(chan struct{}),
		timerDone:    make(chan struct{}),
		timerStop:    make(chan struct{}),
		dayVoteChan:  make(chan DayVoteProviderInterface),
		// Slices.
		startPlayers: &start,
		active:       &active,
		dead:         &dead,
		spectators:   &spectators,
		nightLogs:    make([]NightLog, 0),
		dayLogs:      make([]DayLog, 0),
		infoLogger:   logger,
		// Create a map
		roleChannels:   make(map[*rolesPack.Role]channelPack.RoleChannel),
		votePing:       1,
		errSender:      errChan,
		infoSender:     infoChan,
		errChanDest:    errChan,
		infoChanDest:   infoChan,
		finishFuncOnce: &sync.Once{},
		finishOnce:     &sync.Once{},
		ctx:            ctx,
	}
	messanger := NewGameMessanger(fmtPack.NilFMTInterfaceInstance, newGame)
	newGame.messenger = messanger
	// Set options
	for _, opt := range opts {
		opt(newGame)
	}
	return newGame
}

// ___________________________
// Game.Init validator
// __________________________
/*
	After RegisterGame I must have all information about
		1) Tags and usernames of players
		2) roleChannels info
		3) guildID (Ok, optional)
		4) mainChannel implementation
		5) spectators
		6) And chan s (See GetNewGame)
		7) fmtEr
		8) renameProvider
		9) renameMode

	Let's validate it.
*/

// Init validation Errors.
var (
	EmptyConfigErr                             = errors.New("empty config")
	MismatchPlayersCountAndGamePlayersCountErr = errors.New("mismatch config playersCount and game players")
	NotFullRoleChannelInfoErr                  = errors.New("not full role channel info")
	NotMainChannelInfoErr                      = errors.New("not main channel info")
	EmptyFMTerErr                              = errors.New("empty FMTer")
	EmptyRenameProviderErr                     = errors.New("empty rename provider")
	EmptyRenameModeErr                         = errors.New("empty rename mode")
)

// validationStart is used to validate the game before it is fully initialized.
func (g *Game) validationStart(cfg *configPack.RolesConfig) error {
	g.RLock()
	defer g.RUnlock()

	var err error

	if cfg == nil {
		return EmptyConfigErr
	}

	if cfg.PlayersCount != len(*(g.startPlayers)) {
		err = multierror.Append(err, MismatchPlayersCountAndGamePlayersCountErr)
	}
	if len(g.roleChannels) != len(rolesPack.GetAllNightInteractionRolesNames()) {
		err = multierror.Append(err, NotFullRoleChannelInfoErr)
	}
	if g.mainChannel == nil {
		err = multierror.Append(err, NotMainChannelInfoErr)
	}
	if g.messenger == nil {
		err = multierror.Append(err, EmptyFMTerErr)
	}
	if g.renameMode == NotRenameMode {
		return err
	}
	if g.renameProvider == nil {
		err = multierror.Append(err, EmptyRenameProviderErr)
	}
	switch g.renameMode {
	case RenameInGuildMode:
		return err
	case RenameOnlyInMainChannelMode:
		return err
	case RenameInAllChannelsMode:
		return err
	default:
		err = multierror.Append(err, EmptyRenameModeErr)
	}
	return err
}

// Init
/*
The Init function is used to generate all players, add all players to channels, and rename all players.
It is also used to validate all fields of the game.
This is the penultimate and mandatory function that you must call before starting the game.

Before using it, you must have all options set, all players must have known ServerIDs,
Tags and serverUsernames (all of which must be in startPlayers), and all channels,
both role-based and non-role-based, must be set.
See the realization of the ValidationStart function (line 139)

Also see the file loaders.go in the same package https://github.com/https-whoyan/MafiaBot/blob/main/pkg/core/game/loaders.go.


More references:
https://github.com/https-whoyan/MafiaBot/blob/main/pkg/core/player/loader.go line 50

(DISCORD ONLY): https://github.com/https-whoyan/MafiaBot/blob/main/internal/converter/user.go
*/
func (g *Game) Init(cfg *configPack.RolesConfig) (err error) {
	if err = g.validationStart(cfg); err != nil {
		return err
	}
	// Set fmtEr
	// Set config and players count
	g.SetState(StartingState)
	g.Lock()
	g.rolesConfig = cfg
	g.playersCount = cfg.PlayersCount
	g.timeStart = time.Now()
	g.Unlock()

	// Get Players
	tags := g.startPlayers.GetTags()
	oldNicknames := g.startPlayers.GetUsernames()
	serverUsernames := g.startPlayers.GetServerNicknames()
	players, err := playerPack.GeneratePlayers(tags, oldNicknames, serverUsernames, cfg)
	if err != nil {
		return err
	}
	// And state it to active field
	g.Lock()
	g.active = &players
	g.Unlock()

	g.RLock()
	defer g.RUnlock()
	// ________________
	// Add to channels
	// ________________

	// We need to add spectators and players to channel.
	// First, add users to role channels.
	for _, player := range *g.active {
		if player.Role.NightVoteOrder == -1 {
			continue
		}

		playerChannel := g.roleChannels[player.Role]
		err = playerChannel.AddPlayer(player.Tag)
		if err != nil {
			return
		}
	}

	// Then add spectators to game
	for _, spectator := range *g.spectators {
		for _, interactionChannel := range g.roleChannels {
			err = interactionChannel.AddSpectator(spectator.Tag)
			if err != nil {
				return err
			}
		}
	}

	// Then, add all players to main chat.
	for _, player := range *g.startPlayers {
		err = g.mainChannel.AddPlayer(player.Tag)
		if err != nil {
			return err
		}
	}
	// And spectators.
	for _, spectator := range *g.spectators {
		err = g.mainChannel.AddSpectator(spectator.Tag)
		if err != nil {
			return err
		}
	}

	// _______________
	// Renaming.
	// _______________
	switch g.renameMode {
	case NotRenameMode: // No actions
	case RenameInGuildMode:
		for _, player := range *g.active {
			err = player.RenameAfterGettingID(g.renameProvider, "", g.infoLogger)
			if err != nil {
				return err
			}
		}
		for _, spectator := range *g.spectators {
			err = spectator.RenameToSpectator(g.renameProvider, "", g.infoLogger)
			if err != nil {
				return err
			}
		}
	case RenameOnlyInMainChannelMode:
		mainChannelServerID := g.mainChannel.GetServerID()

		for _, player := range *g.active {
			err = player.RenameAfterGettingID(g.renameProvider, mainChannelServerID, g.infoLogger)
			if err != nil {
				return err
			}
		}
	case RenameInAllChannelsMode:
		// Add to Role Channels.
		for _, player := range *g.active {
			if player.Role.NightVoteOrder == -1 {
				continue
			}

			playerInteractionChannel := g.roleChannels[player.Role]
			playerInteractionChannelIID := playerInteractionChannel.GetServerID()
			err = player.RenameAfterGettingID(g.renameProvider, playerInteractionChannelIID, g.infoLogger)
			if err != nil {
				return err
			}
		}

		// Add to main channel
		mainChannelServerID := g.mainChannel.GetServerID()

		for _, player := range *g.active {
			err = player.RenameAfterGettingID(g.renameProvider, mainChannelServerID, g.infoLogger)
			if err != nil {
				return err
			}
		}
	default:
		return errors.New("invalid rename mode")
	}
	if g.storage != nil {
		deepClone, deepCloneErr := g.GetDeepClone()
		if deepCloneErr != nil {
			return deepCloneErr
		}
		err = g.storage.InitNewGame(g.ctx, deepClone)
		return err
	}

	return nil
}

// ********************
// ____________________
// Main Cycle in game.
// ___________________
// ********************
// ********************

var (
	NilContext            = errors.New("nil context")
	ErrGameAlreadyStarted = errors.New("game already started")
)

// Run
/*
Is used to start the game.

Runs the run method in its goroutine.
Used after Init()

Also call deferred finish() (or FinishAnyway(), if game was stopped by context)

It is recommended to use context.Background()

Return receive chan of Signal type, that informed you about Signal's
*/
func (g *Game) Run(ctx context.Context) (<-chan ErrSignal, <-chan InfoSignal) {
	go func() {
		// Send InteractionMessage About New Game
		err := g.messenger.Init.SendStartMessage(g.mainChannel)
		// Used for participants to familiarize themselves with their roles, and so on.
		time.Sleep(timePack.RoleInfoCount * time.Second)
		safeSendErrSignal(g.errSender, err)
		switch {
		case ctx == nil:
			sendFatalSignal(g.errSender, NilContext)
		case g.IsRunning():
			sendFatalSignal(g.errSender, ErrGameAlreadyStarted)
		default:
			g.Lock()
			g.ctx = ctx
			g.Unlock()

			var (
				finishLog      *FinishLog
				isStoppedByCtx bool
			)

			// Tracing

			defer func() {
				if r := recover(); r != nil {
					isStoppedByCtx = true
					g.errSender <- newErrSignal(
						fmt.Errorf("panic recoved, err %v", err),
					)
					g.FinishAnyway()
				}
			}()

			isStoppedByCtx, finishLog = g.run()

			switch isStoppedByCtx {
			case true:
				g.FinishAnyway()
			case false:
				g.FinishByFinishLog(*finishLog)
			}
		}
	}()

	return g.errChanDest, g.infoChanDest
}

func (g *Game) run() (isStoppedByCtx bool, finishLog *FinishLog) {
	// FinishState will be set when the winner is already clear.
	// This will be determined after the night and after the day's voting.

	for g.state != FinishState {
		isNeedToContinue := true
		select {
		case <-g.ctx.Done():
			isStoppedByCtx = true
			isNeedToContinue = false
			return true, nil
		default:
			g.Lock()
			g.nightCounter++
			g.Unlock()

			// Night

			nightLog := g.Night()
			g.nightLogs = append(g.nightLogs, nightLog)
			g.AffectNight(nightLog)
			if g.storage != nil {
				deepClone, deepCloneErr := g.GetDeepClone()
				safeSendErrSignal(g.errSender, deepCloneErr)
				err := g.storage.SaveNightLog(g.ctx, deepClone, nightLog)
				safeSendErrSignal(g.errSender, err)
			}

			// Validate is final?

			winnerTeam := g.UnderstandWinnerTeam()
			if winnerTeam != nil {
				finishLogValue := g.NewFinishLog(winnerTeam, false)
				finishLog = &finishLogValue
				isNeedToContinue = false
				break
			}

			// Day

			dayLog := g.Day()
			g.dayLogs = append(g.dayLogs, dayLog)
			g.AffectDay(dayLog)
			if g.storage != nil {
				deepClone, deepCloneErr := g.GetDeepClone()
				safeSendErrSignal(g.errSender, deepCloneErr)
				err := g.storage.SaveDayLog(g.ctx, deepClone, dayLog)
				safeSendErrSignal(g.errSender, err)
			}

			// Validate is final?
			winnerTeam = g.UnderstandWinnerTeam()
			fool := (*g.dead.ConvertToPlayers().SearchAllPlayersWithRole(rolesPack.Fool))[0]

			if dayLog.Kicked != nil && *dayLog.Kicked == fool.ID {
				finishLogValue := g.NewFinishLog(nil, true)
				finishLog = &finishLogValue
				isNeedToContinue = false
				break
			} else if winnerTeam != nil {
				finishLogValue := g.NewFinishLog(winnerTeam, false)
				finishLog = &finishLogValue
				isNeedToContinue = false
				break
			}
			g.ClearDayVotes()
		}

		if !isNeedToContinue {
			break
		}
	}
	g.ClearDayVotes()
	return
}

// ********************
// ____________________
// Finishing functions
// ___________________
// ********************

func (g *Game) FinishByFinishLog(l FinishLog) {
	err := g.messenger.Finish.SendMessagesAboutEndOfGame(l, g.mainChannel)
	if err != nil {
		g.errSender <- newErrSignal(err)
	}
	g.finishFuncOnce.Do(func() {
		g.endTime = time.Now()
		g.SetState(FinishState)
		if g.storage != nil {
			deepClone, deepCloneErr := g.GetDeepClone()
			safeSendErrSignal(g.errSender, deepCloneErr)
			loggerErr := g.storage.SaveFinishLog(g.ctx, deepClone, l)
			safeSendErrSignal(g.errSender, loggerErr)
		}
		g.replaceCtx()
		g.finish()
	})
}

func (g *Game) replaceCtx() {
	g.Lock()
	if g.ctx == nil {
		g.ctx = context.Background()
	}
	newCtx, cancel := context.WithCancel(g.ctx)
	g.ctx = newCtx
	g.Unlock()
	cancel()
}

// FinishAnyway is used to end the running game anyway.
func (g *Game) FinishAnyway() {
	g.finishFuncOnce.Do(func() {
		g.endTime = time.Now()
		if g.mainChannel != nil {
			content := "The game was suspended."
			_, err := g.mainChannel.Write([]byte(g.messenger.Finish.f.Bold(content)))
			safeSendErrSignal(g.errSender, err)
		}
		g.SetState(FinishState)
		g.replaceCtx()
		g.finish()
	})
}

func (g *Game) finish() {
	g.finishOnce.Do(func() {
		// Delete from channels
		for _, player := range *g.active {
			if player.Role.NightVoteOrder == -1 {
				continue
			}

			playerChannel := g.roleChannels[player.Role]
			safeSendErrSignal(g.errSender, playerChannel.RemoveUser(player.Tag))
		}

		// Then remove spectators from game
		for _, tag := range playerPack.GetTags(g.dead, g.spectators) {
			for _, interactionChannel := range g.roleChannels {
				safeSendErrSignal(g.errSender, interactionChannel.RemoveUser(tag))
			}
		}

		// Then, remove all players of main chat.
		for _, player := range *g.startPlayers {
			safeSendErrSignal(g.errSender, g.mainChannel.RemoveUser(player.Tag))
		}
		// And spectators.
		for _, spectator := range *g.spectators {
			safeSendErrSignal(g.errSender, g.mainChannel.RemoveUser(spectator.Tag))
		}

		// _______________
		// Renaming.
		// _______________
		activePlayersAndSpectators := append(*g.startPlayers, *g.spectators...)
		switch g.renameMode {
		case NotRenameMode: // No actions
		case RenameInGuildMode:
			for _, player := range activePlayersAndSpectators {
				safeSendErrSignal(g.errSender, player.RenameUserAfterGame(g.renameProvider, "", g.infoLogger))
			}
		case RenameOnlyInMainChannelMode:
			mainChannelServerID := g.mainChannel.GetServerID()

			for _, player := range activePlayersAndSpectators {
				err := player.RenameUserAfterGame(g.renameProvider, mainChannelServerID, g.infoLogger)
				safeSendErrSignal(g.errSender, err)
			}
		case RenameInAllChannelsMode:
			// Rename from Role Channels.
			for _, player := range activePlayersAndSpectators {
				for _, interactionChannel := range g.roleChannels {
					interactionChannelID := interactionChannel.GetServerID()

					err := player.RenameUserAfterGame(g.renameProvider, interactionChannelID, g.infoLogger)
					safeSendErrSignal(g.errSender, err)
				}
			}

			// Rename from main channel
			mainChannelServerID := g.mainChannel.GetServerID()

			for _, player := range activePlayersAndSpectators {
				err := player.RenameUserAfterGame(g.renameProvider, mainChannelServerID, g.infoLogger)
				safeSendErrSignal(g.errSender, err)
			}
		default:
			sendFatalSignal(g.errSender, errors.New("invalid rename mode"))
			return
		}

		g.infoSender <- g.newFinishGameSignal()
	})
}
