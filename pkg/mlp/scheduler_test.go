package mlp

import "testing"

func TestSchedulerPlan(t *testing.T) {
	g := CellGrid{WidthPx: 320, HeightPx: 240, CellSize: 64}
	s := NewScheduler(g)
	cells := []uint32{0, 1, 5, 9}
	out := s.Plan(cells)
	if len(out) != 4 {
		t.Fatal("plan")
	}
	if s.MaxChunksFor(false, 0.1) != 8 {
		t.Fatal("chunks")
	}
	if s.MaxChunksFor(true, 0.1) != 1 {
		t.Fatal("datagram chunks")
	}
}
