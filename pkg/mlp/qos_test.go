package mlp

import "testing"

func TestQos(t *testing.T) {
	if QosRealtime.MlpPriority() >= QosInteractive.MlpPriority() {
		t.Fatal("priority order")
	}
	if QosRealtime.Weight() <= QosInteractive.Weight() {
		t.Fatal("weight order")
	}
}
