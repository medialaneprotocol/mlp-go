package mlp

// SessionPhase is the lane's lifecycle phase.
type SessionPhase int

const (
	PhaseIdle SessionPhase = iota
	PhaseHandshaking
	PhaseActive
	PhaseParked
	PhaseResuming
	PhaseClosing
	PhaseClosed
)

func (p SessionPhase) String() string {
	switch p {
	case PhaseIdle:
		return "idle"
	case PhaseHandshaking:
		return "handshaking"
	case PhaseActive:
		return "active"
	case PhaseParked:
		return "parked"
	case PhaseResuming:
		return "resuming"
	case PhaseClosing:
		return "closing"
	case PhaseClosed:
		return "closed"
	}
	return "unknown"
}

// StateMachine drives session transitions.
type StateMachine struct {
	Phase SessionPhase
}

func NewStateMachine() *StateMachine {
	return &StateMachine{Phase: PhaseIdle}
}

func (sm *StateMachine) Transition(to SessionPhase) bool {
	// Simplified: allow any transition (real impl has rules).
	sm.Phase = to
	return true
}

func (sm *StateMachine) IsActive() bool { return sm.Phase == PhaseActive }
func (sm *StateMachine) IsClosed() bool { return sm.Phase == PhaseClosed }
