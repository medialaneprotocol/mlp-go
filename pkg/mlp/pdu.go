// Package mlp implements the Media Lane Protocol (MLP) wire format.
//
// This is the Go implementation of the protocol's byte-level PDU codec:
// fragment header, TLV header extension, and update metadata. The wire is
// frozen by the specification; this package mirrors mlp-core (Rust)
// byte-for-byte. What it does not implement — the pacing machines, the
// bandwidth estimator, the repair plans — it does not pretend to.
package mlp

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// The fragment magic: "MLP" followed by a zero byte. The zero terminator
// keeps the magic a fixed four bytes without implying a version: the version
// travels in its own header field.
var FragmentMagic = [4]byte{'M', 'L', 'P', 0}

// MediaProtocolVersion is the MLP version carried by every fragment.
const MediaProtocolVersion uint8 = 1

// FragmentFixedHeader is the header size before the extension.
const FragmentFixedHeader = 22

// MaxFragmentCount is the largest fragment_count the wire admits.
const MaxFragmentCount uint16 = 4096

// MaxHeaderExt is the largest header extension length.
const MaxHeaderExt uint16 = 256

// MaxStreamFragment is the payload cap of a stream-carrier fragment message.
const MaxStreamFragment = 65_536

// MaxGrpcCoalesce is the cap of concatenated fragments in one gRPC message.
const MaxGrpcCoalesce = 65_536

// TargetFragmentPayload is the payload target on the datagram carriers.
const TargetFragmentPayload = 1_150

// FragmentClass is the transport class field of a fragment. The wire freezes
// the classes at 0..=4; 5..=255 are reserved and rejected as BAD_CLASS.
type FragmentClass uint8

const (
	ClassControl  FragmentClass = 0 // media-lane control (C0-equivalent)
	ClassCursor   FragmentClass = 1 // cursor (C2)
	ClassMedia    FragmentClass = 2 // media (C3)
	ClassFeedback FragmentClass = 3 // feedback/telemetry (C4)
	ClassPath     FragmentClass = 4 // path/probe control (C5)
)

// ClassFromU8 decodes the wire value; 5..=255 is an error.
func ClassFromU8(v uint8) (FragmentClass, error) {
	if v > 4 {
		return 0, fmt.Errorf("bad class %d", v)
	}
	return FragmentClass(v), nil
}

// String is the stable label spelling.
func (c FragmentClass) String() string {
	switch c {
	case ClassControl:
		return "control"
	case ClassCursor:
		return "cursor"
	case ClassMedia:
		return "media"
	case ClassFeedback:
		return "feedback"
	case ClassPath:
		return "path"
	}
	return "unknown"
}

// Priority: 0..=7, lower is higher. 0-1 are reserved for control, pointer
// and feedback; media may not use them.
type Priority uint8

const (
	PrioControl Priority = 0
	PrioInput   Priority = 1
	PrioReport  Priority = 2
	PrioCursor  Priority = 5
	PrioMedia   Priority = 6
	PrioRefresh Priority = 7
)

func (p Priority) Get() uint8 { return uint8(p) }

// FragmentFlags is the flags byte: bit 7 is reserved and must be zero.
type FragmentFlags uint8

const (
	FlagFin           FragmentFlags = 1 << 0 // last fragment of the object
	FlagAnchor        FragmentFlags = 1 << 1 // absolute writes applying onto any held state
	FlagRetransmit    FragmentFlags = 1 << 2 // a resend by the sender's own policy
	FlagDup           FragmentFlags = 1 << 3 // a deliberate duplicate for echo-critical objects
	FlagEchoRequested FragmentFlags = 1 << 4 // receiver must reply with an echo
	FlagProbe         FragmentFlags = 1 << 5 // a probe object
	FlagUrgent        FragmentFlags = 1 << 6 // control: drain before queued media
	FlagNone          FragmentFlags = 0
)

func (f FragmentFlags) Contains(b FragmentFlags) bool { return f&b != 0 }

// FlagsFromBits decodes the raw byte; bit 7 set is an error.
func FlagsFromBits(bits uint8) (FragmentFlags, error) {
	if bits&0x80 != 0 {
		return 0, errors.New("bad flags: bit 7 is reserved")
	}
	return FragmentFlags(bits), nil
}

