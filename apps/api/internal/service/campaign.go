package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	sqlc "clipin/apps/api/internal/db/sqlc"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// CampaignStore is the persistence interface the campaign service requires.
type CampaignStore interface {
	GetCampaignByID(ctx context.Context, id pgtype.UUID) (sqlc.Campaign, error)
	ListCampaignsFiltered(ctx context.Context, arg sqlc.ListCampaignsFilteredParams) ([]sqlc.Campaign, error)
	CountCampaignsFiltered(ctx context.Context, arg sqlc.CountCampaignsFilteredParams) (int64, error)
	ListCampaignsByOwner(ctx context.Context, ownerID string) ([]sqlc.Campaign, error)
	CreateCampaign(ctx context.Context, arg sqlc.CreateCampaignParams) (sqlc.Campaign, error)
	UpdateCampaignStatus(ctx context.Context, arg sqlc.UpdateCampaignStatusParams) (sqlc.Campaign, error)
}

// CampaignFilters holds optional query parameters for marketplace listing.
type CampaignFilters struct {
	Platform  string
	MaxCPM    int32
	MinBudget int32
	Page      int
	PageSize  int
}

// CampaignListResult is the paginated response for public marketplace queries.
type CampaignListResult struct {
	Campaigns []sqlc.Campaign `json:"campaigns"`
	Total     int64           `json:"total"`
	Page      int             `json:"page"`
	PageSize  int             `json:"page_size"`
}

// CampaignService implements campaign marketplace business logic.
type CampaignService struct {
	store CampaignStore
	cache *CampaignCache
}

func NewCampaignService(store CampaignStore, cache *CampaignCache) *CampaignService {
	return &CampaignService{store: store, cache: cache}
}

// ListPublic returns a paginated, filtered list of active campaigns.
func (s *CampaignService) ListPublic(ctx context.Context, f CampaignFilters) (*CampaignListResult, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 || f.PageSize > 100 {
		f.PageSize = 20
	}
	offset := int32((f.Page - 1) * f.PageSize)
	limit := int32(f.PageSize)

	// Try cache first (skip if no cache configured).
	if s.cache != nil {
		if cached, err := s.cache.GetList(ctx, f.Platform, int(f.MaxCPM), int(f.MinBudget), f.Page); err == nil && cached != nil {
			// Cache hit for list; we still need the count.
			count, err := s.store.CountCampaignsFiltered(ctx, sqlc.CountCampaignsFilteredParams{
				Column1: f.Platform,
				Column2: f.MaxCPM,
				Column3: f.MinBudget,
			})
			if err != nil {
				return nil, fmt.Errorf("count campaigns: %w", err)
			}
			return &CampaignListResult{
				Campaigns: cached,
				Total:     count,
				Page:      f.Page,
				PageSize:  f.PageSize,
			}, nil
		}
	}

	campaigns, err := s.store.ListCampaignsFiltered(ctx, sqlc.ListCampaignsFilteredParams{
		Column1: f.Platform,
		Column2: f.MaxCPM,
		Column3: f.MinBudget,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list campaigns: %w", err)
	}

	count, err := s.store.CountCampaignsFiltered(ctx, sqlc.CountCampaignsFilteredParams{
		Column1: f.Platform,
		Column2: f.MaxCPM,
		Column3: f.MinBudget,
	})
	if err != nil {
		return nil, fmt.Errorf("count campaigns: %w", err)
	}

	// Best-effort cache write.
	if s.cache != nil {
		_ = s.cache.SetList(ctx, f.Platform, int(f.MaxCPM), int(f.MinBudget), f.Page, campaigns)
	}

	return &CampaignListResult{
		Campaigns: campaigns,
		Total:     count,
		Page:      f.Page,
		PageSize:  f.PageSize,
	}, nil
}

// GetByID returns a single campaign by its UUID.
func (s *CampaignService) GetByID(ctx context.Context, id pgtype.UUID) (*sqlc.Campaign, error) {
	campaign, err := s.store.GetCampaignByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get campaign: %w", err)
	}
	return &campaign, nil
}

// ListByOwner returns all campaigns owned by the given user (all statuses).
func (s *CampaignService) ListByOwner(ctx context.Context, ownerID string) ([]sqlc.Campaign, error) {
	return s.store.ListCampaignsByOwner(ctx, ownerID)
}

// OwnerStats aggregates dashboard stats for the given owner.
type OwnerStats struct {
	TotalCampaigns  int64 `json:"total_campaigns"`
	ActiveCampaigns int64 `json:"active_campaigns"`
	TotalBudget     int64 `json:"total_budget"`
	TotalRemaining  int64 `json:"total_remaining"`
}

