package mlp

// BweConfig holds bandwidth estimator policy.
type BweConfig struct {
	PaceFloorBps  uint64
	RequestedFPS  float64
	ViewerLive    bool
}

// Estimate is the BWE's current estimate.
type Estimate struct {
	PaceBps       uint64
	MaxObjectBytes uint32
	LossRatio     float64
	RTTMs         float64
	Mode          string
}

// BandwidthEstimator is a minimal BWE.
type BandwidthEstimator struct {
	Config BweConfig
	RTTMs  float64
}

func NewBandwidthEstimator(cfg BweConfig) *BandwidthEstimator {
	return &BandwidthEstimator{Config: cfg, RTTMs: 50.0}
}

func (e *BandwidthEstimator) Estimate() Estimate {
	return Estimate{
		PaceBps:        e.Config.PaceFloorBps,
		MaxObjectBytes: 8192,
		LossRatio:      0.0,
		RTTMs:          e.RTTMs,
		Mode:           "live",
	}
}

func (e *BandwidthEstimator) RTTMsValue() float64 { return e.RTTMs }
func (e *BandwidthEstimator) SetRTT(ms float64)    { e.RTTMs = ms }
