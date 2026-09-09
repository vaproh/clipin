package poller

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"clipin/services/verifier/internal/api"
	"clipin/services/verifier/internal/monitor"
	"clipin/services/verifier/internal/provider"
)

// Poller polls the main API for pending submissions, fetches metrics, and records snapshots.
type Poller struct {
	Client      *api.Client
	YouTubeKey  string
	BatchSize   int
	Logger      *slog.Logger
	HTTPTimeout time.Duration
	ProviderFor func(string) provider.Provider
	Monitor     *monitor.Monitor
}

// RunOnce executes a single poll cycle. It lists pending submissions, routes each
// to the appropriate provider, fetches metrics, and records successful snapshots.
// Errors on individual submissions are logged and skipped; only a list failure is returned.
func (p *Poller) RunOnce(ctx context.Context) error {
	if p.Monitor != nil {
		p.Monitor.PollStarted()
	}
	submissions, err := p.Client.ListPending(ctx, p.BatchSize)
	if err != nil {
		if p.Monitor != nil {
			p.Monitor.PollFailed()
		}
		return fmt.Errorf("list pending: %w", err)
	}
	if p.Monitor != nil {
		p.Monitor.PollSucceeded()
		p.Monitor.SubmissionsSeen(len(submissions))
	}

	if len(submissions) == 0 {
		return nil
	}

	p.Logger.Info("poll cycle", "pending", len(submissions))

	for _, sub := range submissions {
		if err := ctx.Err(); err != nil {
			return ctx.Err()
		}
		p.processSubmission(ctx, sub)
	}

	return nil
}

func (p *Poller) processSubmission(ctx context.Context, sub api.Submission) {
	providerFor := p.ProviderFor
	if providerFor == nil {
		providerFor = func(postURL string) provider.Provider {
			return provider.RouteToProvider(postURL, p.YouTubeKey, p.httpClient())
		}
	}
	prov := providerFor(sub.PostURL)
	if prov == nil {
		p.Logger.Warn("unsupported platform", "submission_id", sub.SubmissionID, "url", sub.PostURL)
		return
	}

	metrics, err := prov.Fetch(ctx, sub.PostURL)
	if err != nil {
		if p.Monitor != nil {
			p.Monitor.ProviderFailed()
		}
		p.Logger.Info("fetch failed, skipping",
			"submission_id", sub.SubmissionID,
			"url", sub.PostURL,
			"error", err,
		)
		return
	}

	snap := api.Snapshot{
		SubmissionID: sub.SubmissionID,
		Platform:     sub.Platform,
		Views:        metrics.Views,
		Likes:        metrics.Likes,
		Comments:     metrics.Comments,
		Shares:       metrics.Shares,
		CapturedAt:   time.Now().UTC().Format(time.RFC3339),
	}

	if err := p.Client.RecordSnapshot(ctx, snap); err != nil {
		p.Logger.Error("record snapshot failed",
			"submission_id", sub.SubmissionID,
			"error", err,
		)
		return
	}
	if p.Monitor != nil {
		p.Monitor.SnapshotRecorded()
	}

	p.Logger.Info("snapshot recorded",
		"submission_id", sub.SubmissionID,
		"views", metrics.Views,
	)
}

func (p *Poller) httpClient() *http.Client {
	if p.HTTPTimeout > 0 {
		return &http.Client{Timeout: p.HTTPTimeout}
	}
	return &http.Client{Timeout: 30 * time.Second}
}

// Run starts a ticker loop that calls RunOnce at the given interval.
// It blocks until the context is cancelled.
func (p *Poller) Run(ctx context.Context, interval time.Duration) {
	p.Logger.Info("poller started", "interval", interval)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Run immediately on start
	if err := p.RunOnce(ctx); err != nil {
		p.Logger.Error("initial poll failed", "error", err)
	}

	for {
		select {
		case <-ctx.Done():
			p.Logger.Info("poller stopped")
			return
		case <-ticker.C:
			if err := p.RunOnce(ctx); err != nil {
				p.Logger.Error("poll failed", "error", err)
			}
		}
	}
}
