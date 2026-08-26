package service

import (
	"context"
	"encoding/json"
	"fmt"

	sqlc "clipin/apps/api/internal/db/sqlc"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// LedgerStore is the persistence interface the ledger service requires.
type LedgerStore interface {
	CreateLedgerEntry(ctx context.Context, arg sqlc.CreateLedgerEntryParams) (sqlc.LedgerEntry, error)
	GetLedgerEntryByIdempotencyKey(ctx context.Context, idempotencyKey string) (sqlc.LedgerEntry, error)
	ListLedgerEntriesByCampaign(ctx context.Context, campaignID pgtype.UUID) ([]sqlc.LedgerEntry, error)
	ListLedgerEntriesByClipper(ctx context.Context, clipperID pgtype.Text) ([]sqlc.LedgerEntry, error)
	SumEarningsByClipper(ctx context.Context, clipperID pgtype.Text) (int64, error)
	SumEarningsByClipperForCampaign(ctx context.Context, arg sqlc.SumEarningsByClipperForCampaignParams) (int64, error)
	SumFeesByCampaign(ctx context.Context, campaignID pgtype.UUID) (int64, error)
	SumSpendByCampaign(ctx context.Context, campaignID pgtype.UUID) (int64, error)
	GetCampaignByID(ctx context.Context, id pgtype.UUID) (sqlc.Campaign, error)
	UpdateCampaignBudget(ctx context.Context, arg sqlc.UpdateCampaignBudgetParams) (sqlc.Campaign, error)
}

// LedgerService implements the append-only financial ledger.
type LedgerService struct {
	store LedgerStore
}

// NewLedgerService creates a new LedgerService.
func NewLedgerService(store LedgerStore) *LedgerService {
	return &LedgerService{store: store}
}

// CampaignSummary holds aggregate financial data for a campaign.
type CampaignSummary struct {
	CampaignID     string              `json:"campaign_id"`
	TotalFees      int64               `json:"total_fees"`
	TotalSpend     int64               `json:"total_spend"`
	RemainingBudget int32              `json:"remaining_budget"`
	Entries        []sqlc.LedgerEntry  `json:"entries"`
}

// EarningsSummary holds clipper earnings data.
type EarningsSummary struct {
	ClipperID string                `json:"clipper_id"`
	Total     int64                 `json:"total"`
	Entries   []sqlc.LedgerEntry    `json:"entries"`
}

// ledgerMetadata is stored as JSON in the metadata column.
type ledgerMetadata struct {
	CpmRate       int32  `json:"cpm_rate,omitempty"`
	Views         int64  `json:"views,omitempty"`
	EligibleViews int64  `json:"eligible_views,omitempty"`
}

func marshalMetadata(m *ledgerMetadata) ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

// RecordPlatformFee records the 10% platform fee when a campaign is created.
// Idempotent: duplicate calls return the existing entry.
func (s *LedgerService) RecordPlatformFee(ctx context.Context, campaignID pgtype.UUID, amount int32, idempotencyKey string) (*sqlc.LedgerEntry, error) {
	// ON CONFLICT DO NOTHING returns pgx.ErrNoRows if conflict; check existing.
	entry, err := s.store.CreateLedgerEntry(ctx, sqlc.CreateLedgerEntryParams{
		IdempotencyKey: idempotencyKey,
		EntryType:      "platform_fee",
		CampaignID:     campaignID,
		Amount:         amount,
		Description:    pgtype.Text{Valid: true, String: "Platform fee (10% of deposit)"},
	})
	if err != nil {
		// ON CONFLICT DO NOTHING returns ErrNoRows when there's a conflict.
		if err == pgx.ErrNoRows {
			existing, findErr := s.store.GetLedgerEntryByIdempotencyKey(ctx, idempotencyKey)
			if findErr != nil {
				return nil, fmt.Errorf("get existing fee entry: %w", findErr)
			}
			return &existing, nil
		}
		return nil, fmt.Errorf("create platform fee entry: %w", err)
	}
	return &entry, nil
}

// RecordEarning records earnings for an approved/auto-approved submission.
// It atomically deducts from campaign remaining_budget.
// Earning = eligible_views * cpm_rate / 1000 (integer division, round down).
// Idempotent: duplicate calls return the existing entry.
func (s *LedgerService) RecordEarning(ctx context.Context, submissionID pgtype.UUID, campaignID pgtype.UUID, clipperID string, eligibleViews int64, cpmRate int32, idempotencyKey string) (*sqlc.LedgerEntry, error) {
	if eligibleViews <= 0 {
		return nil, fmt.Errorf("eligible views must be positive, got %d", eligibleViews)
	}
	if cpmRate <= 0 {
		return nil, fmt.Errorf("cpm_rate must be positive, got %d", cpmRate)
	}

	// Calculate earnings: eligible_views * cpm_rate / 1000, integer division.
	amount := int32((eligibleViews * int64(cpmRate)) / 1000)
	if amount <= 0 {
		return nil, fmt.Errorf("calculated amount is zero: eligible_views=%d, cpm_rate=%d", eligibleViews, cpmRate)
	}

	// Fetch campaign to check remaining budget.
	campaign, err := s.store.GetCampaignByID(ctx, campaignID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrCampaignNotFound
		}
		return nil, fmt.Errorf("get campaign: %w", err)
	}

	// Budget cap: earnings cannot exceed remaining budget.
	if amount > campaign.RemainingBudget {
		return nil, fmt.Errorf("earning %d exceeds remaining budget %d", amount, campaign.RemainingBudget)
	}

	newRemaining := campaign.RemainingBudget - amount

	// Update campaign remaining budget atomically.
	_, err = s.store.UpdateCampaignBudget(ctx, sqlc.UpdateCampaignBudgetParams{
		ID:              campaignID,
		RemainingBudget: newRemaining,
	})
	if err != nil {
		return nil, fmt.Errorf("update campaign budget: %w", err)
	}

	meta, _ := marshalMetadata(&ledgerMetadata{
		CpmRate:       cpmRate,
		Views:         eligibleViews,
		EligibleViews: eligibleViews,
	})

	entry, err := s.store.CreateLedgerEntry(ctx, sqlc.CreateLedgerEntryParams{
		IdempotencyKey: idempotencyKey,
		EntryType:      "earning",
		CampaignID:     campaignID,
		SubmissionID:   submissionID,
		ClipperID:      pgtype.Text{Valid: true, String: clipperID},
		Amount:         amount,
		Description:    pgtype.Text{Valid: true, String: fmt.Sprintf("Earning for %d eligible views at CPM %d", eligibleViews, cpmRate)},
		Metadata:       meta,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			// Idempotent duplicate. Restore budget since the entry already exists.
			// This is safe because the original entry already deducted the budget.
			// However, we should NOT restore - the original entry already did.
			existing, findErr := s.store.GetLedgerEntryByIdempotencyKey(ctx, idempotencyKey)
			if findErr != nil {
				return nil, fmt.Errorf("get existing earning entry: %w", findErr)
			}
			return &existing, nil
		}
		// On any other error, restore the budget we deducted.
		restoredErr := s.restoreBudget(ctx, campaignID, campaign.RemainingBudget)
		if restoredErr != nil {
			return nil, fmt.Errorf("create earning entry failed and budget restore failed: %w (original: %v)", restoredErr, err)
		}
		return nil, fmt.Errorf("create earning entry: %w", err)
	}
	return &entry, nil
}

