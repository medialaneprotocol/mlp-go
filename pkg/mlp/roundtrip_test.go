package mlp

import "testing"

func TestRoundTrip(t *testing.T) {
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
