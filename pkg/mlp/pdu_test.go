package mlp

import (
	"encoding/hex"
	"testing"
)

func TestGoldenFragment(t *testing.T) {
	meta := &UpdateMeta{
		UpdateID:         7,
		UpdateBase:       6,
		PartIndex:        1,
		PartCount:        4,
		UpdateBytes:      1234,
		ChainEpoch:       3,
		ReflectsInputSeq: 99,
		SendEpochMs:      1000,
	}
	ext := meta.ToExt()
	frag := &Fragment{
		Header: FragmentHeader{
			Version:        MediaProtocolVersion,
			Class:          ClassMedia,
			Flags:          0,
			Priority:       PrioMedia,
			ObjectID:       42,
			FragmentIndex:  0,
			FragmentCount:  2,
			PayloadLen:     4,
			HeaderExtLen:   uint16(ext.EncodedLen()),
		},
		Ext:     ext,
		Payload: []byte{1, 2, 3, 4},
	}
	b, err := frag.Encode()
	if err != nil {
		t.Fatal(err)
	}
	got := hex.EncodeToString(b)
	want := "4d4c5000010200062a00000000000200040000003a0001080700000000000000020806000000000000000304010004000404d2040000050403000000060863000000000000000708e80300000000000001020304"
	if got != want {
		t.Fatalf("mismatch:\n got %s\nwant %s", got, want)
	}
	// round-trip parse
	f2, n, err := ParseFragment(b)
	if err != nil {
		t.Fatal(err)
	}
	if n != len(b) {
		t.Fatalf("consumed %d != %d", n, len(b))
	}
	if f2.Header.ObjectID != 42 {
		t.Fatal("object id")
	}
	m2 := UpdateMetaFromExt(f2.Ext)
	if *m2 != *meta {
		t.Fatalf("meta mismatch: %+v", m2)
	}
}
