package mlp

import "testing"

func TestSenderFragmentsAndStamps(t *testing.T) {
	g := CellGrid{WidthPx: 320, HeightPx: 240, CellSize: 64}
	s := NewMediaSenderMachine(g)
	payload := make([]byte, 5000)
	for i := range payload {
		payload[i] = byte(i)
	}
	outcome, err := s.EnqueueUpdate(
		UpdateRequest{Payload: payload, Cells: []uint32{0}, Anchor: true, EchoCritical: true},
		CarrierWs, 42)
	if err != nil {
		t.Fatal(err)
	}
	if outcome.Fragments != 1 {
		t.Fatalf("a 5000-byte update fits one 64 KiB carrier fragment, got %d", outcome.Fragments)
	}
	f := s.PollFragment()
	if f == nil {
		t.Fatal("the fragment must be queued")
	}
	if !f.Header.Flags.Contains(FlagAnchor) || !f.Header.Flags.Contains(FlagEchoRequested) || !f.Header.Flags.Contains(FlagFin) {
		t.Fatal("a single-fragment update carries ANCHOR, ECHO_REQUESTED and FIN together")
	}
}

func TestSenderDatagramSplitsAndFin(t *testing.T) {
	g := CellGrid{WidthPx: 320, HeightPx: 240, CellSize: 64}
	s := NewMediaSenderMachine(g)
	payload := make([]byte, 4096)
	outcome, err := s.EnqueueUpdate(UpdateRequest{Payload: payload}, CarrierDataChannel, 0)
	if err != nil {
		t.Fatal(err)
	}
	if outcome.Fragments < 2 {
		t.Fatalf("expected a split, got %d", outcome.Fragments)
	}
	last := -1
	for i := 0; i < outcome.Fragments; i++ {
		f := s.PollFragment()
		if f == nil {
			t.Fatalf("fragment %d missing", i)
		}
		b, err := f.Encode()
		if err != nil {
			t.Fatal(err)
		}
		parsed, n, err := ParseFragment(b)
		if err != nil || n != len(b) {
			t.Fatalf("frag %d: %v", i, err)
		}
		meta := UpdateMetaFromExt(parsed.Ext)
		if meta.UpdateID != outcome.UpdateID {
			t.Fatalf("frag %d: update id %d", i, meta.UpdateID)
		}
		last = i
		if i == outcome.Fragments-1 && !parsed.Header.Flags.Contains(FlagFin) {
			t.Fatal("the last fragment must carry FIN")
		}
		if i < outcome.Fragments-1 && parsed.Header.Flags.Contains(FlagFin) {
			t.Fatal("only the last fragment may carry FIN")
		}
	}
	if s.PollFragment() != nil {
		t.Fatal("queue must be empty")
	}
	_ = last
}

func TestSenderAccountingFollowsTheCorpus(t *testing.T) {
	g := CellGrid{WidthPx: 320, HeightPx: 240, CellSize: 64}
	s := NewMediaSenderMachine(g)
	outcome, _ := s.EnqueueUpdate(UpdateRequest{Payload: []byte{1}, Cells: []uint32{3, 4}}, CarrierWs, 0)
	if s.Account.PendingCount() != 2 {
		t.Fatal("pending")
	}
	s.NoteApplied(outcome.UpdateID, []uint32{3, 4})
	if !s.Account.Converged() {
		t.Fatal("converged")
	}
}

func TestCellAccounting(t *testing.T) {
	g := CellGrid{WidthPx: 320, HeightPx: 240, CellSize: 64}
	a := NewCellAccounting(g)
	a.MarkPending([]uint32{0, 1, 2})
	if a.PendingCount() != 3 {
		t.Fatal("pending")
	}
	a.MarkApplied([]uint32{0, 1})
	if a.PendingCount() != 1 || a.Converged() {
		t.Fatal("applied")
	}
	a.MarkApplied([]uint32{2})
	if !a.Converged() {
		t.Fatal("converged")
	}
}
