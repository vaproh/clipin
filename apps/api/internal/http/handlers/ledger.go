package handlers

import (
	"context"
	"errors"
	"fmt"

	sqlc "clipin/apps/api/internal/db/sqlc"
	"clipin/apps/api/internal/service"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgtype"
)

// LedgerServiceInterface is the subset of LedgerService the handlers need.
type LedgerServiceInterface interface {
	GetClipperEarningsSummary(ctx context.Context, clipperID string) (*service.EarningsSummary, error)
	GetCampaignSummary(ctx context.Context, campaignID pgtype.UUID) (*service.CampaignSummary, error)
	GetCampaignLedger(ctx context.Context, campaignID pgtype.UUID) ([]sqlc.LedgerEntry, error)
}

// CampaignOwnershipChecker verifies a user owns a campaign.
type CampaignOwnershipChecker interface {
	GetByID(ctx context.Context, id pgtype.UUID) (*sqlc.Campaign, error)
}

// --- Earnings response types ---

type ledgerEntryItem struct {
	ID             string  `json:"id"`
	IdempotencyKey string  `json:"idempotency_key"`
	EntryType      string  `json:"entry_type"`
	CampaignID     string  `json:"campaign_id"`
	SubmissionID   *string `json:"submission_id,omitempty"`
	ClipperID      *string `json:"clipper_id,omitempty"`
	Amount         int32   `json:"amount"`
	Description    *string `json:"description,omitempty"`
	CreatedAt      string  `json:"created_at"`
}

func toLedgerEntryItem(e sqlc.LedgerEntry) ledgerEntryItem {
	item := ledgerEntryItem{
		ID:             fmt.Sprintf("%x", e.ID.Bytes),
		IdempotencyKey: e.IdempotencyKey,
		EntryType:      e.EntryType,
		CampaignID:     fmt.Sprintf("%x", e.CampaignID.Bytes),
		Amount:         e.Amount,
	}
	if e.SubmissionID.Valid {
		s := fmt.Sprintf("%x", e.SubmissionID.Bytes)
		item.SubmissionID = &s
	}
	if e.ClipperID.Valid {
		item.ClipperID = &e.ClipperID.String
	}
	if e.Description.Valid {
		item.Description = &e.Description.String
	}
	if e.CreatedAt.Valid {
		item.CreatedAt = e.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00")
	}
	return item
}

// --- Earnings summary for clipper ---

type earningsSummaryOutput struct {
	Body struct {
		ClipperID string            `json:"clipper_id"`
		Total     int64             `json:"total"`
		Entries   []ledgerEntryItem `json:"entries"`
	}
}

// --- Campaign ledger summary for owner ---

type campaignLedgerOutput struct {
	Body struct {
		CampaignID      string            `json:"campaign_id"`
		TotalFees       int64             `json:"total_fees"`
		TotalSpend      int64             `json:"total_spend"`
		RemainingBudget int32             `json:"remaining_budget"`
		Entries         []ledgerEntryItem `json:"entries"`
	}
}

// RegisterLedgerHandlers registers ledger/earnings endpoints on the authenticated API.
func RegisterLedgerHandlers(api huma.API, svc LedgerServiceInterface, campaignStore CampaignOwnershipChecker) {
	// GET /me/earnings - clipper earnings summary
	huma.Register(api, huma.Operation{
		OperationID: "get-my-earnings",
		Method:      "GET",
		Path:        "/me/earnings",
		Summary:     "Get my earnings",
		Description: "Returns total lifetime earnings and recent ledger entries for the current clipper.",
		Tags:        []string{"Earnings"},
	}, func(ctx context.Context, input *struct{}) (*earningsSummaryOutput, error) {
		user, err := requireUser(ctx)
		if err != nil {
			return nil, err
		}

		summary, err := svc.GetClipperEarningsSummary(ctx, user.ID)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get earnings")
		}

		resp := &earningsSummaryOutput{}
		resp.Body.ClipperID = summary.ClipperID
		resp.Body.Total = summary.Total
		resp.Body.Entries = make([]ledgerEntryItem, 0, len(summary.Entries))
		for _, e := range summary.Entries {
			resp.Body.Entries = append(resp.Body.Entries, toLedgerEntryItem(e))
		}
		return resp, nil
	})

	// GET /me/campaigns/{id}/ledger - campaign ledger for owner
	huma.Register(api, huma.Operation{
		OperationID: "get-campaign-ledger",
		Method:      "GET",
		Path:        "/me/campaigns/{id}/ledger",
		Summary:     "Get campaign ledger",
		Description: "Returns all ledger entries for a campaign. Only the campaign owner may access.",
		Tags:        []string{"Earnings"},
	}, func(ctx context.Context, input *struct {
		ID string `path:"id" doc:"Campaign UUID"`
	}) (*campaignLedgerOutput, error) {
		user, err := requireOwner(ctx)
		if err != nil {
			return nil, err
		}

		campaignID, err := parseCampaignID(input.ID)
		if err != nil {
			return nil, err
		}

		// Verify ownership.
		campaign, err := campaignStore.GetByID(ctx, campaignID)
		if err != nil {
			return nil, huma.Error404NotFound("campaign not found")
		}
		if campaign.OwnerID != user.ID {
			return nil, huma.Error403Forbidden("not campaign owner")
		}

		summary, err := svc.GetCampaignSummary(ctx, campaignID)
		if err != nil {
			if errors.Is(err, service.ErrCampaignNotFound) {
				return nil, huma.Error404NotFound("campaign not found")
			}
			return nil, huma.Error500InternalServerError("failed to get campaign ledger")
		}

		resp := &campaignLedgerOutput{}
		resp.Body.CampaignID = summary.CampaignID
		resp.Body.TotalFees = summary.TotalFees
		resp.Body.TotalSpend = summary.TotalSpend
		resp.Body.RemainingBudget = summary.RemainingBudget
		resp.Body.Entries = make([]ledgerEntryItem, 0, len(summary.Entries))
		for _, e := range summary.Entries {
			resp.Body.Entries = append(resp.Body.Entries, toLedgerEntryItem(e))
		}
		return resp, nil
	})
}
