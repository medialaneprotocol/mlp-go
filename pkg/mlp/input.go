package mlp

// InputEvent is one viewer input event.
type InputEvent struct {
	Code       string
	Pressed    bool
	Seq        uint64
	DeadlineMs uint64 // absolute, session-relative; 0 = none
}

// HostInputState is the sender-side half of the input path: it sequences
// events, keeps the pressed set, and expires events whose deadline passed.
type HostInputState struct {
	NextSeq   uint64
	Pressed   map[string]bool
	Deadlines map[uint64]uint64
	Order     []uint64
}

func NewHostInputState() *HostInputState {
	return &HostInputState{Pressed: map[string]bool{}, Deadlines: map[uint64]uint64{}}
}

// Submit sequences one event. A release always leaves the queue: a stuck
// key is the one input error a user cannot undo.
func (m *HostInputState) Submit(code string, pressed bool, deadlineMs uint64) InputEvent {
	seq := m.NextSeq
	m.NextSeq++
	if pressed {
		m.Pressed[code] = true
	} else {
		delete(m.Pressed, code)
	}
	if deadlineMs != 0 {
		m.Deadlines[seq] = deadlineMs
	}
	m.Order = append(m.Order, seq)
	return InputEvent{Code: code, Pressed: pressed, Seq: seq, DeadlineMs: deadlineMs}
}

// IsPressed reports the pressed set's own truth.
func (m *HostInputState) IsPressed(code string) bool { return m.Pressed[code] }

// PressedCount is the number of held keys.
func (m *HostInputState) PressedCount() int { return len(m.Pressed) }

// Expired lists sequences whose deadline passed, oldest first.
func (m *HostInputState) Expired(nowMs uint64) []uint64 {
	var out []uint64
	for _, seq := range m.Order {
		if dl, ok := m.Deadlines[seq]; ok && dl != 0 && nowMs > dl {
			out = append(out, seq)
		}
	}
	return out
}
