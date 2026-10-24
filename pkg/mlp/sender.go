package mlp

// CellGrid describes a media surface geometry.
type CellGrid struct {
	WidthPx  uint32
	HeightPx uint32
	CellSize uint32
}

func (g CellGrid) Cells() uint32 {
	cx := (g.WidthPx + g.CellSize - 1) / g.CellSize
	cy := (g.HeightPx + g.CellSize - 1) / g.CellSize
	return cx * cy
}

// UpdateRequest is a sender-side media update.
type UpdateRequest struct {
	Payload []byte
	Cells   []uint32
	Anchor  bool
}

// MediaSenderMachine is a minimal Go port of the Rust sender's enqueue path.
// It produces fragments for one update; pacing/BWE/resume are not yet ported.
type MediaSenderMachine struct {
	Grid      CellGrid
	NextID    uint64
	Objects   [][]byte // assembled fragment payloads (one per part)
}

func NewMediaSenderMachine(grid CellGrid) *MediaSenderMachine {
	return &MediaSenderMachine{Grid: grid, NextID: 1}
}

// EnqueueUpdate builds fragments for a delta update and returns them.
func (s *MediaSenderMachine) EnqueueUpdate(req UpdateRequest, mtu int) ([]*Fragment, error) {
	if mtu <= FragmentFixedHeader {
		return nil, ErrBadLength
	}
	payloadMax := mtu - FragmentFixedHeader - 64 // reserve for ext
	if payloadMax < 64 {
		payloadMax = 64
	}
	// split payload into one part (simplified: single part)
	updateID := s.NextID
	s.NextID++

	meta := &UpdateMeta{
		UpdateID:    updateID,
		UpdateBase:  0,
		PartIndex:   0,
		PartCount:   1,
		UpdateBytes: uint32(len(req.Payload)),
		ChainEpoch:  1,
		SendEpochMs: 0,
	}
	ext := meta.ToExt()

	// fragment the payload
	var frags []*Fragment
	off := 0
	idx := uint16(0)
	total := (len(req.Payload) + payloadMax - 1) / payloadMax
	if total == 0 {
		total = 1
	}
	for off < len(req.Payload) || (len(req.Payload) == 0 && idx == 0) {
		chunk := req.Payload[off:]
		if len(chunk) > payloadMax {
			chunk = chunk[:payloadMax]
		}
		f := &Fragment{
			Header: FragmentHeader{
				Version:        MediaProtocolVersion,
				Class:          ClassMedia,
				Flags:          0,
				Priority:       PrioMedia,
				ObjectID:       uint32(updateID),
				FragmentIndex:  idx,
				FragmentCount:  uint16(total),
				PayloadLen:     uint32(len(chunk)),
				HeaderExtLen:   uint16(ext.EncodedLen()),
			},
			Ext:     ext,
			Payload: append([]byte(nil), chunk...),
		}
		frags = append(frags, f)
		off += len(chunk)
		idx++
		if len(req.Payload) == 0 {
			break
		}
	}
	if req.Anchor && len(frags) > 0 {
		frags[0].Header.Flags |= FlagAnchorBit
	}
	return frags, nil
}

var ErrBadLength = &errBadLength{}

type errBadLength struct{}

func (e *errBadLength) Error() string { return "bad length" }

// CellAccounting tracks per-receiver convergence over a repair-cell grid.
type CellAccounting struct {
	Grid     CellGrid
	Pending  map[uint32]bool
	Applied  map[uint32]bool
}

func NewCellAccounting(grid CellGrid) *CellAccounting {
	return &CellAccounting{Grid: grid, Pending: map[uint32]bool{}, Applied: map[uint32]bool{}}
}

func (a *CellAccounting) MarkPending(cells []uint32) {
	for _, c := range cells {
		a.Pending[c] = true
		delete(a.Applied, c)
	}
}

func (a *CellAccounting) MarkApplied(cells []uint32) {
	for _, c := range cells {
		delete(a.Pending, c)
		a.Applied[c] = true
	}
}

func (a *CellAccounting) PendingCount() int { return len(a.Pending) }
func (a *CellAccounting) Converged() bool    { return len(a.Pending) == 0 }
