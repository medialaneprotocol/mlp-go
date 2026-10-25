package mlp

// The scheduler's own vocabulary is the Rust machine's (priority bands,
// capture cycle, anchor plans); what a Go peer needs from it at the wire
// level is the part-count rule: how many parts a carrier may split an
// update into before the loss of one part costs the whole update.
type Scheduler struct {
	Grid CellGrid
}

func NewScheduler(grid CellGrid) *Scheduler {
	return &Scheduler{Grid: grid}
}

// MaxPartsFor is the largest part count worth sending on a carrier: a
// datagram carrier delivers a part or drops it, so one part per update; a
// stream carrier can pipeline, and eight parts bound the reassembly cost.
func (s *Scheduler) MaxPartsFor(datagram bool, lossRatio float64) uint16 {
	if datagram {
		return 1
	}
	if lossRatio > 0.05 {
		return 2
	}
	return 8
}