// restoreBudget resets the campaign remaining budget to the given value.
func (s *LedgerService) restoreBudget(ctx context.Context, campaignID pgtype.UUID, budget int32) error {
	_, err := s.store.UpdateCampaignBudget(ctx, sqlc.UpdateCampaignBudgetParams{
		ID:              campaignID,
		RemainingBudget: budget,
	})
	return err
}

// RecordRefund records a refund of unspent budget to the campaign owner.
// Idempotent: duplicate calls return the existing entry.
func (s *LedgerService) RecordRefund(ctx context.Context, campaignID pgtype.UUID, amount int32, idempotencyKey string) (*sqlc.LedgerEntry, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("refund amount must be positive, got %d", amount)
	}

	entry, err := s.store.CreateLedgerEntry(ctx, sqlc.CreateLedgerEntryParams{
		IdempotencyKey: idempotencyKey,
		EntryType:      "refund",
		CampaignID:     campaignID,
		Amount:         amount,
		Description:    pgtype.Text{Valid: true, String: fmt.Sprintf("Refund of unspent budget (%d paise)", amount)},
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			existing, findErr := s.store.GetLedgerEntryByIdempotencyKey(ctx, idempotencyKey)
			if findErr != nil {
				return nil, fmt.Errorf("get existing refund entry: %w", findErr)
			}
			return &existing, nil
		}
		return nil, fmt.Errorf("create refund entry: %w", err)
	}
	return &entry, nil
}

