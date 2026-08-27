package handlers_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	sqlc "clipin/apps/api/internal/db/sqlc"
	"clipin/apps/api/internal/http/handlers"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// --- mock clipper store ---

type mockClipperStore struct {
	getProfile func(ctx context.Context, id string) (sqlc.GetUserPublicProfileRow, error)
	getStats   func(ctx context.Context, clipperID string) (sqlc.GetClipperSubmissionStatsRow, error)
	getViews   func(ctx context.Context, clipperID string) (int64, error)
	getEarnings func(ctx context.Context, clipperID pgtype.Text) (int32, error)
	getCampaignCount func(ctx context.Context, clipperID string) (int32, error)
	getSocial  func(ctx context.Context, userID string) ([]sqlc.ListSocialAccountsByUserIDPublicRow, error)
}

func (m *mockClipperStore) GetUserPublicProfile(ctx context.Context, id string) (sqlc.GetUserPublicProfileRow, error) {
	if m.getProfile != nil {
		return m.getProfile(ctx, id)
	}
	return sqlc.GetUserPublicProfileRow{}, nil
}

func (m *mockClipperStore) GetClipperSubmissionStats(ctx context.Context, clipperID string) (sqlc.GetClipperSubmissionStatsRow, error) {
	if m.getStats != nil {
		return m.getStats(ctx, clipperID)
	}
	return sqlc.GetClipperSubmissionStatsRow{}, nil
}

func (m *mockClipperStore) GetClipperTotalViews(ctx context.Context, clipperID string) (int64, error) {
	if m.getViews != nil {
		return m.getViews(ctx, clipperID)
	}
	return 0, nil
}

func (m *mockClipperStore) GetClipperTotalEarnings(ctx context.Context, clipperID pgtype.Text) (int32, error) {
	if m.getEarnings != nil {
		return m.getEarnings(ctx, clipperID)
	}
	return 0, nil
}

func (m *mockClipperStore) GetClipperCampaignCount(ctx context.Context, clipperID string) (int32, error) {
	if m.getCampaignCount != nil {
		return m.getCampaignCount(ctx, clipperID)
	}
	return 0, nil
}

func (m *mockClipperStore) ListSocialAccountsByUserIDPublic(ctx context.Context, userID string) ([]sqlc.ListSocialAccountsByUserIDPublicRow, error) {
	if m.getSocial != nil {
		return m.getSocial(ctx, userID)
	}
	return []sqlc.ListSocialAccountsByUserIDPublicRow{}, nil
}

func clipperRouter(store handlers.ClipperStore) http.Handler {
	r := chi.NewRouter()
	api := humachi.New(r, huma.DefaultConfig("ClipIN API", "1.0.0"))
	handlers.RegisterClipperHandlers(api, store, nil)
	return r
}

func testPublicProfile() sqlc.GetUserPublicProfileRow {
	return sqlc.GetUserPublicProfileRow{
		ID:          "clipper_1",
		DisplayName: pgtype.Text{String: "Test Clipper", Valid: true},
		AvatarUrl:   pgtype.Text{String: "https://example.com/avatar.jpg", Valid: true},
		Bio:         pgtype.Text{String: "I make clips", Valid: true},
		CreatedAt:   pgtype.Timestamptz{Valid: true},
	}
}