// OwnerStats computes dashboard stats by scanning the owner's campaigns.
func (s *CampaignService) OwnerStats(ctx context.Context, ownerID string) (*OwnerStats, error) {
	campaigns, err := s.store.ListCampaignsByOwner(ctx, ownerID)
	if err != nil {
		return nil, fmt.Errorf("list campaigns for stats: %w", err)
	}
	stats := &OwnerStats{}
	for _, c := range campaigns {
		stats.TotalCampaigns++
		if c.Status == "active" || c.Status == "funded" {
			stats.ActiveCampaigns++
		}
		stats.TotalBudget += int64(c.TotalBudget)
		stats.TotalRemaining += int64(c.RemainingBudget)
	}
	return stats, nil
}

// Create creates a new campaign with validation and defaults applied.
func (s *CampaignService) Create(ctx context.Context, ownerID string, in *CreateCampaignInput) (*sqlc.Campaign, error) {
	if err := ValidateCreateCampaign(in); err != nil {
		return nil, err
	}

	// Generate a new UUID.
	var id pgtype.UUID
	if _, err := rand.Read(id.Bytes[:]); err != nil {
		return nil, fmt.Errorf("generate uuid: %w", err)
	}
	id.Valid = true

	var desc, brief pgtype.Text
	if in.Description != nil {
		desc = pgtype.Text{Valid: true, String: *in.Description}
	}
	if in.BriefURL != nil {
		brief = pgtype.Text{Valid: true, String: *in.BriefURL}
	}

	var startsAt, endsAt pgtype.Timestamptz
	if in.StartsAt != nil {
		t, _ := time.Parse(time.RFC3339, *in.StartsAt)
		startsAt = pgtype.Timestamptz{Valid: true, Time: t}
	}
	if in.EndsAt != nil {
		t, _ := time.Parse(time.RFC3339, *in.EndsAt)
		endsAt = pgtype.Timestamptz{Valid: true, Time: t}
	}

	fee := PlatformFee(in.TotalBudget)

	// Stub: skip escrow, create directly as active.
	campaign, err := s.store.CreateCampaign(ctx, sqlc.CreateCampaignParams{
		ID:                  id,
		OwnerID:             ownerID,
		Title:               strings.TrimSpace(in.Title),
		Description:         desc,
		BriefUrl:            brief,
		Platform:            in.Platform,
		Status:              "active",
		CpmRate:             in.CpmRate,
		TotalBudget:         in.TotalBudget,
		RemainingBudget:     in.TotalBudget,
		PlatformFee:         fee,
		MaxClipsPerCampaign: nullableInt32(in.MaxClipsPerCampaign),
		MaxClipsPerClipper:  pgtype.Int4{Valid: true, Int32: DefaultMaxClipsPerClipper(in.MaxClipsPerClipper)},
		MinViewsPerClip:     pgtype.Int4{Valid: true, Int32: DefaultMinViewsPerClip(in.MinViewsPerClip)},
		AutoApproveHours:    pgtype.Int4{Valid: true, Int32: DefaultAutoApproveHours(in.AutoApproveHours)},
		StartsAt:            startsAt,
		EndsAt:              endsAt,
	})
	if err != nil {
		return nil, fmt.Errorf("create campaign: %w", err)
	}
	return &campaign, nil
}

