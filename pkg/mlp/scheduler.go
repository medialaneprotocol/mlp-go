package mlp

// Scheduler decides which cells to send next (greedy rectangles).
type Scheduler struct {
	Grid CellGrid
}

func NewScheduler(grid CellGrid) *Scheduler {
	return &Scheduler{Grid: grid}
}

// Plan produces a list of cell indices to send (all changed cells for now).
func (s *Scheduler) Plan(changed []uint32) []uint32 {
	out := make([]uint32, len(changed))
	copy(out, changed)
	return out
}

// MaxChunksFor returns the max parts for a carrier.
func (s *Scheduler) MaxChunksFor(datagram bool, lossRatio float64) uint16 {
	if datagram {
		return 1
	}
	return 8
}