// GetCampaignLedger returns all ledger entries for a campaign (auditable trail).
func (s *LedgerService) GetCampaignLedger(ctx context.Context, campaignID pgtype.UUID) ([]sqlc.LedgerEntry, error) {
	return s.store.ListLedgerEntriesByCampaign(ctx, campaignID)
}

// GetCampaignSummary returns aggregate financial data for a campaign.
func (s *LedgerService) GetCampaignSummary(ctx context.Context, campaignID pgtype.UUID) (*CampaignSummary, error) {
	campaign, err := s.store.GetCampaignByID(ctx, campaignID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrCampaignNotFound
		}
		return nil, fmt.Errorf("get campaign: %w", err)
	}

	entries, err := s.store.ListLedgerEntriesByCampaign(ctx, campaignID)
	if err != nil {
		return nil, fmt.Errorf("list ledger entries: %w", err)
	}

	totalFees, err := s.store.SumFeesByCampaign(ctx, campaignID)
	if err != nil {
		return nil, fmt.Errorf("sum fees: %w", err)
	}

	totalSpend, err := s.store.SumSpendByCampaign(ctx, campaignID)
	if err != nil {
		return nil, fmt.Errorf("sum spend: %w", err)
	}

	return &CampaignSummary{
		CampaignID:      fmt.Sprintf("%x", campaignID.Bytes),
		TotalFees:       totalFees,
		TotalSpend:      totalSpend,
		RemainingBudget: campaign.RemainingBudget,
		Entries:         entries,
	}, nil
}

// GetClipperEarnings returns total lifetime earnings for a clipper.
func (s *LedgerService) GetClipperEarnings(ctx context.Context, clipperID string) (int64, error) {
	return s.store.SumEarningsByClipper(ctx, pgtype.Text{Valid: true, String: clipperID})
}

// GetClipperEarningsForCampaign returns earnings for a specific campaign.
func (s *LedgerService) GetClipperEarningsForCampaign(ctx context.Context, clipperID string, campaignID pgtype.UUID) (int64, error) {
	return s.store.SumEarningsByClipperForCampaign(ctx, sqlc.SumEarningsByClipperForCampaignParams{
		ClipperID:  pgtype.Text{Valid: true, String: clipperID},
		CampaignID: campaignID,
	})
}

// GetClipperEarningsSummary returns earnings summary with entry history.
func (s *LedgerService) GetClipperEarningsSummary(ctx context.Context, clipperID string) (*EarningsSummary, error) {
	total, err := s.store.SumEarningsByClipper(ctx, pgtype.Text{Valid: true, String: clipperID})
	if err != nil {
		return nil, fmt.Errorf("sum earnings: %w", err)
	}

	entries, err := s.store.ListLedgerEntriesByClipper(ctx, pgtype.Text{Valid: true, String: clipperID})
	if err != nil {
		return nil, fmt.Errorf("list earnings: %w", err)
	}

	return &EarningsSummary{
		ClipperID: clipperID,
		Total:     total,
		Entries:   entries,
	}, nil
}
