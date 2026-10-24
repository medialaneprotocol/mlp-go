package mlp

import "testing"

func TestFEC(t *testing.T) {
	data := [][]byte{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}
	parity := XorParity(data)
	got := [][]byte{data[0], nil, data[2]}
	rec := RecoverOne(1, got, parity)
	if rec == nil || rec[0] != 4 || rec[1] != 5 || rec[2] != 6 {
		t.Fatal("recover")
	}
}
