package mlp

import (
	"testing"
	"time"
)

func TestPacer(t *testing.T) {
	cfg := DefaultPacerConfig()
	p := NewMediaPacer(cfg, time.Now())
	if p.WindowBudget() == 0 {
		t.Fatal("budget")
	}
	if !p.TryAcquire(6, 100, time.Now()) {
		t.Fatal("acquire")
	}
	if p.PaceBps() != 8000 {
		t.Fatal("pace")
	}
}
