package mlp

import "testing"

func TestLossDetector(t *testing.T) {
	ld := NewLossDetector()
	for i := uint64(0); i < 10; i++ {
		ld.OnSend(i)
	}
	ld.OnAck([]uint64{0, 1, 2, 4, 5, 6, 7, 8, 9})
	if !ld.IsLost(3) {
		t.Fatal("3 should be lost")
	}
	if ld.LostCount() != 1 {
		t.Fatalf("lost count %d", ld.LostCount())
	}
}
