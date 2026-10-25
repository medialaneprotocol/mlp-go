package mlp

import "time"

type ProbeResult struct {
	PathID      uint8
	RTTMs       float64
	LossRatio   float64
	GoodputBps  uint64
	CompletedAt time.Time
}

type PathProber struct {
	IntervalMs    uint64
	ProbesPerPath uint32
	Results       []ProbeResult
	LastProbe     *time.Time
}

func NewPathProber(intervalMs uint64, probesPerPath uint32) *PathProber {
	return &PathProber{IntervalMs: intervalMs, ProbesPerPath: probesPerPath}
}

func (p *PathProber) ShouldProbe(now time.Time) bool {
	if p.LastProbe == nil {
		return true
	}
	return now.Sub(*p.LastProbe).Milliseconds() >= int64(p.IntervalMs)
}

func (p *PathProber) Record(r ProbeResult) {
	p.Results = append(p.Results, r)
	t := r.CompletedAt
	p.LastProbe = &t
}

func (p *PathProber) BestPath() *uint8 {
	if len(p.Results) == 0 {
		return nil
	}
	best := p.Results[0]
	for _, r := range p.Results[1:] {
		if r.RTTMs < best.RTTMs || (r.RTTMs == best.RTTMs && r.LossRatio < best.LossRatio) {
			best = r
		}
	}
	id := best.PathID
	return &id
}

func (p *PathProber) PathResults(pathID uint8) []ProbeResult {
	var out []ProbeResult
	for _, r := range p.Results {
		if r.PathID == pathID {
			out = append(out, r)
		}
	}
	return out
}
