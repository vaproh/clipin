package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	sqlc "clipin/apps/api/internal/db/sqlc"
	r "clipin/apps/api/internal/redis"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const templateListCacheTTL = 300 * time.Second

// TemplateStore is the persistence interface the template service requires.
type TemplateStore interface {
	ListCampaignTemplates(ctx context.Context) ([]sqlc.CampaignTemplate, error)
	GetCampaignTemplateByID(ctx context.Context, id pgtype.UUID) (sqlc.CampaignTemplate, error)
	CreateCampaignTemplate(ctx context.Context, arg sqlc.CreateCampaignTemplateParams) (sqlc.CampaignTemplate, error)
	UpdateCampaignTemplate(ctx context.Context, arg sqlc.UpdateCampaignTemplateParams) (sqlc.CampaignTemplate, error)
	DeleteCampaignTemplate(ctx context.Context, id pgtype.UUID) error
	CountCampaignTemplates(ctx context.Context) (int32, error)
}

// TemplateService implements campaign template business logic.
type TemplateService struct {
	store TemplateStore
	redis *r.Client
}

// NewTemplateService creates a new TemplateService.
func NewTemplateService(store TemplateStore) *TemplateService {
	return &TemplateService{store: store}
}

// NewTemplateServiceWithCache creates a TemplateService with Redis caching.
func NewTemplateServiceWithCache(store TemplateStore, redis *r.Client) *TemplateService {
	return &TemplateService{store: store, redis: redis}
}

const templateListCacheKey = "templates:list"

// Sentinel errors for template operations.
var (
	ErrTemplateNotFound = fmt.Errorf("template not found")
)

// ListTemplates returns all campaign templates.
func (s *TemplateService) ListTemplates(ctx context.Context) ([]sqlc.CampaignTemplate, error) {
	if s.redis != nil {
		if cached, err := s.redis.Get(ctx, templateListCacheKey); err == nil && len(cached) > 0 {
			var templates []sqlc.CampaignTemplate
			if json.Unmarshal(cached, &templates) == nil {
				return templates, nil
			}
		}
	}

	templates, err := s.store.ListCampaignTemplates(ctx)
	if err != nil {
		return nil, err
	}

	// Best-effort cache write.
	if s.redis != nil {
		if data, err := json.Marshal(templates); err == nil {
			_ = s.redis.Set(ctx, templateListCacheKey, data, templateListCacheTTL)
		}
	}

	return templates, nil
}

func (s *TemplateService) invalidateListCache(ctx context.Context) {
	if s.redis != nil {
		_ = s.redis.Delete(ctx, templateListCacheKey)
	}
}

// GetTemplateByID returns a single template by ID.
func (s *TemplateService) GetTemplateByID(ctx context.Context, id pgtype.UUID) (*sqlc.CampaignTemplate, error) {
	t, err := s.store.GetCampaignTemplateByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrTemplateNotFound
		}
		return nil, fmt.Errorf("get template: %w", err)
	}
	return &t, nil
}

// CreateTemplate creates a new campaign template.
func (s *TemplateService) CreateTemplate(ctx context.Context, name, platform string, cpmRate, totalBudget int32, maxClipsPerClipper, minViewsPerClip, autoApproveHours int32, descriptionTemplate string) (*sqlc.CampaignTemplate, error) {
	if name == "" {
		return nil, &ValidationError{Errors: []string{"name is required"}}
	}
	if platform == "" {
		platform = "youtube"
	}

	t, err := s.store.CreateCampaignTemplate(ctx, sqlc.CreateCampaignTemplateParams{
		Name:               name,
		Platform:           platform,
		CpmRate:            cpmRate,
		TotalBudget:        totalBudget,
		MaxClipsPerClipper: nullableInt32(&maxClipsPerClipper),
		MinViewsPerClip:    nullableInt32(&minViewsPerClip),
		AutoApproveHours:   nullableInt32(&autoApproveHours),
		DescriptionTemplate: pgtype.Text{Valid: descriptionTemplate != "", String: descriptionTemplate},
	})
	if err != nil {
		return nil, fmt.Errorf("create template: %w", err)
	}
	s.invalidateListCache(ctx)
	return &t, nil
}

