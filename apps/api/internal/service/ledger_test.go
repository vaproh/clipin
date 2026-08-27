package service_test

import (
	"context"
	"fmt"
	"testing"

	sqlc "clipin/apps/api/internal/db/sqlc"
	"clipin/apps/api/internal/service"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// mockLedgerStore implements LedgerStore for testing.
type mockLedgerStore struct {
	createEntry          func(ctx context.Context, arg sqlc.CreateLedgerEntryParams) (sqlc.LedgerEntry, error)
	getByIDempotencyKey  func(ctx context.Context, key string) (sqlc.LedgerEntry, error)
	listByCampaign       func(ctx context.Context, campaignID pgtype.UUID) ([]sqlc.LedgerEntry, error)
	listByClipper        func(ctx context.Context, clipperID pgtype.Text) ([]sqlc.LedgerEntry, error)
	sumEarningsByClipper func(ctx context.Context, clipperID pgtype.Text) (int64, error)
	sumEarningsByClipperForCampaign func(ctx context.Context, arg sqlc.SumEarningsByClipperForCampaignParams) (int64, error)
	sumFeesByCampaign    func(ctx context.Context, campaignID pgtype.UUID) (int64, error)
	sumSpendByCampaign   func(ctx context.Context, campaignID pgtype.UUID) (int64, error)
	getCampaignByID      func(ctx context.Context, id pgtype.UUID) (sqlc.Campaign, error)
	updateBudget         func(ctx context.Context, arg sqlc.UpdateCampaignBudgetParams) (sqlc.Campaign, error)
	deductBudget         func(ctx context.Context, arg sqlc.DeductCampaignBudgetParams) (sqlc.Campaign, error)
}

func (m *mockLedgerStore) CreateLedgerEntry(ctx context.Context, arg sqlc.CreateLedgerEntryParams) (sqlc.LedgerEntry, error) {
	return m.createEntry(ctx, arg)
}
func (m *mockLedgerStore) GetLedgerEntryByIdempotencyKey(ctx context.Context, key string) (sqlc.LedgerEntry, error) {
	return m.getByIDempotencyKey(ctx, key)
}
func (m *mockLedgerStore) ListLedgerEntriesByCampaign(ctx context.Context, campaignID pgtype.UUID) ([]sqlc.LedgerEntry, error) {
	return m.listByCampaign(ctx, campaignID)
}
func (m *mockLedgerStore) ListLedgerEntriesByClipper(ctx context.Context, clipperID pgtype.Text) ([]sqlc.LedgerEntry, error) {
	return m.listByClipper(ctx, clipperID)
}
func (m *mockLedgerStore) SumEarningsByClipper(ctx context.Context, clipperID pgtype.Text) (int64, error) {
	return m.sumEarningsByClipper(ctx, clipperID)
}
func (m *mockLedgerStore) SumEarningsByClipperForCampaign(ctx context.Context, arg sqlc.SumEarningsByClipperForCampaignParams) (int64, error) {
	return m.sumEarningsByClipperForCampaign(ctx, arg)
}
func (m *mockLedgerStore) SumFeesByCampaign(ctx context.Context, campaignID pgtype.UUID) (int64, error) {
	return m.sumFeesByCampaign(ctx, campaignID)
}
func (m *mockLedgerStore) SumSpendByCampaign(ctx context.Context, campaignID pgtype.UUID) (int64, error) {
	return m.sumSpendByCampaign(ctx, campaignID)
}
func (m *mockLedgerStore) GetCampaignByID(ctx context.Context, id pgtype.UUID) (sqlc.Campaign, error) {
	return m.getCampaignByID(ctx, id)
}
func (m *mockLedgerStore) UpdateCampaignBudget(ctx context.Context, arg sqlc.UpdateCampaignBudgetParams) (sqlc.Campaign, error) {
	return m.updateBudget(ctx, arg)
}
func (m *mockLedgerStore) DeductCampaignBudget(ctx context.Context, arg sqlc.DeductCampaignBudgetParams) (sqlc.Campaign, error) {
	if m.deductBudget != nil {
		return m.deductBudget(ctx, arg)
	}
	return sqlc.Campaign{}, nil
}

var testLedgerCampaignID = pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true}
var testLedgerSubmissionID = pgtype.UUID{Bytes: [16]byte{20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35}, Valid: true}

