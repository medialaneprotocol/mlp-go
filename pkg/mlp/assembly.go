package mlp

// ObjectAssembler reassembles fragments into objects.
type ObjectAssembler struct {
	Open map[uint32]*OpenObject
}

type OpenObject struct {
	ObjectID     uint32
	Class        FragmentClass
	Flags        FragmentFlags
	Priority     Priority
	Declared     uint16
	Received     uint16
	Ext          *HeaderExt
	Parts        [][]byte
	Bytes        int
}

func NewObjectAssembler() *ObjectAssembler {
	return &ObjectAssembler{Open: map[uint32]*OpenObject{}}
}

// Accept a fragment; returns the assembled object when complete.
func (a *ObjectAssembler) Accept(f *Fragment) (*OpenObject, bool) {
	oid := f.Header.ObjectID
	obj, ok := a.Open[oid]
	if !ok {
		obj = &OpenObject{
			ObjectID: oid,
			Class:    f.Header.Class,
			Flags:    f.Header.Flags,
			Priority: f.Header.Priority,
			Declared: f.Header.FragmentCount,
			Received: 0,
			Ext:      f.Ext,
			Parts:    make([][]byte, f.Header.FragmentCount),
			Bytes:    0,
		}
		a.Open[oid] = obj
	}
	if f.Header.FragmentCount != obj.Declared {
		return nil, false
	}
	idx := f.Header.FragmentIndex
	if obj.Parts[idx] != nil {
		return nil, false // duplicate
	}
	obj.Parts[idx] = f.Payload
	obj.Received++
	obj.Bytes += len(f.Payload)
	if idx == 0 {
		obj.Ext = f.Ext
	}
	if obj.Received >= obj.Declared {
		delete(a.Open, oid)
		return obj, true
	}
	return nil, false
}

// IsOpen reports whether an object is partially assembled.
func (a *ObjectAssembler) IsOpen(oid uint32) bool {
	_, ok := a.Open[oid]
	return ok
}

// OpenCount reports the number of partially assembled objects.
func (a *ObjectAssembler) OpenCount() int { return len(a.Open) }
