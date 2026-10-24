package mlp

// SurfaceCanvas is a minimal receiver-side surface.
type SurfaceCanvas struct {
	Cells []uint16
	Width  uint32
	Height uint32
}

func NewSurfaceCanvas(w, h uint32) *SurfaceCanvas {
	return &SurfaceCanvas{Cells: make([]uint16, w*h), Width: w, Height: h}
}

// MediaReceiverMachine is a minimal Go port of the receiver's accept path.
type MediaReceiverMachine struct {
	Canvas *SurfaceCanvas
	Seen   map[uint32]bool
}

func NewMediaReceiverMachine(w, h uint32) *MediaReceiverMachine {
	return &MediaReceiverMachine{Canvas: NewSurfaceCanvas(w, h), Seen: map[uint32]bool{}}
}

// AcceptFragment parses and accepts a fragment; returns true if a whole object was assembled.
func (r *MediaReceiverMachine) AcceptFragment(data []byte) (bool, error) {
	f, n, err := ParseFragment(data)
	if err != nil {
		return false, err
	}
	if n != len(data) {
		return false, ErrBadLength
	}
	// Track seen object ids (simplified: one fragment = one object).
	r.Seen[f.Header.ObjectID] = true
	return true, nil
}

// NoteApplied records that an update was applied.
func (r *MediaReceiverMachine) NoteApplied(updateID uint64) {}

// BuildAck produces a minimal ack payload placeholder.
func (r *MediaReceiverMachine) BuildAck() []byte { return []byte{0} }
