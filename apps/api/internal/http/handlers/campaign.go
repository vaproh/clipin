package handlers

import (
	"context"
	"fmt"

	sqlc "clipin/apps/api/internal/db/sqlc"
	"clipin/apps/api/internal/service"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgtype"
)

// CampaignServiceInterface is the subset of CampaignService the handlers need.
type CampaignServiceInterface interface {
	ListPublic(ctx context.Context, f service.CampaignFilters) (*service.CampaignListResult, error)
	GetByID(ctx context.Context, id pgtype.UUID) (*sqlc.Campaign, error)
	ListByOwner(ctx context.Context, ownerID string) ([]sqlc.Campaign, error)
}

// --- Public marketplace ---

type listCampaignsInput struct {
	Platform  string `query:"platform" doc:"Filter by platform (youtube, instagram, tiktok, multi)"`
	MaxCPM    int    `query:"max_cpm" doc:"Max CPM rate in paise per 1000 views"`
	MinBudget int    `query:"min_budget" doc:"Minimum remaining budget in paise"`
	Page      int    `query:"page" doc:"Page number (1-indexed)"`
	PageSize  int    `query:"page_size" doc:"Results per page (max 100)"`
}

type campaignListItem struct {
	ID                 string  `json:"id"`
	OwnerID            string  `json:"owner_id"`
	Title              string  `json:"title"`
	Description        *string `json:"description,omitempty"`
	BriefURL           *string `json:"brief_url,omitempty"`
	Platform           string  `json:"platform"`
	Status             string  `json:"status"`
	CpmRate            int32   `json:"cpm_rate"`
	TotalBudget        int32   `json:"total_budget"`
	RemainingBudget    int32   `json:"remaining_budget"`
	PlatformFee        int32   `json:"platform_fee"`
	MaxClipsPerCampaign *int32 `json:"max_clips_per_campaign,omitempty"`
	MaxClipsPerClipper  *int32 `json:"max_clips_per_clipper,omitempty"`
	MinViewsPerClip     *int32 `json:"min_views_per_clip,omitempty"`
	AutoApproveHours    *int32 `json:"auto_approve_hours,omitempty"`
	StartsAt           *string `json:"starts_at,omitempty"`
	EndsAt             *string `json:"ends_at,omitempty"`
	CreatedAt          string  `json:"created_at"`
	UpdatedAt          string  `json:"updated_at"`
}

type listCampaignsOutput struct {
	Body struct {
		Campaigns []campaignListItem `json:"campaigns"`
		Total     int64              `json:"total"`
		Page      int                `json:"page"`
		PageSize  int                `json:"page_size"`
	}
}

type campaignDetailOutput struct {
	Body campaignListItem
}

type myCampaignsOutput struct {
	Body struct {
		Campaigns []campaignListItem `json:"campaigns"`
	}
}

func toCampaignListItem(c sqlc.Campaign) campaignListItem {
	item := campaignListItem{
		ID:              fmt.Sprintf("%x", c.ID.Bytes),
		OwnerID:         c.OwnerID,
		Title:           c.Title,
		Platform:        c.Platform,
		Status:          c.Status,
		CpmRate:         c.CpmRate,
		TotalBudget:     c.TotalBudget,
		RemainingBudget: c.RemainingBudget,
		PlatformFee:     c.PlatformFee,
		CreatedAt:       c.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       c.UpdatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
	}
	if c.Description.Valid {
		item.Description = &c.Description.String
	}
	if c.BriefUrl.Valid {
		item.BriefURL = &c.BriefUrl.String
	}
	if c.MaxClipsPerCampaign.Valid {
		item.MaxClipsPerCampaign = &c.MaxClipsPerCampaign.Int32
	}
	if c.MaxClipsPerClipper.Valid {
		item.MaxClipsPerClipper = &c.MaxClipsPerClipper.Int32
	}
	if c.MinViewsPerClip.Valid {
		item.MinViewsPerClip = &c.MinViewsPerClip.Int32
	}
	if c.AutoApproveHours.Valid {
		item.AutoApproveHours = &c.AutoApproveHours.Int32
	}
	if c.StartsAt.Valid {
		s := c.StartsAt.Time.Format("2006-01-02T15:04:05Z07:00")
		item.StartsAt = &s
	}
	if c.EndsAt.Valid {
		s := c.EndsAt.Time.Format("2006-01-02T15:04:05Z07:00")
		item.EndsAt = &s
	}
	return item
}

