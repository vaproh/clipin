package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
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
	listPublic  func(ctx context.Context, f service.CampaignFilters) (*service.CampaignListResult, error)
	getByID     func(ctx context.Context, id pgtype.UUID) (*sqlc.Campaign, error)
	listByOwner func(ctx context.Context, ownerID string) ([]sqlc.Campaign, error)
	create      func(ctx context.Context, ownerID string, in *service.CreateCampaignInput) (*sqlc.Campaign, error)
	update      func(ctx context.Context, ownerID string, campaignID pgtype.UUID, in *service.UpdateCampaignInput) (*sqlc.Campaign, error)
	pause       func(ctx context.Context, ownerID string, campaignID pgtype.UUID) (*sqlc.Campaign, error)
	resume      func(ctx context.Context, ownerID string, campaignID pgtype.UUID) (*sqlc.Campaign, error)
	cancel      func(ctx context.Context, ownerID string, campaignID pgtype.UUID) (*sqlc.Campaign, error)
	ownerStats  func(ctx context.Context, ownerID string) (*service.OwnerStats, error)
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

func (m *mockCampaignService) Create(ctx context.Context, ownerID string, in *service.CreateCampaignInput) (*sqlc.Campaign, error) {
	if m.create != nil {
		return m.create(ctx, ownerID, in)
	}
	return nil, nil
}

func (m *mockCampaignService) Update(ctx context.Context, ownerID string, campaignID pgtype.UUID, in *service.UpdateCampaignInput) (*sqlc.Campaign, error) {
	if m.update != nil {
		return m.update(ctx, ownerID, campaignID, in)
	}
	return nil, nil
}

func (m *mockCampaignService) Pause(ctx context.Context, ownerID string, campaignID pgtype.UUID) (*sqlc.Campaign, error) {
	if m.pause != nil {
		return m.pause(ctx, ownerID, campaignID)
	}
	return nil, nil
}

func (m *mockCampaignService) Resume(ctx context.Context, ownerID string, campaignID pgtype.UUID) (*sqlc.Campaign, error) {
	if m.resume != nil {
		return m.resume(ctx, ownerID, campaignID)
	}
	return nil, nil
}

func (m *mockCampaignService) Cancel(ctx context.Context, ownerID string, campaignID pgtype.UUID) (*sqlc.Campaign, error) {
	if m.cancel != nil {
		return m.cancel(ctx, ownerID, campaignID)
	}
	return nil, nil
}

