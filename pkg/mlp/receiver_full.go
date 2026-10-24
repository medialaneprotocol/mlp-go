package mlp

// FullMediaReceiver is a more complete receiver with assembly and apply.
type FullMediaReceiver struct {
	Canvas     *SurfaceCanvas
	Assembler  *ObjectAssembler
	Accounting *CellAccounting
	Applied    uint64
	Seen       map[uint64]bool
}

func NewFullMediaReceiver(w, h uint32) *FullMediaReceiver {
	g := CellGrid{WidthPx: w, HeightPx: h, CellSize: 64}
	return &FullMediaReceiver{
		Canvas:     NewSurfaceCanvas(w, h),
		Assembler:  NewObjectAssembler(),
		Accounting: NewCellAccounting(g),
		Seen:       map[uint64]bool{},
	}
}

func (r *FullMediaReceiver) AcceptFragment(data []byte) (*OpenObject, bool, error) {
	obj, done := r.Assembler.Accept(MustParseFragment(data))
	if done && obj != nil {
		r.Applied++
	}
	return obj, done, nil
}

func MustParseFragment(data []byte) *Fragment {
	f, _, _ := ParseFragment(data)
	return f
}

func (r *FullMediaReceiver) NoteApplied(updateID uint64) {
	r.Applied = updateID
}

func (r *FullMediaReceiver) BuildAck() []byte {
	return []byte{byte(r.Applied)}
}

func (r *FullMediaReceiver) Converged() bool {
	return r.Accounting.Converged()
}
