package mlp

// The park table's TTLs (MLP spec §9.2): a parked session is held for at
// least the floor and at most the ceiling, 300 seconds by default.
const (
	ResumeTTLSeconds    uint64 = 900
	ParkTTLFloorSeconds uint64 = 120
	ParkTTLCeilSeconds  uint64 = 900
	ParkTTLDefault      uint64 = 300
)

// ResumeRejectReason is why a resume claim was refused.
type ResumeRejectReason string

const (
	ResumeAccepted ResumeRejectReason = "accepted"
	ResumeNoSlot   ResumeRejectReason = "no_slot" // the table holds no such session
	ResumeExpired  ResumeRejectReason = "expired" // the TTL passed; the state is gone
	ResumeRebased  ResumeRejectReason = "rebased" // the sender broke the chain; the receiver must reset
)

// ParkedSession is the state a park holds: enough to resume the stream
// without replaying it.
type ParkedSession struct {
	SessionID     uint64
	LastAppliedID uint64
	ChainEpoch    uint32
	PendingCells  []uint32
	ParkedAtMs    uint64
	TTLSeconds    uint64
}

// ParkTable holds parked sessions under their TTLs.
type ParkTable struct {
	Sessions map[uint64]*ParkedSession
	NowMs    func() uint64
}

func NewParkTable(now func() uint64) *ParkTable {
	if now == nil {
		now = func() uint64 { return 0 }
	}
	return &ParkTable{Sessions: map[uint64]*ParkedSession{}, NowMs: now}
}

// Park holds a session for its TTL (clamped to the floor and ceiling).
func (t *ParkTable) Park(s ParkedSession) {
	if s.TTLSeconds < ParkTTLFloorSeconds {
		s.TTLSeconds = ParkTTLFloorSeconds
	}
	if s.TTLSeconds > ParkTTLCeilSeconds {
		s.TTLSeconds = ParkTTLCeilSeconds
	}
	if s.TTLSeconds == 0 {
		s.TTLSeconds = ParkTTLDefault
	}
	s.ParkedAtMs = t.NowMs()
	t.Sessions[s.SessionID] = &s
}

// Claim takes the session out of the table, or names why it cannot.
func (t *ParkTable) Claim(sessionID uint64) (*ParkedSession, ResumeRejectReason) {
	s, ok := t.Sessions[sessionID]
	if !ok {
		return nil, ResumeNoSlot
	}
	delete(t.Sessions, sessionID)
	if t.NowMs()-s.ParkedAtMs > s.TTLSeconds*1000 {
		return nil, ResumeExpired
	}
	return s, ResumeAccepted
}
