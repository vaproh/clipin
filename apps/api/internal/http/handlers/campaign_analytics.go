package handlers

import (
	"context"
	"database/sql"
	"errors"
	"time"

	sqlc "clipin/apps/api/internal/db/sqlc"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgtype"
)

// CampaignAnalyticsStore is the subset of Querier the analytics handler needs.
type CampaignAnalyticsStore interface {
	GetCampaignByID(ctx context.Context, id pgtype.UUID) (sqlc.Campaign, error)
	GetCampaignSubmissionStats(ctx context.Context, campaignID pgtype.UUID) (sqlc.GetCampaignSubmissionStatsRow, error)
	GetCampaignViewStats(ctx context.Context, campaignID pgtype.UUID) (sqlc.GetCampaignViewStatsRow, error)
	GetCampaignFinancialSummary(ctx context.Context, campaignID pgtype.UUID) (sqlc.GetCampaignFinancialSummaryRow, error)
}

type campaignAnalyticsOutput struct {
	Body struct {
		Submissions struct {
			Total    int32 `json:"total"`
			Pending  int32 `json:"pending"`
			Approved int32 `json:"approved"`
			Rejected int32 `json:"rejected"`
		} `json:"submissions"`
		UniqueClippers int32 `json:"unique_clippers"`
		Views          struct {
			Total    int64 `json:"total"`
			Likes    int64 `json:"likes"`
			Comments int64 `json:"comments"`
			Shares   int64 `json:"shares"`
		} `json:"views"`
		Financial struct {
			Earnings int32 `json:"earnings"`
			Fees     int32 `json:"fees"`
			Refunds  int32 `json:"refunds"`
		} `json:"financial"`
		Progress struct {
			BudgetConsumedPercent float64  `json:"budget_consumed_percent"`
			TimeRemaining         *string  `json:"time_remaining,omitempty"`
			CampaignStatus        string   `json:"campaign_status"`
		} `json:"progress"`
	}
}

// RegisterCampaignAnalyticsHandlers registers the campaign analytics endpoint.
func RegisterCampaignAnalyticsHandlers(api huma.API, store CampaignAnalyticsStore) {
	huma.Register(api, huma.Operation{
		OperationID: "campaign-analytics",
		Method:      "GET",
		Path:        "/me/campaigns/{id}/analytics",
		Summary:     "Campaign analytics",
		Description: "Returns analytics for a single campaign: submission stats, view metrics, financial summary, and progress. Only the campaign owner may access this endpoint.",
		Tags:        []string{"Campaigns"},
	}, func(ctx context.Context, input *struct {
		ID string `path:"id" doc:"Campaign UUID"`
	}) (*campaignAnalyticsOutput, error) {
		user, err := requireUser(ctx)
		if err != nil {
			return nil, err
		}

		campaignID, err := parseCampaignID(input.ID)
		if err != nil {
			return nil, err
		}

		campaign, err := store.GetCampaignByID(ctx, campaignID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, huma.Error404NotFound("campaign not found")
			}
			return nil, huma.Error500InternalServerError("failed to get campaign")
		}

		if campaign.OwnerID != user.ID {
			return nil, huma.Error403Forbidden("not campaign owner")
		}

		subStats, err := store.GetCampaignSubmissionStats(ctx, campaignID)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get submission stats")
		}

		viewStats, err := store.GetCampaignViewStats(ctx, campaignID)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get view stats")
		}

		finStats, err := store.GetCampaignFinancialSummary(ctx, campaignID)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get financial summary")
		}

		resp := &campaignAnalyticsOutput{}

		// Submission stats
		resp.Body.Submissions.Total = subStats.TotalSubmissions
		resp.Body.Submissions.Pending = subStats.Pending
		resp.Body.Submissions.Approved = subStats.Approved
		resp.Body.Submissions.Rejected = subStats.Rejected
		resp.Body.UniqueClippers = subStats.UniqueClippers

		// View stats
		resp.Body.Views.Total = viewStats.TotalViews
		resp.Body.Views.Likes = viewStats.TotalLikes
		resp.Body.Views.Comments = viewStats.TotalComments
		resp.Body.Views.Shares = viewStats.TotalShares

		// Financial summary
		resp.Body.Financial.Earnings = finStats.TotalEarnings
		resp.Body.Financial.Fees = finStats.TotalFees
		resp.Body.Financial.Refunds = finStats.TotalRefunds

		// Campaign progress
		resp.Body.Progress.CampaignStatus = campaign.Status
		if campaign.TotalBudget > 0 {
			spent := campaign.TotalBudget - campaign.RemainingBudget
			resp.Body.Progress.BudgetConsumedPercent = float64(spent) / float64(campaign.TotalBudget) * 100
		}
		if campaign.EndsAt.Valid {
			remaining := time.Until(campaign.EndsAt.Time)
			if remaining > 0 {
				s := remaining.Truncate(time.Second).String()
				resp.Body.Progress.TimeRemaining = &s
			}
		}

		return resp, nil
	})
}
