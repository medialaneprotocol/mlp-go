package mlp

// LaneLifecycle tracks the lane's state.
type LaneLifecycle struct {
	State        MediaLaneStateKind
	Parked       bool
	Converged    bool
	PendingFrac  float64
	LastProgress uint64
	StaleBudget  uint64
	NoFeedback   bool
	LossRatio    float64
	LossDegraded bool
	Disabled     bool
}

func NewLaneLifecycle() *LaneLifecycle {
	return &LaneLifecycle{State: LaneStateIdle, StaleBudget: 3000}
}

func (l *LaneLifecycle) SetActive() { l.State = LaneStateActive }
func (l *LaneLifecycle) SetParked() { l.State = LaneStateParked; l.Parked = true }
func (l *LaneLifecycle) IsParked() bool { return l.Parked }
func (l *LaneLifecycle) SetConverged(v bool) { l.Converged = v }
func (l *LaneLifecycle) SetLossRatio(r float64) {
	l.LossRatio = r
	l.LossDegraded = r > 0.05
}
