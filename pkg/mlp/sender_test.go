package mlp

import (
	"testing"
)

func TestSenderEnqueue(t *testing.T) {
	g := CellGrid{WidthPx: 320, HeightPx: 240, CellSize: 64}
	s := NewMediaSenderMachine(g)
	frags, err := s.EnqueueUpdate(UpdateRequest{Payload: []byte{1, 2, 3, 4, 5}}, 1200)
	if err != nil {
		t.Fatal(err)
	}
	if len(frags) == 0 {
		t.Fatal("no fragments")
	}
	// re-parse each fragment
	for i, f := range frags {
		b, err := f.Encode()
		if err != nil {
			t.Fatal(err)
		}
		f2, n, err := ParseFragment(b)
		if err != nil {
			t.Fatalf("frag %d: %v", i, err)
		}
		if n != len(b) {
			t.Fatalf("frag %d: consumed %d != %d", i, n, len(b))
		}
		_ = f2
	}
}
