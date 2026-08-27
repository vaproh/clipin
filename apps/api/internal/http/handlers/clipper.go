package handlers

import (
	"context"
	"database/sql"
	"errors"

	sqlc "clipin/apps/api/internal/db/sqlc"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgtype"
)

// ClipperStore is the subset of the Querier the clipper handler needs.
type ClipperStore interface {
	GetUserPublicProfile(ctx context.Context, id string) (sqlc.GetUserPublicProfileRow, error)
	GetClipperSubmissionStats(ctx context.Context, clipperID string) (sqlc.GetClipperSubmissionStatsRow, error)
	GetClipperTotalViews(ctx context.Context, clipperID string) (int64, error)
	GetClipperTotalEarnings(ctx context.Context, clipperID pgtype.Text) (int32, error)
	GetClipperCampaignCount(ctx context.Context, clipperID string) (int32, error)
	ListSocialAccountsByUserIDPublic(ctx context.Context, userID string) ([]sqlc.ListSocialAccountsByUserIDPublicRow, error)
}

type socialAccountPublic struct {
	Platform string  `json:"platform"`
	Username *string `json:"username,omitempty"`
}

type clipperProfileOutput struct {
	Body struct {
		ID                   string               `json:"id"`
		DisplayName          *string              `json:"display_name,omitempty"`
		AvatarURL            *string              `json:"avatar_url,omitempty"`
		Bio                  *string              `json:"bio,omitempty"`
		CreatedAt            string               `json:"created_at"`
		TotalSubmissions     int32                `json:"total_submissions"`
		ApprovedSubmissions  int32                `json:"approved_submissions"`
		PendingSubmissions   int32                `json:"pending_submissions"`
		RejectedSubmissions  int32                `json:"rejected_submissions"`
		TotalViews           int64                `json:"total_views"`
		TotalEarnings        int32                `json:"total_earnings"`
		CampaignsParticipated int32               `json:"campaigns_participated"`
		SocialAccounts       []socialAccountPublic `json:"social_accounts"`
	}
}

// RegisterClipperHandlers registers public clipper profile endpoints (no auth required).
func RegisterClipperHandlers(api huma.API, store ClipperStore) {
	huma.Register(api, huma.Operation{
		OperationID: "get-clipper-profile",
		Method:      "GET",
		Path:        "/clippers/{id}",
		Summary:     "Get clipper public profile",
		Description: "Returns a public profile with aggregated stats for a clipper.",
		Tags:        []string{"Clippers"},
	}, func(ctx context.Context, input *struct {
		ID string `path:"id" doc:"Clipper user ID"`
	}) (*clipperProfileOutput, error) {
		profile, err := store.GetUserPublicProfile(ctx, input.ID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, huma.Error404NotFound("clipper not found")
			}
			return nil, huma.Error500InternalServerError("failed to get clipper profile")
		}

		stats, err := store.GetClipperSubmissionStats(ctx, input.ID)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get submission stats")
		}

		totalViews, err := store.GetClipperTotalViews(ctx, input.ID)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get total views")
		}

		totalEarnings, err := store.GetClipperTotalEarnings(ctx, pgtype.Text{String: input.ID, Valid: true})
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get total earnings")
		}

		campaignCount, err := store.GetClipperCampaignCount(ctx, input.ID)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get campaign count")
		}

		socialAccounts, err := store.ListSocialAccountsByUserIDPublic(ctx, input.ID)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get social accounts")
		}

		resp := &clipperProfileOutput{}
		resp.Body.ID = profile.ID
		resp.Body.CreatedAt = profile.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00")
		resp.Body.TotalSubmissions = stats.TotalSubmissions
		resp.Body.ApprovedSubmissions = stats.ApprovedSubmissions
		resp.Body.PendingSubmissions = stats.PendingSubmissions
		resp.Body.RejectedSubmissions = stats.RejectedSubmissions
		resp.Body.TotalViews = totalViews
		resp.Body.TotalEarnings = totalEarnings
		resp.Body.CampaignsParticipated = campaignCount

		if profile.DisplayName.Valid {
			resp.Body.DisplayName = &profile.DisplayName.String
		}
		if profile.AvatarUrl.Valid {
			resp.Body.AvatarURL = &profile.AvatarUrl.String
		}
		if profile.Bio.Valid {
			resp.Body.Bio = &profile.Bio.String
		}

		resp.Body.SocialAccounts = make([]socialAccountPublic, 0, len(socialAccounts))
		for _, sa := range socialAccounts {
			item := socialAccountPublic{Platform: sa.Platform}
			if sa.PlatformUsername.Valid {
				item.Username = &sa.PlatformUsername.String
			}
			resp.Body.SocialAccounts = append(resp.Body.SocialAccounts, item)
		}

		return resp, nil
	})
}
