package service_test

import (
	"context"
	"testing"

	sqlc "clipin/apps/api/internal/db/sqlc"
	"clipin/apps/api/internal/service"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type mockTemplateStore struct {
	list   func(ctx context.Context) ([]sqlc.CampaignTemplate, error)
	getByID func(ctx context.Context, id pgtype.UUID) (sqlc.CampaignTemplate, error)
	create func(ctx context.Context, arg sqlc.CreateCampaignTemplateParams) (sqlc.CampaignTemplate, error)
}

func (m *mockTemplateStore) ListCampaignTemplates(ctx context.Context) ([]sqlc.CampaignTemplate, error) {
	if m.list != nil {
		return m.list(ctx)
	}
	return nil, nil
}

func (m *mockTemplateStore) GetCampaignTemplateByID(ctx context.Context, id pgtype.UUID) (sqlc.CampaignTemplate, error) {
	if m.getByID != nil {
		return m.getByID(ctx, id)
	}
	return sqlc.CampaignTemplate{}, nil
}

func (m *mockTemplateStore) CreateCampaignTemplate(ctx context.Context, arg sqlc.CreateCampaignTemplateParams) (sqlc.CampaignTemplate, error) {
	if m.create != nil {
		return m.create(ctx, arg)
	}
	return sqlc.CampaignTemplate{
		ID:       pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
		Name:     arg.Name,
		Platform: arg.Platform,
		CpmRate:  arg.CpmRate,
	}, nil
}

func TestTemplateList(t *testing.T) {
	expected := []sqlc.CampaignTemplate{
		{ID: pgtype.UUID{Bytes: [16]byte{1}, Valid: true}, Name: "YT Starter", Platform: "youtube", CpmRate: 100},
	}
	store := &mockTemplateStore{
		list: func(_ context.Context) ([]sqlc.CampaignTemplate, error) {
			return expected, nil
		},
	}
	svc := service.NewTemplateService(store)
	templates, err := svc.ListTemplates(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(templates) != 1 {
		t.Fatalf("expected 1, got %d", len(templates))
	}
	if templates[0].Name != "YT Starter" {
		t.Errorf("expected YT Starter, got %s", templates[0].Name)
	}
}

func TestTemplateGetByID(t *testing.T) {
	id := pgtype.UUID{Bytes: [16]byte{42}, Valid: true}
	store := &mockTemplateStore{
		getByID: func(_ context.Context, gotID pgtype.UUID) (sqlc.CampaignTemplate, error) {
			if gotID != id {
				t.Errorf("expected ID to match")
			}
			return sqlc.CampaignTemplate{
				ID:       gotID,
				Name:     "Test Template",
				Platform: "youtube",
				CpmRate:  200,
			}, nil
		},
	}
	svc := service.NewTemplateService(store)
	tmpl, err := svc.GetTemplateByID(context.Background(), id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tmpl.Name != "Test Template" {
		t.Errorf("expected Test Template, got %s", tmpl.Name)
	}
}

func TestTemplateGetByID_NotFound(t *testing.T) {
	store := &mockTemplateStore{
		getByID: func(_ context.Context, _ pgtype.UUID) (sqlc.CampaignTemplate, error) {
			return sqlc.CampaignTemplate{}, pgx.ErrNoRows
		},
	}
	svc := service.NewTemplateService(store)
	_, err := svc.GetTemplateByID(context.Background(), pgtype.UUID{Bytes: [16]byte{99}, Valid: true})
	if err != service.ErrTemplateNotFound {
		t.Errorf("expected ErrTemplateNotFound, got %v", err)
	}
}

func TestTemplateCreate(t *testing.T) {
	store := &mockTemplateStore{
		create: func(_ context.Context, arg sqlc.CreateCampaignTemplateParams) (sqlc.CampaignTemplate, error) {
			if arg.Name != "New Template" {
				t.Errorf("expected New Template, got %s", arg.Name)
			}
			if arg.CpmRate != 150 {
				t.Errorf("expected cpm 150, got %d", arg.CpmRate)
			}
			if arg.TotalBudget != 10000 {
				t.Errorf("expected budget 10000, got %d", arg.TotalBudget)
			}
			return sqlc.CampaignTemplate{
				ID:       pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
				Name:     arg.Name,
				Platform: arg.Platform,
				CpmRate:  arg.CpmRate,
			}, nil
		},
	}
	svc := service.NewTemplateService(store)
	tmpl, err := svc.CreateTemplate(context.Background(), "New Template", "youtube", 150, 10000, 3, 1000, 48, "Description here")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tmpl.Name != "New Template" {
		t.Errorf("expected New Template, got %s", tmpl.Name)
	}
}

func TestTemplateCreateEmptyName(t *testing.T) {
	store := &mockTemplateStore{}
	svc := service.NewTemplateService(store)
	_, err := svc.CreateTemplate(context.Background(), "", "youtube", 150, 10000, 3, 1000, 48, "")
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestTemplateCreateDefaultPlatform(t *testing.T) {
	store := &mockTemplateStore{
		create: func(_ context.Context, arg sqlc.CreateCampaignTemplateParams) (sqlc.CampaignTemplate, error) {
			if arg.Platform != "youtube" {
				t.Errorf("expected default youtube platform, got %s", arg.Platform)
			}
			return sqlc.CampaignTemplate{}, nil
		},
	}
	svc := service.NewTemplateService(store)
	_, err := svc.CreateTemplate(context.Background(), "Test", "", 150, 10000, 3, 1000, 48, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