// The header-extension TLV type registry. Kinds 0x10..=0xEF are skipped
// silently (forward compatibility); 0xF0..=0xFF are experimental: skipped
// and counted.
var tlv = struct {
	UpdateID         uint8
	UpdateBase       uint8
	Part             uint8
	UpdateBytes      uint8
	ChainEpoch       uint8
	ReflectsInputSeq uint8
	SendEpochMs      uint8
	DeadlineEpochMs  uint8
	EchoOfObjectID   uint8
	BudgetBytes      uint8
	PaceBps          uint8
	PathHint         uint8
	InputSeq         uint8
	CellCount        uint8
	CostMs           uint8
	StreamID         uint8
	FecGroupID       uint8
	FecK             uint8
	PathID           uint8
	Grease           uint8
}{
	UpdateID:         0x01,
	UpdateBase:       0x02,
	Part:             0x03,
	UpdateBytes:      0x04,
	ChainEpoch:       0x05,
	ReflectsInputSeq: 0x06,
	SendEpochMs:      0x07,
	DeadlineEpochMs:  0x08,
	EchoOfObjectID:   0x09,
	BudgetBytes:      0x0A,
	PaceBps:          0x0B,
	PathHint:         0x0C,
	InputSeq:         0x0D,
	CellCount:        0x0E,
	CostMs:           0x0F,
	StreamID:         0x10,
	FecGroupID:       0x11,
	FecK:             0x12,
	PathID:           0x13,
	Grease:           0x7F,
}

// TlvKindName returns the registry's name for a kind, if it has one.
func TlvKindName(kind uint8) (string, bool) {
	names := map[uint8]string{
		tlv.UpdateID:         "UPDATE_ID",
		tlv.UpdateBase:       "UPDATE_BASE",
		tlv.Part:             "PART",
		tlv.UpdateBytes:      "UPDATE_BYTES",
		tlv.ChainEpoch:       "CHAIN_EPOCH",
		tlv.ReflectsInputSeq: "REFLECTS_INPUT_SEQ",
		tlv.SendEpochMs:      "SEND_EPOCH_MS",
		tlv.DeadlineEpochMs:  "DEADLINE_EPOCH_MS",
		tlv.EchoOfObjectID:   "ECHO_OF_OBJECT_ID",
		tlv.BudgetBytes:      "BUDGET_BYTES",
		tlv.PaceBps:          "PACE_BPS",
		tlv.PathHint:         "PATH_HINT",
		tlv.InputSeq:         "INPUT_SEQ",
		tlv.CellCount:        "CELL_COUNT",
		tlv.CostMs:           "COST_MS",
		tlv.FecGroupID:       "FEC_GROUP_ID",
		tlv.FecK:             "FEC_K",
}
	n, ok := names[kind]
	return n, ok
}

// TlvEntry is one typed length value.
type TlvEntry struct {
	Kind  uint8
	Value []byte
}

// HeaderExt is a TLV container.
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

func (h *HeaderExt) PushU16(kind uint8, v uint16) {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, v)
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
	if len(v) != 8 {
		return 0, false
	}
	return binary.LittleEndian.Uint64(v), true
}

func (h *HeaderExt) GetU32(kind uint8) (uint32, bool) {
	v := h.Get(kind)
	if len(v) != 4 {
		return 0, false
	}
	return binary.LittleEndian.Uint32(v), true
}

func (h *HeaderExt) GetU16(kind uint8) (uint16, bool) {
	v := h.Get(kind)
	if len(v) != 2 {
		return 0, false
	}
	return binary.LittleEndian.Uint16(v), true
}

func (h *HeaderExt) GetU8(kind uint8) (uint8, bool) {
	v := h.Get(kind)
	if len(v) != 1 {
		return 0, false
	}
	return v[0], true
}

