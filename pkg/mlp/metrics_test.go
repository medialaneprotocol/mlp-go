package mlp

import "testing"

func TestMetrics(t *testing.T) {
	tm := &PrometheusTelemetry{}
	tm.RecordFragmentSent(100)
	tm.RecordUpdateApplied()
	out := tm.Render()
	if !contains(out, "mlp_fragments_sent_total 1") {
		t.Fatal("missing fragments")
	}
	if !contains(out, "mlp_bytes_sent_total 100") {
		t.Fatal("missing bytes")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})()
}
