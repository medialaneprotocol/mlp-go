package mlp

import "fmt"

// PrometheusTelemetry records MLP metrics.
type PrometheusTelemetry struct {
	FragmentsSent     uint64
	FragmentsReceived uint64
	BytesSent         uint64
	BytesReceived     uint64
	UpdatesApplied    uint64
	Convergences      uint64
	Losses            uint64
	Echoes            uint64
}

func (t *PrometheusTelemetry) RecordFragmentSent(n uint64) {
	t.FragmentsSent++
	t.BytesSent += n
}
func (t *PrometheusTelemetry) RecordFragmentReceived(n uint64) {
	t.FragmentsReceived++
	t.BytesReceived += n
}
func (t *PrometheusTelemetry) RecordUpdateApplied() { t.UpdatesApplied++ }
func (t *PrometheusTelemetry) RecordConvergence()   { t.Convergences++ }
func (t *PrometheusTelemetry) RecordLoss()          { t.Losses++ }
func (t *PrometheusTelemetry) RecordEcho()          { t.Echoes++ }

func (t *PrometheusTelemetry) Render() string {
	return fmt.Sprintf(
		"# HELP mlp_fragments_sent_total Total fragments sent.\n"+
			"# TYPE mlp_fragments_sent_total counter\n"+
			"mlp_fragments_sent_total %d\n"+
			"# HELP mlp_bytes_sent_total Total bytes sent.\n"+
			"# TYPE mlp_bytes_sent_total counter\n"+
			"mlp_bytes_sent_total %d\n"+
			"# HELP mlp_updates_applied_total Total updates applied.\n"+
			"# TYPE mlp_updates_applied_total counter\n"+
			"mlp_updates_applied_total %d\n"+
			"# HELP mlp_convergences_total Total convergences.\n"+
			"# TYPE mlp_convergences_total counter\n"+
			"mlp_convergences_total %d\n",
		t.FragmentsSent, t.BytesSent, t.UpdatesApplied, t.Convergences)
}
