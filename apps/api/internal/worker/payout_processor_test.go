package worker

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	sqlc "clipin/apps/api/internal/db/sqlc"
	"clipin/apps/api/internal/payout"
	"clipin/apps/api/internal/service"

	"github.com/jackc/pgx/v5/pgtype"
)

// mockPayoutStore implements service.PayoutStore for testing.
type mockPayoutStore struct {
	listPendingPayoutRequests func(ctx context.Context) ([]sqlc.PayoutRequest, error)
}

func (m *mockPayoutStore) GetUserByID(_ context.Context, _ string) (sqlc.User, error) {
	return sqlc.User{}, nil
}
func (m *mockPayoutStore) UpdateUserUPI(_ context.Context, _ sqlc.UpdateUserUPIParams) (sqlc.User, error) {
	return sqlc.User{}, nil
}
func (m *mockPayoutStore) CreatePayoutRequest(_ context.Context, _ sqlc.CreatePayoutRequestParams) (sqlc.PayoutRequest, error) {
	return sqlc.PayoutRequest{}, nil
}
func (m *mockPayoutStore) GetPayoutRequestByID(_ context.Context, _ pgtype.UUID) (sqlc.PayoutRequest, error) {
	return sqlc.PayoutRequest{}, nil
}
func (m *mockPayoutStore) ListPayoutRequestsByClipper(_ context.Context, _ string) ([]sqlc.PayoutRequest, error) {
	return nil, nil
}
func (m *mockPayoutStore) UpdatePayoutRequestStatus(_ context.Context, _ sqlc.UpdatePayoutRequestStatusParams) (sqlc.PayoutRequest, error) {
	return sqlc.PayoutRequest{}, nil
}
func (m *mockPayoutStore) SumPendingPayoutsByClipper(_ context.Context, _ string) (int64, error) {
	return 0, nil
}
func (m *mockPayoutStore) SumEarningsByClipper(_ context.Context, _ pgtype.Text) (int64, error) {
	return 0, nil
}
func (m *mockPayoutStore) CreateLedgerEntry(_ context.Context, _ sqlc.CreateLedgerEntryParams) (sqlc.LedgerEntry, error) {
	return sqlc.LedgerEntry{}, nil
}
func (m *mockPayoutStore) GetLedgerEntryByIdempotencyKey(_ context.Context, _ string) (sqlc.LedgerEntry, error) {
	return sqlc.LedgerEntry{}, nil
}
func (m *mockPayoutStore) ListPendingPayoutRequests(ctx context.Context) ([]sqlc.PayoutRequest, error) {
	if m.listPendingPayoutRequests != nil {
		return m.listPendingPayoutRequests(ctx)
	}
	return nil, nil
}

// stubPayoutProvider is a no-op PayoutProvider for testing.
type stubPayoutProvider struct{}

func (s *stubPayoutProvider) CreateTransfer(_ context.Context, _ payout.TransferRequest) (*payout.TransferResponse, error) {
	return &payout.TransferResponse{Status: "completed"}, nil
}
func (s *stubPayoutProvider) GetTransferStatus(_ context.Context, _ string) (*payout.TransferResponse, error) {
	return &payout.TransferResponse{Status: "completed"}, nil
}

func TestPayoutProcessorWorker_CallsProcessPendingPayouts(t *testing.T) {
	var calls atomic.Int32

	store := &mockPayoutStore{
		listPendingPayoutRequests: func(_ context.Context) ([]sqlc.PayoutRequest, error) {
			calls.Add(1)
			return nil, nil
		},
	}
	svc := service.NewPayoutService(store, &stubPayoutProvider{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	StartPayoutProcessorWorker(ctx, svc, 10*time.Millisecond)

	deadline := time.After(2 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatal("worker did not call ProcessPendingPayouts within timeout")
		default:
			if calls.Load() > 0 {
				return
			}
			time.Sleep(5 * time.Millisecond)
		}
	}
}

func TestPayoutProcessorWorker_Shutdown(t *testing.T) {
	var calls atomic.Int32

	store := &mockPayoutStore{
		listPendingPayoutRequests: func(_ context.Context) ([]sqlc.PayoutRequest, error) {
			calls.Add(1)
			return nil, nil
		},
	}
	svc := service.NewPayoutService(store, &stubPayoutProvider{})

	ctx, cancel := context.WithCancel(context.Background())
	StartPayoutProcessorWorker(ctx, svc, 10*time.Millisecond)

	// Let at least one tick fire.
	time.Sleep(30 * time.Millisecond)
	before := calls.Load()

	cancel()
	time.Sleep(50 * time.Millisecond)
	after := calls.Load()

	// Verify no new calls after cancellation.
	time.Sleep(100 * time.Millisecond)
	afterDelay := calls.Load()
	if afterDelay > after {
		t.Errorf("worker continued after cancellation: calls went from %d to %d", after, afterDelay)
	}
	_ = before
}