// UpdateTemplate updates an existing campaign template.
func (s *TemplateService) UpdateTemplate(ctx context.Context, id pgtype.UUID, name, platform string, cpmRate, totalBudget int32, maxClipsPerClipper, minViewsPerClip, autoApproveHours int32, descriptionTemplate string) (*sqlc.CampaignTemplate, error) {
	if name == "" {
		return nil, &ValidationError{Errors: []string{"name is required"}}
	}

	t, err := s.store.UpdateCampaignTemplate(ctx, sqlc.UpdateCampaignTemplateParams{
		ID:                  id,
		Name:                name,
		Platform:            platform,
		CpmRate:             cpmRate,
		TotalBudget:         totalBudget,
		MaxClipsPerClipper:  nullableInt32(&maxClipsPerClipper),
		MinViewsPerClip:     nullableInt32(&minViewsPerClip),
		AutoApproveHours:    nullableInt32(&autoApproveHours),
		DescriptionTemplate: pgtype.Text{Valid: descriptionTemplate != "", String: descriptionTemplate},
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrTemplateNotFound
		}
		return nil, fmt.Errorf("update template: %w", err)
	}
	s.invalidateListCache(ctx)
	return &t, nil
}

// DeleteTemplate deletes a campaign template by ID.
func (s *TemplateService) DeleteTemplate(ctx context.Context, id pgtype.UUID) error {
	err := s.store.DeleteCampaignTemplate(ctx, id)
	if err == nil {
		s.invalidateListCache(ctx)
	}
	return err
}

// CountTemplates returns the total number of templates.
func (s *TemplateService) CountTemplates(ctx context.Context) (int32, error) {
	return s.store.CountCampaignTemplates(ctx)
}

// defaultTemplate is the definition of a seed template.
type defaultTemplate struct {
	Name                string
	Platform            string
	CpmRate             int32
	TotalBudget         int32
	MaxClipsPerClipper  int32
	MinViewsPerClip     int32
	AutoApproveHours    int32
	DescriptionTemplate string
}

var defaultTemplates = []defaultTemplate{
	{
		Name:                "YouTube Short - Gaming",
		Platform:            "youtube",
		CpmRate:             2000, // Rs 20 in paise
		TotalBudget:         500000,
		MaxClipsPerClipper:  3,
		MinViewsPerClip:     1000,
		AutoApproveHours:    48,
		DescriptionTemplate: "Create a short-form clip highlighting exciting gameplay moments.",
	},
	{
		Name:                "Instagram Reel - Lifestyle",
		Platform:            "instagram",
		CpmRate:             1500, // Rs 15 in paise
		TotalBudget:         300000,
		MaxClipsPerClipper:  5,
		MinViewsPerClip:     500,
		AutoApproveHours:    48,
		DescriptionTemplate: "Create an engaging Reel featuring lifestyle content.",
	},
	{
		Name:                "TikTok - Trending",
		Platform:            "tiktok",
		CpmRate:             1000, // Rs 10 in paise
		TotalBudget:         200000,
		MaxClipsPerClipper:  5,
		MinViewsPerClip:     500,
		AutoApproveHours:    24,
		DescriptionTemplate: "Create a trending TikTok clip with viral potential.",
	},
	{
		Name:                "Multi-platform - Brand",
		Platform:            "multi",
		CpmRate:             2500, // Rs 25 in paise
		TotalBudget:         1000000,
		MaxClipsPerClipper:  3,
		MinViewsPerClip:     1000,
		AutoApproveHours:    48,
		DescriptionTemplate: "Create cross-platform clips for brand campaigns.",
	},
}

// SeedDefaultTemplates inserts default templates if none exist. Idempotent.
func (s *TemplateService) SeedDefaultTemplates(ctx context.Context) error {
	count, err := s.store.CountCampaignTemplates(ctx)
	if err != nil {
		return fmt.Errorf("count templates: %w", err)
	}
	if count > 0 {
		return nil
	}

	for _, dt := range defaultTemplates {
		_, err := s.store.CreateCampaignTemplate(ctx, sqlc.CreateCampaignTemplateParams{
			Name:                dt.Name,
			Platform:            dt.Platform,
			CpmRate:             dt.CpmRate,
			TotalBudget:         dt.TotalBudget,
			MaxClipsPerClipper:  nullableInt32(&dt.MaxClipsPerClipper),
			MinViewsPerClip:     nullableInt32(&dt.MinViewsPerClip),
			AutoApproveHours:    nullableInt32(&dt.AutoApproveHours),
			DescriptionTemplate: pgtype.Text{Valid: true, String: dt.DescriptionTemplate},
		})
		if err != nil {
			return fmt.Errorf("seed template %q: %w", dt.Name, err)
		}
	}
	return nil
}
