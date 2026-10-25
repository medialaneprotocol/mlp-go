package mlp

// UpdateRequest is one media update the sender is asked to carry. Cells are
// the update's covered repair cells; the accounting advances on them.
type UpdateRequest struct {
	Payload      []byte
	Cells        []uint32
	Anchor       bool
	EchoCritical bool
}

// EnqueueOutcome says what the sender did with an update.
type EnqueueOutcome struct {
	UpdateID  uint64
	Fragments int
}

// SenderStats counts what left the sender.
type SenderStats struct {
	UpdatesEnqueued uint64
	FragmentsSent   uint64
	BytesSent       uint64
}

// MediaSenderMachine is the Go sender's enqueue-and-fragment path: it cuts
// an update into parts of at most the carrier's fragment payload, stamps
// every part with the update's metadata, and marks the last part FIN. The
// pacer, the estimator and the repair plans are the Rust machine's; this
// one is the byte-level half a Go peer needs to speak first.
type MediaSenderMachine struct {
	Grid      CellGrid
	Account   *CellAccounting
	NextObjID uint32
	NextUpdID uint64
	Queue     [][]byte // encoded fragments waiting to leave
	Stats     SenderStats
}

func NewMediaSenderMachine(grid CellGrid) *MediaSenderMachine {
	return &MediaSenderMachine{
		Grid:      grid,
		Account:   NewCellAccounting(grid),
		NextObjID: 1,
		NextUpdID: 1,
	}
}

// EnqueueUpdate fragments the update for a carrier and queues the parts.
// An update larger than the carrier can ever carry is refused, not split
// past the wire's own cap.
func (s *MediaSenderMachine) EnqueueUpdate(req UpdateRequest, carrier Carrier, nowMs uint64) (EnqueueOutcome, error) {
	payloadMax := carrier.FragmentPayloadMax() - 64 // reserve for the metadata extension
	if payloadMax < 64 {
		payloadMax = 64
	}
	if len(req.Payload) > payloadMax*int(MaxFragmentCount) {
		return EnqueueOutcome{}, errBadLength
	}

	updateID := s.NextUpdID
	s.NextUpdID++
	objectID := s.NextObjID
	s.NextObjID++
	s.Account.MarkPending(req.Cells)

	total := (len(req.Payload) + payloadMax - 1) / payloadMax
	if total == 0 {
		total = 1
	}
	flags := FlagNone
	if req.Anchor {
		flags |= FlagAnchor
	}
	if req.EchoCritical {
		flags |= FlagEchoRequested
	}

	meta := &UpdateMeta{
		UpdateID:    updateID,
		UpdateBase:  0,
		PartIndex:   0,
		PartCount:   1,
		UpdateBytes: uint32(len(req.Payload)),
		ChainEpoch:  1,
		SendEpochMs: nowMs,
	}

	for idx := 0; idx < total; idx++ {
		off := idx * payloadMax
		end := off + payloadMax
		if end > len(req.Payload) {
			end = len(req.Payload)
		}
		chunk := req.Payload[off:end]
		partFlags := flags
		if idx == total-1 {
			partFlags |= FlagFin
		}
		meta.PartIndex = uint16(idx)
		meta.PartCount = 1 // one part per fragment: each fragment is a whole part
		ext := meta.ToExt()
		f := &Fragment{
			Header: FragmentHeader{
				Version:       MediaProtocolVersion,
				Class:         ClassMedia,
				Flags:         partFlags,
				Priority:      PrioMedia,
				ObjectID:      objectID,
				FragmentIndex: uint16(idx),
				FragmentCount: uint16(total),
				PayloadLen:    uint32(len(chunk)),
				HeaderExtLen:  uint16(ext.EncodedLen()),
			},
			Ext:     ext,
			Payload: append([]byte(nil), chunk...),
		}
		encoded, err := f.Encode()
		if err != nil {
			return EnqueueOutcome{}, err
		}
		s.Queue = append(s.Queue, encoded)
		s.Stats.FragmentsSent++
		s.Stats.BytesSent += uint64(len(chunk))
	}
	s.Stats.UpdatesEnqueued++
	return EnqueueOutcome{UpdateID: updateID, Fragments: total}, nil
}

// PollFragment returns the next queued fragment, or nil when the queue is
// empty.
func (s *MediaSenderMachine) PollFragment() *Fragment {
	if len(s.Queue) == 0 {
		return nil
	}
	raw := s.Queue[0]
	s.Queue = s.Queue[1:]
	f, _, err := ParseFragment(raw)
	if err != nil {
		return nil
	}
	return f
}

// NoteApplied settles the update's cells once the receiver applied it.
func (s *MediaSenderMachine) NoteApplied(updateID uint64, cells []uint32) {
	s.Account.MarkApplied(cells)
}

var errBadLength = assemblyError("bad length")
