package handlers_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	sqlc "clipin/apps/api/internal/db/sqlc"
	"clipin/apps/api/internal/http/handlers"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type mockCampaignAnalyticsStore struct {
	getCampaignByID         func(ctx context.Context, id pgtype.UUID) (sqlc.Campaign, error)
	getCampaignSubmissionStats func(ctx context.Context, campaignID pgtype.UUID) (sqlc.GetCampaignSubmissionStatsRow, error)
	getCampaignViewStats    func(ctx context.Context, campaignID pgtype.UUID) (sqlc.GetCampaignViewStatsRow, error)
	getCampaignFinancialSummary func(ctx context.Context, campaignID pgtype.UUID) (sqlc.GetCampaignFinancialSummaryRow, error)
}

func (m *mockCampaignAnalyticsStore) GetCampaignByID(ctx context.Context, id pgtype.UUID) (sqlc.Campaign, error) {
	if m.getCampaignByID != nil {
		return m.getCampaignByID(ctx, id)
	}
	return sqlc.Campaign{}, sql.ErrNoRows
}

func (m *mockCampaignAnalyticsStore) GetCampaignSubmissionStats(ctx context.Context, campaignID pgtype.UUID) (sqlc.GetCampaignSubmissionStatsRow, error) {
	if m.getCampaignSubmissionStats != nil {
		return m.getCampaignSubmissionStats(ctx, campaignID)
	}
	return sqlc.GetCampaignSubmissionStatsRow{}, nil
}

func (m *mockCampaignAnalyticsStore) GetCampaignViewStats(ctx context.Context, campaignID pgtype.UUID) (sqlc.GetCampaignViewStatsRow, error) {
	if m.getCampaignViewStats != nil {
		return m.getCampaignViewStats(ctx, campaignID)
	}
	return sqlc.GetCampaignViewStatsRow{}, nil
}

func (m *mockCampaignAnalyticsStore) GetCampaignFinancialSummary(ctx context.Context, campaignID pgtype.UUID) (sqlc.GetCampaignFinancialSummaryRow, error) {
	if m.getCampaignFinancialSummary != nil {
		return m.getCampaignFinancialSummary(ctx, campaignID)
	}
	return sqlc.GetCampaignFinancialSummaryRow{}, nil
}

func analyticsRouter(store handlers.CampaignAnalyticsStore, user *sqlc.User) http.Handler {
	r := chi.NewRouter()
	api := humachi.New(r, huma.DefaultConfig("ClipIN API", "1.0.0"))
	handlers.RegisterCampaignAnalyticsHandlers(api, store)
	return withUser(user)(r)
}

func analyticsRouterNoAuth(store handlers.CampaignAnalyticsStore) http.Handler {
	r := chi.NewRouter()
	api := humachi.New(r, huma.DefaultConfig("ClipIN API", "1.0.0"))
	handlers.RegisterCampaignAnalyticsHandlers(api, store)
	return r
}

