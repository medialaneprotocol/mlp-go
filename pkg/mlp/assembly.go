package mlp

// The object assembly budget, frozen by the specification (MLP spec §6.4):
// at most MAX_OPEN_OBJECTS objects open at once, at most
// MAX_ASSEMBLY_BYTES of partial payloads among them, and a dedup window of
// DEDUP_WINDOW object ids so a replayed object is refused whole.
const (
	MaxOpenObjects   = 4
	MaxAssemblyBytes = 1 << 20
	DedupWindow      = 4096
)

// OpenObject is an object under assembly.
type OpenObject struct {
	ObjectID uint32
	Class    FragmentClass
	Flags    FragmentFlags
	Priority Priority
	Declared uint16
	Received uint16
	Ext      *HeaderExt
	Parts    [][]byte
	Bytes    int
	SawFin   bool
}

// Whole reports whether the object has every fragment and its FIN.
func (o *OpenObject) Whole() bool {
	return o.SawFin && o.Received >= o.Declared
}

// Payload concatenates the parts in fragment order.
func (o *OpenObject) Payload() []byte {
	out := make([]byte, 0, o.Bytes)
	for _, p := range o.Parts {
		out = append(out, p...)
	}
	return out
}

// ObjectAssembler reassembles fragments into objects. An object is whole
// when its FIN fragment has arrived and every declared fragment is present.
type ObjectAssembler struct {
	Open   map[uint32]*OpenObject
	Seen   map[uint32]bool // dedup by object id, windowed
	Window []uint32        // FIFO of the dedup window
	Bytes  int             // partial payloads held
}

func NewObjectAssembler() *ObjectAssembler {
	return &ObjectAssembler{Open: map[uint32]*OpenObject{}, Seen: map[uint32]bool{}}
}

// Accept takes one fragment. It returns the assembled object when the
// object became whole on this fragment, and an error for a fragment that
// contradicts an open object (a different declared count) or exceeds the
// budget.
func (a *ObjectAssembler) Accept(f *Fragment) (*OpenObject, bool, error) {
	oid := f.Header.ObjectID
	if a.Seen[oid] {
		return nil, false, nil // a replayed object is refused whole
	}
	obj, ok := a.Open[oid]
	if !ok {
		if len(a.Open) >= MaxOpenObjects {
			return nil, false, errAssemblyBudget
		}
		obj = &OpenObject{
			ObjectID: oid,
			Class:    f.Header.Class,
			Flags:    f.Header.Flags,
			Priority: f.Header.Priority,
			Declared: f.Header.FragmentCount,
			Ext:      f.Ext,
			Parts:    make([][]byte, f.Header.FragmentCount),
		}
		a.Open[oid] = obj
	}
	if f.Header.FragmentCount != obj.Declared {
		return nil, false, errMetaMismatch
	}
	idx := f.Header.FragmentIndex
	if obj.Parts[idx] != nil {
		return nil, false, nil // duplicate fragment, already held
	}
	if a.Bytes+len(f.Payload) > MaxAssemblyBytes {
		return nil, false, errAssemblyBudget
	}
	obj.Parts[idx] = f.Payload
	obj.Received++
	obj.Bytes += len(f.Payload)
	a.Bytes += len(f.Payload)
	if idx == 0 {
		obj.Ext = f.Ext
	}
	if f.Header.Flags.Contains(FlagFin) {
		obj.SawFin = true
	}
	if obj.Whole() {
		delete(a.Open, oid)
		a.Bytes -= obj.Bytes
		a.remember(oid)
		return obj, true, nil
	}
	return nil, false, nil
}

func (a *ObjectAssembler) remember(oid uint32) {
	a.Seen[oid] = true
	a.Window = append(a.Window, oid)
	for len(a.Window) > DedupWindow {
		old := a.Window[0]
		a.Window = a.Window[1:]
		delete(a.Seen, old)
	}
}

// IsOpen reports whether an object is partially assembled.
func (a *ObjectAssembler) IsOpen(oid uint32) bool {
	_, ok := a.Open[oid]
	return ok
}

// OpenCount reports the number of partially assembled objects.
func (a *ObjectAssembler) OpenCount() int { return len(a.Open) }

type assemblyError string

func (e assemblyError) Error() string { return string(e) }

const (
	errAssemblyBudget assemblyError = "assembly budget exceeded"
	errMetaMismatch   assemblyError = "meta mismatch"
)
