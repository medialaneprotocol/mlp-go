// Package mlp implements the Media Lane Protocol (MLP) wire format.
//
// This is the Go reference implementation of the protocol's byte-level PDU
// codec: fragment header, TLV header extension, and update metadata. It
// mirrors mlp-core (Rust) byte-for-byte.
package mlp

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// Wire constants.
var FragmentMagic = [4]byte{'M', 'L', 'P', 0}

const MediaProtocolVersion uint8 = 1
const FragmentFixedHeader = 22
const MaxHeaderExt uint16 = 256

// Fragment class.
type FragmentClass uint8

const (
	ClassControl FragmentClass = 0
	ClassCursor  FragmentClass = 1
	ClassMedia   FragmentClass = 2
	ClassFeedback FragmentClass = 3
	ClassPath    FragmentClass = 4
	ClassDatagram FragmentClass = 5
	ClassFec      FragmentClass = 6
)

// Priority.
type Priority uint8

const (
	PrioControl  Priority = 0
	PrioInput    Priority = 1
	PrioReport   Priority = 2
	PrioCursor   Priority = 5
	PrioMedia    Priority = 6
	PrioRefresh  Priority = 7
)

func (p Priority) Get() uint8 { return uint8(p) }

// Fragment flags.
type FragmentFlags uint8

const (
	FlagAnchor       FragmentFlags = 1 << 0
	FlagEchoRequested FragmentFlags = 1 << 1
	FlagDuplicate     FragmentFlags = 1 << 2
)

func (f FragmentFlags) Contains(b FragmentFlags) bool { return f&b != 0 }

// Traffic class (for pacer) - not on wire but used internally; kept for parity.
type TrafficClass uint8

// TLV kinds.
var tlv = struct {
	UpdateID         uint8
	UpdateBase       uint8
	Part             uint8
	UpdateBytes      uint8
	ChainEpoch       uint8
	ReflectsInputSeq uint8
	SendEpochMs      uint8
	CellCount        uint8
	StreamID         uint8
	PathID           uint8
	Grease           uint8
	BudgetBytes      uint8
}{
	UpdateID:         0x01,
	UpdateBase:       0x02,
	Part:             0x03,
	UpdateBytes:      0x04,
	ChainEpoch:       0x05,
	ReflectsInputSeq: 0x06,
	SendEpochMs:      0x07,
	CellCount:        0x08,
	StreamID:         0x10,
	PathID:           0x13,
	Grease:           0x7F,
	BudgetBytes:      0x09,
}

// HeaderExt is a TLV container.
type TlvEntry struct {
	Kind  uint8
	Value []byte
}

type HeaderExt struct {
	Entries []TlvEntry
}

func NewHeaderExt() *HeaderExt { return &HeaderExt{} }

func (h *HeaderExt) PushU64(kind uint8, v uint64) {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, v)
	h.Entries = append(h.Entries, TlvEntry{Kind: kind, Value: b})
}

func (h *HeaderExt) PushU32(kind uint8, v uint32) {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, v)
	h.Entries = append(h.Entries, TlvEntry{Kind: kind, Value: b})
}

func (h *HeaderExt) PushU8(kind uint8, v uint8) {
	h.Entries = append(h.Entries, TlvEntry{Kind: kind, Value: []byte{v}})
}

func (h *HeaderExt) PushU16Pair(kind uint8, a, b uint16) {
	v := make([]byte, 4)
	binary.LittleEndian.PutUint16(v[0:2], a)
	binary.LittleEndian.PutUint16(v[2:4], b)
	h.Entries = append(h.Entries, TlvEntry{Kind: kind, Value: v})
}

func (h *HeaderExt) Get(kind uint8) []byte {
	for _, e := range h.Entries {
		if e.Kind == kind {
			return e.Value
		}
	}
	return nil
}

func (h *HeaderExt) GetU64(kind uint8) (uint64, bool) {
	v := h.Get(kind)
	if v == nil || len(v) != 8 {
		return 0, false
	}
	return binary.LittleEndian.Uint64(v), true
}

func (h *HeaderExt) GetU32(kind uint8) (uint32, bool) {
	v := h.Get(kind)
	if v == nil || len(v) != 4 {
		return 0, false
	}
	return binary.LittleEndian.Uint32(v), true
}

func (h *HeaderExt) GetU8(kind uint8) (uint8, bool) {
	v := h.Get(kind)
	if v == nil || len(v) != 1 {
		return 0, false
	}
	return v[0], true
}