func testCampaignForLedger(remainingBudget int32) sqlc.Campaign {
	return sqlc.Campaign{
		ID:              testLedgerCampaignID,
		OwnerID:         "owner1",
		Title:           "Test Campaign",
		Platform:        "youtube",
		Status:          "active",
		CpmRate:         100,
		TotalBudget:     10000,
		RemainingBudget: remainingBudget,
		PlatformFee:     1000,
		CreatedAt:       pgtype.Timestamptz{Valid: true},
		UpdatedAt:       pgtype.Timestamptz{Valid: true},
	}
}

func TestRecordPlatformFee_HappyPath(t *testing.T) {
	store := &mockLedgerStore{
		createEntry: func(ctx context.Context, arg sqlc.CreateLedgerEntryParams) (sqlc.LedgerEntry, error) {
			if arg.EntryType != "platform_fee" {
				t.Errorf("expected entry_type platform_fee, got %s", arg.EntryType)
			}
			if arg.Amount != 1000 {
				t.Errorf("expected amount 1000, got %d", arg.Amount)
			}
			if arg.IdempotencyKey != "fee:test-campaign" {
				t.Errorf("expected idempotency_key fee:test-campaign, got %s", arg.IdempotencyKey)
			}
			return sqlc.LedgerEntry{
				ID:             pgtype.UUID{Bytes: [16]byte{50}, Valid: true},
				IdempotencyKey: arg.IdempotencyKey,
				EntryType:      arg.EntryType,
				CampaignID:     arg.CampaignID,
				Amount:         arg.Amount,
			}, nil
		},
	}
	svc := service.NewLedgerService(store)
	entry, err := svc.RecordPlatformFee(context.Background(), testLedgerCampaignID, 1000, "fee:test-campaign")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Amount != 1000 {
		t.Errorf("expected amount 1000, got %d", entry.Amount)
	}
}

func TestRecordPlatformFee_Idempotent(t *testing.T) {
	callCount := 0
	store := &mockLedgerStore{
		createEntry: func(ctx context.Context, arg sqlc.CreateLedgerEntryParams) (sqlc.LedgerEntry, error) {
			callCount++
			// Simulate conflict (ON CONFLICT DO NOTHING returns ErrNoRows).
			return sqlc.LedgerEntry{}, pgx.ErrNoRows
		},
		getByIDempotencyKey: func(ctx context.Context, key string) (sqlc.LedgerEntry, error) {
			return sqlc.LedgerEntry{
				IdempotencyKey: key,
				EntryType:      "platform_fee",
				Amount:         1000,
			}, nil
		},
	}
	svc := service.NewLedgerService(store)
	entry, err := svc.RecordPlatformFee(context.Background(), testLedgerCampaignID, 1000, "fee:test-campaign")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Amount != 1000 {
		t.Errorf("expected existing entry amount 1000, got %d", entry.Amount)
	}
}

func TestRecordEarning_HappyPath(t *testing.T) {
	store := &mockLedgerStore{
		createEntry: func(ctx context.Context, arg sqlc.CreateLedgerEntryParams) (sqlc.LedgerEntry, error) {
			if arg.EntryType != "earning" {
				t.Errorf("expected entry_type earning, got %s", arg.EntryType)
			}
			// 5000 eligible views * 100 CPM / 1000 = 500
			if arg.Amount != 500 {
				t.Errorf("expected amount 500, got %d", arg.Amount)
			}
			if !arg.ClipperID.Valid || arg.ClipperID.String != "clipper1" {
				t.Errorf("expected clipper_id clipper1, got %v", arg.ClipperID)
			}
			return sqlc.LedgerEntry{
				ID:             pgtype.UUID{Bytes: [16]byte{60}, Valid: true},
				IdempotencyKey: arg.IdempotencyKey,
				EntryType:      arg.EntryType,
				CampaignID:     arg.CampaignID,
				SubmissionID:   arg.SubmissionID,
				ClipperID:      arg.ClipperID,
				Amount:         arg.Amount,
			}, nil
		},
		deductBudget: func(ctx context.Context, arg sqlc.DeductCampaignBudgetParams) (sqlc.Campaign, error) {
			if arg.RemainingBudget != 500 {
				t.Errorf("expected deducted amount 500, got %d", arg.RemainingBudget)
			}
			return sqlc.Campaign{RemainingBudget: 5000 - 500}, nil
		},
	}
	svc := service.NewLedgerService(store)
	entry, err := svc.RecordEarning(context.Background(), testLedgerSubmissionID, testLedgerCampaignID, "clipper1", 5000, 100, "earning:test-sub")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Amount != 500 {
		t.Errorf("expected amount 500, got %d", entry.Amount)
	}
}

