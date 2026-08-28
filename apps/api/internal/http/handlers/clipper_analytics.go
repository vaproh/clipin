package handlers

import (
	"context"
	"time"

	sqlc "clipin/apps/api/internal/db/sqlc"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgtype"
)

// ClipperAnalyticsStore is the subset of Querier the analytics handler needs.
type ClipperAnalyticsStore interface {
	GetClipperEarningsByDay(ctx context.Context, arg sqlc.GetClipperEarningsByDayParams) ([]sqlc.GetClipperEarningsByDayRow, error)
	GetClipperEarningsByCampaign(ctx context.Context, clipperID string) ([]sqlc.GetClipperEarningsByCampaignRow, error)
	GetClipperRecentSubmissions(ctx context.Context, arg sqlc.GetClipperRecentSubmissionsParams) ([]sqlc.GetClipperRecentSubmissionsRow, error)
	GetClipperTotalEarnings(ctx context.Context, clipperID pgtype.Text) (int32, error)
	GetClipperSubmissionStats(ctx context.Context, clipperID string) (sqlc.GetClipperSubmissionStatsRow, error)
	GetClipperTotalViews(ctx context.Context, clipperID string) (int64, error)
	GetClipperCampaignCount(ctx context.Context, clipperID string) (int32, error)
}

type earningsByDayItem struct {
	Day   string `json:"day"`
	Total int32  `json:"total"`
}

type earningsByCampaignItem struct {
	CampaignID        string  `json:"campaign_id"`
	Title             string  `json:"title"`
	Platform          string  `json:"platform"`
	CpmRate           int32   `json:"cpm_rate"`
	TotalEarnings     int32   `json:"total_earnings"`
	TotalSubmissions  int32   `json:"total_submissions"`
	ApprovedSubmissions int32 `json:"approved_submissions"`
	TotalViews        int64   `json:"total_views"`
}

type recentSubmissionItem struct {
	ID            string  `json:"id"`
	PostURL       string  `json:"post_url"`
	Platform      string  `json:"platform"`
	Status        string  `json:"status"`
	CreatedAt     string  `json:"created_at"`
	CampaignTitle string  `json:"campaign_title"`
	CpmRate       int32   `json:"cpm_rate"`
	LatestViews   int64   `json:"latest_views"`
}

type clipperAnalyticsSummary struct {
	TotalEarnings        int32   `json:"total_earnings"`
	TotalSubmissions     int32   `json:"total_submissions"`
	ApprovedSubmissions  int32   `json:"approved_submissions"`
	TotalViews           int64   `json:"total_views"`
	ApprovalRate         float64 `json:"approval_rate"`
	AvgEarningsPerClip   int32   `json:"avg_earnings_per_clip"`
	CampaignsParticipated int32  `json:"campaigns_participated"`
}

type clipperAnalyticsOutput struct {
	Body struct {
		EarningsByDay     []earningsByDayItem     `json:"earnings_by_day"`
		EarningsByCampaign []earningsByCampaignItem `json:"earnings_by_campaign"`
		RecentSubmissions []recentSubmissionItem   `json:"recent_submissions"`
		Summary           clipperAnalyticsSummary  `json:"summary"`
	}
}

// RegisterClipperAnalyticsHandlers registers the clipper analytics endpoint.
func RegisterClipperAnalyticsHandlers(api huma.API, store ClipperAnalyticsStore) {
	huma.Register(api, huma.Operation{
		OperationID: "clipper-analytics",
		Method:      "GET",
		Path:        "/me/analytics",
		Summary:     "Clipper analytics dashboard",
		Description: "Returns earnings trends, campaign performance, recent submissions, and summary stats for the authenticated clipper.",
		Tags:        []string{"Analytics"},
	}, func(ctx context.Context, input *struct{}) (*clipperAnalyticsOutput, error) {
		user, err := requireUser(ctx)
		if err != nil {
			return nil, err
		}

		clipperID := user.ID
		pgClipperID := pgtype.Text{String: clipperID, Valid: true}

		resp := &clipperAnalyticsOutput{}

		// Earnings by day (last 30 days)
		startDate := time.Now().AddDate(0, 0, -30)
		earningsByDay, err := store.GetClipperEarningsByDay(ctx, sqlc.GetClipperEarningsByDayParams{
			ClipperID: pgClipperID,
			CreatedAt: pgtype.Timestamptz{Time: startDate, Valid: true},
		})
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get earnings by day")
		}
		resp.Body.EarningsByDay = make([]earningsByDayItem, 0, len(earningsByDay))
		for _, row := range earningsByDay {
			resp.Body.EarningsByDay = append(resp.Body.EarningsByDay, earningsByDayItem{
				Day:   row.Day.Time.Format("2006-01-02"),
				Total: row.Total,
			})
		}

		// Earnings by campaign
		earningsByCampaign, err := store.GetClipperEarningsByCampaign(ctx, clipperID)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get earnings by campaign")
		}
		resp.Body.EarningsByCampaign = make([]earningsByCampaignItem, 0, len(earningsByCampaign))
		for _, row := range earningsByCampaign {
			resp.Body.EarningsByCampaign = append(resp.Body.EarningsByCampaign, earningsByCampaignItem{
				CampaignID:        row.CampaignID.String(),
				Title:             row.Title,
				Platform:          row.Platform,
				CpmRate:           row.CpmRate,
				TotalEarnings:     row.TotalEarnings,
				TotalSubmissions:  row.TotalSubmissions,
				ApprovedSubmissions: row.ApprovedSubmissions,
				TotalViews:        row.TotalViews,
			})
		}

		// Recent submissions (last 10)
		recentSubs, err := store.GetClipperRecentSubmissions(ctx, sqlc.GetClipperRecentSubmissionsParams{
			ClipperID: clipperID,
			Limit:     10,
		})
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get recent submissions")
		}
		resp.Body.RecentSubmissions = make([]recentSubmissionItem, 0, len(recentSubs))
		for _, row := range recentSubs {
			resp.Body.RecentSubmissions = append(resp.Body.RecentSubmissions, recentSubmissionItem{
				ID:            row.ID.String(),
				PostURL:       row.PostUrl,
				Platform:      row.Platform,
				Status:        row.Status,
				CreatedAt:     row.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
				CampaignTitle: row.CampaignTitle,
				CpmRate:       row.CpmRate,
				LatestViews:   row.LatestViews,
			})
		}

		// Summary stats
		totalEarnings, err := store.GetClipperTotalEarnings(ctx, pgClipperID)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get total earnings")
		}
		subStats, err := store.GetClipperSubmissionStats(ctx, clipperID)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get submission stats")
		}
		totalViews, err := store.GetClipperTotalViews(ctx, clipperID)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get total views")
		}
		campaignCount, err := store.GetClipperCampaignCount(ctx, clipperID)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get campaign count")
		}

		summary := clipperAnalyticsSummary{
			TotalEarnings:        totalEarnings,
			TotalSubmissions:     subStats.TotalSubmissions,
			ApprovedSubmissions:  subStats.ApprovedSubmissions,
			TotalViews:           totalViews,
			CampaignsParticipated: campaignCount,
		}
		if subStats.TotalSubmissions > 0 {
			summary.ApprovalRate = float64(subStats.ApprovedSubmissions) / float64(subStats.TotalSubmissions) * 100
		}
		if subStats.ApprovedSubmissions > 0 {
			summary.AvgEarningsPerClip = totalEarnings / subStats.ApprovedSubmissions
		}
		resp.Body.Summary = summary

		return resp, nil
	})
}
