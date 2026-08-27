package service_test

import (
	"context"
	"fmt"
	"testing"

	sqlc "clipin/apps/api/internal/db/sqlc"
	"clipin/apps/api/internal/payout"
	"clipin/apps/api/internal/service"

	"github.com/jackc/pgx/v5/pgtype"
)

// mockPayoutStore implements PayoutStore for testing.
type mockPayoutStore struct {
	getUserByID           func(ctx context.Context, id string) (sqlc.User, error)
	updateUserUPI         func(ctx context.Context, arg sqlc.UpdateUserUPIParams) (sqlc.User, error)
	createPayoutRequest   func(ctx context.Context, arg sqlc.CreatePayoutRequestParams) (sqlc.PayoutRequest, error)
	getPayoutRequestByID  func(ctx context.Context, id pgtype.UUID) (sqlc.PayoutRequest, error)
	listPayoutRequestsByClipper func(ctx context.Context, clipperID string) ([]sqlc.PayoutRequest, error)
	updatePayoutRequestStatus   func(ctx context.Context, arg sqlc.UpdatePayoutRequestStatusParams) (sqlc.PayoutRequest, error)
	sumPendingPayoutsByClipper  func(ctx context.Context, clipperID string) (int64, error)
	sumEarningsByClipper  func(ctx context.Context, clipperID pgtype.Text) (int64, error)
	createLedgerEntry     func(ctx context.Context, arg sqlc.CreateLedgerEntryParams) (sqlc.LedgerEntry, error)
	getLedgerEntryByIdempotencyKey func(ctx context.Context, idempotencyKey string) (sqlc.LedgerEntry, error)
	listPendingPayoutRequests func(ctx context.Context) ([]sqlc.PayoutRequest, error)
}

func (m *mockPayoutStore) GetUserByID(ctx context.Context, id string) (sqlc.User, error) {
	return m.getUserByID(ctx, id)
}
func (m *mockPayoutStore) UpdateUserUPI(ctx context.Context, arg sqlc.UpdateUserUPIParams) (sqlc.User, error) {
	return m.updateUserUPI(ctx, arg)
}
func (m *mockPayoutStore) CreatePayoutRequest(ctx context.Context, arg sqlc.CreatePayoutRequestParams) (sqlc.PayoutRequest, error) {
	return m.createPayoutRequest(ctx, arg)
}
func (m *mockPayoutStore) GetPayoutRequestByID(ctx context.Context, id pgtype.UUID) (sqlc.PayoutRequest, error) {
	return m.getPayoutRequestByID(ctx, id)
}
func (m *mockPayoutStore) ListPayoutRequestsByClipper(ctx context.Context, clipperID string) ([]sqlc.PayoutRequest, error) {
	return m.listPayoutRequestsByClipper(ctx, clipperID)
}
func (m *mockPayoutStore) UpdatePayoutRequestStatus(ctx context.Context, arg sqlc.UpdatePayoutRequestStatusParams) (sqlc.PayoutRequest, error) {
	return m.updatePayoutRequestStatus(ctx, arg)
}
func (m *mockPayoutStore) SumPendingPayoutsByClipper(ctx context.Context, clipperID string) (int64, error) {
	return m.sumPendingPayoutsByClipper(ctx, clipperID)
}
func (m *mockPayoutStore) SumEarningsByClipper(ctx context.Context, clipperID pgtype.Text) (int64, error) {
	return m.sumEarningsByClipper(ctx, clipperID)
}
func (m *mockPayoutStore) CreateLedgerEntry(ctx context.Context, arg sqlc.CreateLedgerEntryParams) (sqlc.LedgerEntry, error) {
	return m.createLedgerEntry(ctx, arg)
}
func (m *mockPayoutStore) GetLedgerEntryByIdempotencyKey(ctx context.Context, idempotencyKey string) (sqlc.LedgerEntry, error) {
	return m.getLedgerEntryByIdempotencyKey(ctx, idempotencyKey)
}
func (m *mockPayoutStore) ListPendingPayoutRequests(ctx context.Context) ([]sqlc.PayoutRequest, error) {
	return m.listPendingPayoutRequests(ctx)
}

func testUserWithUPI(upiID string) sqlc.User {
	u := sqlc.User{
		ID:    "clipper1",
		Email: "clipper@test.com",
		Role:  "clipper",
	}
	if upiID != "" {
		u.UpiID = pgtype.Text{Valid: true, String: upiID}
	}
	return u
}

