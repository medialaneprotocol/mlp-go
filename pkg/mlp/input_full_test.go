package mlp

import "testing"

func TestInputManager(t *testing.T) {
	m := NewInputManager()
	seq := m.Submit("KeyA", true, 1000)
	if !m.IsPressed("KeyA") {
		t.Fatal("pressed")
	}
	if m.PressedCount() != 1 {
		t.Fatal("count")
	}
	m.Submit("KeyA", false, 1000)
	if m.IsPressed("KeyA") {
		t.Fatal("released")
	}
	if len(m.Expired(2000)) < 1 {
		t.Fatal("expired")
	}
	_ = seq
}