func (h *HeaderExt) GetU16Pair(kind uint8) (uint16, uint16, bool) {
	v := h.Get(kind)
	if v == nil || len(v) != 4 {
		return 0, 0, false
	}
	return binary.LittleEndian.Uint16(v[0:2]), binary.LittleEndian.Uint16(v[2:4]), true
}

func (h *HeaderExt) EncodedLen() int {
	n := 0
	for _, e := range h.Entries {
		n += 2 + len(e.Value)
	}
	return n
}

func (h *HeaderExt) Encode() ([]byte, error) {
	out := make([]byte, 0, h.EncodedLen())
	for _, e := range h.Entries {
		if len(e.Value) > 255 {
			return nil, errors.New("tlv value too long")
		}
		out = append(out, e.Kind, uint8(len(e.Value)))
		out = append(out, e.Value...)
	}
	return out, nil
}

func ParseHeaderExt(b []byte) (*HeaderExt, error) {
	h := &HeaderExt{}
	off := 0
	for off < len(b) {
		if off+2 > len(b) {
			return nil, errors.New("bad ext")
		}
		kind := b[off]
		l := int(b[off+1])
		off += 2
		if off+l > len(b) {
			return nil, errors.New("bad ext")
		}
		v := make([]byte, l)
		copy(v, b[off:off+l])
		h.Entries = append(h.Entries, TlvEntry{Kind: kind, Value: v})
		off += l
	}
	return h, nil
}

// FragmentHeader.
type FragmentHeader struct {
	Version        uint8
	Class          FragmentClass
	Flags          FragmentFlags
	Priority       Priority
	ObjectID       uint32
	FragmentIndex  uint16
	FragmentCount  uint16
	PayloadLen     uint32
	HeaderExtLen   uint16
}

// Fragment.
type Fragment struct {
	Header  FragmentHeader
	Ext     *HeaderExt
	Payload []byte
}

func (f *Fragment) Encode() ([]byte, error) {
	extBytes, err := f.Ext.Encode()
	if err != nil {
		return nil, err
	}
	if len(extBytes) != int(f.Header.HeaderExtLen) {
		// recompute
		f.Header.HeaderExtLen = uint16(len(extBytes))
	}
	out := make([]byte, FragmentFixedHeader+len(extBytes)+len(f.Payload))
	copy(out[0:4], FragmentMagic[:])
	out[4] = f.Header.Version
	out[5] = uint8(f.Header.Class)
	out[6] = uint8(f.Header.Flags)
	out[7] = uint8(f.Header.Priority)
	binary.LittleEndian.PutUint32(out[8:12], f.Header.ObjectID)
	binary.LittleEndian.PutUint16(out[12:14], f.Header.FragmentIndex)
	binary.LittleEndian.PutUint16(out[14:16], f.Header.FragmentCount)
	binary.LittleEndian.PutUint32(out[16:20], uint32(len(f.Payload)))
	binary.LittleEndian.PutUint16(out[20:22], uint16(len(extBytes)))
	copy(out[22:22+len(extBytes)], extBytes)
	copy(out[22+len(extBytes):], f.Payload)
	return out, nil
}

