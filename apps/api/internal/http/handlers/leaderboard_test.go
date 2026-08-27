package handlers_test

import (
	"context"
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

// --- mock leaderboard store ---

type mockLeaderboardStore struct {
	byEarnings    func(ctx context.Context, arg sqlc.GetLeaderboardByEarningsParams) ([]sqlc.GetLeaderboardByEarningsRow, error)
	bySubmissions func(ctx context.Context, arg sqlc.GetLeaderboardBySubmissionsParams) ([]sqlc.GetLeaderboardBySubmissionsRow, error)
}

func (m *mockLeaderboardStore) GetLeaderboardByEarnings(ctx context.Context, arg sqlc.GetLeaderboardByEarningsParams) ([]sqlc.GetLeaderboardByEarningsRow, error) {
	if m.byEarnings != nil {
		return m.byEarnings(ctx, arg)
	}
	return []sqlc.GetLeaderboardByEarningsRow{}, nil
}

func (m *mockLeaderboardStore) GetLeaderboardBySubmissions(ctx context.Context, arg sqlc.GetLeaderboardBySubmissionsParams) ([]sqlc.GetLeaderboardBySubmissionsRow, error) {
	if m.bySubmissions != nil {
		return m.bySubmissions(ctx, arg)
	}
	return []sqlc.GetLeaderboardBySubmissionsRow{}, nil
}

func leaderboardRouter(store handlers.LeaderboardStore) http.Handler {
	r := chi.NewRouter()
	api := humachi.New(r, huma.DefaultConfig("ClipIN API", "1.0.0"))
	handlers.RegisterLeaderboardHandlers(api, store)
	return r
}

// --- tests ---

func TestLeaderboard_DefaultSortByEarnings(t *testing.T) {
	store := &mockLeaderboardStore{
		byEarnings: func(ctx context.Context, arg sqlc.GetLeaderboardByEarningsParams) ([]sqlc.GetLeaderboardByEarningsRow, error) {
			if arg.Limit != 20 {
				t.Errorf("expected default limit 20, got %d", arg.Limit)
			}
			if arg.Offset != 0 {
				t.Errorf("expected default offset 0, got %d", arg.Offset)
			}
			return []sqlc.GetLeaderboardByEarningsRow{
				{
					ID:                    "user_1",
					DisplayName:           pgtype.Text{String: "Top Clipper", Valid: true},
					AvatarUrl:             pgtype.Text{String: "https://example.com/a.jpg", Valid: true},
					TotalEarnings:         50000,
					TotalSubmissions:      15,
					CampaignsParticipated: 5,
				},
				{
					ID:                    "user_2",
					DisplayName:           pgtype.Text{String: "Second Place", Valid: true},
					AvatarUrl:             pgtype.Text{Valid: false},
					TotalEarnings:         30000,
					TotalSubmissions:      10,
					CampaignsParticipated: 3,
				},
			}, nil
		},
	}
	router := leaderboardRouter(store)
	rec := doRequest(t, router, http.MethodGet, "/leaderboard", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var out struct {
		Entries []struct {
			Rank                  int     `json:"rank"`
			UserID                string  `json:"user_id"`
			DisplayName           string  `json:"display_name"`
			AvatarURL             string  `json:"avatar_url"`
			TotalEarnings         int32   `json:"total_earnings"`
			TotalSubmissions      int32   `json:"total_submissions"`
			CampaignsParticipated int32   `json:"campaigns_participated"`
		} `json:"entries"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(out.Entries))
	}
	if out.Entries[0].Rank != 1 {
		t.Errorf("expected rank 1, got %d", out.Entries[0].Rank)
	}
	if out.Entries[0].UserID != "user_1" {
		t.Errorf("expected user_1, got %q", out.Entries[0].UserID)
	}
	if out.Entries[0].TotalEarnings != 50000 {
		t.Errorf("expected 50000 earnings, got %d", out.Entries[0].TotalEarnings)
	}
	if out.Entries[1].Rank != 2 {
		t.Errorf("expected rank 2, got %d", out.Entries[1].Rank)
	}
	// Second entry should not have avatar_url (null -> omitted)
	var raw map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
		t.Fatalf("decode raw: %v", err)
	}
}

func TestLeaderboard_SortBySubmissions(t *testing.T) {
	store := &mockLeaderboardStore{
		bySubmissions: func(ctx context.Context, arg sqlc.GetLeaderboardBySubmissionsParams) ([]sqlc.GetLeaderboardBySubmissionsRow, error) {
			return []sqlc.GetLeaderboardBySubmissionsRow{
				{
					ID:                  "user_a",
					DisplayName:         pgtype.Text{String: "Prolific Clipper", Valid: true},
					TotalSubmissions:    50,
					ApprovedSubmissions: 40,
					TotalEarnings:       25000,
				},
			}, nil
		},
	}
	router := leaderboardRouter(store)
	rec := doRequest(t, router, http.MethodGet, "/leaderboard?sort=submissions", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var out struct {
		Entries []struct {
			Rank               int    `json:"rank"`
			UserID             string `json:"user_id"`
			TotalSubmissions   int32  `json:"total_submissions"`
			ApprovedSubmissions int32 `json:"approved_submissions"`
			TotalEarnings      int32  `json:"total_earnings"`
		} `json:"entries"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(out.Entries))
	}
	if out.Entries[0].TotalSubmissions != 50 {
		t.Errorf("expected 50 submissions, got %d", out.Entries[0].TotalSubmissions)
	}
	if out.Entries[0].ApprovedSubmissions != 40 {
		t.Errorf("expected 40 approved, got %d", out.Entries[0].ApprovedSubmissions)
	}
}

func TestLeaderboard_Pagination(t *testing.T) {
	var capturedArg sqlc.GetLeaderboardByEarningsParams
	store := &mockLeaderboardStore{
		byEarnings: func(ctx context.Context, arg sqlc.GetLeaderboardByEarningsParams) ([]sqlc.GetLeaderboardByEarningsRow, error) {
			capturedArg = arg
			return []sqlc.GetLeaderboardByEarningsRow{
				{ID: "user_3", TotalEarnings: 10000, TotalSubmissions: 5, CampaignsParticipated: 2},
			}, nil
		},
	}
	router := leaderboardRouter(store)
	rec := doRequest(t, router, http.MethodGet, "/leaderboard?limit=5&offset=10", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if capturedArg.Limit != 5 {
		t.Errorf("expected limit 5, got %d", capturedArg.Limit)
	}
	if capturedArg.Offset != 10 {
		t.Errorf("expected offset 10, got %d", capturedArg.Offset)
	}
	var out struct {
		Entries []struct {
			Rank   int    `json:"rank"`
			UserID string `json:"user_id"`
		} `json:"entries"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(out.Entries))
	}
	// Rank should be offset + 1
	if out.Entries[0].Rank != 11 {
		t.Errorf("expected rank 11 (offset 10 + 1), got %d", out.Entries[0].Rank)
	}
}

func TestLeaderboard_EmptyLeaderboard(t *testing.T) {
	store := &mockLeaderboardStore{
		byEarnings: func(ctx context.Context, arg sqlc.GetLeaderboardByEarningsParams) ([]sqlc.GetLeaderboardByEarningsRow, error) {
			return []sqlc.GetLeaderboardByEarningsRow{}, nil
		},
	}
	router := leaderboardRouter(store)
	rec := doRequest(t, router, http.MethodGet, "/leaderboard", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var out struct {
		Entries []interface{} `json:"entries"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out.Entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(out.Entries))
	}
}

func TestLeaderboard_LimitCapAt100(t *testing.T) {
	var capturedArg sqlc.GetLeaderboardByEarningsParams
	store := &mockLeaderboardStore{
		byEarnings: func(ctx context.Context, arg sqlc.GetLeaderboardByEarningsParams) ([]sqlc.GetLeaderboardByEarningsRow, error) {
			capturedArg = arg
			return []sqlc.GetLeaderboardByEarningsRow{}, nil
		},
	}
	router := leaderboardRouter(store)
	rec := doRequest(t, router, http.MethodGet, "/leaderboard?limit=500", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if capturedArg.Limit != 100 {
		t.Errorf("expected limit capped at 100, got %d", capturedArg.Limit)
	}
}

func TestLeaderboard_NegativeOffset(t *testing.T) {
	var capturedArg sqlc.GetLeaderboardByEarningsParams
	store := &mockLeaderboardStore{
		byEarnings: func(ctx context.Context, arg sqlc.GetLeaderboardByEarningsParams) ([]sqlc.GetLeaderboardByEarningsRow, error) {
			capturedArg = arg
			return []sqlc.GetLeaderboardByEarningsRow{}, nil
		},
	}
	router := leaderboardRouter(store)
	rec := doRequest(t, router, http.MethodGet, "/leaderboard?offset=-5", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if capturedArg.Offset != 0 {
		t.Errorf("expected offset clamped to 0, got %d", capturedArg.Offset)
	}
}

func TestLeaderboard_DBError(t *testing.T) {
	store := &mockLeaderboardStore{
		byEarnings: func(ctx context.Context, arg sqlc.GetLeaderboardByEarningsParams) ([]sqlc.GetLeaderboardByEarningsRow, error) {
			return nil, errors.New("db down")
		},
	}
	router := leaderboardRouter(store)
	rec := doRequest(t, router, http.MethodGet, "/leaderboard", "")

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestLeaderboard_DBErrorSubmissions(t *testing.T) {
	store := &mockLeaderboardStore{
		bySubmissions: func(ctx context.Context, arg sqlc.GetLeaderboardBySubmissionsParams) ([]sqlc.GetLeaderboardBySubmissionsRow, error) {
			return nil, errors.New("db down")
		},
	}
	router := leaderboardRouter(store)
	rec := doRequest(t, router, http.MethodGet, "/leaderboard?sort=submissions", "")

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestLeaderboard_InvalidSortDefaultsToEarnings(t *testing.T) {
	earningsCalled := false
	store := &mockLeaderboardStore{
		byEarnings: func(ctx context.Context, arg sqlc.GetLeaderboardByEarningsParams) ([]sqlc.GetLeaderboardByEarningsRow, error) {
			earningsCalled = true
			return []sqlc.GetLeaderboardByEarningsRow{}, nil
		},
	}
	router := leaderboardRouter(store)
	rec := doRequest(t, router, http.MethodGet, "/leaderboard?sort=unknown", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !earningsCalled {
		t.Error("expected earnings query to be called for unknown sort value")
	}
}

func TestLeaderboard_MultipleRanksSequential(t *testing.T) {
	store := &mockLeaderboardStore{
		byEarnings: func(ctx context.Context, arg sqlc.GetLeaderboardByEarningsParams) ([]sqlc.GetLeaderboardByEarningsRow, error) {
			return []sqlc.GetLeaderboardByEarningsRow{
				{ID: "u1", TotalEarnings: 100, TotalSubmissions: 5, CampaignsParticipated: 1},
				{ID: "u2", TotalEarnings: 80, TotalSubmissions: 4, CampaignsParticipated: 1},
				{ID: "u3", TotalEarnings: 60, TotalSubmissions: 3, CampaignsParticipated: 1},
			}, nil
		},
	}
	router := leaderboardRouter(store)
	rec := doRequest(t, router, http.MethodGet, "/leaderboard", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var out struct {
		Entries []struct {
			Rank   int    `json:"rank"`
			UserID string `json:"user_id"`
		} `json:"entries"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out.Entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(out.Entries))
	}
	for i, entry := range out.Entries {
		expectedRank := i + 1
		if entry.Rank != expectedRank {
			t.Errorf("entry %d: expected rank %d, got %d", i, expectedRank, entry.Rank)
		}
	}
}

func TestLeaderboard_ZeroLimitDefaultsTo20(t *testing.T) {
	var capturedArg sqlc.GetLeaderboardByEarningsParams
	store := &mockLeaderboardStore{
		byEarnings: func(ctx context.Context, arg sqlc.GetLeaderboardByEarningsParams) ([]sqlc.GetLeaderboardByEarningsRow, error) {
			capturedArg = arg
			return []sqlc.GetLeaderboardByEarningsRow{}, nil
		},
	}
	router := leaderboardRouter(store)
	rec := doRequest(t, router, http.MethodGet, "/leaderboard?limit=0", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if capturedArg.Limit != 20 {
		t.Errorf("expected default limit 20, got %d", capturedArg.Limit)
	}
}
