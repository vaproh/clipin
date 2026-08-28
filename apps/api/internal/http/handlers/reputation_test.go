package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"clipin/apps/api/internal/http/handlers"
	"clipin/apps/api/internal/service"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
)

// --- mock reputation service ---

type mockReputationService struct {
	getReputation func(ctx context.Context, userID string) (*service.ClipperReputation, error)
}

func (m *mockReputationService) GetReputation(ctx context.Context, userID string) (*service.ClipperReputation, error) {
	return m.getReputation(ctx, userID)
}

func reputationRouter(svc handlers.ReputationServiceInterface) http.Handler {
	r := chi.NewRouter()
	api := humachi.New(r, huma.DefaultConfig("ClipIN API", "1.0.0"))
	handlers.RegisterReputationHandlers(api, svc)
	return r
}

func TestGetClipperReputation_Success(t *testing.T) {
	svc := &mockReputationService{
		getReputation: func(ctx context.Context, userID string) (*service.ClipperReputation, error) {
			return &service.ClipperReputation{
				UserID:                userID,
				Tier:                  service.ComputeTier(500000),
				TotalEarnings:         500000,
				TotalSubmissions:      20,
				CampaignsParticipated: 5,
				MemberSince:           "2025-01-01T00:00:00Z",
			}, nil
		},
	}
	router := reputationRouter(svc)
	rec := doRequest(t, router, http.MethodGet, "/clippers/clipper1/reputation", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var out struct {
		UserID                string `json:"user_id"`
		Tier                  struct {
			Name        string `json:"name"`
			Label       string `json:"label"`
			MinEarnings int64  `json:"min_earnings"`
			NextTier    string `json:"next_tier"`
			NextMin     int64  `json:"next_min_earnings"`
		} `json:"tier"`
		TotalEarnings         int64 `json:"total_earnings"`
		TotalSubmissions      int   `json:"total_submissions"`
		CampaignsParticipated int   `json:"campaigns_participated"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.UserID != "clipper1" {
		t.Errorf("expected user_id clipper1, got %q", out.UserID)
	}
	if out.Tier.Name != "silver" {
		t.Errorf("expected tier name silver, got %q", out.Tier.Name)
	}
	if out.Tier.Label != "Active Clipper" {
		t.Errorf("expected tier label Active Clipper, got %q", out.Tier.Label)
	}
	if out.Tier.MinEarnings != 100000 {
		t.Errorf("expected min_earnings 100000, got %d", out.Tier.MinEarnings)
	}
	if out.Tier.NextTier != "gold" {
		t.Errorf("expected next_tier gold, got %q", out.Tier.NextTier)
	}
	if out.TotalEarnings != 500000 {
		t.Errorf("expected total_earnings 500000, got %d", out.TotalEarnings)
	}
	if out.TotalSubmissions != 20 {
		t.Errorf("expected total_submissions 20, got %d", out.TotalSubmissions)
	}
}

func TestGetClipperReputation_NotFound(t *testing.T) {
	svc := &mockReputationService{
		getReputation: func(ctx context.Context, userID string) (*service.ClipperReputation, error) {
			return nil, service.ErrSubmissionNotFound // reuse an existing error type for testing
		},
	}
	router := reputationRouter(svc)
	rec := doRequest(t, router, http.MethodGet, "/clippers/nonexistent/reputation", "")

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestGetClipperReputation_BronzeTier(t *testing.T) {
	svc := &mockReputationService{
		getReputation: func(ctx context.Context, userID string) (*service.ClipperReputation, error) {
			return &service.ClipperReputation{
				UserID:                userID,
				Tier:                  service.ComputeTier(0),
				TotalEarnings:         0,
				TotalSubmissions:      2,
				CampaignsParticipated: 1,
				MemberSince:           "2025-06-01T00:00:00Z",
			}, nil
		},
	}
	router := reputationRouter(svc)
	rec := doRequest(t, router, http.MethodGet, "/clippers/newbie/reputation", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var out struct {
		Tier struct {
			Name     string  `json:"name"`
			Label    string  `json:"label"`
			NextTier *string `json:"next_tier"`
		} `json:"tier"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Tier.Name != "bronze" {
		t.Errorf("expected bronze, got %q", out.Tier.Name)
	}
	if out.Tier.NextTier == nil || *out.Tier.NextTier != "silver" {
		t.Errorf("expected next_tier silver, got %v", out.Tier.NextTier)
	}
}

func TestGetClipperReputation_PlatinumTier(t *testing.T) {
	svc := &mockReputationService{
		getReputation: func(ctx context.Context, userID string) (*service.ClipperReputation, error) {
			return &service.ClipperReputation{
				UserID:                userID,
				Tier:                  service.ComputeTier(5000000),
				TotalEarnings:         5000000,
				TotalSubmissions:      100,
				CampaignsParticipated: 20,
				MemberSince:           "2024-01-01T00:00:00Z",
			}, nil
		},
	}
	router := reputationRouter(svc)
	rec := doRequest(t, router, http.MethodGet, "/clippers/elite/reputation", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var out struct {
		Tier struct {
			Name    string  `json:"name"`
			NextTier *string `json:"next_tier"`
		} `json:"tier"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Tier.Name != "platinum" {
		t.Errorf("expected platinum, got %q", out.Tier.Name)
	}
	if out.Tier.NextTier != nil {
		t.Errorf("expected no next_tier for platinum, got %v", out.Tier.NextTier)
	}
}
