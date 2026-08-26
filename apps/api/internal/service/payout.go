package service

import (
	"context"
	"crypto/rand"
	"fmt"

	sqlc "clipin/apps/api/internal/db/sqlc"
	"clipin/apps/api/internal/payout"

	"github.com/jackc/pgx/v5/pgtype"
)

// PayoutStore is the persistence interface the payout service requires.
type PayoutStore interface {
	GetUserByID(ctx context.Context, id string) (sqlc.User, error)
	UpdateUserUPI(ctx context.Context, arg sqlc.UpdateUserUPIParams) (sqlc.User, error)
	CreatePayoutRequest(ctx context.Context, arg sqlc.CreatePayoutRequestParams) (sqlc.PayoutRequest, error)
	GetPayoutRequestByID(ctx context.Context, id pgtype.UUID) (sqlc.PayoutRequest, error)
	ListPayoutRequestsByClipper(ctx context.Context, clipperID string) ([]sqlc.PayoutRequest, error)
	UpdatePayoutRequestStatus(ctx context.Context, arg sqlc.UpdatePayoutRequestStatusParams) (sqlc.PayoutRequest, error)
	SumPendingPayoutsByClipper(ctx context.Context, clipperID string) (int64, error)
	SumEarningsByClipper(ctx context.Context, clipperID pgtype.Text) (int64, error)
	CreateLedgerEntry(ctx context.Context, arg sqlc.CreateLedgerEntryParams) (sqlc.LedgerEntry, error)
	GetLedgerEntryByIdempotencyKey(ctx context.Context, idempotencyKey string) (sqlc.LedgerEntry, error)
	ListPendingPayoutRequests(ctx context.Context) ([]sqlc.PayoutRequest, error)
}

// PayoutService implements UPI payout request business logic.
type PayoutService struct {
	store    PayoutStore
	provider payout.PayoutProvider
}

// NewPayoutService creates a new PayoutService.
func NewPayoutService(store PayoutStore, provider payout.PayoutProvider) *PayoutService {
	return &PayoutService{store: store, provider: provider}
}

// Minimum payout threshold: 50000 paise = ₹500.
const minPayoutAmount = 50000

// Sentinel errors for payout operations.
var (
	ErrNoUPIID           = fmt.Errorf("UPI ID is not set on your account")
	ErrBelowThreshold    = fmt.Errorf("payout amount must be at least ₹500 (50000 paise)")
	ErrInsufficientFunds = fmt.Errorf("insufficient available balance for this payout")
	ErrPayoutNotFound    = fmt.Errorf("payout request not found")
)

// UpdateUPI saves the user's UPI details.
func (s *PayoutService) UpdateUPI(ctx context.Context, userID, upiID string) (*sqlc.User, error) {
	if upiID == "" {
		return nil, fmt.Errorf("upi_id is required")
	}
	user, err := s.store.UpdateUserUPI(ctx, sqlc.UpdateUserUPIParams{
		ID:    userID,
		UpiID: pgtype.Text{Valid: true, String: upiID},
	})
	if err != nil {
		return nil, fmt.Errorf("update upi: %w", err)
	}
	return &user, nil
}

