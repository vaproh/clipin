package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	sqlc "clipin/apps/api/internal/db/sqlc"
	"clipin/apps/api/internal/http/handlers"
	"clipin/apps/api/internal/service"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// mockVerificationService implements VerificationServiceInterface for testing.
type mockVerificationService struct {
	recordSnapshot     func(ctx context.Context, submissionID pgtype.UUID, platform string, views, likes, comments, shares int64, capturedAt *time.Time) (*sqlc.MetricSnapshot, error)
	getVerificationStatus func(ctx context.Context, submissionID pgtype.UUID) (*service.VerificationStatus, error)
	getNeedingVerification func(ctx context.Context, limit int) ([]sqlc.GetSubmissionsNeedingVerificationRow, error)
}

func (m *mockVerificationService) RecordSnapshot(ctx context.Context, submissionID pgtype.UUID, platform string, views, likes, comments, shares int64, capturedAt *time.Time) (*sqlc.MetricSnapshot, error) {
	if m.recordSnapshot != nil {
		return m.recordSnapshot(ctx, submissionID, platform, views, likes, comments, shares, capturedAt)
	}
	return &sqlc.MetricSnapshot{}, nil
}

func (m *mockVerificationService) ComputeEligibleViews(ctx context.Context, submissionID pgtype.UUID) (*service.EligibleViewsResult, error) {
	return &service.EligibleViewsResult{}, nil
}

func (m *mockVerificationService) GetVerificationStatus(ctx context.Context, submissionID pgtype.UUID) (*service.VerificationStatus, error) {
	if m.getVerificationStatus != nil {
		return m.getVerificationStatus(ctx, submissionID)
	}
	return &service.VerificationStatus{}, nil
}

func (m *mockVerificationService) GetSubmissionsNeedingVerification(ctx context.Context, limit int) ([]sqlc.GetSubmissionsNeedingVerificationRow, error) {
	if m.getNeedingVerification != nil {
		return m.getNeedingVerification(ctx, limit)
	}
	return nil, nil
}

func internalVerificationRouter(svc handlers.VerificationServiceInterface) http.Handler {
	r := chi.NewRouter()
	api := humachi.New(r, huma.DefaultConfig("ClipIN API", "1.0.0"))
	handlers.RegisterInternalVerificationHandlers(api, svc)
	return r
}

func authVerificationRouter(svc handlers.VerificationServiceInterface, user *sqlc.User) http.Handler {
	r := chi.NewRouter()
	api := humachi.New(r, huma.DefaultConfig("ClipIN API", "1.0.0"))
	handlers.RegisterVerificationStatusHandlers(api, svc)
	return withUser(user)(r)
}

