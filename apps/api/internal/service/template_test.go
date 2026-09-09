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
	list    func(ctx context.Context) ([]sqlc.CampaignTemplate, error)
	getByID func(ctx context.Context, id pgtype.UUID) (sqlc.CampaignTemplate, error)
	create  func(ctx context.Context, arg sqlc.CreateCampaignTemplateParams) (sqlc.CampaignTemplate, error)
	update  func(ctx context.Context, arg sqlc.UpdateCampaignTemplateParams) (sqlc.CampaignTemplate, error)
	delete  func(ctx context.Context, id pgtype.UUID) error
	count   func(ctx context.Context) (int32, error)
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

func (m *mockTemplateStore) UpdateCampaignTemplate(ctx context.Context, arg sqlc.UpdateCampaignTemplateParams) (sqlc.CampaignTemplate, error) {
	if m.update != nil {
		return m.update(ctx, arg)
	}
	return sqlc.CampaignTemplate{
		ID:       arg.ID,
		Name:     arg.Name,
		Platform: arg.Platform,
		CpmRate:  arg.CpmRate,
	}, nil
}

func (m *mockTemplateStore) DeleteCampaignTemplate(ctx context.Context, id pgtype.UUID) error {
	if m.delete != nil {
		return m.delete(ctx, id)
	}
	return nil
}

func (m *mockTemplateStore) CountCampaignTemplates(ctx context.Context) (int32, error) {
	if m.count != nil {
		return m.count(ctx)
	}
	return 0, nil
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

func TestTemplateUpdate(t *testing.T) {
	id := pgtype.UUID{Bytes: [16]byte{7}, Valid: true}
	store := &mockTemplateStore{
		update: func(_ context.Context, arg sqlc.UpdateCampaignTemplateParams) (sqlc.CampaignTemplate, error) {
			if arg.ID != id {
				t.Errorf("expected ID %v, got %v", id, arg.ID)
			}
			if arg.Name != "Updated Name" {
				t.Errorf("expected Updated Name, got %s", arg.Name)
			}
			if arg.CpmRate != 300 {
				t.Errorf("expected cpm 300, got %d", arg.CpmRate)
			}
			return sqlc.CampaignTemplate{
				ID:       arg.ID,
				Name:     arg.Name,
				Platform: arg.Platform,
				CpmRate:  arg.CpmRate,
			}, nil
		},
	}
	svc := service.NewTemplateService(store)
	tmpl, err := svc.UpdateTemplate(context.Background(), id, "Updated Name", "youtube", 300, 50000, 3, 1000, 48, "new desc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tmpl.Name != "Updated Name" {
		t.Errorf("expected Updated Name, got %s", tmpl.Name)
	}
	if tmpl.CpmRate != 300 {
		t.Errorf("expected cpm 300, got %d", tmpl.CpmRate)
	}
}

func TestTemplateUpdateEmptyName(t *testing.T) {
	store := &mockTemplateStore{}
	svc := service.NewTemplateService(store)
	_, err := svc.UpdateTemplate(context.Background(), pgtype.UUID{Bytes: [16]byte{1}, Valid: true}, "", "youtube", 150, 10000, 3, 1000, 48, "")
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestTemplateUpdateNotFound(t *testing.T) {
	store := &mockTemplateStore{
		update: func(_ context.Context, _ sqlc.UpdateCampaignTemplateParams) (sqlc.CampaignTemplate, error) {
			return sqlc.CampaignTemplate{}, pgx.ErrNoRows
		},
	}
	svc := service.NewTemplateService(store)
	_, err := svc.UpdateTemplate(context.Background(), pgtype.UUID{Bytes: [16]byte{99}, Valid: true}, "Name", "youtube", 150, 10000, 3, 1000, 48, "")
	if err != service.ErrTemplateNotFound {
		t.Errorf("expected ErrTemplateNotFound, got %v", err)
	}
}

func TestTemplateDelete(t *testing.T) {
	id := pgtype.UUID{Bytes: [16]byte{5}, Valid: true}
	var deletedID pgtype.UUID
	store := &mockTemplateStore{
		delete: func(_ context.Context, delID pgtype.UUID) error {
			deletedID = delID
			return nil
		},
	}
	svc := service.NewTemplateService(store)
	err := svc.DeleteTemplate(context.Background(), id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deletedID != id {
		t.Errorf("expected delete ID %v, got %v", id, deletedID)
	}
}

func TestSeedDefaultTemplates(t *testing.T) {
	var created []string
	store := &mockTemplateStore{
		count: func(_ context.Context) (int32, error) {
			return 0, nil
		},
		create: func(_ context.Context, arg sqlc.CreateCampaignTemplateParams) (sqlc.CampaignTemplate, error) {
			created = append(created, arg.Name)
			return sqlc.CampaignTemplate{
				ID:       pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
				Name:     arg.Name,
				Platform: arg.Platform,
				CpmRate:  arg.CpmRate,
			}, nil
		},
	}
	svc := service.NewTemplateService(store)
	err := svc.SeedDefaultTemplates(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(created) != 3 {
		t.Fatalf("expected 3 templates seeded, got %d", len(created))
	}
	expected := []string{
		"YouTube Short - Gaming",
		"Instagram Reel - Lifestyle",
		"Multi-platform - Brand",
	}
	for i, name := range expected {
		if created[i] != name {
			t.Errorf("template %d: expected %q, got %q", i, name, created[i])
		}
	}
}

func TestSeedDefaultTemplates_AlreadySeeded(t *testing.T) {
	createCalled := false
	store := &mockTemplateStore{
		count: func(_ context.Context) (int32, error) {
			return 3, nil
		},
		create: func(_ context.Context, _ sqlc.CreateCampaignTemplateParams) (sqlc.CampaignTemplate, error) {
			createCalled = true
			return sqlc.CampaignTemplate{}, nil
		},
	}
	svc := service.NewTemplateService(store)
	err := svc.SeedDefaultTemplates(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if createCalled {
		t.Error("expected no creates when templates already exist")
	}
}
