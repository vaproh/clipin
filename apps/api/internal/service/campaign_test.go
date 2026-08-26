package service_test

import (
	"context"
	"testing"

	sqlc "clipin/apps/api/internal/db/sqlc"
	"clipin/apps/api/internal/service"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// mockCampaignStore implements CampaignStore for testing.
type mockCampaignStore struct {
	getByID         func(ctx context.Context, id pgtype.UUID) (sqlc.Campaign, error)
	listFiltered    func(ctx context.Context, arg sqlc.ListCampaignsFilteredParams) ([]sqlc.Campaign, error)
	countFiltered   func(ctx context.Context, arg sqlc.CountCampaignsFilteredParams) (int64, error)
	listByOwner     func(ctx context.Context, ownerID string) ([]sqlc.Campaign, error)
}

func (m *mockCampaignStore) GetCampaignByID(ctx context.Context, id pgtype.UUID) (sqlc.Campaign, error) {
	if m.getByID != nil {
		return m.getByID(ctx, id)
	}
	return sqlc.Campaign{}, nil
}

func (m *mockCampaignStore) ListCampaignsFiltered(ctx context.Context, arg sqlc.ListCampaignsFilteredParams) ([]sqlc.Campaign, error) {
	if m.listFiltered != nil {
		return m.listFiltered(ctx, arg)
	}
	return nil, nil
}

func (m *mockCampaignStore) CountCampaignsFiltered(ctx context.Context, arg sqlc.CountCampaignsFilteredParams) (int64, error) {
	if m.countFiltered != nil {
		return m.countFiltered(ctx, arg)
	}
	return 0, nil
}

func (m *mockCampaignStore) ListCampaignsByOwner(ctx context.Context, ownerID string) ([]sqlc.Campaign, error) {
	if m.listByOwner != nil {
		return m.listByOwner(ctx, ownerID)
	}
	return nil, nil
}

func testCampaign(id string, ownerID string, title string, status string) sqlc.Campaign {
	return sqlc.Campaign{
		ID:              pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true},
		OwnerID:         ownerID,
		Title:           title,
		Platform:        "youtube",
		Status:          status,
		CpmRate:         100,
		TotalBudget:     10000,
		RemainingBudget: 5000,
		PlatformFee:     1000,
		CreatedAt:       pgtype.Timestamptz{Valid: true},
		UpdatedAt:       pgtype.Timestamptz{Valid: true},
	}
}