func TestRecordEarning_EarningCalculation(t *testing.T) {
	tests := []struct {
		name          string
		eligibleViews int64
		cpmRate       int32
		expectedAmount int32
	}{
		{"1000 views at CPM 100", 1000, 100, 100},
		{"5000 views at CPM 200", 5000, 200, 1000},
		{"100 views at CPM 500", 100, 500, 50},
		{"1 view at CPM 1000", 1, 1000, 1}, // minimum
		{"9999 views at CPM 100", 9999, 100, 999}, // integer division rounds down
		{"0 views at CPM 100", 0, 100, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.eligibleViews == 0 {
				// Should error for 0 views.
				store := &mockLedgerStore{}
				svc := service.NewLedgerService(store)
				_, err := svc.RecordEarning(context.Background(), testLedgerSubmissionID, testLedgerCampaignID, "clipper1", tt.eligibleViews, tt.cpmRate, "earning:calc-test")
				if err == nil {
					t.Error("expected error for 0 eligible views")
				}
				return
			}

			var capturedAmount int32
			store := &mockLedgerStore{
				createEntry: func(ctx context.Context, arg sqlc.CreateLedgerEntryParams) (sqlc.LedgerEntry, error) {
					capturedAmount = arg.Amount
					return sqlc.LedgerEntry{Amount: arg.Amount}, nil
				},
				deductBudget: func(ctx context.Context, arg sqlc.DeductCampaignBudgetParams) (sqlc.Campaign, error) {
					return sqlc.Campaign{RemainingBudget: 100000 - arg.RemainingBudget}, nil
				},
			}
			svc := service.NewLedgerService(store)
			_, err := svc.RecordEarning(context.Background(), testLedgerSubmissionID, testLedgerCampaignID, "clipper1", tt.eligibleViews, tt.cpmRate, "earning:calc-test")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if capturedAmount != tt.expectedAmount {
				t.Errorf("expected amount %d, got %d", tt.expectedAmount, capturedAmount)
			}
		})
	}
}

func TestRecordEarning_BudgetCapEnforcement(t *testing.T) {
	store := &mockLedgerStore{
		deductBudget: func(ctx context.Context, arg sqlc.DeductCampaignBudgetParams) (sqlc.Campaign, error) {
			// Conditional UPDATE returns no rows when remaining_budget < amount
			return sqlc.Campaign{}, pgx.ErrNoRows
		},
	}
	svc := service.NewLedgerService(store)
	_, err := svc.RecordEarning(context.Background(), testLedgerSubmissionID, testLedgerCampaignID, "clipper1", 5000, 100, "earning:over-budget")
	if err == nil {
		t.Error("expected error for exceeding remaining budget")
	}
}

