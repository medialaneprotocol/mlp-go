package mlp

import "time"

// CompleteSender is a full-featured sender with pacing, BWE, cell tracking.
type CompleteSender struct {
	Grid        CellGrid
	Accounting  *CellAccounting
	Scheduler   *Scheduler
	Estimator   *BandwidthEstimator
	FlowCtrl    *FlowControl
	Qos         QosTier
	NextUpdate  uint64
	NextObject  uint32
	Objects     []*OutboundObject
	Started     time.Time
	Stats       SenderStats
}

type SenderStats struct {
	UpdatesEnqueued uint64
	ObjectsEnqueued uint64
	BytesSent       uint64
	FragmentsSent   uint64
}

func NewCompleteSender(grid CellGrid) *CompleteSender {
	return &CompleteSender{
		Grid:       grid,
		Accounting: NewCellAccounting(grid),
		Scheduler:  NewScheduler(grid),
		Estimator:  NewBandwidthEstimator(BweConfig{PaceFloorBps: 100000}),
		FlowCtrl:   NewFlowControl(1000000, 100000),
		Qos:        QosInteractive,
		NextUpdate: 1,
		NextObject: 1,
		Started:    time.Now(),
	}
}

func (s *CompleteSender) EnqueueUpdate(req UpdateRequest) (uint64, error) {
	uid := s.NextUpdate
	s.NextUpdate++
	s.Accounting.MarkPending(req.Cells)
	s.Stats.UpdatesEnqueued++
	// Create outbound object
	obj := &OutboundObject{
		ObjectID:    s.NextObject,
		UpdateID:    uid,
		Payload:     req.Payload,
		EchoCritical: req.EchoCritical,
		EnqueuedAt:  time.Now(),
	}
	s.NextObject++
	s.Objects = append(s.Objects, obj)
	s.Stats.ObjectsEnqueued++
	return uid, nil
}

func (s *CompleteSender) PollFragment(mtu int) (*Fragment, error) {
	for _, obj := range s.Objects {
		if obj.NextFrag < obj.FragCount || obj.FragCount == 0 {
			if obj.FragCount == 0 {
				obj.FragCount = 1
			}
			payloadMax := mtu - FragmentFixedHeader - 64
			if payloadMax < 64 {
				payloadMax = 64
			}
			chunk := obj.Payload
			if len(chunk) > payloadMax {
				off := int(obj.NextFrag) * payloadMax
				end := off + payloadMax
				if end > len(chunk) {
					end = len(chunk)
				}
				if off >= len(chunk) {
					obj.NextFrag++
					continue
				}
				chunk = chunk[off:end]
			}
			f := &Fragment{
				Header: FragmentHeader{
					Version:        MediaProtocolVersion,
					Class:          ClassMedia,
					Flags:          0,
					Priority:       PrioMedia,
					ObjectID:       obj.ObjectID,
					FragmentIndex:  obj.NextFrag,
					FragmentCount:  obj.FragCount,
					PayloadLen:     uint32(len(chunk)),
					HeaderExtLen:   0,
				},
				Ext:     NewHeaderExt(),
				Payload: append([]byte(nil), chunk...),
			}
			obj.NextFrag++
			s.Stats.FragmentsSent++
			s.Stats.BytesSent += uint64(len(chunk))
			return f, nil
		}
	}
	return nil, nil
}

func (s *CompleteSender) InFlight() int { return len(s.Objects) }
func (s *CompleteSender) QueuedObjects() int { return len(s.Objects) }

func (s *CompleteSender) NoteApplied(updateID uint64) {
	s.Accounting.MarkApplied([]uint32{uint32(updateID)})
}

func (s *CompleteSender) GetStats() SenderStats { return s.Stats }
