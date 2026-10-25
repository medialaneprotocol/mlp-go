package mlp

// RoundTripResult is what one sender-to-receiver round trip proved.
type RoundTripResult struct {
	FragmentsSent     int
	FragmentsReceived int
	ObjectsAssembled  int
	UpdateID          uint64
	AckHighestApplied uint64
	PayloadMatch      bool
}

// RunRoundTrip sends one update through the wire and assembles it back.
func RunRoundTrip(w, h uint32, payload []byte) (*RoundTripResult, error) {
	grid := CellGrid{WidthPx: w, HeightPx: h, CellSize: 64}
	sender := NewMediaSenderMachine(grid)
	receiver := NewMediaReceiverMachine()

	outcome, err := sender.EnqueueUpdate(
		UpdateRequest{Payload: payload, Cells: []uint32{0, 1, 2}, Anchor: true},
		CarrierWs, 0)
	if err != nil {
		return nil, err
	}

	result := &RoundTripResult{FragmentsSent: outcome.Fragments, UpdateID: outcome.UpdateID}
	var assembled []byte
	for {
		f := sender.PollFragment()
		if f == nil {
			break
		}
		raw, err := f.Encode()
		if err != nil {
			return nil, err
		}
		obj, err := receiver.AcceptFragment(raw)
		if err != nil {
			return nil, err
		}
		result.FragmentsReceived++
		if obj != nil {
			assembled = obj.Payload()
			result.ObjectsAssembled++
			meta := UpdateMetaFromExt(obj.Ext)
			receiver.NoteApplied(meta.UpdateID)
		}
	}

	result.AckHighestApplied = receiver.Ack.HighestAppliedUpdateId()
	result.PayloadMatch = string(assembled) == string(payload)
	return result, nil
}
