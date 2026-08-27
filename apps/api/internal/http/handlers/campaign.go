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

// CampaignServiceInterface is the subset of CampaignService the handlers need.
type CampaignServiceInterface interface {
	ListPublic(ctx context.Context, f service.CampaignFilters) (*service.CampaignListResult, error)
	GetByID(ctx context.Context, id pgtype.UUID) (*sqlc.Campaign, error)
	ListByOwner(ctx context.Context, ownerID string) ([]sqlc.Campaign, error)
	Create(ctx context.Context, ownerID string, in *service.CreateCampaignInput) (*sqlc.Campaign, error)
	Update(ctx context.Context, ownerID string, campaignID pgtype.UUID, in *service.UpdateCampaignInput) (*sqlc.Campaign, error)
	Pause(ctx context.Context, ownerID string, campaignID pgtype.UUID) (*sqlc.Campaign, error)
	Resume(ctx context.Context, ownerID string, campaignID pgtype.UUID) (*sqlc.Campaign, error)
	Cancel(ctx context.Context, ownerID string, campaignID pgtype.UUID) (*sqlc.Campaign, error)
	OwnerStats(ctx context.Context, ownerID string) (*service.OwnerStats, error)
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

// --- Owner campaign management ---

type createCampaignInput struct {
	Body struct {
		Title               string  `json:"title" doc:"Campaign title (1-200 chars)"`
		Description         *string `json:"description,omitempty" doc:"Campaign description"`
		BriefURL            *string `json:"brief_url,omitempty" doc:"Link to campaign brief"`
		Platform            string  `json:"platform" doc:"Platform: youtube, instagram, tiktok, multi"`
		CpmRate             int32   `json:"cpm_rate" doc:"Rate per 1000 views in paise"`
		TotalBudget         int32   `json:"total_budget" doc:"Total budget in paise"`
		MaxClipsPerCampaign *int32  `json:"max_clips_per_campaign,omitempty" doc:"Max clips for this campaign"`
		MaxClipsPerClipper  *int32  `json:"max_clips_per_clipper,omitempty" doc:"Max clips per clipper (default 3)"`
		MinViewsPerClip     *int32  `json:"min_views_per_clip,omitempty" doc:"Min views per clip (default 1000)"`
		AutoApproveHours    *int32  `json:"auto_approve_hours,omitempty" doc:"Auto-approve after hours (default 48)"`
		StartsAt            *string `json:"starts_at,omitempty" doc:"Campaign start time (RFC3339)"`
		EndsAt              *string `json:"ends_at,omitempty" doc:"Campaign end time (RFC3339)"`
	}
}

type updateCampaignInput struct {
	ID   string `path:"id" doc:"Campaign UUID"`
	Body struct {
		Title               *string `json:"title,omitempty" doc:"Campaign title"`
		Description         *string `json:"description,omitempty" doc:"Campaign description"`
		BriefURL            *string `json:"brief_url,omitempty" doc:"Link to campaign brief"`
		MaxClipsPerCampaign *int32  `json:"max_clips_per_campaign,omitempty"`
		MaxClipsPerClipper  *int32  `json:"max_clips_per_clipper,omitempty"`
		MinViewsPerClip     *int32  `json:"min_views_per_clip,omitempty"`
		AutoApproveHours    *int32  `json:"auto_approve_hours,omitempty"`
		EndsAt              *string `json:"ends_at,omitempty"`
	}
}

type campaignActionInput struct {
	ID string `path:"id" doc:"Campaign UUID"`
}

type ownerStatsOutput struct {
	Body struct {
		TotalCampaigns  int64 `json:"total_campaigns"`
		ActiveCampaigns int64 `json:"active_campaigns"`
		TotalBudget     int64 `json:"total_budget"`
		TotalRemaining  int64 `json:"total_remaining"`
	}
}

// requireOwner checks the user is authenticated and has the "owner" role.
func requireOwner(ctx context.Context) (*sqlc.User, error) {
	user, err := requireUser(ctx)
	if err != nil {
		return nil, err
	}
	if user.Role != "owner" {
		return nil, huma.Error403Forbidden("owner role required")
	}
	return user, nil
}

// parseCampaignID is a small helper to parse a path param into pgtype.UUID.
func parseCampaignID(raw string) (pgtype.UUID, error) {
	var id pgtype.UUID
	if err := id.Scan(raw); err != nil {
		return pgtype.UUID{}, huma.Error422UnprocessableEntity("invalid campaign id")
	}
	return id, nil
}

// RegisterCampaignOwnerHandlers registers owner-only campaign management endpoints.
// These must be mounted on the authenticated router group.
func RegisterCampaignOwnerHandlers(api huma.API, svc CampaignServiceInterface) {
	// POST /campaigns - create campaign
	huma.Register(api, huma.Operation{
		OperationID: "create-campaign",
		Method:      "POST",
		Path:        "/campaigns",
		Summary:     "Create a campaign",
		Description: "Creates a new campaign. Only users with the owner role may create campaigns.",
		Tags:        []string{"Campaigns"},
	}, func(ctx context.Context, input *createCampaignInput) (*campaignDetailOutput, error) {
		user, err := requireOwner(ctx)
		if err != nil {
			return nil, err
		}

		campaign, err := svc.Create(ctx, user.ID, &service.CreateCampaignInput{
			Title:               input.Body.Title,
			Description:         input.Body.Description,
			BriefURL:            input.Body.BriefURL,
			Platform:            input.Body.Platform,
			CpmRate:             input.Body.CpmRate,
			TotalBudget:         input.Body.TotalBudget,
			MaxClipsPerCampaign: input.Body.MaxClipsPerCampaign,
			MaxClipsPerClipper:  input.Body.MaxClipsPerClipper,
			MinViewsPerClip:     input.Body.MinViewsPerClip,
			AutoApproveHours:    input.Body.AutoApproveHours,
			StartsAt:            input.Body.StartsAt,
			EndsAt:              input.Body.EndsAt,
		})
		if err != nil {
			if ve, ok := err.(*service.ValidationError); ok {
				return nil, huma.Error422UnprocessableEntity(ve.Error())
			}
			return nil, huma.Error500InternalServerError("failed to create campaign")
		}

		resp := &campaignDetailOutput{}
		resp.Body = toCampaignListItem(*campaign)
		return resp, nil
	})

	// PATCH /campaigns/{id} - update campaign
	huma.Register(api, huma.Operation{
		OperationID: "update-campaign",
		Method:      "PATCH",
		Path:        "/campaigns/{id}",
		Summary:     "Update a campaign",
		Description: "Updates editable fields of a draft or paused campaign. Only the campaign owner may update.",
		Tags:        []string{"Campaigns"},
	}, func(ctx context.Context, input *updateCampaignInput) (*campaignDetailOutput, error) {
		user, err := requireOwner(ctx)
		if err != nil {
			return nil, err
		}

		campaignID, err := parseCampaignID(input.ID)
		if err != nil {
			return nil, err
		}

		campaign, err := svc.Update(ctx, user.ID, campaignID, &service.UpdateCampaignInput{
			Title:               input.Body.Title,
			Description:         input.Body.Description,
			BriefURL:            input.Body.BriefURL,
			MaxClipsPerCampaign: input.Body.MaxClipsPerCampaign,
			MaxClipsPerClipper:  input.Body.MaxClipsPerClipper,
			MinViewsPerClip:     input.Body.MinViewsPerClip,
			AutoApproveHours:    input.Body.AutoApproveHours,
			EndsAt:              input.Body.EndsAt,
		})
		if err != nil {
			if ve, ok := err.(*service.ValidationError); ok {
				return nil, huma.Error422UnprocessableEntity(ve.Error())
			}
			if errors.Is(err, service.ErrCampaignNotFound) {
				return nil, huma.Error404NotFound("campaign not found")
			}
			if errors.Is(err, service.ErrNotOwner) {
				return nil, huma.Error403Forbidden("not campaign owner")
			}
			return nil, huma.Error500InternalServerError(err.Error())
		}

		resp := &campaignDetailOutput{}
		resp.Body = toCampaignListItem(*campaign)
		return resp, nil
	})

	// POST /campaigns/{id}/pause
	huma.Register(api, huma.Operation{
		OperationID: "pause-campaign",
		Method:      "POST",
		Path:        "/campaigns/{id}/pause",
		Summary:     "Pause a campaign",
		Description: "Pauses an active campaign. Only the campaign owner may pause.",
		Tags:        []string{"Campaigns"},
	}, func(ctx context.Context, input *campaignActionInput) (*campaignDetailOutput, error) {
		user, err := requireOwner(ctx)
		if err != nil {
			return nil, err
		}
		campaignID, err := parseCampaignID(input.ID)
		if err != nil {
			return nil, err
		}
		campaign, err := svc.Pause(ctx, user.ID, campaignID)
		if err != nil {
			if errors.Is(err, service.ErrCampaignNotFound) {
				return nil, huma.Error404NotFound("campaign not found")
			}
			if errors.Is(err, service.ErrNotOwner) {
				return nil, huma.Error403Forbidden("not campaign owner")
			}
			return nil, huma.Error409Conflict(err.Error())
		}
		resp := &campaignDetailOutput{}
		resp.Body = toCampaignListItem(*campaign)
		return resp, nil
	})

	// POST /campaigns/{id}/resume
	huma.Register(api, huma.Operation{
		OperationID: "resume-campaign",
		Method:      "POST",
		Path:        "/campaigns/{id}/resume",
		Summary:     "Resume a campaign",
		Description: "Resumes a paused campaign. Only the campaign owner may resume.",
		Tags:        []string{"Campaigns"},
	}, func(ctx context.Context, input *campaignActionInput) (*campaignDetailOutput, error) {
		user, err := requireOwner(ctx)
		if err != nil {
			return nil, err
		}
		campaignID, err := parseCampaignID(input.ID)
		if err != nil {
			return nil, err
		}
		campaign, err := svc.Resume(ctx, user.ID, campaignID)
		if err != nil {
			if errors.Is(err, service.ErrCampaignNotFound) {
				return nil, huma.Error404NotFound("campaign not found")
			}
			if errors.Is(err, service.ErrNotOwner) {
				return nil, huma.Error403Forbidden("not campaign owner")
			}
			return nil, huma.Error409Conflict(err.Error())
		}
		resp := &campaignDetailOutput{}
		resp.Body = toCampaignListItem(*campaign)
		return resp, nil
	})

	// POST /campaigns/{id}/cancel
	huma.Register(api, huma.Operation{
		OperationID: "cancel-campaign",
		Method:      "POST",
		Path:        "/campaigns/{id}/cancel",
		Summary:     "Cancel a campaign",
		Description: "Cancels a draft, paused, or active campaign. Only the campaign owner may cancel.",
		Tags:        []string{"Campaigns"},
	}, func(ctx context.Context, input *campaignActionInput) (*campaignDetailOutput, error) {
		user, err := requireOwner(ctx)
		if err != nil {
			return nil, err
		}
		campaignID, err := parseCampaignID(input.ID)
		if err != nil {
			return nil, err
		}
		campaign, err := svc.Cancel(ctx, user.ID, campaignID)
		if err != nil {
			if errors.Is(err, service.ErrCampaignNotFound) {
				return nil, huma.Error404NotFound("campaign not found")
			}
			if errors.Is(err, service.ErrNotOwner) {
				return nil, huma.Error403Forbidden("not campaign owner")
			}
			return nil, huma.Error409Conflict(err.Error())
		}
		resp := &campaignDetailOutput{}
		resp.Body = toCampaignListItem(*campaign)
		return resp, nil
	})

	// GET /me/campaigns/stats
	huma.Register(api, huma.Operation{
		OperationID: "owner-campaign-stats",
		Method:      "GET",
		Path:        "/me/campaigns/stats",
		Summary:     "Owner campaign stats",
		Description: "Returns aggregate stats for the current owner's campaigns.",
		Tags:        []string{"Campaigns"},
	}, func(ctx context.Context, input *struct{}) (*ownerStatsOutput, error) {
		user, err := requireOwner(ctx)
		if err != nil {
			return nil, err
		}
		stats, err := svc.OwnerStats(ctx, user.ID)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to compute stats")
		}
		resp := &ownerStatsOutput{}
		resp.Body.TotalCampaigns = stats.TotalCampaigns
		resp.Body.ActiveCampaigns = stats.ActiveCampaigns
		resp.Body.TotalBudget = stats.TotalBudget
		resp.Body.TotalRemaining = stats.TotalRemaining
		return resp, nil
	})
}