// ParseFragment parses one fragment from buf, returns fragment and consumed bytes.
func ParseFragment(buf []byte) (*Fragment, int, error) {
	if len(buf) < FragmentFixedHeader {
		return nil, 0, errors.New("bad length")
	}
	if buf[0] != FragmentMagic[0] || buf[1] != FragmentMagic[1] || buf[2] != FragmentMagic[2] || buf[3] != FragmentMagic[3] {
		return nil, 0, errors.New("bad magic")
	}
	version := buf[4]
	if version != MediaProtocolVersion {
		return nil, 0, errors.New("unsupported version")
	}
	class := FragmentClass(buf[5])
	flags := FragmentFlags(buf[6])
	prioRaw := buf[7]
	if prioRaw > 7 {
		return nil, 0, errors.New("bad priority")
	}
	priority := Priority(prioRaw)
	objectID := binary.LittleEndian.Uint32(buf[8:12])
	fragIndex := binary.LittleEndian.Uint16(buf[12:14])
	fragCount := binary.LittleEndian.Uint16(buf[14:16])
	payloadLen := binary.LittleEndian.Uint32(buf[16:20])
	extLen := binary.LittleEndian.Uint16(buf[20:22])

	if fragCount == 0 || fragIndex >= fragCount {
		return nil, 0, errors.New("bad fragment")
	}
	if class == ClassMedia && priority <= PrioInput {
		return nil, 0, errors.New("bad priority")
	}
	if extLen != 0 && extLen < 2 {
		return nil, 0, errors.New("bad ext")
	}
	if extLen > MaxHeaderExt {
		return nil, 0, errors.New("bad ext")
	}
	if int(payloadLen) > 65535 { // carrier check loose
	}
	extEnd := FragmentFixedHeader + int(extLen)
	payloadEnd := extEnd + int(payloadLen)
	if len(buf) < payloadEnd {
		return nil, 0, errors.New("bad length")
	}
	ext, err := ParseHeaderExt(buf[FragmentFixedHeader:extEnd])
	if err != nil {
		return nil, 0, err
	}
	payload := make([]byte, payloadLen)
	copy(payload, buf[extEnd:payloadEnd])
	f := &Fragment{
		Header: FragmentHeader{
			Version:        version,
			Class:          class,
			Flags:          flags,
			Priority:       priority,
			ObjectID:       objectID,
			FragmentIndex:  fragIndex,
			FragmentCount:  fragCount,
			PayloadLen:     payloadLen,
			HeaderExtLen:   extLen,
		},
		Ext:     ext,
		Payload: payload,
	}
	return f, payloadEnd, nil
}

// UpdateMeta.
type UpdateMeta struct {
	UpdateID         uint64
	UpdateBase       uint64
	PartIndex        uint16
	PartCount        uint16
	UpdateBytes      uint32
	ChainEpoch       uint32
	ReflectsInputSeq uint64
	StreamID         uint32
	PathID           uint8
	SendEpochMs      uint64
}

func (m *UpdateMeta) ToExt() *HeaderExt {
	h := NewHeaderExt()
	h.PushU64(tlv.UpdateID, m.UpdateID)
	h.PushU64(tlv.UpdateBase, m.UpdateBase)
	h.PushU16Pair(tlv.Part, m.PartIndex, m.PartCount)
	h.PushU32(tlv.UpdateBytes, m.UpdateBytes)
	h.PushU32(tlv.ChainEpoch, m.ChainEpoch)
	h.PushU64(tlv.ReflectsInputSeq, m.ReflectsInputSeq)
	h.PushU32(tlv.StreamID, m.StreamID)
	h.PushU8(tlv.PathID, m.PathID)
	h.PushU64(tlv.SendEpochMs, m.SendEpochMs)
	return h
}

func UpdateMetaFromExt(ext *HeaderExt) *UpdateMeta {
	partIndex, partCount, _ := ext.GetU16Pair(tlv.Part)
	uid, _ := ext.GetU64(tlv.UpdateID)
	ub, _ := ext.GetU64(tlv.UpdateBase)
	ub2, _ := ext.GetU32(tlv.UpdateBytes)
	ce, _ := ext.GetU32(tlv.ChainEpoch)
	ris, _ := ext.GetU64(tlv.ReflectsInputSeq)
	sid, _ := ext.GetU32(tlv.StreamID)
	pid, _ := ext.GetU8(tlv.PathID)
	sem, _ := ext.GetU64(tlv.SendEpochMs)
	return &UpdateMeta{
		UpdateID:         uid,
		UpdateBase:       ub,
		PartIndex:        partIndex,
		PartCount:        partCount,
		UpdateBytes:      ub2,
		ChainEpoch:       ce,
		ReflectsInputSeq: ris,
		StreamID:         sid,
		PathID:           pid,
		SendEpochMs:      sem,
	}
}

func (m *UpdateMeta) CheckAgainst(other *UpdateMeta) error {
	if *m != *other {
		return errors.New("meta mismatch")
	}
	return nil
}

var _ = fmt.Sprintf

// FragmentFlags constants (wire).
const (
	FlagAnchorBit       = 1 << 0
	FlagEchoRequestedBit = 1 << 1
	FlagDuplicateBit    = 1 << 2
)

// Carrier identifies the transport carrier.
type Carrier uint8

const (
	CarrierWs   Carrier = 0
	CarrierGrpc Carrier = 1
)

func (c Carrier) FragmentPayloadMax() int {
	switch c {
	case CarrierWs:
		return 1024
	case CarrierGrpc:
		return 4096
	}
	return 1024
}
