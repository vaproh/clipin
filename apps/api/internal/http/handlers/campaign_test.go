package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	sqlc "clipin/apps/api/internal/db/sqlc"
	"clipin/apps/api/internal/http/handlers"
	"clipin/apps/api/internal/service"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// mockCampaignService implements CampaignServiceInterface for testing.
type mockCampaignService struct {
	listPublic func(ctx context.Context, f service.CampaignFilters) (*service.CampaignListResult, error)
	getByID    func(ctx context.Context, id pgtype.UUID) (*sqlc.Campaign, error)
	listByOwner func(ctx context.Context, ownerID string) ([]sqlc.Campaign, error)
}

func (m *mockCampaignService) ListPublic(ctx context.Context, f service.CampaignFilters) (*service.CampaignListResult, error) {
	if m.listPublic != nil {
		return m.listPublic(ctx, f)
	}
	return &service.CampaignListResult{Campaigns: []sqlc.Campaign{}, Total: 0, Page: 1, PageSize: 20}, nil
}

func (m *mockCampaignService) GetByID(ctx context.Context, id pgtype.UUID) (*sqlc.Campaign, error) {
	if m.getByID != nil {
		return m.getByID(ctx, id)
	}
	return nil, nil
}

func (m *mockCampaignService) ListByOwner(ctx context.Context, ownerID string) ([]sqlc.Campaign, error) {
	if m.listByOwner != nil {
		return m.listByOwner(ctx, ownerID)
	}
	return []sqlc.Campaign{}, nil
}

func testCampaignModel() sqlc.Campaign {
	return sqlc.Campaign{
		ID:              pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true},
		OwnerID:         "owner_1",
		Title:           "Test Campaign",
		Platform:        "youtube",
		Status:          "active",
		CpmRate:         150,
		TotalBudget:     50000,
		RemainingBudget: 30000,
		PlatformFee:     5000,
		CreatedAt:       pgtype.Timestamptz{Valid: true},
		UpdatedAt:       pgtype.Timestamptz{Valid: true},
	}
}

func campaignRouter(svc handlers.CampaignServiceInterface) http.Handler {
	r := chi.NewRouter()
	api := humachi.New(r, huma.DefaultConfig("ClipIN API", "1.0.0"))
	handlers.RegisterCampaignHandlers(api, svc)
	return r
}

func TestListCampaigns_Empty(t *testing.T) {
	svc := &mockCampaignService{}
	router := campaignRouter(svc)
	rec := doRequest(t, router, http.MethodGet, "/campaigns", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var out struct {
		Campaigns []interface{} `json:"campaigns"`
		Total     int64         `json:"total"`
		Page      int           `json:"page"`
		PageSize  int           `json:"page_size"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if out.Page != 1 {
		t.Errorf("expected page 1, got %d", out.Page)
	}
	if out.PageSize != 20 {
		t.Errorf("expected page_size 20, got %d", out.PageSize)
	}
}

func TestListCampaigns_WithResults(t *testing.T) {
	c := testCampaignModel()
	svc := &mockCampaignService{
		listPublic: func(ctx context.Context, f service.CampaignFilters) (*service.CampaignListResult, error) {
			return &service.CampaignListResult{
				Campaigns: []sqlc.Campaign{c},
				Total:     1,
				Page:      1,
				PageSize:  20,
			}, nil
		},
	}
	router := campaignRouter(svc)
	rec := doRequest(t, router, http.MethodGet, "/campaigns?page=1&page_size=10", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var out struct {
		Campaigns []struct {
			ID         string `json:"id"`
			Title      string `json:"title"`
			Platform   string `json:"platform"`
			CpmRate    int32  `json:"cpm_rate"`
			TotalBudget int32 `json:"total_budget"`
		} `json:"campaigns"`
		Total int64 `json:"total"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(out.Campaigns) != 1 {
		t.Fatalf("expected 1 campaign, got %d", len(out.Campaigns))
	}
	if out.Campaigns[0].Title != "Test Campaign" {
		t.Errorf("expected title Test Campaign, got %q", out.Campaigns[0].Title)
	}
	if out.Total != 1 {
		t.Errorf("expected total 1, got %d", out.Total)
	}
}

func TestGetCampaign_NotFound(t *testing.T) {
	svc := &mockCampaignService{
		getByID: func(ctx context.Context, id pgtype.UUID) (*sqlc.Campaign, error) {
			return nil, nil
		},
	}
	router := campaignRouter(svc)
	rec := doRequest(t, router, http.MethodGet, "/campaigns/01020304-0506-0708-090a-0b0c0d0e0f10", "")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestGetCampaign_Found(t *testing.T) {
	c := testCampaignModel()
	svc := &mockCampaignService{
		getByID: func(ctx context.Context, id pgtype.UUID) (*sqlc.Campaign, error) {
			return &c, nil
		},
	}
	router := campaignRouter(svc)
	rec := doRequest(t, router, http.MethodGet, "/campaigns/01020304-0506-0708-090a-0b0c0d0e0f10", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var out struct {
		Title      string `json:"title"`
		Platform   string `json:"platform"`
		CpmRate    int32  `json:"cpm_rate"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if out.Title != "Test Campaign" {
		t.Errorf("expected title Test Campaign, got %q", out.Title)
	}
	if out.CpmRate != 150 {
		t.Errorf("expected cpm_rate 150, got %d", out.CpmRate)
	}
}

func TestMyCampaigns_Unauthorized(t *testing.T) {
	svc := &mockCampaignService{}
	router := campaignRouter(svc)
	rec := doRequest(t, router, http.MethodGet, "/me/campaigns", "")

	// Without auth middleware, requireUser returns 401
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestMyCampaigns_Authorized(t *testing.T) {
	c := testCampaignModel()
	svc := &mockCampaignService{
		listByOwner: func(ctx context.Context, ownerID string) ([]sqlc.Campaign, error) {
			return []sqlc.Campaign{c}, nil
		},
	}

	// Create router with auth middleware that injects a user
	r := chi.NewRouter()
	api := humachi.New(r, huma.DefaultConfig("ClipIN API", "1.0.0"))
	handlers.RegisterCampaignHandlers(api, svc)

	// Wrap with user middleware
	handler := withUser(&sqlc.User{
		ID:    "owner_1",
		Email: "owner@test.com",
		Role:  "owner",
	})(r)

	rec := doRequest(t, handler, http.MethodGet, "/me/campaigns", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var out struct {
		Campaigns []struct {
			Title string `json:"title"`
		} `json:"campaigns"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(out.Campaigns) != 1 {
		t.Fatalf("expected 1 campaign, got %d", len(out.Campaigns))
	}
}