// RequestPayout creates a payout request for the given clipper.
func (s *PayoutService) RequestPayout(ctx context.Context, userID string, amount int64, idempotencyKey string) (*sqlc.PayoutRequest, error) {
	// Validate threshold.
	if amount < minPayoutAmount {
		return nil, ErrBelowThreshold
	}

	// Fetch user to check UPI ID.
	user, err := s.store.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if !user.UpiID.Valid || user.UpiID.String == "" {
		return nil, ErrNoUPIID
	}

	// Available balance = total earnings - pending/processing payouts.
	totalEarnings, err := s.store.SumEarningsByClipper(ctx, pgtype.Text{Valid: true, String: userID})
	if err != nil {
		return nil, fmt.Errorf("sum earnings: %w", err)
	}
	pendingPayouts, err := s.store.SumPendingPayoutsByClipper(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("sum pending payouts: %w", err)
	}
	available := totalEarnings - pendingPayouts
	if available < amount {
		return nil, ErrInsufficientFunds
	}

	// Generate UUID for payout request.
	var id pgtype.UUID
	if _, err := rand.Read(id.Bytes[:]); err != nil {
		return nil, fmt.Errorf("generate uuid: %w", err)
	}
	id.Valid = true

	// Create payout request. ON CONFLICT handles idempotency.
	pr, err := s.store.CreatePayoutRequest(ctx, sqlc.CreatePayoutRequestParams{
		ID:             id,
		ClipperID:      userID,
		Amount:         int32(amount),
		UpiID:          user.UpiID.String,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		return nil, fmt.Errorf("create payout request: %w", err)
	}

	// ON CONFLICT DO NOTHING returns a zero-value row.
	if !pr.ID.Valid {
		// Fetch existing by querying clipper's payouts and finding the matching one.
		payouts, listErr := s.store.ListPayoutRequestsByClipper(ctx, userID)
		if listErr != nil {
			return nil, fmt.Errorf("list payouts for idempotency: %w", listErr)
		}
		for i := range payouts {
			if payouts[i].IdempotencyKey == idempotencyKey {
				return &payouts[i], nil
			}
		}
		return nil, fmt.Errorf("payout request created but not found (idempotent)")
	}

	// Call provider to initiate transfer.
	resp, err := s.provider.CreateTransfer(ctx, payout.TransferRequest{
		Amount:    amount,
		Currency:  "INR",
		UPITarget: user.UpiID.String,
		Reference: idempotencyKey,
	})
	if err != nil {
		// Mark as failed.
		_, _ = s.store.UpdatePayoutRequestStatus(ctx, sqlc.UpdatePayoutRequestStatusParams{
			ID:            pr.ID,
			Status:        "failed",
			FailureReason: pgtype.Text{Valid: true, String: err.Error()},
		})
		return nil, fmt.Errorf("provider transfer: %w", err)
	}

	// Update status from provider response.
	updated, err := s.store.UpdatePayoutRequestStatus(ctx, sqlc.UpdatePayoutRequestStatusParams{
		ID:          pr.ID,
		Status:      resp.Status,
		ProviderRef: pgtype.Text{Valid: true, String: resp.TransferID},
	})
	if err != nil {
		return nil, fmt.Errorf("update payout status: %w", err)
	}

	// If completed, record a debit ledger entry.
	if resp.Status == "completed" {
		_, _ = s.store.CreateLedgerEntry(ctx, sqlc.CreateLedgerEntryParams{
			IdempotencyKey: fmt.Sprintf("payout:%s", idempotencyKey),
			EntryType:      "payout",
			CampaignID:     pgtype.UUID{}, // payout is not campaign-specific
			Amount:         -int32(amount),
			Description:    pgtype.Text{Valid: true, String: fmt.Sprintf("Payout of %d paise to %s", amount, user.UpiID.String)},
			ClipperID:      pgtype.Text{Valid: true, String: userID},
		})
	}

	return &updated, nil
}

// ListMyPayouts returns all payout requests for a clipper.
func (s *PayoutService) ListMyPayouts(ctx context.Context, userID string) ([]sqlc.PayoutRequest, error) {
	return s.store.ListPayoutRequestsByClipper(ctx, userID)
}

// ProcessPendingPayouts processes all pending payout requests.
// Returns the count of processed payouts.
func (s *PayoutService) ProcessPendingPayouts(ctx context.Context) (int, error) {
	pending, err := s.store.ListPendingPayoutRequests(ctx)
	if err != nil {
		return 0, fmt.Errorf("list pending payouts: %w", err)
	}

	count := 0
	for _, pr := range pending {
		// Mark as processing.
		_, err := s.store.UpdatePayoutRequestStatus(ctx, sqlc.UpdatePayoutRequestStatusParams{
			ID:     pr.ID,
			Status: "processing",
		})
		if err != nil {
			continue
		}

		// Call provider.
		resp, err := s.provider.CreateTransfer(ctx, payout.TransferRequest{
			Amount:    int64(pr.Amount),
			Currency:  "INR",
			UPITarget: pr.UpiID,
			Reference: pr.IdempotencyKey,
		})
		if err != nil {
			_, _ = s.store.UpdatePayoutRequestStatus(ctx, sqlc.UpdatePayoutRequestStatusParams{
				ID:            pr.ID,
				Status:        "failed",
				FailureReason: pgtype.Text{Valid: true, String: err.Error()},
			})
			continue
		}

		// Update with provider response.
		_, err = s.store.UpdatePayoutRequestStatus(ctx, sqlc.UpdatePayoutRequestStatusParams{
			ID:          pr.ID,
			Status:      resp.Status,
			ProviderRef: pgtype.Text{Valid: true, String: resp.TransferID},
		})
		if err != nil {
			continue
		}

		// If completed, record ledger entry.
		if resp.Status == "completed" {
			_, _ = s.store.CreateLedgerEntry(ctx, sqlc.CreateLedgerEntryParams{
				IdempotencyKey: fmt.Sprintf("payout:%s", pr.IdempotencyKey),
				EntryType:      "payout",
				CampaignID:     pgtype.UUID{},
				Amount:         -pr.Amount,
				Description:    pgtype.Text{Valid: true, String: fmt.Sprintf("Payout of %d paise to %s", pr.Amount, pr.UpiID)},
				ClipperID:      pgtype.Text{Valid: true, String: pr.ClipperID},
			})
		}

		count++
	}

	return count, nil
}
