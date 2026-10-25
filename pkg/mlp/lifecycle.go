package mlp

// MediaLaneStateKind is the lane's state, as it travels on the wire
// (MLP spec §7). The receiver's UI maps these values to the honest
// strings of the specification.
type MediaLaneStateKind uint32

const (
	LaneStateStarting         MediaLaneStateKind = 0 // waiting for MediaHello/MediaHelloAck
	LaneStateLive             MediaLaneStateKind = 1 // healthy: converged or converging quickly
	LaneStateDegraded         MediaLaneStateKind = 2 // delivering, below the link's capability
	LaneStateConverging       MediaLaneStateKind = 3 // draining a plan or a large pending set
	LaneStateStale            MediaLaneStateKind = 4 // no applied update for a long time
	LaneStateParked           MediaLaneStateKind = 5 // the peer is away, the state is held
	LaneStateDisabled         MediaLaneStateKind = 6 // media off by policy or kill switch
	LaneStateDecodeBound      MediaLaneStateKind = 7 // the receiver is decode-bound
	LaneStateNoFeedback       MediaLaneStateKind = 8 // no feedback from the receiver
	LaneStateMediaUnsupported MediaLaneStateKind = 9 // the peer does not speak MLP v1
)

// LaneStateFromU32 decodes the wire value; unknown values are an error.
func LaneStateFromU32(v uint32) (MediaLaneStateKind, error) {
	if v > 9 {
		return 0, errMetaMismatch
	}
	return MediaLaneStateKind(v), nil
}

// String is the stable label spelling.
func (s MediaLaneStateKind) String() string {
	names := [...]string{
		"starting", "live", "degraded", "converging", "stale",
		"parked", "disabled", "decode_bound", "no_feedback", "media_unsupported",
	}
	if int(s) < len(names) {
		return names[s]
	}
	return "unknown"
}

// MediaCloseCode is a close's reason, as it travels on the wire.
type MediaCloseCode uint32

const (
	CloseNormal         MediaCloseCode = 0
	CloseServerShutdown MediaCloseCode = 1
	ClosePolicy         MediaCloseCode = 2
	CloseSuperseded     MediaCloseCode = 3
	ClosePark           MediaCloseCode = 4 // park: the "come back" signal
)

// CloseCodeFromU32 decodes the wire value.
func CloseCodeFromU32(v uint32) (MediaCloseCode, error) {
	if v > 4 {
		return 0, errMetaMismatch
	}
	return MediaCloseCode(v), nil
}

// MediaUnsupportedReason is why a peer cannot serve the lane.
type MediaUnsupportedReason string

const (
	UnsupportedProtocol           MediaUnsupportedReason = "protocol"
	UnsupportedDisabledByPolicy   MediaUnsupportedReason = "disabled_by_policy"
	UnsupportedNoSenderCapability MediaUnsupportedReason = "no_sender_capability"
)

// LaneMode is the lane's operating mode.
type LaneMode int

const (
	LaneModeSingle LaneMode = iota // both lanes share the control connection
	LaneModeSplit                  // the media lane is its own connection
)

// LaneLifecycle holds the evidence a state is decided on. A state is a
// fact both ends can measure — freshness, convergence, loss — and neither
// end's claim overrides the other's staleness.
type LaneLifecycle struct {
	State         MediaLaneStateKind
	SinceMs       uint64
	Converged     bool
	PendingCells  uint64
	LastAppliedMs uint64
	StaleBudgetMs uint64
	NoFeedback    bool
	LossRatio     float64
}

func NewLaneLifecycle() *LaneLifecycle {
	return &LaneLifecycle{State: LaneStateStarting, StaleBudgetMs: 3000}
}

// SetApplied records an applied update: progress is the one thing that
// clears stale.
func (l *LaneLifecycle) SetApplied(nowMs uint64, pending uint64) {
	l.LastAppliedMs = nowMs
	l.PendingCells = pending
	l.Converged = pending == 0
	l.reconsider(nowMs)
}

// SetLoss records the measured loss ratio.
func (l *LaneLifecycle) SetLoss(ratio float64) {
	l.LossRatio = ratio
}

func (l *LaneLifecycle) reconsider(nowMs uint64) {
	if l.State == LaneStateParked || l.State == LaneStateDisabled ||
		l.State == LaneStateMediaUnsupported {
		return
	}
	switch {
	case l.NoFeedback:
		l.State = LaneStateNoFeedback
	case l.Converged:
		l.State = LaneStateLive
	case l.LastAppliedMs > 0 && nowMs-l.LastAppliedMs > l.StaleBudgetMs:
		l.State = LaneStateStale
	case l.PendingCells > 0:
		l.State = LaneStateConverging
	default:
		l.State = LaneStateLive
	}
	l.SinceMs = nowMs
}
