package mlp

import "testing"

func TestAssembly(t *testing.T) {
	a := NewObjectAssembler()
	frag1 := &Fragment{
		Header: FragmentHeader{ObjectID: 1, FragmentIndex: 0, FragmentCount: 2, PayloadLen: 4},
		Ext:    NewHeaderExt(),
		Payload: []byte{1, 2, 3, 4},
	}
	frag2 := &Fragment{
		Header: FragmentHeader{ObjectID: 1, FragmentIndex: 1, FragmentCount: 2, PayloadLen: 4},
		Ext:    NewHeaderExt(),
		Payload: []byte{5, 6, 7, 8},
	}
	obj, done := a.Accept(frag1)
	if done {
		t.Fatal("not done yet")
	}
	_ = obj
	obj, done = a.Accept(frag2)
	if !done {
		t.Fatal("should be done")
	}
	if obj.Received != 2 {
		t.Fatal("received")
	}
	if a.OpenCount() != 0 {
		t.Fatal("open count")
	}
}
