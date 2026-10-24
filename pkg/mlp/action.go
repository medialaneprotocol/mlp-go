package mlp

// Action is an outbound action.
type Action struct {
	Kind string
	Data []byte
}

// ConnectionEvent is a connection-level event.
type ConnectionEvent struct {
	Kind string
	Peer string
}

// Receive is a received message.
type Receive struct {
	Data []byte
	At   int64
}

// Transmit is an outbound transmission.
type Transmit struct {
	To   string
	Data []byte
}
