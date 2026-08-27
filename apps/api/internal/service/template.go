package service

import (
	"context"
	"fmt"

	sqlc "clipin/apps/api/internal/db/sqlc"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// TemplateStore is the persistence interface the template service requires.
type TemplateStore interface {
	ListCampaignTemplates(ctx context.Context) ([]sqlc.CampaignTemplate, error)
	GetCampaignTemplateByID(ctx context.Context, id pgtype.UUID) (sqlc.CampaignTemplate, error)
	CreateCampaignTemplate(ctx context.Context, arg sqlc.CreateCampaignTemplateParams) (sqlc.CampaignTemplate, error)
}

// TemplateService implements campaign template business logic.
type TemplateService struct {
	store TemplateStore
}

// NewTemplateService creates a new TemplateService.
func NewTemplateService(store TemplateStore) *TemplateService {
	return &TemplateService{store: store}
}

// Sentinel errors for template operations.
var (
	ErrTemplateNotFound = fmt.Errorf("template not found")
)

// ListTemplates returns all campaign templates.
func (s *TemplateService) ListTemplates(ctx context.Context) ([]sqlc.CampaignTemplate, error) {
	return s.store.ListCampaignTemplates(ctx)
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
	return &t, nil
}
