package mlp

// InputEvent is a viewer input event.
type InputEvent struct {
	Code    string
	Pressed bool
	Seq     uint64
}

// InputQueue holds pending inputs.
type InputQueue struct {
	Events []InputEvent
}

func NewInputQueue() *InputQueue { return &InputQueue{} }

func (q *InputQueue) Submit(e InputEvent) { q.Events = append(q.Events, e) }
func (q *InputQueue) Len() int           { return len(q.Events) }
