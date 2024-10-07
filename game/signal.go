package game

import (
	"github.com/https-whoyan/MafiaCore/roles"
	"time"
)

// _____________________
// ErrSignal
// _____________________

type ErrSignal struct {
	InitialTime   time.Time
	ErrSignalType ErrSignalType
	Err           error
}

func (s ErrSignal) GetTime() time.Time { return s.InitialTime }

type ErrSignalType uint8

const (
	ErrorSignal ErrSignalType = iota
	FatalSignal               // After a fatal signal, the channel will close immediately.
)

// __________________
// InfoSignal
// __________________

type InfoSignal struct {
	InitialTime    time.Time
	InfoSignalType InfoSignalType
	Info           infoSignalInterface
}

type InfoSignalType uint8

const (
	SwitchStateSignal InfoSignalType = iota
	SwitchVotingRoleSignal
	FinishGameSignal
)

// See SwitchStateInfo, SwitchVotingRoleInfo, FinishGameInfo
type infoSignalInterface interface {
	infoSignalInterfacePrivateMethod()
}

type SwitchStateInfo struct {
	DayCounter    int
	PreviousState State
	NewState      State
}

func (SwitchStateInfo) infoSignalInterfacePrivateMethod() {}

type SwitchVotingRoleInfo struct {
	CurrVotingRole *roles.Role
}

func (SwitchVotingRoleInfo) infoSignalInterfacePrivateMethod() {}

type FinishGameInfo struct{}

func (FinishGameInfo) infoSignalInterfacePrivateMethod() {}

// InternalCode

// errs

func newErrSignal(err error) ErrSignal {
	return ErrSignal{
		InitialTime:   time.Now(),
		ErrSignalType: ErrorSignal,
		Err:           err,
	}
}

func safeSendErrSignal(ch chan<- ErrSignal, err error) {
	if err != nil {
		ch <- newErrSignal(err)
	}
}

func sendFatalSignal(ch chan<- ErrSignal, err error) {
	if err == nil {
		return
	}
	ch <- ErrSignal{
		InitialTime:   time.Now(),
		ErrSignalType: FatalSignal,
		Err:           err,
	}
	close(ch)
}

// info

func (g *Game) newSwitchStateSignal() InfoSignal {
	return InfoSignal{
		InitialTime:    time.Now(),
		InfoSignalType: SwitchStateSignal,
		Info: SwitchStateInfo{
			DayCounter:    g.nightCounter,
			PreviousState: g.previousState,
			NewState:      g.state,
		},
	}
}

func (g *Game) newSwitchVotingRoleSignal() InfoSignal {
	return InfoSignal{
		InitialTime:    time.Now(),
		InfoSignalType: SwitchVotingRoleSignal,
		Info: SwitchVotingRoleInfo{
			CurrVotingRole: g.nightVoting,
		},
	}
}

func (g *Game) newFinishGameSignal() InfoSignal {
	return InfoSignal{
		InitialTime:    time.Now(),
		InfoSignalType: FinishGameSignal,
		Info:           FinishGameInfo{},
	}
}
