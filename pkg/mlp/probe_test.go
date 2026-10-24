package mlp

import (
	"testing"
	"time"
)

func TestProbe(t *testing.T) {
	pp := NewPathProber(1000, 3)
	now := time.Now()
	pp.Record(ProbeResult{PathID: 0, RTTMs: 100, LossRatio: 0.05, GoodputBps: 1000000, CompletedAt: now})
	pp.Record(ProbeResult{PathID: 1, RTTMs: 50, LossRatio: 0.01, GoodputBps: 2000000, CompletedAt: now})
	bp := pp.BestPath()
	if bp == nil || *bp != 1 {
		t.Fatal("best path")
	}
	if !pp.ShouldProbe(now) {
		// initially should probe
	}
}
