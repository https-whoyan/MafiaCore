package game

type State int8

const (
	_ State = iota
	NonDefinedState
	RegisterState
	InitState
	StartingState
	NightState
	DayState
	FinishState
)

func (g *Game) IsFinished() bool {
	return g.state == FinishState
}

func (g *Game) IsRunning() bool {
	return g.state == NightState || g.state == DayState
}

// _________________
// States functions
// _________________

func (g *Game) getNextState() State {
	g.RLock()
	defer g.RUnlock()
	switch g.state {
	case NonDefinedState:
		return RegisterState
	case RegisterState:
		return InitState
	case InitState:
		return StartingState
	case StartingState:
		return NightState
	case NightState:
		return DayState
	case DayState:
		return NightState
	default:
		panic("unknown game state")
	}

	return g.previousState
}

func (g *Game) SetState(state State) {
	g.Lock()
	defer g.Unlock()
	currGState := g.state
	g.previousState = currGState
	g.state = state
}

func (g *Game) SwitchState() {
	nextState := g.getNextState()
	g.SetState(nextState)
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
	FinishState:     "is finished",
}

func (s State) String() string {
	str, ok := stateDefinition[s]
	if !ok {
		return "Unknown"
	}
	return str
}
