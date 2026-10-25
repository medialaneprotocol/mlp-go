package mlp

import "time"

// PacerConfig holds pacer configuration.
type PacerConfig struct {
	PaceBps             uint64
	MaxObjectBytes      uint32
	ReservedFloorBps    uint64
	ReservedFraction    float64
	Window              time.Duration
	BurstAllowanceRatio float64
	MaxGrantBytes       uint64
}

func DefaultPacerConfig() PacerConfig {
	return PacerConfig{
		PaceBps:             8000,
		MaxObjectBytes:      8192,
		ReservedFloorBps:    2000,
		ReservedFraction:    0.10,
		Window:              time.Second,
		BurstAllowanceRatio: 1.25,
		MaxGrantBytes:       16384,
	}
}

// TokenBucket is a deterministic token bucket.
type TokenBucket struct {
	Capacity     float64
	RefillPerSec float64
	Tokens       float64
	LastRefill   time.Time
}

func NewTokenBucket(capacity, refillPerSec float64, now time.Time) *TokenBucket {
	return &TokenBucket{Capacity: capacity, RefillPerSec: refillPerSec, Tokens: capacity, LastRefill: now}
}

func (b *TokenBucket) Refill(now time.Time) {
	elapsed := now.Sub(b.LastRefill).Seconds()
	b.Tokens = min(b.Capacity, b.Tokens+elapsed*b.RefillPerSec)
	b.LastRefill = now
}

func (b *TokenBucket) TryConsume(amount float64) bool {
	if b.Tokens >= amount {
		b.Tokens -= amount
		return true
	}
	return false
}

func (b *TokenBucket) TokensAvailable() float64 { return b.Tokens }

// MediaPacer wraps a token bucket with a window.
type MediaPacer struct {
	Config PacerConfig
	Bucket *TokenBucket
	Window []struct {
		At    time.Time
		Bytes uint64
	}
}

func NewMediaPacer(cfg PacerConfig, now time.Time) *MediaPacer {
	cap := float64(cfg.PaceBps) / 8.0 * cfg.Window.Seconds() * cfg.BurstAllowanceRatio
	return &MediaPacer{
		Config: cfg,
		Bucket: NewTokenBucket(cap, float64(cfg.PaceBps)/8.0, now),
	}
}

func (p *MediaPacer) WindowBudget() uint64 {
	rate := float64(p.Config.PaceBps) / 8.0
	allowance := rate * p.Config.Window.Seconds() * p.Config.BurstAllowanceRatio
	budget := uint64(allowance)
	if budget < p.Config.MaxGrantBytes {
		budget = p.Config.MaxGrantBytes
	}
	return budget
}

func (p *MediaPacer) WindowUsed(now time.Time) uint64 {
	cutoff := now.Add(-p.Config.Window)
	var total uint64
	for _, w := range p.Window {
		if w.At.After(cutoff) {
			total += w.Bytes
		}
	}
	return total
}

func (p *MediaPacer) TryAcquire(class int, bytes uint64, now time.Time) bool {
	p.Bucket.Refill(now)
	budget := p.WindowBudget()
	used := p.WindowUsed(now)
	if used+bytes > budget {
		return false
	}
	watermark := uint64(0)
	if class >= 5 { // media
		watermark = uint64(float64(p.Config.PaceBps) / 8.0 * 0.1)
	}
	if p.Bucket.TokensAvailable() < float64(bytes)+float64(watermark) {
		return false
	}
	p.Bucket.TryConsume(float64(bytes))
	p.Window = append(p.Window, struct {
		At    time.Time
		Bytes uint64
	}{At: now, Bytes: bytes})
	return true
}

func (p *MediaPacer) PaceBps() uint64       { return p.Config.PaceBps }
func (p *MediaPacer) MaxGrantBytes() uint64 { return p.Config.MaxGrantBytes }

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