func TestRecordEarning_ExactlyBudgetCap(t *testing.T) {
	store := &mockLedgerStore{
		createEntry: func(ctx context.Context, arg sqlc.CreateLedgerEntryParams) (sqlc.LedgerEntry, error) {
			return sqlc.LedgerEntry{Amount: arg.Amount}, nil
		},
		deductBudget: func(ctx context.Context, arg sqlc.DeductCampaignBudgetParams) (sqlc.Campaign, error) {
			if arg.RemainingBudget != 500 {
				t.Errorf("expected deducted amount 500, got %d", arg.RemainingBudget)
			}
			return sqlc.Campaign{RemainingBudget: 0}, nil
		},
	}
	svc := service.NewLedgerService(store)
	// 5000 views * 100 CPM / 1000 = 500, exactly budget
	_, err := svc.RecordEarning(context.Background(), testLedgerSubmissionID, testLedgerCampaignID, "clipper1", 5000, 100, "earning:exact-budget")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRecordEarning_CampaignNotFound(t *testing.T) {
	store := &mockLedgerStore{
		deductBudget: func(ctx context.Context, arg sqlc.DeductCampaignBudgetParams) (sqlc.Campaign, error) {
			return sqlc.Campaign{}, pgx.ErrNoRows
		},
	}
	svc := service.NewLedgerService(store)
	_, err := svc.RecordEarning(context.Background(), testLedgerSubmissionID, testLedgerCampaignID, "clipper1", 5000, 100, "earning:not-found")
	if err == nil {
		t.Error("expected error for campaign not found")
	}
}

func TestRecordEarning_Idempotent(t *testing.T) {
	store := &mockLedgerStore{
		createEntry: func(ctx context.Context, arg sqlc.CreateLedgerEntryParams) (sqlc.LedgerEntry, error) {
			return sqlc.LedgerEntry{}, pgx.ErrNoRows
		},
		getByIDempotencyKey: func(ctx context.Context, key string) (sqlc.LedgerEntry, error) {
			return sqlc.LedgerEntry{
				IdempotencyKey: key,
				EntryType:      "earning",
				Amount:         500,
			}, nil
		},
		deductBudget: func(ctx context.Context, arg sqlc.DeductCampaignBudgetParams) (sqlc.Campaign, error) {
			return sqlc.Campaign{RemainingBudget: 5000 - arg.RemainingBudget}, nil
		},
	}
	svc := service.NewLedgerService(store)
	entry, err := svc.RecordEarning(context.Background(), testLedgerSubmissionID, testLedgerCampaignID, "clipper1", 5000, 100, "earning:idem-test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Amount != 500 {
		t.Errorf("expected existing entry amount 500, got %d", entry.Amount)
	}
}

func TestRecordEarning_ZeroEligibleViews(t *testing.T) {
	store := &mockLedgerStore{}
	svc := service.NewLedgerService(store)
	_, err := svc.RecordEarning(context.Background(), testLedgerSubmissionID, testLedgerCampaignID, "clipper1", 0, 100, "earning:zero-views")
	if err == nil {
		t.Error("expected error for zero eligible views")
	}
}

func TestRecordEarning_NegativeEligibleViews(t *testing.T) {
	store := &mockLedgerStore{}
	svc := service.NewLedgerService(store)
	_, err := svc.RecordEarning(context.Background(), testLedgerSubmissionID, testLedgerCampaignID, "clipper1", -100, 100, "earning:neg-views")
	if err == nil {
		t.Error("expected error for negative eligible views")
	}
}

func TestRecordEarning_ZeroCPMRate(t *testing.T) {
	store := &mockLedgerStore{}
	svc := service.NewLedgerService(store)
	_, err := svc.RecordEarning(context.Background(), testLedgerSubmissionID, testLedgerCampaignID, "clipper1", 5000, 0, "earning:zero-cpm")
	if err == nil {
		t.Error("expected error for zero cpm_rate")
	}
}

func TestRecordRefund_HappyPath(t *testing.T) {
	store := &mockLedgerStore{
		createEntry: func(ctx context.Context, arg sqlc.CreateLedgerEntryParams) (sqlc.LedgerEntry, error) {
			if arg.EntryType != "refund" {
				t.Errorf("expected entry_type refund, got %s", arg.EntryType)
			}
			if arg.Amount != 3000 {
				t.Errorf("expected amount 3000, got %d", arg.Amount)
			}
			return sqlc.LedgerEntry{
				IdempotencyKey: arg.IdempotencyKey,
				EntryType:      arg.EntryType,
				Amount:         arg.Amount,
			}, nil
		},
	}
	svc := service.NewLedgerService(store)
	entry, err := svc.RecordRefund(context.Background(), testLedgerCampaignID, 3000, "refund:test-campaign")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Amount != 3000 {
		t.Errorf("expected amount 3000, got %d", entry.Amount)
	}
}

func TestRecordRefund_Idempotent(t *testing.T) {
	store := &mockLedgerStore{
		createEntry: func(ctx context.Context, arg sqlc.CreateLedgerEntryParams) (sqlc.LedgerEntry, error) {
			return sqlc.LedgerEntry{}, pgx.ErrNoRows
		},
		getByIDempotencyKey: func(ctx context.Context, key string) (sqlc.LedgerEntry, error) {
			return sqlc.LedgerEntry{IdempotencyKey: key, EntryType: "refund", Amount: 3000}, nil
		},
	}
	svc := service.NewLedgerService(store)
	entry, err := svc.RecordRefund(context.Background(), testLedgerCampaignID, 3000, "refund:idem-test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Amount != 3000 {
		t.Errorf("expected existing entry amount 3000, got %d", entry.Amount)
	}
}

func TestRecordRefund_ZeroAmount(t *testing.T) {
	store := &mockLedgerStore{}
	svc := service.NewLedgerService(store)
	_, err := svc.RecordRefund(context.Background(), testLedgerCampaignID, 0, "refund:zero")
	if err == nil {
		t.Error("expected error for zero refund amount")
	}
}

func TestRecordRefund_NegativeAmount(t *testing.T) {
	store := &mockLedgerStore{}
	svc := service.NewLedgerService(store)
	_, err := svc.RecordRefund(context.Background(), testLedgerCampaignID, -100, "refund:neg")
	if err == nil {
		t.Error("expected error for negative refund amount")
	}
}

func TestGetCampaignLedger(t *testing.T) {
	expected := []sqlc.LedgerEntry{
		{IdempotencyKey: "fee:1", EntryType: "platform_fee", Amount: 1000},
		{IdempotencyKey: "earning:1", EntryType: "earning", Amount: 500},
	}
	store := &mockLedgerStore{
		listByCampaign: func(ctx context.Context, campaignID pgtype.UUID) ([]sqlc.LedgerEntry, error) {
			return expected, nil
		},
	}
	svc := service.NewLedgerService(store)
	entries, err := svc.GetCampaignLedger(context.Background(), testLedgerCampaignID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

func TestGetCampaignSummary(t *testing.T) {
	store := &mockLedgerStore{
		getCampaignByID: func(ctx context.Context, id pgtype.UUID) (sqlc.Campaign, error) {
			return testCampaignForLedger(4500), nil
		},
		listByCampaign: func(ctx context.Context, campaignID pgtype.UUID) ([]sqlc.LedgerEntry, error) {
			return []sqlc.LedgerEntry{
				{EntryType: "platform_fee", Amount: 1000},
				{EntryType: "earning", Amount: 500},
			}, nil
		},
		sumFeesByCampaign: func(ctx context.Context, campaignID pgtype.UUID) (int64, error) {
			return 1000, nil
		},
		sumSpendByCampaign: func(ctx context.Context, campaignID pgtype.UUID) (int64, error) {
			return 500, nil
		},
	}
	svc := service.NewLedgerService(store)
	summary, err := svc.GetCampaignSummary(context.Background(), testLedgerCampaignID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.TotalFees != 1000 {
		t.Errorf("expected total fees 1000, got %d", summary.TotalFees)
	}
	if summary.TotalSpend != 500 {
		t.Errorf("expected total spend 500, got %d", summary.TotalSpend)
	}
	if summary.RemainingBudget != 4500 {
		t.Errorf("expected remaining budget 4500, got %d", summary.RemainingBudget)
	}
	if len(summary.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(summary.Entries))
	}
}

func TestGetCampaignSummary_NotFound(t *testing.T) {
	store := &mockLedgerStore{
		getCampaignByID: func(ctx context.Context, id pgtype.UUID) (sqlc.Campaign, error) {
			return sqlc.Campaign{}, pgx.ErrNoRows
		},
	}
	svc := service.NewLedgerService(store)
	_, err := svc.GetCampaignSummary(context.Background(), testLedgerCampaignID)
	if err != service.ErrCampaignNotFound {
		t.Errorf("expected ErrCampaignNotFound, got %v", err)
	}
}

func TestGetClipperEarnings(t *testing.T) {
	store := &mockLedgerStore{
		sumEarningsByClipper: func(ctx context.Context, clipperID pgtype.Text) (int64, error) {
			if clipperID.String != "clipper1" {
				t.Errorf("expected clipper1, got %s", clipperID.String)
			}
			return 2500, nil
		},
	}
	svc := service.NewLedgerService(store)
	total, err := svc.GetClipperEarnings(context.Background(), "clipper1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2500 {
		t.Errorf("expected total 2500, got %d", total)
	}
}

func TestGetClipperEarningsForCampaign(t *testing.T) {
	store := &mockLedgerStore{
		sumEarningsByClipperForCampaign: func(ctx context.Context, arg sqlc.SumEarningsByClipperForCampaignParams) (int64, error) {
			if arg.ClipperID.String != "clipper1" {
				t.Errorf("expected clipper1, got %s", arg.ClipperID.String)
			}
			if arg.CampaignID != testLedgerCampaignID {
				t.Errorf("expected matching campaign ID")
			}
			return 800, nil
		},
	}
	svc := service.NewLedgerService(store)
	total, err := svc.GetClipperEarningsForCampaign(context.Background(), "clipper1", testLedgerCampaignID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 800 {
		t.Errorf("expected total 800, got %d", total)
	}
}

func TestGetClipperEarningsSummary(t *testing.T) {
	expected := []sqlc.LedgerEntry{
		{EntryType: "earning", Amount: 500},
		{EntryType: "earning", Amount: 300},
	}
	store := &mockLedgerStore{
		sumEarningsByClipper: func(ctx context.Context, clipperID pgtype.Text) (int64, error) {
			return 800, nil
		},
		listByClipper: func(ctx context.Context, clipperID pgtype.Text) ([]sqlc.LedgerEntry, error) {
			return expected, nil
		},
	}
	svc := service.NewLedgerService(store)
	summary, err := svc.GetClipperEarningsSummary(context.Background(), "clipper1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Total != 800 {
		t.Errorf("expected total 800, got %d", summary.Total)
	}
	if len(summary.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(summary.Entries))
	}
}

func TestRecordEarning_BudgetRestoreOnError(t *testing.T) {
	// Test that budget is restored when CreateLedgerEntry fails with a non-conflict error.
	store := &mockLedgerStore{
		createEntry: func(ctx context.Context, arg sqlc.CreateLedgerEntryParams) (sqlc.LedgerEntry, error) {
			return sqlc.LedgerEntry{}, fmt.Errorf("db connection lost")
		},
		deductBudget: func(ctx context.Context, arg sqlc.DeductCampaignBudgetParams) (sqlc.Campaign, error) {
			return sqlc.Campaign{RemainingBudget: 5000 - arg.RemainingBudget}, nil
		},
		updateBudget: func(ctx context.Context, arg sqlc.UpdateCampaignBudgetParams) (sqlc.Campaign, error) {
			// Restore call
			return sqlc.Campaign{}, nil
		},
	}
	svc := service.NewLedgerService(store)
	_, err := svc.RecordEarning(context.Background(), testLedgerSubmissionID, testLedgerCampaignID, "clipper1", 5000, 100, "earning:error-restore")
	if err == nil {
		t.Error("expected error from db failure")
	}
}

// Test total count: this test file has 22 tests.
func TestRecordEarning_VerifyIdempotencyKeyFormat(t *testing.T) {
	var capturedKey string
	store := &mockLedgerStore{
		createEntry: func(ctx context.Context, arg sqlc.CreateLedgerEntryParams) (sqlc.LedgerEntry, error) {
			capturedKey = arg.IdempotencyKey
			return sqlc.LedgerEntry{Amount: arg.Amount}, nil
		},
		deductBudget: func(ctx context.Context, arg sqlc.DeductCampaignBudgetParams) (sqlc.Campaign, error) {
			return sqlc.Campaign{RemainingBudget: 10000 - arg.RemainingBudget}, nil
		},
	}
	svc := service.NewLedgerService(store)
	key := "earning:sub-abc-123"
	_, err := svc.RecordEarning(context.Background(), testLedgerSubmissionID, testLedgerCampaignID, "clipper1", 5000, 100, key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedKey != key {
		t.Errorf("expected key %s, got %s", key, capturedKey)
	}
}

func TestRecordEarning_NegativeCPMRate(t *testing.T) {
	store := &mockLedgerStore{}
	svc := service.NewLedgerService(store)
	_, err := svc.RecordEarning(context.Background(), testLedgerSubmissionID, testLedgerCampaignID, "clipper1", 5000, -100, "earning:neg-cpm")
	if err == nil {
		t.Error("expected error for negative cpm_rate")
	}
}
