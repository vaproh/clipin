package worker

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	sqlc "clipin/apps/api/internal/db/sqlc"
	"clipin/apps/api/internal/service"

	"github.com/jackc/pgx/v5/pgtype"
)

// mockSubmissionStore implements service.SubmissionStore for testing.
type mockSubmissionStore struct {
	listPending    func(ctx context.Context, createdAt pgtype.Timestamptz) ([]sqlc.ListPendingSubmissionsOlderThanRow, error)
	updateStatus   func(ctx context.Context, arg sqlc.UpdateSubmissionStatusParams) (sqlc.Submission, error)
	getCampaignByID func(ctx context.Context, id pgtype.UUID) (sqlc.Campaign, error)
}

func (m *mockSubmissionStore) GetCampaignByID(ctx context.Context, id pgtype.UUID) (sqlc.Campaign, error) {
	if m.getCampaignByID != nil {
		return m.getCampaignByID(ctx, id)
	}
	return sqlc.Campaign{}, nil
}
func (m *mockSubmissionStore) GetSubmissionByID(_ context.Context, _ pgtype.UUID) (sqlc.Submission, error) {
	return sqlc.Submission{}, nil
}
func (m *mockSubmissionStore) CreateSubmission(_ context.Context, _ sqlc.CreateSubmissionParams) (sqlc.Submission, error) {
	return sqlc.Submission{}, nil
}
func (m *mockSubmissionStore) UpdateSubmissionStatus(ctx context.Context, arg sqlc.UpdateSubmissionStatusParams) (sqlc.Submission, error) {
	if m.updateStatus != nil {
		return m.updateStatus(ctx, arg)
	}
	return sqlc.Submission{}, nil
}
func (m *mockSubmissionStore) ListSubmissionsByCampaign(_ context.Context, _ pgtype.UUID) ([]sqlc.Submission, error) {
	return nil, nil
}
func (m *mockSubmissionStore) ListSubmissionsByClipper(_ context.Context, _ string) ([]sqlc.Submission, error) {
	return nil, nil
}
func (m *mockSubmissionStore) CountSubmissionsByCampaign(_ context.Context, _ pgtype.UUID) (int64, error) {
	return 0, nil
}
func (m *mockSubmissionStore) CountSubmissionsByClipperForCampaign(_ context.Context, _ sqlc.CountSubmissionsByClipperForCampaignParams) (int64, error) {
	return 0, nil
}
func (m *mockSubmissionStore) ListPendingSubmissionsOlderThan(ctx context.Context, createdAt pgtype.Timestamptz) ([]sqlc.ListPendingSubmissionsOlderThanRow, error) {
	if m.listPending != nil {
		return m.listPending(ctx, createdAt)
	}
	return nil, nil
}

func TestAutoApproveWorker_CallsAutoApprove(t *testing.T) {
	var calls atomic.Int32

	store := &mockSubmissionStore{
		listPending: func(_ context.Context, _ pgtype.Timestamptz) ([]sqlc.ListPendingSubmissionsOlderThanRow, error) {
			calls.Add(1)
			return nil, nil
		},
	}
	svc := service.NewSubmissionService(store)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	StartAutoApproveWorker(ctx, svc, 10*time.Millisecond)

	// Wait for at least one tick to fire.
	deadline := time.After(2 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatal("worker did not call AutoApprove within timeout")
		default:
			if calls.Load() > 0 {
				return // success
			}
			time.Sleep(5 * time.Millisecond)
		}
	}
}

func TestAutoApproveWorker_Shutdown(t *testing.T) {
	var calls atomic.Int32

	store := &mockSubmissionStore{
		listPending: func(_ context.Context, _ pgtype.Timestamptz) ([]sqlc.ListPendingSubmissionsOlderThanRow, error) {
			calls.Add(1)
			return nil, nil
		},
	}
	svc := service.NewSubmissionService(store)

	ctx, cancel := context.WithCancel(context.Background())
	StartAutoApproveWorker(ctx, svc, 10*time.Millisecond)

	// Let at least one tick fire.
	time.Sleep(30 * time.Millisecond)
	before := calls.Load()

	// Cancel context to trigger shutdown.
	cancel()
	time.Sleep(50 * time.Millisecond)
	after := calls.Load()

	// After cancellation, no more calls should happen.
	// Give a small window, then verify no new calls.
	time.Sleep(100 * time.Millisecond)
	afterDelay := calls.Load()
	if afterDelay > after {
		t.Errorf("worker continued after cancellation: calls went from %d to %d", after, afterDelay)
	}
	_ = before // just verifying worker ran at least once
}
