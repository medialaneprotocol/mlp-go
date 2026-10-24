package mlp

import "time"

const TicketTTLSeconds = 3600

type SessionTicket struct {
	Payload    []byte
	IssuedAtMs uint64
	Nonce      uint64
}

func (t *SessionTicket) IsValid(nowMs uint64) bool {
	return nowMs >= t.IssuedAtMs && nowMs-t.IssuedAtMs <= TicketTTLSeconds*1000
}

func IssueTicket(lastAppliedUpdateID uint64, chainEpoch uint32) SessionTicket {
	payload := make([]byte, 12)
	// little-endian
	for i := 0; i < 8; i++ {
		payload[i] = byte(lastAppliedUpdateID >> (8 * i))
	}
	for i := 0; i < 4; i++ {
		payload[8+i] = byte(chainEpoch >> (8 * i))
	}
	return SessionTicket{Payload: payload, IssuedAtMs: uint64(time.Now().UnixMilli()), Nonce: uint64(time.Now().UnixNano())}
}

func ValidateTicket(t *SessionTicket, nowMs uint64) (uint64, uint32, bool) {
	if !t.IsValid(nowMs) {
		return 0, 0, false
	}
	if len(t.Payload) < 12 {
		return 0, 0, false
	}
	var uid uint64
	for i := 0; i < 8; i++ {
		uid |= uint64(t.Payload[i]) << (8 * i)
	}
	var epoch uint32
	for i := 0; i < 4; i++ {
		epoch |= uint32(t.Payload[8+i]) << (8 * i)
	}
	return uid, epoch, true
}
