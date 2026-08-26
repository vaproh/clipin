package service

import (
	"context"
	"fmt"

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
