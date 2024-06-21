package game

type State int

const (
	NonDefinedState State = 1
	RegisterState   State = 2
	StartingState   State = 3
	NightState      State = 4
	DayState        State = 5
	VotingState     State = 6
	PausedState     State = 7
	FinishState     State = 8
)

func (g *Game) IsFinished() bool {
	return g.State == FinishState
}

func (g *Game) IsRunning() bool {
	return g.State == NightState || g.State == DayState || g.State == VotingState || g.State == PausedState
}

// _________________
// States functions
// _________________

func (g *Game) getNextState() State {
	g.RLock()
	defer g.RUnlock()
	switch g.State {
	case NonDefinedState:
		return RegisterState
	case RegisterState:
		return StartingState
	case StartingState:
		return NightState
	case NightState:
		return DayState
	case DayState:
		return VotingState
	case VotingState:
		return NightState
	}

	return g.PreviousState
}

func (g *Game) SetState(state State) {
	g.Lock()
	currGState := g.State
	defer g.Unlock()
	g.PreviousState = currGState
	g.State = state
}

func (g *Game) SwitchState() {
	nextState := g.getNextState()
	g.SetState(nextState)
}

func (g *Game) ChangeStateToPause() {
	g.Lock()
	defer g.Unlock()
	currGState := g.State
	g.PreviousState = currGState
	g.State = PausedState
}

// _______________
// For format
// _______________

var stateDefinition = map[State]string{
	NonDefinedState: "is full raw (nothing is known)",
	RegisterState:   "is waited for registration",
	StartingState:   "is prepared for starting",
	NightState:      "is in night state",
	DayState:        "is in day state",
	VotingState:     "is in day voting state",
	PausedState:     "is in paused state",
	FinishState:     "is finished",
}

func (s State) String() string {
	str, ok := stateDefinition[s]
	if !ok {
		return "Unknown"
	}
	return str
}