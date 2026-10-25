package mlp

// AckWindowBits is the ack bitmap's width: the ack names the highest
// applied update and the 63 updates below it.
const AckWindowBits = 64

// MediaAckWindow is the 64-update window an ack carries. Bit 0 names the
// highest applied update; bit i names the update i below it. Update id 0 is
// not an update id: the sender's ids start at 1.
type MediaAckWindow struct {
	highest uint64
	bits    uint64
}

// HighestAppliedUpdateId is the newest update the window names.
func (w *MediaAckWindow) HighestAppliedUpdateId() uint64 { return w.highest }

// Bitmap is the window's 64 bits, exactly.
func (w *MediaAckWindow) Bitmap() uint64 { return w.bits }

// Reset empties the window.
func (w *MediaAckWindow) Reset() {
	w.highest = 0
	w.bits = 0
}

// Has reports whether an update inside the window is marked applied.
func (w *MediaAckWindow) Has(updateID uint64) bool {
	if updateID < 1 || updateID > w.highest {
		return false
	}
	distance := w.highest - updateID
	if distance >= AckWindowBits {
		return false
	}
	return w.bits&(1<<distance) != 0
}

// Mark declares an update applied. A newer id slides the window; a jump of
// 64 ids or more leaves the window holding the new update alone.
func (w *MediaAckWindow) Mark(updateID uint64) {
	if updateID > w.highest {
		shift := updateID - w.highest
		if shift >= AckWindowBits {
			w.bits = 1
		} else {
			w.bits = (w.bits<<shift | 1) & (AckWindowBitsMask())
		}
		w.highest = updateID
		return
	}
	distance := w.highest - updateID
	if distance < AckWindowBits {
		w.bits |= 1 << distance
	}
}

// SeedFromResume seeds the window from a resume ack: the highest applied
// update is known, nothing below it is.
func (w *MediaAckWindow) SeedFromResume(highestApplied uint64) {
	w.Reset()
	w.highest = highestApplied
	if w.highest > 0 {
		w.bits = 1
	}
}

func AckWindowBitsMask() uint64 { return ^uint64(0) }

// ReceiverStats counts what the receiver did.
type ReceiverStats struct {
	FragmentsAccepted uint64
	FragmentsRefused  uint64
	ObjectsAssembled  uint64
	ObjectsReplayed   uint64
	UpdatesApplied    uint64
}

// MediaReceiverMachine is the Go receiver's accept path: parse, dedup,
// assemble, and hold the ack window over the applied update ids.
type MediaReceiverMachine struct {
	Assembler *ObjectAssembler
	Ack       MediaAckWindow
	Stats     ReceiverStats
}

func NewMediaReceiverMachine() *MediaReceiverMachine {
	return &MediaReceiverMachine{Assembler: NewObjectAssembler()}
}

// AcceptFragment takes one encoded fragment. A whole object is returned
// exactly once, on the fragment that completed it.
func (r *MediaReceiverMachine) AcceptFragment(data []byte) (*OpenObject, error) {
	f, n, err := ParseFragment(data)
	if err != nil {
		r.Stats.FragmentsRefused++
		return nil, err
	}
	if n != len(data) {
		r.Stats.FragmentsRefused++
		return nil, errBadLength
	}
	r.Stats.FragmentsAccepted++
	obj, whole, err := r.Assembler.Accept(f)
	if err != nil {
		r.Stats.FragmentsRefused++
		return nil, err
	}
	if whole {
		r.Stats.ObjectsAssembled++
	}
	return obj, nil
}

// NoteApplied marks an update applied: it takes its bit in the ack window.
func (r *MediaReceiverMachine) NoteApplied(updateID uint64) {
	r.Stats.UpdatesApplied++
	r.Ack.Mark(updateID)
}
