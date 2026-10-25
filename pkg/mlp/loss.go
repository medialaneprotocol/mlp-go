package mlp

// LossDetector detects lost packets based on gaps in acks.
type LossDetector struct {
	LargestAcked *uint64
	Sent         map[uint64]bool
	Acked        map[uint64]bool
	Lost         map[uint64]bool
	Threshold    uint64
}

func NewLossDetector() *LossDetector {
	return &LossDetector{Sent: map[uint64]bool{}, Acked: map[uint64]bool{}, Lost: map[uint64]bool{}, Threshold: 3}
}

func (ld *LossDetector) OnSend(pn uint64) { ld.Sent[pn] = true }

func (ld *LossDetector) OnAck(acked []uint64) {
	for _, pn := range acked {
		ld.Acked[pn] = true
		if ld.LargestAcked == nil || pn > *ld.LargestAcked {
			v := pn
			ld.LargestAcked = &v
		}
	}
	ld.detect()
}

func (ld *LossDetector) detect() {
	if ld.LargestAcked == nil {
		return
	}
	largest := *ld.LargestAcked
	for pn := range ld.Sent {
		if pn >= largest {
			continue
		}
		if ld.Acked[pn] || ld.Lost[pn] {
			continue
		}
		if largest-pn >= ld.Threshold {
			ld.Lost[pn] = true
		}
	}
}

func (ld *LossDetector) IsLost(pn uint64) bool { return ld.Lost[pn] }
func (ld *LossDetector) LostCount() int        { return len(ld.Lost) }