func (m *mockCampaignService) OwnerStats(ctx context.Context, ownerID string) (*service.OwnerStats, error) {
	if m.ownerStats != nil {
		return m.ownerStats(ctx, ownerID)
	}
	return &service.OwnerStats{}, nil
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

// ownerRouter registers both public and owner endpoints, wrapping with auth.
func ownerRouter(svc handlers.CampaignServiceInterface, user *sqlc.User) http.Handler {
	r := chi.NewRouter()
	api := humachi.New(r, huma.DefaultConfig("ClipIN API", "1.0.0"))
	handlers.RegisterCampaignHandlers(api, svc)
	handlers.RegisterCampaignOwnerHandlers(api, svc)
	return withUser(user)(r)
}

// ownerRouterNoAuth registers owner endpoints without auth (for 401 tests).
func ownerRouterNoAuth(svc handlers.CampaignServiceInterface) http.Handler {
	r := chi.NewRouter()
	api := humachi.New(r, huma.DefaultConfig("ClipIN API", "1.0.0"))
	handlers.RegisterCampaignHandlers(api, svc)
	handlers.RegisterCampaignOwnerHandlers(api, svc)
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
			ID          string `json:"id"`
			Title       string `json:"title"`
			Platform    string `json:"platform"`
			CpmRate     int32  `json:"cpm_rate"`
			TotalBudget int32  `json:"total_budget"`
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
		Title   string `json:"title"`
		Platform string `json:"platform"`
		CpmRate int32  `json:"cpm_rate"`
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

	r := chi.NewRouter()
	api := humachi.New(r, huma.DefaultConfig("ClipIN API", "1.0.0"))
	handlers.RegisterCampaignHandlers(api, svc)

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

// --- Owner endpoint tests ---

var ownerUser = &sqlc.User{
	ID:    "owner_1",
	Email: "owner@test.com",
	Role:  "owner",
}

var clipperUser = &sqlc.User{
	ID:    "clipper_1",
	Email: "clipper@test.com",
	Role:  "clipper",
}

const campaignIDHex = "0102030405060708090a0b0c0d0e0f10"

// doRequestJSON is like doRequest but sets Content-Type for JSON bodies.
func doRequestJSON(t *testing.T, handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func TestCreateCampaign_Unauthorized(t *testing.T) {
	svc := &mockCampaignService{}
	router := ownerRouterNoAuth(svc)
	body := `{"title":"Test","platform":"youtube","cpm_rate":150,"total_budget":50000}`
	rec := doRequestJSON(t, router, http.MethodPost, "/campaigns", body)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestCreateCampaign_ClipperForbidden(t *testing.T) {
	svc := &mockCampaignService{}
	router := ownerRouter(svc, clipperUser)
	body := `{"title":"Test","platform":"youtube","cpm_rate":150,"total_budget":50000}`
	rec := doRequestJSON(t, router, http.MethodPost, "/campaigns", body)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCreateCampaign_HappyPath(t *testing.T) {
	c := testCampaignModel()
	svc := &mockCampaignService{
		create: func(ctx context.Context, ownerID string, in *service.CreateCampaignInput) (*sqlc.Campaign, error) {
			if ownerID != "owner_1" {
				t.Errorf("expected owner_1, got %q", ownerID)
			}
			if in.Title != "New Campaign" {
				t.Errorf("expected title 'New Campaign', got %q", in.Title)
			}
			return &c, nil
		},
	}
	router := ownerRouter(svc, ownerUser)
	body := `{"title":"New Campaign","platform":"youtube","cpm_rate":150,"total_budget":50000}`
	rec := doRequestJSON(t, router, http.MethodPost, "/campaigns", body)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var out struct {
		Title    string `json:"title"`
		Platform string `json:"platform"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Title != "Test Campaign" {
		t.Errorf("expected title Test Campaign, got %q", out.Title)
	}
}

func TestCreateCampaign_ValidationError(t *testing.T) {
	svc := &mockCampaignService{
		create: func(ctx context.Context, ownerID string, in *service.CreateCampaignInput) (*sqlc.Campaign, error) {
			return nil, &service.ValidationError{Errors: []string{"title is required", "platform must be one of: youtube, instagram, multi"}}
		},
	}
	router := ownerRouter(svc, ownerUser)
	body := `{"title":"","platform":"bad","cpm_rate":0,"total_budget":0}`
	rec := doRequestJSON(t, router, http.MethodPost, "/campaigns", body)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestUpdateCampaign_Unauthorized(t *testing.T) {
	svc := &mockCampaignService{}
	router := ownerRouterNoAuth(svc)
	body := `{"title":"Updated"}`
	rec := doRequestJSON(t, router, http.MethodPatch, "/campaigns/"+campaignIDHex, body)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestUpdateCampaign_NotFound(t *testing.T) {
	svc := &mockCampaignService{
		update: func(ctx context.Context, ownerID string, id pgtype.UUID, in *service.UpdateCampaignInput) (*sqlc.Campaign, error) {
			return nil, service.ErrCampaignNotFound
		},
	}
	router := ownerRouter(svc, ownerUser)
	body := `{"title":"Updated"}`
	rec := doRequestJSON(t, router, http.MethodPatch, "/campaigns/"+campaignIDHex, body)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestUpdateCampaign_Forbidden(t *testing.T) {
	svc := &mockCampaignService{
		update: func(ctx context.Context, ownerID string, id pgtype.UUID, in *service.UpdateCampaignInput) (*sqlc.Campaign, error) {
			return nil, service.ErrNotOwner
		},
	}
	router := ownerRouter(svc, ownerUser)
	body := `{"title":"Updated"}`
	rec := doRequestJSON(t, router, http.MethodPatch, "/campaigns/"+campaignIDHex, body)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestPauseCampaign_HappyPath(t *testing.T) {
	c := testCampaignModel()
	c.Status = "paused"
	svc := &mockCampaignService{
		pause: func(ctx context.Context, ownerID string, id pgtype.UUID) (*sqlc.Campaign, error) {
			return &c, nil
		},
	}
	router := ownerRouter(svc, ownerUser)
	rec := doRequestJSON(t, router, http.MethodPost, "/campaigns/"+campaignIDHex+"/pause", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestPauseCampaign_Conflict(t *testing.T) {
	svc := &mockCampaignService{
		pause: func(ctx context.Context, ownerID string, id pgtype.UUID) (*sqlc.Campaign, error) {
			return nil, fmt.Errorf("campaign must be active to pause, current status: draft")
		},
	}
	router := ownerRouter(svc, ownerUser)
	rec := doRequestJSON(t, router, http.MethodPost, "/campaigns/"+campaignIDHex+"/pause", "")

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestResumeCampaign_HappyPath(t *testing.T) {
	c := testCampaignModel()
	c.Status = "active"
	svc := &mockCampaignService{
		resume: func(ctx context.Context, ownerID string, id pgtype.UUID) (*sqlc.Campaign, error) {
			return &c, nil
		},
	}
	router := ownerRouter(svc, ownerUser)
	rec := doRequestJSON(t, router, http.MethodPost, "/campaigns/"+campaignIDHex+"/resume", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCancelCampaign_HappyPath(t *testing.T) {
	c := testCampaignModel()
	c.Status = "cancelled"
	svc := &mockCampaignService{
		cancel: func(ctx context.Context, ownerID string, id pgtype.UUID) (*sqlc.Campaign, error) {
			return &c, nil
		},
	}
	router := ownerRouter(svc, ownerUser)
	rec := doRequestJSON(t, router, http.MethodPost, "/campaigns/"+campaignIDHex+"/cancel", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestOwnerStats_HappyPath(t *testing.T) {
	svc := &mockCampaignService{
		ownerStats: func(ctx context.Context, ownerID string) (*service.OwnerStats, error) {
			return &service.OwnerStats{
				TotalCampaigns:  5,
				ActiveCampaigns: 3,
				TotalBudget:     250000,
				TotalRemaining:  150000,
			}, nil
		},
	}
	router := ownerRouter(svc, ownerUser)
	rec := doRequest(t, router, http.MethodGet, "/me/campaigns/stats", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var out struct {
		TotalCampaigns  int64 `json:"total_campaigns"`
		ActiveCampaigns int64 `json:"active_campaigns"`
		TotalBudget     int64 `json:"total_budget"`
		TotalRemaining  int64 `json:"total_remaining"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.TotalCampaigns != 5 {
		t.Errorf("expected 5 total, got %d", out.TotalCampaigns)
	}
	if out.ActiveCampaigns != 3 {
		t.Errorf("expected 3 active, got %d", out.ActiveCampaigns)
	}
}

func TestOwnerStats_Unauthorized(t *testing.T) {
	svc := &mockCampaignService{}
	router := ownerRouterNoAuth(svc)
	rec := doRequest(t, router, http.MethodGet, "/me/campaigns/stats", "")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestOwnerStats_ClipperForbidden(t *testing.T) {
	svc := &mockCampaignService{}
	router := ownerRouter(svc, clipperUser)
	rec := doRequest(t, router, http.MethodGet, "/me/campaigns/stats", "")

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body=%s", rec.Code, rec.Body.String())
	}
}

// --- Search tests ---

func TestListCampaigns_SearchByQ(t *testing.T) {
	svc := &mockCampaignService{
		listPublic: func(ctx context.Context, f service.CampaignFilters) (*service.CampaignListResult, error) {
			if f.Search != "promo" {
				t.Errorf("expected search 'promo', got %q", f.Search)
			}
			c := testCampaignModel()
			c.Title = "Summer Promo"
			return &service.CampaignListResult{
				Campaigns: []sqlc.Campaign{c},
				Total:     1,
				Page:      1,
				PageSize:  20,
			}, nil
		},
	}
	router := campaignRouter(svc)
	rec := doRequest(t, router, http.MethodGet, "/campaigns?q=promo", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var out struct {
		Campaigns []struct {
			Title string `json:"title"`
		} `json:"campaigns"`
		Total int64 `json:"total"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Total != 1 {
		t.Errorf("expected total 1, got %d", out.Total)
	}
	if out.Campaigns[0].Title != "Summer Promo" {
		t.Errorf("expected title 'Summer Promo', got %q", out.Campaigns[0].Title)
	}
}

func TestListCampaigns_SearchEmpty(t *testing.T) {
	svc := &mockCampaignService{
		listPublic: func(ctx context.Context, f service.CampaignFilters) (*service.CampaignListResult, error) {
			if f.Search != "" {
				t.Errorf("expected empty search, got %q", f.Search)
			}
			return &service.CampaignListResult{
				Campaigns: []sqlc.Campaign{},
				Total:     0,
				Page:      1,
				PageSize:  20,
			}, nil
		},
	}
	router := campaignRouter(svc)
	rec := doRequest(t, router, http.MethodGet, "/campaigns", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestListCampaigns_SearchWithFilters(t *testing.T) {
	svc := &mockCampaignService{
		listPublic: func(ctx context.Context, f service.CampaignFilters) (*service.CampaignListResult, error) {
			if f.Search != "gaming" {
				t.Errorf("expected search 'gaming', got %q", f.Search)
			}
			if f.Platform != "youtube" {
				t.Errorf("expected platform 'youtube', got %q", f.Platform)
			}
			if f.MaxCPM != 200 {
				t.Errorf("expected maxCPM 200, got %d", f.MaxCPM)
			}
			return &service.CampaignListResult{
				Campaigns: []sqlc.Campaign{},
				Total:     0,
				Page:      1,
				PageSize:  20,
			}, nil
		},
	}
	router := campaignRouter(svc)
	rec := doRequest(t, router, http.MethodGet, "/campaigns?q=gaming&platform=youtube&max_cpm=200", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestListCampaigns_SearchNoResults(t *testing.T) {
	svc := &mockCampaignService{
		listPublic: func(ctx context.Context, f service.CampaignFilters) (*service.CampaignListResult, error) {
			return &service.CampaignListResult{
				Campaigns: []sqlc.Campaign{},
				Total:     0,
				Page:      1,
				PageSize:  20,
			}, nil
		},
	}
	router := campaignRouter(svc)
	rec := doRequest(t, router, http.MethodGet, "/campaigns?q=nonexistent", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var out struct {
		Campaigns []interface{} `json:"campaigns"`
		Total     int64         `json:"total"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Total != 0 {
		t.Errorf("expected total 0, got %d", out.Total)
	}
	if len(out.Campaigns) != 0 {
		t.Errorf("expected 0 campaigns, got %d", len(out.Campaigns))
	}
}