func TestListPublic_DefaultPagination(t *testing.T) {
	store := &mockCampaignStore{
		listFiltered: func(ctx context.Context, arg sqlc.ListCampaignsFilteredParams) ([]sqlc.Campaign, error) {
			if arg.Limit != 20 {
				t.Errorf("expected default limit 20, got %d", arg.Limit)
			}
			if arg.Offset != 0 {
				t.Errorf("expected default offset 0, got %d", arg.Offset)
			}
			return []sqlc.Campaign{testCampaign("1", "owner1", "Campaign 1", "active")}, nil
		},
		countFiltered: func(ctx context.Context, arg sqlc.CountCampaignsFilteredParams) (int64, error) {
			return 1, nil
		},
	}
	svc := service.NewCampaignService(store, nil)
	result, err := svc.ListPublic(context.Background(), service.CampaignFilters{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Page != 1 {
		t.Errorf("expected page 1, got %d", result.Page)
	}
	if result.PageSize != 20 {
		t.Errorf("expected page_size 20, got %d", result.PageSize)
	}
	if result.Total != 1 {
		t.Errorf("expected total 1, got %d", result.Total)
	}
	if len(result.Campaigns) != 1 {
		t.Fatalf("expected 1 campaign, got %d", len(result.Campaigns))
	}
}

func TestListPublic_CustomPagination(t *testing.T) {
	store := &mockCampaignStore{
		listFiltered: func(ctx context.Context, arg sqlc.ListCampaignsFilteredParams) ([]sqlc.Campaign, error) {
			if arg.Limit != 10 {
				t.Errorf("expected limit 10, got %d", arg.Limit)
			}
			if arg.Offset != 20 {
				t.Errorf("expected offset 20 (page 3, size 10), got %d", arg.Offset)
			}
			return nil, nil
		},
		countFiltered: func(ctx context.Context, arg sqlc.CountCampaignsFilteredParams) (int64, error) {
			return 0, nil
		},
	}
	svc := service.NewCampaignService(store, nil)
	result, err := svc.ListPublic(context.Background(), service.CampaignFilters{Page: 3, PageSize: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Page != 3 {
		t.Errorf("expected page 3, got %d", result.Page)
	}
}

func TestListPublic_CapsPageSize(t *testing.T) {
	store := &mockCampaignStore{
		listFiltered: func(ctx context.Context, arg sqlc.ListCampaignsFilteredParams) ([]sqlc.Campaign, error) {
			if arg.Limit > 100 {
				t.Errorf("expected max limit 100, got %d", arg.Limit)
			}
			return nil, nil
		},
		countFiltered: func(ctx context.Context, arg sqlc.CountCampaignsFilteredParams) (int64, error) {
			return 0, nil
		},
	}
	svc := service.NewCampaignService(store, nil)
	result, err := svc.ListPublic(context.Background(), service.CampaignFilters{PageSize: 999})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PageSize != 20 {
		t.Errorf("expected page_size 20 (capped), got %d", result.PageSize)
	}
}

func TestListPublic_PassesFilters(t *testing.T) {
	store := &mockCampaignStore{
		listFiltered: func(ctx context.Context, arg sqlc.ListCampaignsFilteredParams) ([]sqlc.Campaign, error) {
			if arg.Column1 != "youtube" {
				t.Errorf("expected platform youtube, got %q", arg.Column1)
			}
			if arg.Column2 != 200 {
				t.Errorf("expected maxCPM 200, got %d", arg.Column2)
			}
			if arg.Column3 != 5000 {
				t.Errorf("expected minBudget 5000, got %d", arg.Column3)
			}
			return nil, nil
		},
		countFiltered: func(ctx context.Context, arg sqlc.CountCampaignsFilteredParams) (int64, error) {
			return 0, nil
		},
	}
	svc := service.NewCampaignService(store, nil)
	_, err := svc.ListPublic(context.Background(), service.CampaignFilters{
		Platform:  "youtube",
		MaxCPM:    200,
		MinBudget: 5000,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetByID_Found(t *testing.T) {
	c := testCampaign("1", "owner1", "My Campaign", "active")
	store := &mockCampaignStore{
		getByID: func(ctx context.Context, id pgtype.UUID) (sqlc.Campaign, error) {
			return c, nil
		},
	}
	svc := service.NewCampaignService(store, nil)
	result, err := svc.GetByID(context.Background(), c.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected campaign, got nil")
	}
	if result.Title != "My Campaign" {
		t.Errorf("expected title My Campaign, got %q", result.Title)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	store := &mockCampaignStore{
		getByID: func(ctx context.Context, id pgtype.UUID) (sqlc.Campaign, error) {
			return sqlc.Campaign{}, pgx.ErrNoRows
		},
	}
	svc := service.NewCampaignService(store, nil)
	result, err := svc.GetByID(context.Background(), pgtype.UUID{Bytes: [16]byte{1}, Valid: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil, got campaign")
	}
}

func TestListByOwner(t *testing.T) {
	store := &mockCampaignStore{
		listByOwner: func(ctx context.Context, ownerID string) ([]sqlc.Campaign, error) {
			if ownerID != "owner1" {
				t.Errorf("expected owner owner1, got %q", ownerID)
			}
			return []sqlc.Campaign{
				testCampaign("1", "owner1", "C1", "draft"),
				testCampaign("2", "owner1", "C2", "active"),
			}, nil
		},
	}
	svc := service.NewCampaignService(store, nil)
	result, err := svc.ListByOwner(context.Background(), "owner1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 campaigns, got %d", len(result))
	}
}
