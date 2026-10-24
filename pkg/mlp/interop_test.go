package mlp

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"
)

// TestInteropRoundTrip verifies that Go can parse all Rust golden vectors.
func TestInteropRoundTrip(t *testing.T) {
	data, _ := os.ReadFile("conformance/vectors.json")
	var cf struct {
		Vectors []struct {
			Name string `json:"name"`
			Hex  string `json:"hex"`
		} `json:"vectors"`
	}
	json.Unmarshal(data, &cf)
	for _, v := range cf.Vectors {
		b, _ := hex.DecodeString(v.Hex)
		f, n, err := ParseFragment(b)
		if err != nil {
			t.Fatalf("%s: parse error: %v", v.Name, err)
		}
		if n != len(b) {
			t.Fatalf("%s: consumed %d != %d", v.Name, n, len(b))
		}
		// Re-encode and verify byte-for-byte.
		enc, err := f.Encode()
		if err != nil {
			t.Fatalf("%s: encode error: %v", v.Name, err)
		}
		if hex.EncodeToString(enc) != v.Hex {
			t.Fatalf("%s: re-encode mismatch", v.Name)
		}
	}
}

// TestSenderReceiverRoundTrip verifies a full round trip.
func TestSenderReceiverRoundTrip(t *testing.T) {
	res, err := RunRoundTrip(320, 240, []byte{1, 2, 3, 4, 5})
	if err != nil {
		t.Fatal(err)
	}
	if res.FragmentsSent == 0 {
		t.Fatal("no fragments sent")
	}
	if res.FragmentsReceived != res.FragmentsSent {
		t.Fatalf("received %d != sent %d", res.FragmentsReceived, res.FragmentsSent)
	}
}