func testPayoutRequest(id pgtype.UUID, status string, amount int32) sqlc.PayoutRequest {
	return sqlc.PayoutRequest{
		ID:             id,
		ClipperID:      "clipper1",
		Amount:         amount,
		UpiID:          "test@upi",
		Status:         status,
		IdempotencyKey: "idem-1",
	}
}

// --- UpdateUPI tests ---

func TestUpdateUPI_HappyPath(t *testing.T) {
	store := &mockPayoutStore{
		updateUserUPI: func(ctx context.Context, arg sqlc.UpdateUserUPIParams) (sqlc.User, error) {
			if arg.ID != "clipper1" {
				t.Errorf("expected user clipper1, got %s", arg.ID)
			}
			if !arg.UpiID.Valid || arg.UpiID.String != "new@upi" {
				t.Errorf("expected upi_id new@upi, got %v", arg.UpiID)
			}
			return testUserWithUPI("new@upi"), nil
		},
	}
	svc := service.NewPayoutService(store, &payout.RazorpayStub{})
	user, err := svc.UpdateUPI(context.Background(), "clipper1", "new@upi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !user.UpiID.Valid || user.UpiID.String != "new@upi" {
		t.Errorf("expected upi_id new@upi, got %s", user.UpiID.String)
	}
}

func TestUpdateUPI_EmptyUpiID(t *testing.T) {
	svc := service.NewPayoutService(&mockPayoutStore{}, &payout.RazorpayStub{})
	_, err := svc.UpdateUPI(context.Background(), "clipper1", "")
	if err == nil {
		t.Error("expected error for empty UPI ID")
	}
}

func TestUpdateUPI_StoreError(t *testing.T) {
	store := &mockPayoutStore{
		updateUserUPI: func(ctx context.Context, arg sqlc.UpdateUserUPIParams) (sqlc.User, error) {
			return sqlc.User{}, fmt.Errorf("db error")
		},
	}
	svc := service.NewPayoutService(store, &payout.RazorpayStub{})
	_, err := svc.UpdateUPI(context.Background(), "clipper1", "test@upi")
	if err == nil {
		t.Error("expected error from store")
	}
}

func TestUpdateUPI_InvalidFormat_NoAt(t *testing.T) {
	svc := service.NewPayoutService(&mockPayoutStore{}, &payout.RazorpayStub{})
	_, err := svc.UpdateUPI(context.Background(), "clipper1", "invalidupi")
	if err == nil {
		t.Error("expected error for UPI ID without @")
	}
}

func TestUpdateUPI_InvalidFormat_TooShort(t *testing.T) {
	svc := service.NewPayoutService(&mockPayoutStore{}, &payout.RazorpayStub{})
	_, err := svc.UpdateUPI(context.Background(), "clipper1", "a@")
	if err == nil {
		t.Error("expected error for too-short UPI ID")
	}
}

func TestUpdateUPI_InvalidFormat_TooLong(t *testing.T) {
	svc := service.NewPayoutService(&mockPayoutStore{}, &payout.RazorpayStub{})
	longUpi := string(make([]byte, 40))
	_, err := svc.UpdateUPI(context.Background(), "clipper1", longUpi)
	if err == nil {
		t.Error("expected error for too-long UPI ID")
	}
}

// --- RequestPayout tests ---

func TestRequestPayout_HappyPath(t *testing.T) {
	var capturedStatus sqlc.UpdatePayoutRequestStatusParams
	var capturedLedger sqlc.CreateLedgerEntryParams
	store := &mockPayoutStore{
		getUserByID: func(ctx context.Context, id string) (sqlc.User, error) {
			return testUserWithUPI("test@upi"), nil
		},
		sumEarningsByClipper: func(ctx context.Context, clipperID pgtype.Text) (int64, error) {
			return 100000, nil // ₹1000 earned
		},
		sumPendingPayoutsByClipper: func(ctx context.Context, clipperID string) (int64, error) {
			return 0, nil
		},
		createPayoutRequest: func(ctx context.Context, arg sqlc.CreatePayoutRequestParams) (sqlc.PayoutRequest, error) {
			return testPayoutRequest(arg.ID, "pending", arg.Amount), nil
		},
		updatePayoutRequestStatus: func(ctx context.Context, arg sqlc.UpdatePayoutRequestStatusParams) (sqlc.PayoutRequest, error) {
			capturedStatus = arg
			return testPayoutRequest(arg.ID, arg.Status, 50000), nil
		},
		createLedgerEntry: func(ctx context.Context, arg sqlc.CreateLedgerEntryParams) (sqlc.LedgerEntry, error) {
			capturedLedger = arg
			return sqlc.LedgerEntry{}, nil
		},
	}
	svc := service.NewPayoutService(store, &payout.RazorpayStub{})
	pr, err := svc.RequestPayout(context.Background(), "clipper1", 50000, "idem-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pr.Status != "completed" {
		t.Errorf("expected status completed, got %s", pr.Status)
	}
	if capturedStatus.Status != "completed" {
		t.Errorf("expected status update to completed, got %s", capturedStatus.Status)
	}
	// Verify ledger entry was created for payout.
	if capturedLedger.EntryType != "payout" {
		t.Errorf("expected ledger entry type 'payout', got %s", capturedLedger.EntryType)
	}
}

