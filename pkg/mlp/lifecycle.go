package mlp

// LaneMode is the lane's operating mode.
type LaneMode int

const (
	LaneModeActive LaneMode = iota
	LaneModeParked
	LaneModeDegraded
)

// MediaLaneStateKind is the lane state.
type MediaLaneStateKind int

const (
	LaneStateIdle MediaLaneStateKind = iota
	LaneStateHandshaking
	LaneStateActive
	LaneStateParked
	LaneStateClosing
	LaneStateClosed
)

// MediaLanePolicy holds lane policy.
type MediaLanePolicy struct {
	MediaLaneEnabled bool
}

func DefaultMediaLanePolicy() MediaLanePolicy {
	return MediaLanePolicy{MediaLaneEnabled: true}
}