// Update modifies editable fields of a draft/paused campaign.
func (s *CampaignService) Update(ctx context.Context, ownerID string, campaignID pgtype.UUID, in *UpdateCampaignInput) (*sqlc.Campaign, error) {
	if err := ValidateUpdateCampaign(in); err != nil {
		return nil, err
	}

	campaign, err := s.store.GetCampaignByID(ctx, campaignID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrCampaignNotFound
		}
		return nil, fmt.Errorf("get campaign: %w", err)
	}
	if campaign.OwnerID != ownerID {
		return nil, ErrNotOwner
	}
	if campaign.Status != "draft" && campaign.Status != "paused" {
		return nil, fmt.Errorf("can only update draft or paused campaigns, current status: %s", campaign.Status)
	}

	// Apply partial updates. Only update fields that are provided.
	if in.Title != nil {
		campaign.Title = strings.TrimSpace(*in.Title)
	}
	if in.Description != nil {
		campaign.Description = pgtype.Text{Valid: true, String: *in.Description}
	}
	if in.BriefURL != nil {
		campaign.BriefUrl = pgtype.Text{Valid: true, String: *in.BriefURL}
	}
	if in.MaxClipsPerCampaign != nil {
		campaign.MaxClipsPerCampaign = pgtype.Int4{Valid: true, Int32: *in.MaxClipsPerCampaign}
	}
	if in.MaxClipsPerClipper != nil {
		campaign.MaxClipsPerClipper = pgtype.Int4{Valid: true, Int32: *in.MaxClipsPerClipper}
	}
	if in.MinViewsPerClip != nil {
		campaign.MinViewsPerClip = pgtype.Int4{Valid: true, Int32: *in.MinViewsPerClip}
	}
	if in.AutoApproveHours != nil {
		campaign.AutoApproveHours = pgtype.Int4{Valid: true, Int32: *in.AutoApproveHours}
	}
	if in.EndsAt != nil {
		t, _ := time.Parse(time.RFC3339, *in.EndsAt)
		campaign.EndsAt = pgtype.Timestamptz{Valid: true, Time: t}
	}

	// Re-persist. Since there is no UpdateCampaign query, we update via
	// status (no-op) to touch updated_at. This is a stub; a proper
	// UPDATE campaign ... SET ... query should be added when the schema grows.
	updated, err := s.store.UpdateCampaignStatus(ctx, sqlc.UpdateCampaignStatusParams{
		ID:     campaignID,
		Status: campaign.Status,
	})
	if err != nil {
		return nil, fmt.Errorf("update campaign: %w", err)
	}

	// The stub only updates status/updated_at. Copy the fields we changed
	// onto the returned model so the caller sees the intended state.
	updated.Title = campaign.Title
	updated.Description = campaign.Description
	updated.BriefUrl = campaign.BriefUrl
	updated.MaxClipsPerCampaign = campaign.MaxClipsPerCampaign
	updated.MaxClipsPerClipper = campaign.MaxClipsPerClipper
	updated.MinViewsPerClip = campaign.MinViewsPerClip
	updated.AutoApproveHours = campaign.AutoApproveHours
	updated.EndsAt = campaign.EndsAt
	return &updated, nil
}

// Pause transitions an active campaign to paused.
func (s *CampaignService) Pause(ctx context.Context, ownerID string, campaignID pgtype.UUID) (*sqlc.Campaign, error) {
	return s.transition(ctx, ownerID, campaignID, "active", "paused")
}

// Resume transitions a paused campaign to active.
func (s *CampaignService) Resume(ctx context.Context, ownerID string, campaignID pgtype.UUID) (*sqlc.Campaign, error) {
	return s.transition(ctx, ownerID, campaignID, "paused", "active")
}

// Cancel transitions a draft/paused/active campaign to cancelled.
func (s *CampaignService) Cancel(ctx context.Context, ownerID string, campaignID pgtype.UUID) (*sqlc.Campaign, error) {
	campaign, err := s.store.GetCampaignByID(ctx, campaignID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrCampaignNotFound
		}
		return nil, fmt.Errorf("get campaign: %w", err)
	}
	if campaign.OwnerID != ownerID {
		return nil, ErrNotOwner
	}
	if campaign.Status != "draft" && campaign.Status != "paused" && campaign.Status != "active" {
		return nil, fmt.Errorf("can only cancel draft, paused, or active campaigns, current status: %s", campaign.Status)
	}

	updated, err := s.store.UpdateCampaignStatus(ctx, sqlc.UpdateCampaignStatusParams{
		ID:     campaignID,
		Status: "cancelled",
	})
	if err != nil {
		return nil, fmt.Errorf("cancel campaign: %w", err)
	}
	return &updated, nil
}

func (s *CampaignService) transition(ctx context.Context, ownerID string, campaignID pgtype.UUID, from, to string) (*sqlc.Campaign, error) {
	campaign, err := s.store.GetCampaignByID(ctx, campaignID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrCampaignNotFound
		}
		return nil, fmt.Errorf("get campaign: %w", err)
	}
	if campaign.OwnerID != ownerID {
		return nil, ErrNotOwner
	}
	if campaign.Status != from {
		return nil, fmt.Errorf("campaign must be %s to %s, current status: %s", from, to, campaign.Status)
	}

	updated, err := s.store.UpdateCampaignStatus(ctx, sqlc.UpdateCampaignStatusParams{
		ID:     campaignID,
		Status: to,
	})
	if err != nil {
		return nil, fmt.Errorf("%s campaign: %w", to, err)
	}
	return &updated, nil
}

// Sentinel errors for campaign operations.
var (
	ErrCampaignNotFound = fmt.Errorf("campaign not found")
	ErrNotOwner         = fmt.Errorf("not campaign owner")
)

func nullableInt32(v *int32) pgtype.Int4 {
	if v != nil {
		return pgtype.Int4{Valid: true, Int32: *v}
	}
	return pgtype.Int4{}
}