func TestRequestPayout_BelowThreshold(t *testing.T) {
	svc := service.NewPayoutService(&mockPayoutStore{}, &payout.RazorpayStub{})
	_, err := svc.RequestPayout(context.Background(), "clipper1", 49999, "idem-1")
	if err != service.ErrBelowThreshold {
		t.Errorf("expected ErrBelowThreshold, got %v", err)
	}
}

func TestRequestPayout_ExactlyThreshold(t *testing.T) {
	store := &mockPayoutStore{
		getUserByID: func(ctx context.Context, id string) (sqlc.User, error) {
			return testUserWithUPI("test@upi"), nil
		},
		sumEarningsByClipper: func(ctx context.Context, clipperID pgtype.Text) (int64, error) {
			return 100000, nil
		},
		sumPendingPayoutsByClipper: func(ctx context.Context, clipperID string) (int64, error) {
			return 0, nil
		},
		createPayoutRequest: func(ctx context.Context, arg sqlc.CreatePayoutRequestParams) (sqlc.PayoutRequest, error) {
			return testPayoutRequest(arg.ID, "pending", arg.Amount), nil
		},
		updatePayoutRequestStatus: func(ctx context.Context, arg sqlc.UpdatePayoutRequestStatusParams) (sqlc.PayoutRequest, error) {
			return testPayoutRequest(arg.ID, arg.Status, 50000), nil
		},
		createLedgerEntry: func(ctx context.Context, arg sqlc.CreateLedgerEntryParams) (sqlc.LedgerEntry, error) {
			return sqlc.LedgerEntry{}, nil
		},
	}
	svc := service.NewPayoutService(store, &payout.RazorpayStub{})
	_, err := svc.RequestPayout(context.Background(), "clipper1", 50000, "idem-2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRequestPayout_NoUPIID(t *testing.T) {
	store := &mockPayoutStore{
		getUserByID: func(ctx context.Context, id string) (sqlc.User, error) {
			return testUserWithUPI(""), nil
		},
	}
	svc := service.NewPayoutService(store, &payout.RazorpayStub{})
	_, err := svc.RequestPayout(context.Background(), "clipper1", 50000, "idem-3")
	if err != service.ErrNoUPIID {
		t.Errorf("expected ErrNoUPIID, got %v", err)
	}
}

func TestRequestPayout_InsufficientFunds(t *testing.T) {
	store := &mockPayoutStore{
		getUserByID: func(ctx context.Context, id string) (sqlc.User, error) {
			return testUserWithUPI("test@upi"), nil
		},
		sumEarningsByClipper: func(ctx context.Context, clipperID pgtype.Text) (int64, error) {
			return 10000, nil // only ₹100 earned
		},
		sumPendingPayoutsByClipper: func(ctx context.Context, clipperID string) (int64, error) {
			return 0, nil
		},
	}
	svc := service.NewPayoutService(store, &payout.RazorpayStub{})
	_, err := svc.RequestPayout(context.Background(), "clipper1", 50000, "idem-4")
	if err != service.ErrInsufficientFunds {
		t.Errorf("expected ErrInsufficientFunds, got %v", err)
	}
}

func TestRequestPayout_InsufficientFunds_WithPending(t *testing.T) {
	store := &mockPayoutStore{
		getUserByID: func(ctx context.Context, id string) (sqlc.User, error) {
			return testUserWithUPI("test@upi"), nil
		},
		sumEarningsByClipper: func(ctx context.Context, clipperID pgtype.Text) (int64, error) {
			return 100000, nil // ₹1000 earned
		},
		sumPendingPayoutsByClipper: func(ctx context.Context, clipperID string) (int64, error) {
			return 70000, nil // ₹700 pending
		},
	}
	svc := service.NewPayoutService(store, &payout.RazorpayStub{})
	_, err := svc.RequestPayout(context.Background(), "clipper1", 50000, "idem-5")
	// available = 100000 - 70000 = 30000, < 50000
	if err != service.ErrInsufficientFunds {
		t.Errorf("expected ErrInsufficientFunds, got %v", err)
	}
}

func TestRequestPayout_ProviderFailure(t *testing.T) {
	store := &mockPayoutStore{
		getUserByID: func(ctx context.Context, id string) (sqlc.User, error) {
			return testUserWithUPI("test@upi"), nil
		},
		sumEarningsByClipper: func(ctx context.Context, clipperID pgtype.Text) (int64, error) {
			return 100000, nil
		},
		sumPendingPayoutsByClipper: func(ctx context.Context, clipperID string) (int64, error) {
			return 0, nil
		},
		createPayoutRequest: func(ctx context.Context, arg sqlc.CreatePayoutRequestParams) (sqlc.PayoutRequest, error) {
			return testPayoutRequest(arg.ID, "pending", arg.Amount), nil
		},
		updatePayoutRequestStatus: func(ctx context.Context, arg sqlc.UpdatePayoutRequestStatusParams) (sqlc.PayoutRequest, error) {
			return testPayoutRequest(arg.ID, arg.Status, 50000), nil
		},
	}
	// Use a provider that always fails.
	failProvider := &failingProvider{err: fmt.Errorf("provider down")}
	svc := service.NewPayoutService(store, failProvider)
	_, err := svc.RequestPayout(context.Background(), "clipper1", 50000, "idem-6")
	if err == nil {
		t.Error("expected error from provider")
	}
}

func TestRequestPayout_Idempotent(t *testing.T) {
	callCount := 0
	store := &mockPayoutStore{
		getUserByID: func(ctx context.Context, id string) (sqlc.User, error) {
			return testUserWithUPI("test@upi"), nil
		},
		sumEarningsByClipper: func(ctx context.Context, clipperID pgtype.Text) (int64, error) {
			return 100000, nil
		},
		sumPendingPayoutsByClipper: func(ctx context.Context, clipperID string) (int64, error) {
			return 0, nil
		},
		createPayoutRequest: func(ctx context.Context, arg sqlc.CreatePayoutRequestParams) (sqlc.PayoutRequest, error) {
			// First call: return zero-value (simulating DO NOTHING).
			callCount++
			if callCount == 1 {
				return sqlc.PayoutRequest{}, nil
			}
			return testPayoutRequest(arg.ID, "pending", arg.Amount), nil
		},
		listPayoutRequestsByClipper: func(ctx context.Context, clipperID string) ([]sqlc.PayoutRequest, error) {
			return []sqlc.PayoutRequest{
				{ID: pgtype.UUID{Bytes: [16]byte{1}, Valid: true}, IdempotencyKey: "idem-7", Status: "completed", Amount: 50000},
			}, nil
		},
	}
	svc := service.NewPayoutService(store, &payout.RazorpayStub{})
	pr, err := svc.RequestPayout(context.Background(), "clipper1", 50000, "idem-7")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pr.IdempotencyKey != "idem-7" {
		t.Errorf("expected idempotency key idem-7, got %s", pr.IdempotencyKey)
	}
}

// --- ListMyPayouts tests ---

func TestListMyPayouts(t *testing.T) {
	expected := []sqlc.PayoutRequest{
		{Status: "completed", Amount: 50000},
		{Status: "pending", Amount: 75000},
	}
	store := &mockPayoutStore{
		listPayoutRequestsByClipper: func(ctx context.Context, clipperID string) ([]sqlc.PayoutRequest, error) {
			return expected, nil
		},
	}
	svc := service.NewPayoutService(store, &payout.RazorpayStub{})
	payouts, err := svc.ListMyPayouts(context.Background(), "clipper1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(payouts) != 2 {
		t.Fatalf("expected 2 payouts, got %d", len(payouts))
	}
}

func TestListMyPayouts_Empty(t *testing.T) {
	store := &mockPayoutStore{
		listPayoutRequestsByClipper: func(ctx context.Context, clipperID string) ([]sqlc.PayoutRequest, error) {
			return []sqlc.PayoutRequest{}, nil
		},
	}
	svc := service.NewPayoutService(store, &payout.RazorpayStub{})
	payouts, err := svc.ListMyPayouts(context.Background(), "clipper1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(payouts) != 0 {
		t.Errorf("expected 0 payouts, got %d", len(payouts))
	}
}

// --- ProcessPendingPayouts tests ---

func TestProcessPendingPayouts_HappyPath(t *testing.T) {
	var processedIDs []string
	store := &mockPayoutStore{
		listPendingPayoutRequests: func(ctx context.Context) ([]sqlc.PayoutRequest, error) {
			return []sqlc.PayoutRequest{
				{
					ID:             pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
					ClipperID:      "clipper1",
					Amount:         50000,
					UpiID:          "test@upi",
					Status:         "pending",
					IdempotencyKey: "p1",
				},
				{
					ID:             pgtype.UUID{Bytes: [16]byte{2}, Valid: true},
					ClipperID:      "clipper2",
					Amount:         75000,
					UpiID:          "user2@upi",
					Status:         "pending",
					IdempotencyKey: "p2",
				},
			}, nil
		},
		updatePayoutRequestStatus: func(ctx context.Context, arg sqlc.UpdatePayoutRequestStatusParams) (sqlc.PayoutRequest, error) {
			processedIDs = append(processedIDs, arg.Status)
			return testPayoutRequest(arg.ID, arg.Status, 50000), nil
		},
		createLedgerEntry: func(ctx context.Context, arg sqlc.CreateLedgerEntryParams) (sqlc.LedgerEntry, error) {
			return sqlc.LedgerEntry{}, nil
		},
	}
	svc := service.NewPayoutService(store, &payout.RazorpayStub{})
	count, err := svc.ProcessPendingPayouts(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 processed, got %d", count)
	}
}

func TestProcessPendingPayouts_Empty(t *testing.T) {
	store := &mockPayoutStore{
		listPendingPayoutRequests: func(ctx context.Context) ([]sqlc.PayoutRequest, error) {
			return []sqlc.PayoutRequest{}, nil
		},
	}
	svc := service.NewPayoutService(store, &payout.RazorpayStub{})
	count, err := svc.ProcessPendingPayouts(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 processed, got %d", count)
	}
}

func TestProcessPendingPayouts_ProviderFailure(t *testing.T) {
	store := &mockPayoutStore{
		listPendingPayoutRequests: func(ctx context.Context) ([]sqlc.PayoutRequest, error) {
			return []sqlc.PayoutRequest{
				{
					ID:             pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
					ClipperID:      "clipper1",
					Amount:         50000,
					UpiID:          "test@upi",
					Status:         "pending",
					IdempotencyKey: "p1",
				},
			}, nil
		},
		updatePayoutRequestStatus: func(ctx context.Context, arg sqlc.UpdatePayoutRequestStatusParams) (sqlc.PayoutRequest, error) {
			return testPayoutRequest(arg.ID, arg.Status, 50000), nil
		},
	}
	failProvider := &failingProvider{err: fmt.Errorf("network error")}
	svc := service.NewPayoutService(store, failProvider)
	count, err := svc.ProcessPendingPayouts(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 processed (provider failed), got %d", count)
	}
}

// --- RequestPayout negative amount ---

func TestRequestPayout_NegativeAmount(t *testing.T) {
	svc := service.NewPayoutService(&mockPayoutStore{}, &payout.RazorpayStub{})
	_, err := svc.RequestPayout(context.Background(), "clipper1", -100, "idem-neg")
	if err != service.ErrBelowThreshold {
		t.Errorf("expected ErrBelowThreshold, got %v", err)
	}
}

// --- RequestPayout zero amount ---

func TestRequestPayout_ZeroAmount(t *testing.T) {
	svc := service.NewPayoutService(&mockPayoutStore{}, &payout.RazorpayStub{})
	_, err := svc.RequestPayout(context.Background(), "clipper1", 0, "idem-zero")
	if err != service.ErrBelowThreshold {
		t.Errorf("expected ErrBelowThreshold, got %v", err)
	}
}

// failingProvider always returns an error.
type failingProvider struct {
	err error
}

func (f *failingProvider) CreateTransfer(ctx context.Context, req payout.TransferRequest) (*payout.TransferResponse, error) {
	return nil, f.err
}

func (f *failingProvider) GetTransferStatus(ctx context.Context, transferID string) (*payout.TransferResponse, error) {
	return nil, f.err
}
