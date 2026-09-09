package monitor

import (
	"strings"
	"testing"
)

func TestWritePrometheusIncludesCounters(t *testing.T) {
	m := &Monitor{}
	m.PollStarted()
	m.SubmissionsSeen(2)
	m.ProviderFailed()
	m.SnapshotRecorded()

	var output strings.Builder
	err := m.WritePrometheus(&output)
	if err != nil {
		t.Fatalf("WritePrometheus: %v", err)
	}
	for _, metric := range []string{
		"clipin_verifier_poll_cycles_total 1",
		"clipin_verifier_submissions_seen_total 2",
		"clipin_verifier_provider_failures_total 1",
		"clipin_verifier_snapshots_recorded_total 1",
	} {
		if !strings.Contains(output.String(), metric) {
			t.Errorf("metrics missing %q", metric)
		}
	}
}
