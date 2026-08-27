package handlers

import (
	"context"

	sqlc "clipin/apps/api/internal/db/sqlc"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgtype"
)

// LeaderboardStore is the subset of the Querier the leaderboard handler needs.
type LeaderboardStore interface {
	GetLeaderboardByEarnings(ctx context.Context, arg sqlc.GetLeaderboardByEarningsParams) ([]sqlc.GetLeaderboardByEarningsRow, error)
	GetLeaderboardBySubmissions(ctx context.Context, arg sqlc.GetLeaderboardBySubmissionsParams) ([]sqlc.GetLeaderboardBySubmissionsRow, error)
}

type leaderboardInput struct {
	Sort   string `query:"sort" doc:"Sort by 'earnings' (default) or 'submissions'"`
	Limit  int    `query:"limit" doc:"Max results (default 20, max 100)"`
	Offset int    `query:"offset" doc:"Offset for pagination (default 0)"`
}

type leaderboardItem struct {
	Rank                  int     `json:"rank"`
	UserID                string  `json:"user_id"`
	DisplayName           *string `json:"display_name,omitempty"`
	AvatarURL             *string `json:"avatar_url,omitempty"`
	TotalEarnings         int32   `json:"total_earnings"`
	TotalSubmissions      int32   `json:"total_submissions"`
	ApprovedSubmissions   *int32  `json:"approved_submissions,omitempty"`
	CampaignsParticipated *int32  `json:"campaigns_participated,omitempty"`
}

type leaderboardOutput struct {
	Body struct {
		Entries []leaderboardItem `json:"entries"`
	}
}

func textPtr(t pgtype.Text) *string {
	if t.Valid {
		return &t.String
	}
	return nil
}

// RegisterLeaderboardHandlers registers the public leaderboard endpoint.
func RegisterLeaderboardHandlers(api huma.API, store LeaderboardStore) {
	huma.Register(api, huma.Operation{
		OperationID: "get-leaderboard",
		Method:      "GET",
		Path:        "/leaderboard",
		Summary:     "Get public leaderboard",
		Description: "Returns top clippers ranked by earnings or approved submissions. Public endpoint, no auth required.",
		Tags:        []string{"Leaderboard"},
	}, func(ctx context.Context, input *leaderboardInput) (*leaderboardOutput, error) {
		// Defaults and caps
		sort := input.Sort
		if sort != "submissions" {
			sort = "earnings"
		}
		limit := input.Limit
		if limit <= 0 {
			limit = 20
		}
		if limit > 100 {
			limit = 100
		}
		offset := input.Offset
		if offset < 0 {
			offset = 0
		}

		resp := &leaderboardOutput{}
		resp.Body.Entries = make([]leaderboardItem, 0)

		if sort == "submissions" {
			rows, err := store.GetLeaderboardBySubmissions(ctx, sqlc.GetLeaderboardBySubmissionsParams{
				Limit:  int32(limit),
				Offset: int32(offset),
			})
			if err != nil {
				return nil, huma.Error500InternalServerError("failed to get leaderboard")
			}
			for i, row := range rows {
				approved := row.ApprovedSubmissions
				resp.Body.Entries = append(resp.Body.Entries, leaderboardItem{
					Rank:                offset + i + 1,
					UserID:              row.ID,
					DisplayName:         textPtr(row.DisplayName),
					AvatarURL:           textPtr(row.AvatarUrl),
					TotalEarnings:       row.TotalEarnings,
					TotalSubmissions:    row.TotalSubmissions,
					ApprovedSubmissions: &approved,
				})
			}
		} else {
			rows, err := store.GetLeaderboardByEarnings(ctx, sqlc.GetLeaderboardByEarningsParams{
				Limit:  int32(limit),
				Offset: int32(offset),
			})
			if err != nil {
				return nil, huma.Error500InternalServerError("failed to get leaderboard")
			}
			for i, row := range rows {
				campaigns := row.CampaignsParticipated
				resp.Body.Entries = append(resp.Body.Entries, leaderboardItem{
					Rank:                  offset + i + 1,
					UserID:                row.ID,
					DisplayName:           textPtr(row.DisplayName),
					AvatarURL:             textPtr(row.AvatarUrl),
					TotalEarnings:         row.TotalEarnings,
					TotalSubmissions:      row.TotalSubmissions,
					CampaignsParticipated: &campaigns,
				})
			}
		}

		return resp, nil
	})
}