// RegisterCampaignHandlers registers public and authenticated campaign endpoints.
func RegisterCampaignHandlers(api huma.API, svc CampaignServiceInterface) {
	// GET /campaigns - public marketplace listing
	huma.Register(api, huma.Operation{
		OperationID: "list-campaigns",
		Method:      "GET",
		Path:        "/campaigns",
		Summary:     "List marketplace campaigns",
		Description: "Returns a paginated list of active/funded campaigns with optional filters.",
		Tags:        []string{"Campaigns"},
	}, func(ctx context.Context, input *listCampaignsInput) (*listCampaignsOutput, error) {
		page := input.Page
		if page < 1 {
			page = 1
		}
		pageSize := input.PageSize
		if pageSize < 1 || pageSize > 100 {
			pageSize = 20
		}

		result, err := svc.ListPublic(ctx, service.CampaignFilters{
			Platform:  input.Platform,
			MaxCPM:    int32(input.MaxCPM),
			MinBudget: int32(input.MinBudget),
			Page:      page,
			PageSize:  pageSize,
		})
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to list campaigns")
		}

		resp := &listCampaignsOutput{}
		resp.Body.Total = result.Total
		resp.Body.Page = result.Page
		resp.Body.PageSize = result.PageSize
		resp.Body.Campaigns = make([]campaignListItem, 0, len(result.Campaigns))
		for _, c := range result.Campaigns {
			resp.Body.Campaigns = append(resp.Body.Campaigns, toCampaignListItem(c))
		}
		return resp, nil
	})

	// GET /campaigns/{id} - campaign detail
	huma.Register(api, huma.Operation{
		OperationID: "get-campaign",
		Method:      "GET",
		Path:        "/campaigns/{id}",
		Summary:     "Get campaign details",
		Description: "Returns full details for a single campaign.",
		Tags:        []string{"Campaigns"},
	}, func(ctx context.Context, input *struct {
		ID string `path:"id" doc:"Campaign UUID"`
	}) (*campaignDetailOutput, error) {
		var pgUUID pgtype.UUID
		if err := pgUUID.Scan(input.ID); err != nil {
			return nil, huma.Error422UnprocessableEntity("invalid campaign id")
		}

		campaign, err := svc.GetByID(ctx, pgUUID)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get campaign")
		}
		if campaign == nil {
			return nil, huma.Error404NotFound("campaign not found")
		}

		resp := &campaignDetailOutput{}
		resp.Body = toCampaignListItem(*campaign)
		return resp, nil
	})

	// GET /me/campaigns - authenticated owner's campaigns
	huma.Register(api, huma.Operation{
		OperationID: "list-my-campaigns",
		Method:      "GET",
		Path:        "/me/campaigns",
		Summary:     "List my campaigns",
		Description: "Returns all campaigns owned by the current user (all statuses).",
		Tags:        []string{"Campaigns"},
	}, func(ctx context.Context, input *struct{}) (*myCampaignsOutput, error) {
		user, err := requireUser(ctx)
		if err != nil {
			return nil, err
		}

		campaigns, err := svc.ListByOwner(ctx, user.ID)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to list campaigns")
		}

		resp := &myCampaignsOutput{}
		resp.Body.Campaigns = make([]campaignListItem, 0, len(campaigns))
		for _, c := range campaigns {
			resp.Body.Campaigns = append(resp.Body.Campaigns, toCampaignListItem(c))
		}
		return resp, nil
	})
}

