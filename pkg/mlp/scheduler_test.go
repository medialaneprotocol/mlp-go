package mlp

import "testing"

func TestMaxPartsFollowsTheCarrier(t *testing.T) {
	s := NewScheduler(CellGrid{WidthPx: 320, HeightPx: 240, CellSize: 64})
	if got := s.MaxPartsFor(true, 0.0); got != 1 {
		t.Fatalf("datagram carrier: %d", got)
	}
	if got := s.MaxPartsFor(false, 0.0); got != 8 {
		t.Fatalf("clean stream carrier: %d", got)
	}
	if got := s.MaxPartsFor(false, 0.1); got != 2 {
		t.Fatalf("lossy stream carrier: %d", got)
	}
}
