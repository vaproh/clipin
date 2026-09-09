package monitor

import (
	"fmt"
	"io"
	"sync/atomic"
	"time"
)

// Monitor stores process-local verifier counters for health and scraping.
type Monitor struct {
	pollCycles        atomic.Int64
	pollFailures      atomic.Int64
	submissionsSeen   atomic.Int64
	providerFailures  atomic.Int64
	snapshotsRecorded atomic.Int64
	lastPollUnix      atomic.Int64
	lastSuccessUnix   atomic.Int64
}

func (m *Monitor) PollStarted() {
	m.pollCycles.Add(1)
	m.lastPollUnix.Store(time.Now().Unix())
}

func (m *Monitor) PollFailed() {
	m.pollFailures.Add(1)
}

func (m *Monitor) PollSucceeded() {
	m.lastSuccessUnix.Store(time.Now().Unix())
}

func (m *Monitor) SubmissionsSeen(n int) {
	m.submissionsSeen.Add(int64(n))
}

func (m *Monitor) ProviderFailed() {
	m.providerFailures.Add(1)
}

func (m *Monitor) SnapshotRecorded() {
	m.snapshotsRecorded.Add(1)
}

// WritePrometheus writes a dependency-free Prometheus text exposition.
func (m *Monitor) WritePrometheus(w io.Writer) error {
	values := []struct {
		name  string
		help  string
		type_ string
		value int64
	}{
		{"clipin_verifier_poll_cycles_total", "Completed poll cycles started by the verifier.", "counter", m.pollCycles.Load()},
		{"clipin_verifier_poll_failures_total", "Poll cycles that failed to list submissions.", "counter", m.pollFailures.Load()},
		{"clipin_verifier_submissions_seen_total", "Submissions returned by the main API.", "counter", m.submissionsSeen.Load()},
		{"clipin_verifier_provider_failures_total", "Submission metric fetches that failed.", "counter", m.providerFailures.Load()},
		{"clipin_verifier_snapshots_recorded_total", "Snapshots accepted by the main API.", "counter", m.snapshotsRecorded.Load()},
		{"clipin_verifier_last_poll_timestamp_seconds", "Unix timestamp of the latest poll cycle.", "gauge", m.lastPollUnix.Load()},
		{"clipin_verifier_last_success_timestamp_seconds", "Unix timestamp of the latest successful list request.", "gauge", m.lastSuccessUnix.Load()},
	}
	for _, value := range values {
		_, err := fmt.Fprintf(w, "# HELP %s %s\n# TYPE %s %s\n%s %d\n", value.name, value.help, value.name, value.type_, value.name, value.value)
		if err != nil {
			return err
		}
	}
	return nil
}