func TestGetClipperProfile_Found(t *testing.T) {
	profile := testPublicProfile()
	store := &mockClipperStore{
		getProfile: func(ctx context.Context, id string) (sqlc.GetUserPublicProfileRow, error) {
			if id != "clipper_1" {
				t.Errorf("expected clipper_1, got %q", id)
			}
			return profile, nil
		},
		getStats: func(ctx context.Context, clipperID string) (sqlc.GetClipperSubmissionStatsRow, error) {
			return sqlc.GetClipperSubmissionStatsRow{
				TotalSubmissions:    10,
				ApprovedSubmissions: 7,
				PendingSubmissions:  2,
				RejectedSubmissions: 1,
			}, nil
		},
		getViews: func(ctx context.Context, clipperID string) (int64, error) {
			return 50000, nil
		},
		getEarnings: func(ctx context.Context, clipperID pgtype.Text) (int32, error) {
			return 12500, nil
		},
		getCampaignCount: func(ctx context.Context, clipperID string) (int32, error) {
			return 4, nil
		},
		getSocial: func(ctx context.Context, userID string) ([]sqlc.ListSocialAccountsByUserIDPublicRow, error) {
			return []sqlc.ListSocialAccountsByUserIDPublicRow{
				{Platform: "youtube", PlatformUsername: pgtype.Text{String: "@testclipper", Valid: true}},
				{Platform: "instagram", PlatformUsername: pgtype.Text{String: "testclipper", Valid: true}},
			}, nil
		},
	}
	router := clipperRouter(store)
	rec := doRequest(t, router, http.MethodGet, "/clippers/clipper_1", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var out struct {
		ID                   string  `json:"id"`
		DisplayName          string  `json:"display_name"`
		AvatarURL            string  `json:"avatar_url"`
		Bio                  string  `json:"bio"`
		TotalSubmissions     int32   `json:"total_submissions"`
		ApprovedSubmissions  int32   `json:"approved_submissions"`
		PendingSubmissions   int32   `json:"pending_submissions"`
		RejectedSubmissions  int32   `json:"rejected_submissions"`
		TotalViews           int64   `json:"total_views"`
		TotalEarnings        int32   `json:"total_earnings"`
		CampaignsParticipated int32  `json:"campaigns_participated"`
		SocialAccounts       []struct {
			Platform string `json:"platform"`
			Username string `json:"username"`
		} `json:"social_accounts"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.ID != "clipper_1" {
		t.Errorf("expected id clipper_1, got %q", out.ID)
	}
	if out.DisplayName != "Test Clipper" {
		t.Errorf("expected display name 'Test Clipper', got %q", out.DisplayName)
	}
	if out.AvatarURL != "https://example.com/avatar.jpg" {
		t.Errorf("expected avatar_url, got %q", out.AvatarURL)
	}
	if out.Bio != "I make clips" {
		t.Errorf("expected bio, got %q", out.Bio)
	}
	if out.TotalSubmissions != 10 {
		t.Errorf("expected 10 total_submissions, got %d", out.TotalSubmissions)
	}
	if out.ApprovedSubmissions != 7 {
		t.Errorf("expected 7 approved_submissions, got %d", out.ApprovedSubmissions)
	}
	if out.TotalViews != 50000 {
		t.Errorf("expected 50000 total_views, got %d", out.TotalViews)
	}
	if out.TotalEarnings != 12500 {
		t.Errorf("expected 12500 total_earnings, got %d", out.TotalEarnings)
	}
	if out.CampaignsParticipated != 4 {
		t.Errorf("expected 4 campaigns_participated, got %d", out.CampaignsParticipated)
	}
	if len(out.SocialAccounts) != 2 {
		t.Fatalf("expected 2 social accounts, got %d", len(out.SocialAccounts))
	}
	if out.SocialAccounts[0].Platform != "youtube" {
		t.Errorf("expected youtube platform, got %q", out.SocialAccounts[0].Platform)
	}
}

func TestGetClipperProfile_NotFound(t *testing.T) {
	store := &mockClipperStore{
		getProfile: func(ctx context.Context, id string) (sqlc.GetUserPublicProfileRow, error) {
			return sqlc.GetUserPublicProfileRow{}, sql.ErrNoRows
		},
	}
	router := clipperRouter(store)
	rec := doRequest(t, router, http.MethodGet, "/clippers/nonexistent", "")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestGetClipperProfile_DBError(t *testing.T) {
	store := &mockClipperStore{
		getProfile: func(ctx context.Context, id string) (sqlc.GetUserPublicProfileRow, error) {
			return sqlc.GetUserPublicProfileRow{}, errors.New("db down")
		},
	}
	router := clipperRouter(store)
	rec := doRequest(t, router, http.MethodGet, "/clippers/clipper_1", "")

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestGetClipperProfile_StatsError(t *testing.T) {
	store := &mockClipperStore{
		getProfile: func(ctx context.Context, id string) (sqlc.GetUserPublicProfileRow, error) {
			return testPublicProfile(), nil
		},
		getStats: func(ctx context.Context, clipperID string) (sqlc.GetClipperSubmissionStatsRow, error) {
			return sqlc.GetClipperSubmissionStatsRow{}, errors.New("stats query failed")
		},
	}
	router := clipperRouter(store)
	rec := doRequest(t, router, http.MethodGet, "/clippers/clipper_1", "")

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestGetClipperProfile_ViewsError(t *testing.T) {
	store := &mockClipperStore{
		getProfile: func(ctx context.Context, id string) (sqlc.GetUserPublicProfileRow, error) {
			return testPublicProfile(), nil
		},
		getStats: func(ctx context.Context, clipperID string) (sqlc.GetClipperSubmissionStatsRow, error) {
			return sqlc.GetClipperSubmissionStatsRow{}, nil
		},
		getViews: func(ctx context.Context, clipperID string) (int64, error) {
			return 0, errors.New("views query failed")
		},
	}
	router := clipperRouter(store)
	rec := doRequest(t, router, http.MethodGet, "/clippers/clipper_1", "")

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestGetClipperProfile_EarningsError(t *testing.T) {
	store := &mockClipperStore{
		getProfile: func(ctx context.Context, id string) (sqlc.GetUserPublicProfileRow, error) {
			return testPublicProfile(), nil
		},
		getStats: func(ctx context.Context, clipperID string) (sqlc.GetClipperSubmissionStatsRow, error) {
			return sqlc.GetClipperSubmissionStatsRow{}, nil
		},
		getViews: func(ctx context.Context, clipperID string) (int64, error) {
			return 0, nil
		},
		getEarnings: func(ctx context.Context, clipperID pgtype.Text) (int32, error) {
			return 0, errors.New("earnings query failed")
		},
	}
	router := clipperRouter(store)
	rec := doRequest(t, router, http.MethodGet, "/clippers/clipper_1", "")

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestGetClipperProfile_CampaignCountError(t *testing.T) {
	store := &mockClipperStore{
		getProfile: func(ctx context.Context, id string) (sqlc.GetUserPublicProfileRow, error) {
			return testPublicProfile(), nil
		},
		getStats: func(ctx context.Context, clipperID string) (sqlc.GetClipperSubmissionStatsRow, error) {
			return sqlc.GetClipperSubmissionStatsRow{}, nil
		},
		getViews: func(ctx context.Context, clipperID string) (int64, error) {
			return 0, nil
		},
		getEarnings: func(ctx context.Context, clipperID pgtype.Text) (int32, error) {
			return 0, nil
		},
		getCampaignCount: func(ctx context.Context, clipperID string) (int32, error) {
			return 0, errors.New("count query failed")
		},
	}
	router := clipperRouter(store)
	rec := doRequest(t, router, http.MethodGet, "/clippers/clipper_1", "")

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestGetClipperProfile_SocialAccountsError(t *testing.T) {
	store := &mockClipperStore{
		getProfile: func(ctx context.Context, id string) (sqlc.GetUserPublicProfileRow, error) {
			return testPublicProfile(), nil
		},
		getStats: func(ctx context.Context, clipperID string) (sqlc.GetClipperSubmissionStatsRow, error) {
			return sqlc.GetClipperSubmissionStatsRow{}, nil
		},
		getViews: func(ctx context.Context, clipperID string) (int64, error) {
			return 0, nil
		},
		getEarnings: func(ctx context.Context, clipperID pgtype.Text) (int32, error) {
			return 0, nil
		},
		getCampaignCount: func(ctx context.Context, clipperID string) (int32, error) {
			return 0, nil
		},
		getSocial: func(ctx context.Context, userID string) ([]sqlc.ListSocialAccountsByUserIDPublicRow, error) {
			return nil, errors.New("social query failed")
		},
	}
	router := clipperRouter(store)
	rec := doRequest(t, router, http.MethodGet, "/clippers/clipper_1", "")

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestGetClipperProfile_OmitsSensitiveFields(t *testing.T) {
	store := &mockClipperStore{
		getProfile: func(ctx context.Context, id string) (sqlc.GetUserPublicProfileRow, error) {
			return sqlc.GetUserPublicProfileRow{
				ID:          "clipper_1",
				DisplayName: pgtype.Text{String: "Safe Name", Valid: true},
			}, nil
		},
	}
	router := clipperRouter(store)
	rec := doRequest(t, router, http.MethodGet, "/clippers/clipper_1", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
		t.Fatalf("decode: %v", err)
	}
	for _, key := range []string{"email", "upi_id", "role", "updated_at"} {
		if _, ok := raw[key]; ok {
			t.Errorf("sensitive field %q should not be in response", key)
		}
	}
}

func TestGetClipperProfile_ZeroStats(t *testing.T) {
	store := &mockClipperStore{
		getProfile: func(ctx context.Context, id string) (sqlc.GetUserPublicProfileRow, error) {
			return sqlc.GetUserPublicProfileRow{
				ID:          "new_clipper",
				DisplayName: pgtype.Text{String: "New Clipper", Valid: true},
			}, nil
		},
		getStats: func(ctx context.Context, clipperID string) (sqlc.GetClipperSubmissionStatsRow, error) {
			return sqlc.GetClipperSubmissionStatsRow{}, nil
		},
		getViews: func(ctx context.Context, clipperID string) (int64, error) {
			return 0, nil
		},
		getEarnings: func(ctx context.Context, clipperID pgtype.Text) (int32, error) {
			return 0, nil
		},
		getCampaignCount: func(ctx context.Context, clipperID string) (int32, error) {
			return 0, nil
		},
		getSocial: func(ctx context.Context, userID string) ([]sqlc.ListSocialAccountsByUserIDPublicRow, error) {
			return []sqlc.ListSocialAccountsByUserIDPublicRow{}, nil
		},
	}
	router := clipperRouter(store)
	rec := doRequest(t, router, http.MethodGet, "/clippers/new_clipper", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var out struct {
		ID                   string `json:"id"`
		TotalSubmissions     int32  `json:"total_submissions"`
		TotalViews           int64  `json:"total_views"`
		TotalEarnings        int32  `json:"total_earnings"`
		CampaignsParticipated int32 `json:"campaigns_participated"`
		SocialAccounts       []struct{} `json:"social_accounts"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.TotalSubmissions != 0 {
		t.Errorf("expected 0 submissions, got %d", out.TotalSubmissions)
	}
	if out.TotalViews != 0 {
		t.Errorf("expected 0 views, got %d", out.TotalViews)
	}
	if out.TotalEarnings != 0 {
		t.Errorf("expected 0 earnings, got %d", out.TotalEarnings)
	}
	if out.CampaignsParticipated != 0 {
		t.Errorf("expected 0 campaigns, got %d", out.CampaignsParticipated)
	}
	if len(out.SocialAccounts) != 0 {
		t.Errorf("expected 0 social accounts, got %d", len(out.SocialAccounts))
	}
}

func TestGetClipperProfile_NilOptionalFields(t *testing.T) {
	store := &mockClipperStore{
		getProfile: func(ctx context.Context, id string) (sqlc.GetUserPublicProfileRow, error) {
			return sqlc.GetUserPublicProfileRow{
				ID:          "clipper_1",
				DisplayName: pgtype.Text{Valid: false},
				AvatarUrl:   pgtype.Text{Valid: false},
				Bio:         pgtype.Text{Valid: false},
			}, nil
		},
	}
	router := clipperRouter(store)
	rec := doRequest(t, router, http.MethodGet, "/clippers/clipper_1", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
		t.Fatalf("decode: %v", err)
	}
	for _, key := range []string{"display_name", "avatar_url", "bio"} {
		if v, ok := raw[key]; ok && v != nil {
			t.Errorf("null optional field %q should be omitted or null, got %v", key, v)
		}
	}
}
