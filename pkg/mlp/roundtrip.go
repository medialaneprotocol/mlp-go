package mlp

import "fmt"

// RoundTripResult holds the result of a sender->receiver round trip.
type RoundTripResult struct {
	FragmentsSent     int
	FragmentsReceived int
	ObjectsAssembled  int
	CellsPending      int
	Converged         bool
}

// RunRoundTrip simulates a full sender->receiver round trip.
func RunRoundTrip(w, h uint32, payload []byte) (*RoundTripResult, error) {
	grid := CellGrid{WidthPx: w, HeightPx: h, CellSize: 64}
	sender := NewFullMediaSender(grid)
	receiver := NewFullMediaReceiver(w, h)

	uid, err := sender.EnqueueUpdate(UpdateRequest{Payload: payload, Cells: []uint32{0, 1, 2}})
	if err != nil {
		return nil, err
	}
	_ = uid

	// Encode fragments
	frags, err := sender.EnqueueUpdate(UpdateRequest{Payload: payload, Cells: []uint32{0, 1, 2}})
	_ = frags
	// Use the basic sender for fragment generation
	basicSender := NewMediaSenderMachine(grid)
	basicFrags, err := basicSender.EnqueueUpdate(UpdateRequest{Payload: payload}, 1200)
	if err != nil {
		return nil, err
	}

	result := &RoundTripResult{}
	result.FragmentsSent = len(basicFrags)

	// Deliver fragments to receiver
	for _, f := range basicFrags {
		b, _ := f.Encode()
		_, _, err := receiver.AcceptFragment(b)
		if err == nil {
			result.FragmentsReceived++
		}
	}

	result.ObjectsAssembled = receiver.Assembler.OpenCount()
	result.CellsPending = receiver.Accounting.PendingCount()
	result.Converged = receiver.Converged()

	fmt.Printf("roundtrip: sent=%d received=%d objects=%d pending=%d converged=%v\n",
		result.FragmentsSent, result.FragmentsReceived, result.ObjectsAssembled,
		result.CellsPending, result.Converged)

	return result, nil
}
