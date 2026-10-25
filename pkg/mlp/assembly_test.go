package mlp

import "testing"

func makePart(oid uint32, idx, count uint16, fin bool, payload []byte) *Fragment {
	flags := FlagNone
	if fin {
		flags |= FlagFin
	}
	return &Fragment{
		Header: FragmentHeader{
			ObjectID: oid, FragmentIndex: idx, FragmentCount: count,
			Flags: flags, PayloadLen: uint32(len(payload)),
		},
		Ext:     NewHeaderExt(),
		Payload: payload,
	}
}

func TestAssemblyWaitsForFin(t *testing.T) {
	a := NewObjectAssembler()
	if _, done, _ := a.Accept(makePart(1, 0, 2, false, []byte{1, 2})); done {
		t.Fatal("not done: no FIN yet")
	}
	if _, done, _ := a.Accept(makePart(1, 1, 2, false, []byte{3, 4})); done {
		t.Fatal("not done: every fragment held, still no FIN")
	}
	obj, done, _ := a.Accept(makePart(1, 1, 2, true, []byte{3, 4}))
	if done {
		t.Fatal("duplicate fragment cannot complete the object")
	}
	// The FIN arrived on a fragment already held: the object needs the FIN
	// bit on a *new* fragment of the same object, which the wire never
	// sends twice — so exercise the real shape: re-send index 1 with FIN is
	// a duplicate; instead finish with a re-declared count? No: the wire's
	// own shape is FIN on the last fragment. Rebuild with that shape.
	a = NewObjectAssembler()
	a.Accept(makePart(1, 0, 2, false, []byte{1, 2}))
	obj, done, _ = a.Accept(makePart(1, 1, 2, true, []byte{3, 4}))
	if !done || obj == nil {
		t.Fatal("FIN on the last fragment completes the object")
	}
	if got := obj.Payload(); string(got) != "\x01\x02\x03\x04" {
		t.Fatalf("payload %v", got)
	}
	if a.OpenCount() != 0 {
		t.Fatal("open count")
	}
}

func TestAssemblyRefusesReplay(t *testing.T) {
	a := NewObjectAssembler()
	a.Accept(makePart(7, 0, 1, true, []byte{9}))
	if _, done, _ := a.Accept(makePart(7, 0, 1, true, []byte{9})); done {
		t.Fatal("a replayed object is refused whole")
	}
}

func TestAssemblyBudgetsOpenObjects(t *testing.T) {
	a := NewObjectAssembler()
	for oid := uint32(1); oid <= MaxOpenObjects; oid++ {
		if _, _, err := a.Accept(makePart(oid, 0, 2, false, []byte{1})); err != nil {
			t.Fatalf("object %d: %v", oid, err)
		}
	}
	if _, _, err := a.Accept(makePart(99, 0, 2, false, []byte{1})); err == nil {
		t.Fatal("the fifth open object must exceed the budget")
	}
}
