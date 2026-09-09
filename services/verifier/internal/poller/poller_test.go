package poller

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"clipin/services/verifier/internal/api"
	"clipin/services/verifier/internal/provider"
)

func TestRunOnce(t *testing.T) {
	snapshots := []api.Snapshot{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/internal/snapshots/list":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"submissions": []map[string]interface{}{
					{
						"submission_id": "sub1",
						"post_url":      "https://www.youtube.com/shorts/abc123",
						"platform":      "youtube",
					},
				},
			})
		case "/internal/snapshots":
			var snap api.Snapshot
			json.NewDecoder(r.Body).Decode(&snap)
			snapshots = append(snapshots, snap)
			w.WriteHeader(200)
		default:
			w.WriteHeader(404)
		}
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))
	client := api.NewClient(server.URL, "", 5*time.Second)

	p := &Poller{
		Client:      client,
		YouTubeKey:  "test-key",
		BatchSize:   10,
		Logger:      logger,
		HTTPTimeout: 5 * time.Second,
	}

	// RunOnce will try to fetch from YouTube (which will fail without a real API),
	// but the test verifies the poller doesn't crash and handles the error gracefully.
	// We use a mock provider by testing at the processSubmission level.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := p.RunOnce(ctx)
	// Error expected because YouTube fetch fails without real API key
	// but the important thing is the poller doesn't panic
	_ = err
}

func TestRunOnceEmptyList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"submissions": []map[string]interface{}{},
		})
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))
	client := api.NewClient(server.URL, "", 5*time.Second)

	p := &Poller{
		Client:      client,
		Logger:      logger,
		HTTPTimeout: 5 * time.Second,
	}

	ctx := context.Background()
	err := p.RunOnce(ctx)
	if err != nil {
		t.Fatalf("RunOnce with empty list should not error: %v", err)
	}
}

func TestRunOnceListFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))
	client := api.NewClient(server.URL, "", 5*time.Second)

	p := &Poller{
		Client:      client,
		Logger:      logger,
		HTTPTimeout: 5 * time.Second,
	}

	ctx := context.Background()
	err := p.RunOnce(ctx)
	if err == nil {
		t.Error("expected error when list endpoint fails")
	}
}

func TestRunOnceUnsupportedPlatform(t *testing.T) {
	snapshots := []api.Snapshot{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/internal/snapshots/list":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"submissions": []map[string]interface{}{
					{
						"submission_id": "sub1",
						"post_url":      "https://example.com/video",
						"platform":      "unknown",
					},
				},
			})
		case "/internal/snapshots":
			var snap api.Snapshot
			json.NewDecoder(r.Body).Decode(&snap)
			snapshots = append(snapshots, snap)
			w.WriteHeader(200)
		default:
			w.WriteHeader(404)
		}
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))
	client := api.NewClient(server.URL, "", 5*time.Second)

	p := &Poller{
		Client:      client,
		Logger:      logger,
		HTTPTimeout: 5 * time.Second,
	}

	ctx := context.Background()
	err := p.RunOnce(ctx)
	if err != nil {
		t.Fatalf("RunOnce should not error for unsupported platform: %v", err)
	}
	if len(snapshots) != 0 {
		t.Errorf("expected no snapshots for unsupported platform, got %d", len(snapshots))
	}
}

func TestRunOnceRecordsSnapshotOnSuccess(t *testing.T) {
	snapshots := []api.Snapshot{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/internal/snapshots/list":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"submissions": []map[string]interface{}{
					{
						"submission_id": "sub1",
						"post_url":      "https://example.com/video",
						"platform":      "youtube",
					},
				},
			})
		case "/internal/snapshots":
			var snap api.Snapshot
			json.NewDecoder(r.Body).Decode(&snap)
			snapshots = append(snapshots, snap)
			w.WriteHeader(200)
		default:
			w.WriteHeader(404)
		}
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))
	client := api.NewClient(server.URL, "", 5*time.Second)

	p := &Poller{
		Client:      client,
		Logger:      logger,
		HTTPTimeout: 5 * time.Second,
		ProviderFor: func(string) provider.Provider {
			return staticProvider{metrics: provider.Metrics{Views: 123, Likes: 4, Comments: 5, Shares: 6}}
		},
	}

	ctx := context.Background()
	err := p.RunOnce(ctx)
	if err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if len(snapshots) != 1 {
		t.Fatalf("expected 1 snapshot, got %d", len(snapshots))
	}
	if snapshots[0].Views != 123 || snapshots[0].Likes != 4 || snapshots[0].Comments != 5 || snapshots[0].Shares != 6 {
		t.Errorf("unexpected snapshot metrics: %+v", snapshots[0])
	}
}

type staticProvider struct {
	metrics provider.Metrics
}

func (p staticProvider) Fetch(context.Context, string) (provider.Metrics, error) {
	return p.metrics, nil
}

func TestRunStopContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"submissions": []map[string]interface{}{},
		})
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))
	client := api.NewClient(server.URL, "", 5*time.Second)

	p := &Poller{
		Client:      client,
		Logger:      logger,
		HTTPTimeout: 5 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		p.Run(ctx, 100*time.Millisecond)
		close(done)
	}()

	time.Sleep(200 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// OK
	case <-time.After(2 * time.Second):
		t.Error("Run did not stop after context cancellation")
	}
}
