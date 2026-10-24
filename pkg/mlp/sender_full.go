package mlp

import "time"

// FullMediaSender is a more complete sender with pacing and cell tracking.
type FullMediaSender struct {
	Grid        CellGrid
	Accounting  *CellAccounting
	Scheduler   *Scheduler
	Estimator   *BandwidthEstimator
	NextUpdate  uint64
	NextObject  uint32
	Objects     []*OutboundObject
	PendingFrags [][]byte
}

type OutboundObject struct {
	ObjectID    uint32
	UpdateID    uint64
	NextFrag    uint16
	FragCount   uint16
	Payload     []byte
	EchoCritical bool
	EnqueuedAt  time.Time
}

func NewFullMediaSender(grid CellGrid) *FullMediaSender {
	return &FullMediaSender{
		Grid:       grid,
		Accounting: NewCellAccounting(grid),
		Scheduler:  NewScheduler(grid),
		Estimator:  NewBandwidthEstimator(BweConfig{PaceFloorBps: 100000}),
		NextUpdate: 1,
		NextObject: 1,
	}
}

func (s *FullMediaSender) EnqueueUpdate(req UpdateRequest) (uint64, error) {
	uid := s.NextUpdate
	s.NextUpdate++
	cells := req.Cells
	s.Accounting.MarkPending(cells)
	return uid, nil
}

func (s *FullMediaSender) PollFragment(mtu int) (*Fragment, error) {
	if len(s.PendingFrags) == 0 {
		return nil, nil
	}
	f := s.PendingFrags[0]
	s.PendingFrags = s.PendingFrags[1:]
	parsed, _, err := ParseFragment(f)
	return parsed, err
}

func (s *FullMediaSender) Stats() (updates, bytes uint64) {
	return s.NextUpdate - 1, 0
}