func testCampaignForAnalytics(ownerID string) sqlc.Campaign {
	return sqlc.Campaign{
		ID:              pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true},
		OwnerID:         ownerID,
		Title:           "Analytics Campaign",
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

const analyticsCampaignID = "01020304-0506-0708-090a-0b0c0d0e0f10"

func TestCampaignAnalytics_Unauthorized(t *testing.T) {
	store := &mockCampaignAnalyticsStore{}
	router := analyticsRouterNoAuth(store)
	rec := doRequest(t, router, http.MethodGet, "/me/campaigns/"+analyticsCampaignID+"/analytics", "")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestCampaignAnalytics_NotFound(t *testing.T) {
	store := &mockCampaignAnalyticsStore{
		getCampaignByID: func(ctx context.Context, id pgtype.UUID) (sqlc.Campaign, error) {
			return sqlc.Campaign{}, sql.ErrNoRows
		},
	}
	router := analyticsRouter(store, ownerUser)
	rec := doRequest(t, router, http.MethodGet, "/me/campaigns/"+analyticsCampaignID+"/analytics", "")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCampaignAnalytics_Forbidden(t *testing.T) {
	c := testCampaignForAnalytics("other_owner")
	store := &mockCampaignAnalyticsStore{
		getCampaignByID: func(ctx context.Context, id pgtype.UUID) (sqlc.Campaign, error) {
			return c, nil
		},
	}
	router := analyticsRouter(store, ownerUser)
	rec := doRequest(t, router, http.MethodGet, "/me/campaigns/"+analyticsCampaignID+"/analytics", "")

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCampaignAnalytics_HappyPath(t *testing.T) {
	c := testCampaignForAnalytics("owner_1")
	store := &mockCampaignAnalyticsStore{
		getCampaignByID: func(ctx context.Context, id pgtype.UUID) (sqlc.Campaign, error) {
			return c, nil
		},
		getCampaignSubmissionStats: func(ctx context.Context, campaignID pgtype.UUID) (sqlc.GetCampaignSubmissionStatsRow, error) {
			return sqlc.GetCampaignSubmissionStatsRow{
				TotalSubmissions: 10,
				Pending:          3,
				Approved:         5,
				Rejected:         2,
				UniqueClippers:   7,
			}, nil
		},
		getCampaignViewStats: func(ctx context.Context, campaignID pgtype.UUID) (sqlc.GetCampaignViewStatsRow, error) {
			return sqlc.GetCampaignViewStatsRow{
				TotalViews:    150000,
				TotalLikes:    12000,
				TotalComments: 3500,
				TotalShares:   800,
			}, nil
		},
		getCampaignFinancialSummary: func(ctx context.Context, campaignID pgtype.UUID) (sqlc.GetCampaignFinancialSummaryRow, error) {
			return sqlc.GetCampaignFinancialSummaryRow{
				TotalEarnings: 20000,
				TotalFees:     5000,
				TotalRefunds:  1000,
			}, nil
		},
	}
	router := analyticsRouter(store, ownerUser)
	rec := doRequest(t, router, http.MethodGet, "/me/campaigns/"+analyticsCampaignID+"/analytics", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var out struct {
		Submissions struct {
			Total    int32 `json:"total"`
			Pending  int32 `json:"pending"`
			Approved int32 `json:"approved"`
			Rejected int32 `json:"rejected"`
		} `json:"submissions"`
		UniqueClippers int32 `json:"unique_clippers"`
		Views struct {
			Total    int64 `json:"total"`
			Likes    int64 `json:"likes"`
			Comments int64 `json:"comments"`
			Shares   int64 `json:"shares"`
		} `json:"views"`
		Financial struct {
			Earnings int32 `json:"earnings"`
			Fees     int32 `json:"fees"`
			Refunds  int32 `json:"refunds"`
		} `json:"financial"`
		Progress struct {
			BudgetConsumedPercent float64 `json:"budget_consumed_percent"`
			CampaignStatus        string  `json:"campaign_status"`
		} `json:"progress"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if out.Submissions.Total != 10 {
		t.Errorf("expected 10 total submissions, got %d", out.Submissions.Total)
	}
	if out.Submissions.Pending != 3 {
		t.Errorf("expected 3 pending, got %d", out.Submissions.Pending)
	}
	if out.Submissions.Approved != 5 {
		t.Errorf("expected 5 approved, got %d", out.Submissions.Approved)
	}
	if out.Submissions.Rejected != 2 {
		t.Errorf("expected 2 rejected, got %d", out.Submissions.Rejected)
	}
	if out.UniqueClippers != 7 {
		t.Errorf("expected 7 unique clippers, got %d", out.UniqueClippers)
	}
	if out.Views.Total != 150000 {
		t.Errorf("expected 150000 views, got %d", out.Views.Total)
	}
	if out.Views.Likes != 12000 {
		t.Errorf("expected 12000 likes, got %d", out.Views.Likes)
	}
	if out.Views.Comments != 3500 {
		t.Errorf("expected 3500 comments, got %d", out.Views.Comments)
	}
	if out.Views.Shares != 800 {
		t.Errorf("expected 800 shares, got %d", out.Views.Shares)
	}
	if out.Financial.Earnings != 20000 {
		t.Errorf("expected 20000 earnings, got %d", out.Financial.Earnings)
	}
	if out.Financial.Fees != 5000 {
		t.Errorf("expected 5000 fees, got %d", out.Financial.Fees)
	}
	if out.Financial.Refunds != 1000 {
		t.Errorf("expected 1000 refunds, got %d", out.Financial.Refunds)
	}
	if out.Progress.CampaignStatus != "active" {
		t.Errorf("expected status 'active', got %q", out.Progress.CampaignStatus)
	}
}

func TestCampaignAnalytics_EmptyCampaign(t *testing.T) {
	c := testCampaignForAnalytics("owner_1")
	store := &mockCampaignAnalyticsStore{
		getCampaignByID: func(ctx context.Context, id pgtype.UUID) (sqlc.Campaign, error) {
			return c, nil
		},
		getCampaignSubmissionStats: func(ctx context.Context, campaignID pgtype.UUID) (sqlc.GetCampaignSubmissionStatsRow, error) {
			return sqlc.GetCampaignSubmissionStatsRow{}, nil
		},
		getCampaignViewStats: func(ctx context.Context, campaignID pgtype.UUID) (sqlc.GetCampaignViewStatsRow, error) {
			return sqlc.GetCampaignViewStatsRow{}, nil
		},
		getCampaignFinancialSummary: func(ctx context.Context, campaignID pgtype.UUID) (sqlc.GetCampaignFinancialSummaryRow, error) {
			return sqlc.GetCampaignFinancialSummaryRow{}, nil
		},
	}
	router := analyticsRouter(store, ownerUser)
	rec := doRequest(t, router, http.MethodGet, "/me/campaigns/"+analyticsCampaignID+"/analytics", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var out struct {
		Submissions struct {
			Total int32 `json:"total"`
		} `json:"submissions"`
		UniqueClippers int32 `json:"unique_clippers"`
		Views          struct {
			Total int64 `json:"total"`
		} `json:"views"`
		Financial struct {
			Earnings int32 `json:"earnings"`
		} `json:"financial"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Submissions.Total != 0 {
		t.Errorf("expected 0 submissions, got %d", out.Submissions.Total)
	}
	if out.UniqueClippers != 0 {
		t.Errorf("expected 0 clippers, got %d", out.UniqueClippers)
	}
	if out.Views.Total != 0 {
		t.Errorf("expected 0 views, got %d", out.Views.Total)
	}
	if out.Financial.Earnings != 0 {
		t.Errorf("expected 0 earnings, got %d", out.Financial.Earnings)
	}
}

func TestCampaignAnalytics_BudgetConsumedPercent(t *testing.T) {
	c := testCampaignForAnalytics("owner_1")
	c.TotalBudget = 100000
	c.RemainingBudget = 40000
	store := &mockCampaignAnalyticsStore{
		getCampaignByID: func(ctx context.Context, id pgtype.UUID) (sqlc.Campaign, error) {
			return c, nil
		},
		getCampaignSubmissionStats: func(ctx context.Context, campaignID pgtype.UUID) (sqlc.GetCampaignSubmissionStatsRow, error) {
			return sqlc.GetCampaignSubmissionStatsRow{}, nil
		},
		getCampaignViewStats: func(ctx context.Context, campaignID pgtype.UUID) (sqlc.GetCampaignViewStatsRow, error) {
			return sqlc.GetCampaignViewStatsRow{}, nil
		},
		getCampaignFinancialSummary: func(ctx context.Context, campaignID pgtype.UUID) (sqlc.GetCampaignFinancialSummaryRow, error) {
			return sqlc.GetCampaignFinancialSummaryRow{}, nil
		},
	}
	router := analyticsRouter(store, ownerUser)
	rec := doRequest(t, router, http.MethodGet, "/me/campaigns/"+analyticsCampaignID+"/analytics", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var out struct {
		Progress struct {
			BudgetConsumedPercent float64 `json:"budget_consumed_percent"`
		} `json:"progress"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Progress.BudgetConsumedPercent != 60.0 {
		t.Errorf("expected 60%% budget consumed, got %f", out.Progress.BudgetConsumedPercent)
	}
}

func TestCampaignAnalytics_TimeRemaining(t *testing.T) {
	c := testCampaignForAnalytics("owner_1")
	c.EndsAt = pgtype.Timestamptz{Time: time.Now().Add(48 * time.Hour), Valid: true}
	store := &mockCampaignAnalyticsStore{
		getCampaignByID: func(ctx context.Context, id pgtype.UUID) (sqlc.Campaign, error) {
			return c, nil
		},
		getCampaignSubmissionStats: func(ctx context.Context, campaignID pgtype.UUID) (sqlc.GetCampaignSubmissionStatsRow, error) {
			return sqlc.GetCampaignSubmissionStatsRow{}, nil
		},
		getCampaignViewStats: func(ctx context.Context, campaignID pgtype.UUID) (sqlc.GetCampaignViewStatsRow, error) {
			return sqlc.GetCampaignViewStatsRow{}, nil
		},
		getCampaignFinancialSummary: func(ctx context.Context, campaignID pgtype.UUID) (sqlc.GetCampaignFinancialSummaryRow, error) {
			return sqlc.GetCampaignFinancialSummaryRow{}, nil
		},
	}
	router := analyticsRouter(store, ownerUser)
	rec := doRequest(t, router, http.MethodGet, "/me/campaigns/"+analyticsCampaignID+"/analytics", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var out struct {
		Progress struct {
			TimeRemaining *string `json:"time_remaining"`
		} `json:"progress"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Progress.TimeRemaining == nil {
		t.Fatal("expected time_remaining to be set")
	}
	if *out.Progress.TimeRemaining == "" {
		t.Error("expected non-empty time_remaining")
	}
}

func TestCampaignAnalytics_ExpiredCampaignNoTimeRemaining(t *testing.T) {
	c := testCampaignForAnalytics("owner_1")
	c.EndsAt = pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour), Valid: true}
	store := &mockCampaignAnalyticsStore{
		getCampaignByID: func(ctx context.Context, id pgtype.UUID) (sqlc.Campaign, error) {
			return c, nil
		},
		getCampaignSubmissionStats: func(ctx context.Context, campaignID pgtype.UUID) (sqlc.GetCampaignSubmissionStatsRow, error) {
			return sqlc.GetCampaignSubmissionStatsRow{}, nil
		},
		getCampaignViewStats: func(ctx context.Context, campaignID pgtype.UUID) (sqlc.GetCampaignViewStatsRow, error) {
			return sqlc.GetCampaignViewStatsRow{}, nil
		},
		getCampaignFinancialSummary: func(ctx context.Context, campaignID pgtype.UUID) (sqlc.GetCampaignFinancialSummaryRow, error) {
			return sqlc.GetCampaignFinancialSummaryRow{}, nil
		},
	}
	router := analyticsRouter(store, ownerUser)
	rec := doRequest(t, router, http.MethodGet, "/me/campaigns/"+analyticsCampaignID+"/analytics", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var out struct {
		Progress struct {
			TimeRemaining *string `json:"time_remaining"`
		} `json:"progress"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Progress.TimeRemaining != nil {
		t.Errorf("expected nil time_remaining for expired campaign, got %q", *out.Progress.TimeRemaining)
	}
}

func TestCampaignAnalytics_ClipperForbidden(t *testing.T) {
	c := testCampaignForAnalytics("owner_1")
	store := &mockCampaignAnalyticsStore{
		getCampaignByID: func(ctx context.Context, id pgtype.UUID) (sqlc.Campaign, error) {
			return c, nil
		},
	}
	router := analyticsRouter(store, clipperUser)
	rec := doRequest(t, router, http.MethodGet, "/me/campaigns/"+analyticsCampaignID+"/analytics", "")

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body=%s", rec.Code, rec.Body.String())
	}
}
