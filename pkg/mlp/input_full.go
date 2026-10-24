package mlp

// InputManager handles input sequencing, deadlines, and pressed set.
type InputManager struct {
	NextSeq     uint64
	Pressed     map[string]bool
	Deadlines   map[uint64]uint64 // seq -> deadline_ms
	Submitted   []uint64
}

func NewInputManager() *InputManager {
	return &InputManager{Pressed: map[string]bool{}, Deadlines: map[uint64]uint64{}}
}

func (m *InputManager) Submit(code string, pressed bool, deadlineMs uint64) uint64 {
	seq := m.NextSeq
	m.NextSeq++
	if pressed {
		m.Pressed[code] = true
	} else {
		delete(m.Pressed, code)
	}
	m.Deadlines[seq] = deadlineMs
	m.Submitted = append(m.Submitted, seq)
	return seq
}

func (m *InputManager) IsPressed(code string) bool {
	return m.Pressed[code]
}

func (m *InputManager) PressedCount() int { return len(m.Pressed) }

func (m *InputManager) Expired(nowMs uint64) []uint64 {
	var out []uint64
	for seq, dl := range m.Deadlines {
		if dl != 0 && nowMs > dl {
			out = append(out, seq)
		}
	}
	return out
}