func TestCreateSnapshot_HappyPath(t *testing.T) {
	svc := &mockVerificationService{
		recordSnapshot: func(ctx context.Context, submissionID pgtype.UUID, platform string, views, likes, comments, shares int64, capturedAt *time.Time) (*sqlc.MetricSnapshot, error) {
			return &sqlc.MetricSnapshot{
				ID:           pgtype.UUID{Bytes: [16]byte{20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35}, Valid: true},
				SubmissionID: submissionID,
				Platform:     platform,
				Views:        views,
				Likes:        likes,
				Comments:     comments,
				Shares:       shares,
				CapturedAt:   pgtype.Timestamptz{Valid: true, Time: time.Now()},
				CreatedAt:    pgtype.Timestamptz{Valid: true, Time: time.Now()},
			}, nil
		},
	}
	router := internalVerificationRouter(svc)
	body := `{"submission_id":"0102030405060708090a0b0c0d0e0f10","platform":"youtube","views":5000,"likes":500,"comments":100,"shares":50}`
	rec := doRequestJSON(t, router, http.MethodPost, "/internal/snapshots", body)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var out struct {
		Platform string `json:"platform"`
		Views    int64  `json:"views"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Platform != "youtube" {
		t.Errorf("expected platform youtube, got %q", out.Platform)
	}
	if out.Views != 5000 {
		t.Errorf("expected views 5000, got %d", out.Views)
	}
}

func TestCreateSnapshot_MissingPlatform(t *testing.T) {
	svc := &mockVerificationService{}
	router := internalVerificationRouter(svc)
	body := `{"submission_id":"0102030405060708090a0b0c0d0e0f10","views":5000}`
	rec := doRequestJSON(t, router, http.MethodPost, "/internal/snapshots", body)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCreateSnapshot_NegativeViews(t *testing.T) {
	svc := &mockVerificationService{}
	router := internalVerificationRouter(svc)
	body := `{"submission_id":"0102030405060708090a0b0c0d0e0f10","platform":"youtube","views":-1}`
	rec := doRequestJSON(t, router, http.MethodPost, "/internal/snapshots", body)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCreateSnapshot_InvalidSubmissionID(t *testing.T) {
	svc := &mockVerificationService{}
	router := internalVerificationRouter(svc)
	body := `{"submission_id":"not-a-uuid","platform":"youtube","views":100}`
	rec := doRequestJSON(t, router, http.MethodPost, "/internal/snapshots", body)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestGetVerificationStatus_Unauthorized(t *testing.T) {
	svc := &mockVerificationService{}
	r := chi.NewRouter()
	api := humachi.New(r, huma.DefaultConfig("ClipIN API", "1.0.0"))
	handlers.RegisterVerificationStatusHandlers(api, svc)

	rec := doRequest(t, r, http.MethodGet, "/submissions/0102030405060708090a0b0c0d0e0f10/verification", "")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestGetVerificationStatus_HappyPath(t *testing.T) {
	svc := &mockVerificationService{
		getVerificationStatus: func(ctx context.Context, id pgtype.UUID) (*service.VerificationStatus, error) {
			return &service.VerificationStatus{
				SubmissionID:  id,
				Status:        "approved",
				HasSnapshots:  true,
				SnapshotCount: 3,
				InitialViews:  100,
				LatestViews:   5000,
				GrowthViews:   4900,
				EligibleViews: 4900,
			}, nil
		},
	}
	router := authVerificationRouter(svc, ownerUser)
	rec := doRequest(t, router, http.MethodGet, "/submissions/0102030405060708090a0b0c0d0e0f10/verification", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var out struct {
		Status        string `json:"status"`
		HasSnapshots  bool   `json:"has_snapshots"`
		SnapshotCount int64  `json:"snapshot_count"`
		EligibleViews int64  `json:"eligible_views"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Status != "approved" {
		t.Errorf("expected status approved, got %q", out.Status)
	}
	if !out.HasSnapshots {
		t.Error("expected has_snapshots true")
	}
	if out.EligibleViews != 4900 {
		t.Errorf("expected eligible views 4900, got %d", out.EligibleViews)
	}
}

func TestGetVerificationStatus_NotFound(t *testing.T) {
	svc := &mockVerificationService{
		getVerificationStatus: func(ctx context.Context, id pgtype.UUID) (*service.VerificationStatus, error) {
			return nil, service.ErrSubmissionNotFound
		},
	}
	router := authVerificationRouter(svc, ownerUser)
	rec := doRequest(t, router, http.MethodGet, "/submissions/0102030405060708090a0b0c0d0e0f10/verification", "")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestListNeedsVerification_HappyPath(t *testing.T) {
	svc := &mockVerificationService{
		getNeedingVerification: func(ctx context.Context, limit int) ([]sqlc.GetSubmissionsNeedingVerificationRow, error) {
			return []sqlc.GetSubmissionsNeedingVerificationRow{
				{
					ID:         pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
					CampaignID: pgtype.UUID{Bytes: [16]byte{2}, Valid: true},
					ClipperID:  "clipper1",
					PostUrl:    "https://youtube.com/watch?v=abc",
					Platform:   "youtube",
					Status:     "approved",
					CpmRate:    150,
				},
			}, nil
		},
	}
	router := internalVerificationRouter(svc)
	body := `{"limit":10}`
	rec := doRequestJSON(t, router, http.MethodPost, "/internal/snapshots/list", body)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var out struct {
		Submissions []struct {
			SubmissionID string `json:"submission_id"`
			Platform     string `json:"platform"`
			CPMRate      int32  `json:"cpm_rate"`
		} `json:"submissions"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out.Submissions) != 1 {
		t.Fatalf("expected 1 submission, got %d", len(out.Submissions))
	}
	if out.Submissions[0].Platform != "youtube" {
		t.Errorf("expected platform youtube, got %q", out.Submissions[0].Platform)
	}
}

func TestListNeedsVerification_DefaultLimit(t *testing.T) {
	svc := &mockVerificationService{
		getNeedingVerification: func(ctx context.Context, limit int) ([]sqlc.GetSubmissionsNeedingVerificationRow, error) {
			if limit != 20 {
				t.Errorf("expected default limit 20, got %d", limit)
			}
			return nil, nil
		},
	}
	router := internalVerificationRouter(svc)
	body := `{"limit":0}`
	rec := doRequestJSON(t, router, http.MethodPost, "/internal/snapshots/list", body)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}