func (h *HeaderExt) GetU16Pair(kind uint8) (uint16, uint16, bool) {
	v := h.Get(kind)
	if len(v) != 4 {
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

// ParseHeaderExt parses a whole TLV sequence; a partial trailing TLV is an
// error, not a truncation to keep.
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

// FragmentHeader is the fixed 22-byte header.
type FragmentHeader struct {
	Version       uint8
	Class         FragmentClass
	Flags         FragmentFlags
	Priority      Priority
	ObjectID      uint32
	FragmentIndex uint16
	FragmentCount uint16
	PayloadLen    uint32
	HeaderExtLen  uint16
}

// Fragment is a header, its extension, and the payload.
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
	if len(extBytes) > int(MaxHeaderExt) {
		return nil, errors.New("bad ext: too long")
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

// ParseFragment parses one fragment from buf and returns the bytes consumed.
func ParseFragment(buf []byte) (*Fragment, int, error) {
	if len(buf) < FragmentFixedHeader {
		return nil, 0, errors.New("bad length")
	}
	if buf[0] != FragmentMagic[0] || buf[1] != FragmentMagic[1] ||
		buf[2] != FragmentMagic[2] || buf[3] != FragmentMagic[3] {
		return nil, 0, errors.New("bad magic")
	}
	version := buf[4]
	if version != MediaProtocolVersion {
		return nil, 0, errors.New("unsupported version")
	}
	class, err := ClassFromU8(buf[5])
	if err != nil {
		return nil, 0, err
	}
	flags, err := FlagsFromBits(buf[6])
	if err != nil {
		return nil, 0, err
	}
	prioRaw := buf[7]
	if prioRaw > 7 {
		return nil, 0, errors.New("bad priority")
	}
	priority := Priority(prioRaw)
	if class == ClassMedia && priority <= PrioInput {
		return nil, 0, errors.New("bad priority: media may not use the reserved band")
	}
	objectID := binary.LittleEndian.Uint32(buf[8:12])
	fragIndex := binary.LittleEndian.Uint16(buf[12:14])
	fragCount := binary.LittleEndian.Uint16(buf[14:16])
	payloadLen := binary.LittleEndian.Uint32(buf[16:20])
	extLen := binary.LittleEndian.Uint16(buf[20:22])

	if fragCount == 0 || fragCount > MaxFragmentCount || fragIndex >= fragCount {
		return nil, 0, errors.New("bad fragment")
	}
	if extLen != 0 && extLen < 2 {
		return nil, 0, errors.New("bad ext")
	}
	if extLen > MaxHeaderExt {
		return nil, 0, errors.New("bad ext")
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
			Version:       version,
			Class:         class,
			Flags:         flags,
			Priority:      priority,
			ObjectID:      objectID,
			FragmentIndex: fragIndex,
			FragmentCount: fragCount,
			PayloadLen:    payloadLen,
			HeaderExtLen:  extLen,
		},
		Ext:     ext,
		Payload: payload,
	}
	return f, payloadEnd, nil
}

// UpdateMeta is the media update metadata a fragment part carries.
type UpdateMeta struct {
	UpdateID         uint64
	UpdateBase       uint64 // 0 = self-contained
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
	return &UpdateMeta{
		UpdateID:         getU64Or(ext, tlv.UpdateID),
		UpdateBase:       getU64Or(ext, tlv.UpdateBase),
		PartIndex:        partIndex,
		PartCount:        partCount,
		UpdateBytes:      getU32Or(ext, tlv.UpdateBytes),
		ChainEpoch:       getU32Or(ext, tlv.ChainEpoch),
		ReflectsInputSeq: getU64Or(ext, tlv.ReflectsInputSeq),
		StreamID:         getU32Or(ext, tlv.StreamID),
		PathID:           getU8Or(ext, tlv.PathID),
		SendEpochMs:      getU64Or(ext, tlv.SendEpochMs),
	}
}

func getU64Or(h *HeaderExt, kind uint8) uint64 { v, _ := h.GetU64(kind); return v }
func getU32Or(h *HeaderExt, kind uint8) uint32 { v, _ := h.GetU32(kind); return v }
func getU8Or(h *HeaderExt, kind uint8) uint8   { v, _ := h.GetU8(kind); return v }

// CheckAgainst reports whether two metas of the same object agree.
func (m *UpdateMeta) CheckAgainst(other *UpdateMeta) error {
	if *m != *other {
		return errors.New("meta mismatch")
	}
	return nil
}

// Carrier identifies the transport carrier.
type Carrier uint8

const (
	CarrierWs             Carrier = 0 // WebSocket: one fragment per message, 64 KiB cap
	CarrierGrpc           Carrier = 1 // gRPC: 1..N fragments per message, 64 KiB coalesce cap
	CarrierDataChannel    Carrier = 2 // WebRTC data channel rm.media
	CarrierRepairChannel  Carrier = 3 // WebRTC data channel rm.repair
	CarrierControlChannel Carrier = 4 // WebRTC data channel rm.control
)

// FragmentPayloadMax is the largest payload a single fragment may carry.
func (c Carrier) FragmentPayloadMax() int {
	switch c {
	case CarrierWs, CarrierGrpc:
		return MaxStreamFragment
	case CarrierDataChannel, CarrierRepairChannel, CarrierControlChannel:
		return TargetFragmentPayload
	}
	return TargetFragmentPayload
}

// MessageCap is the largest concatenated fragment message this carrier takes.
func (c Carrier) MessageCap() int {
	if c == CarrierGrpc {
		return MaxGrpcCoalesce
	}
	return c.FragmentPayloadMax()
}

// IsDatagram reports whether this carrier is datagram-class (one fragment
// per message).
func (c Carrier) IsDatagram() bool {
	switch c {
	case CarrierDataChannel, CarrierRepairChannel, CarrierControlChannel:
		return true
	}
	return false
}
