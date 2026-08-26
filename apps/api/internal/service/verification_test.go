package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	sqlc "clipin/apps/api/internal/db/sqlc"
	"clipin/apps/api/internal/service"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// mockVerificationStore implements VerificationStore for testing.
type mockVerificationStore struct {
	getByID               func(ctx context.Context, id pgtype.UUID) (sqlc.Submission, error)
	getCampaignByID       func(ctx context.Context, id pgtype.UUID) (sqlc.Campaign, error)
	createSnapshot        func(ctx context.Context, arg sqlc.CreateMetricSnapshotParams) (sqlc.MetricSnapshot, error)
	getLatestSnapshot     func(ctx context.Context, submissionID pgtype.UUID) (sqlc.MetricSnapshot, error)
	getInitialSnapshot    func(ctx context.Context, submissionID pgtype.UUID) (sqlc.MetricSnapshot, error)
	listSnapshots         func(ctx context.Context, submissionID pgtype.UUID) ([]sqlc.MetricSnapshot, error)
	countSnapshots        func(ctx context.Context, submissionID pgtype.UUID) (int64, error)
	getSubWithCampaign    func(ctx context.Context, id pgtype.UUID) (sqlc.GetSubmissionWithCampaignRow, error)
	getNeedingVerification func(ctx context.Context, limit int32) ([]sqlc.GetSubmissionsNeedingVerificationRow, error)
}

func (m *mockVerificationStore) GetSubmissionByID(ctx context.Context, id pgtype.UUID) (sqlc.Submission, error) {
	return m.getByID(ctx, id)
}

func (m *mockVerificationStore) GetCampaignByID(ctx context.Context, id pgtype.UUID) (sqlc.Campaign, error) {
	return m.getCampaignByID(ctx, id)
}

func (m *mockVerificationStore) CreateMetricSnapshot(ctx context.Context, arg sqlc.CreateMetricSnapshotParams) (sqlc.MetricSnapshot, error) {
	return m.createSnapshot(ctx, arg)
}

func (m *mockVerificationStore) GetLatestSnapshotForSubmission(ctx context.Context, submissionID pgtype.UUID) (sqlc.MetricSnapshot, error) {
	return m.getLatestSnapshot(ctx, submissionID)
}

func (m *mockVerificationStore) GetInitialSnapshotForSubmission(ctx context.Context, submissionID pgtype.UUID) (sqlc.MetricSnapshot, error) {
	return m.getInitialSnapshot(ctx, submissionID)
}

func (m *mockVerificationStore) ListSnapshotsBySubmission(ctx context.Context, submissionID pgtype.UUID) ([]sqlc.MetricSnapshot, error) {
	return m.listSnapshots(ctx, submissionID)
}

func (m *mockVerificationStore) CountSnapshotsBySubmission(ctx context.Context, submissionID pgtype.UUID) (int64, error) {
	return m.countSnapshots(ctx, submissionID)
}

func (m *mockVerificationStore) GetSubmissionWithCampaign(ctx context.Context, id pgtype.UUID) (sqlc.GetSubmissionWithCampaignRow, error) {
	return m.getSubWithCampaign(ctx, id)
}

func (m *mockVerificationStore) GetSubmissionsNeedingVerification(ctx context.Context, limit int32) ([]sqlc.GetSubmissionsNeedingVerificationRow, error) {
	return m.getNeedingVerification(ctx, limit)
}

func testSnapshotID() pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte{20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35}, Valid: true}
}

func testSnap(views int64) sqlc.MetricSnapshot {
	return sqlc.MetricSnapshot{
		ID:           testSnapshotID(),
		SubmissionID: testCampaignID,
		Platform:     "youtube",
		Views:        views,
		Likes:        views / 10,
		Comments:     views / 20,
		Shares:       views / 50,
		CapturedAt:   pgtype.Timestamptz{Valid: true, Time: time.Now()},
		CreatedAt:    pgtype.Timestamptz{Valid: true, Time: time.Now()},
	}
}

func testSubWithCampaign(minViews int32) sqlc.GetSubmissionWithCampaignRow {
	return sqlc.GetSubmissionWithCampaignRow{
		ID:              testCampaignID,
		CampaignID:      testCampaignID,
		ClipperID:       "clipper1",
		PostUrl:         "https://youtube.com/watch?v=abc",
		Platform:        "youtube",
		Status:          "approved",
		MinViewsPerClip: pgtype.Int4{Valid: true, Int32: minViews},
		CpmRate:         150,
		OwnerID:         "owner1",
	}
}

func TestRecordSnapshot_HappyPath(t *testing.T) {
	store := &mockVerificationStore{
		createSnapshot: func(ctx context.Context, arg sqlc.CreateMetricSnapshotParams) (sqlc.MetricSnapshot, error) {
			if arg.SubmissionID != testCampaignID {
				t.Errorf("expected submission ID %x, got %x", testCampaignID.Bytes, arg.SubmissionID.Bytes)
			}
			if arg.Platform != "youtube" {
				t.Errorf("expected platform youtube, got %s", arg.Platform)
			}
			if arg.Views != 5000 {
				t.Errorf("expected views 5000, got %d", arg.Views)
			}
			if arg.Likes != 500 {
				t.Errorf("expected likes 500, got %d", arg.Likes)
			}
			return sqlc.MetricSnapshot{
				ID:           testSnapshotID(),
				SubmissionID: arg.SubmissionID,
				Platform:     arg.Platform,
				Views:        arg.Views,
				Likes:        arg.Likes,
				Comments:     arg.Comments,
				Shares:       arg.Shares,
				CapturedAt:   arg.CapturedAt,
				CreatedAt:    pgtype.Timestamptz{Valid: true, Time: time.Now()},
			}, nil
		},
	}
	svc := service.NewVerificationService(store)
	snap, err := svc.RecordSnapshot(context.Background(), testCampaignID, "youtube", 5000, 500, 100, 50, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snap.Views != 5000 {
		t.Errorf("expected views 5000, got %d", snap.Views)
	}
}

func TestRecordSnapshot_WithCapturedAt(t *testing.T) {
	captured := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)
	store := &mockVerificationStore{
		createSnapshot: func(ctx context.Context, arg sqlc.CreateMetricSnapshotParams) (sqlc.MetricSnapshot, error) {
			if !arg.CapturedAt.Valid {
				t.Error("expected captured_at to be valid")
			}
			if !arg.CapturedAt.Time.Equal(captured) {
				t.Errorf("expected captured_at %v, got %v", captured, arg.CapturedAt.Time)
			}
			return sqlc.MetricSnapshot{CapturedAt: arg.CapturedAt}, nil
		},
	}
	svc := service.NewVerificationService(store)
	_, err := svc.RecordSnapshot(context.Background(), testCampaignID, "youtube", 100, 10, 5, 2, &captured)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRecordSnapshot_StoreError(t *testing.T) {
	store := &mockVerificationStore{
		createSnapshot: func(ctx context.Context, arg sqlc.CreateMetricSnapshotParams) (sqlc.MetricSnapshot, error) {
			return sqlc.MetricSnapshot{}, fmt.Errorf("db error")
		},
	}
	svc := service.NewVerificationService(store)
	_, err := svc.RecordSnapshot(context.Background(), testCampaignID, "youtube", 100, 10, 5, 2, nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestComputeEligibleViews_NoSnapshots(t *testing.T) {
	store := &mockVerificationStore{
		getSubWithCampaign: func(ctx context.Context, id pgtype.UUID) (sqlc.GetSubmissionWithCampaignRow, error) {
			return testSubWithCampaign(1000), nil
		},
		getInitialSnapshot: func(ctx context.Context, id pgtype.UUID) (sqlc.MetricSnapshot, error) {
			return sqlc.MetricSnapshot{}, pgx.ErrNoRows
		},
	}
	svc := service.NewVerificationService(store)
	result, err := svc.ComputeEligibleViews(context.Background(), testCampaignID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.HasSnapshots {
		t.Error("expected has_snapshots to be false")
	}
	if result.EligibleViews != 0 {
		t.Errorf("expected eligible views 0, got %d", result.EligibleViews)
	}
}

func TestComputeEligibleViews_SubmissionNotFound(t *testing.T) {
	store := &mockVerificationStore{
		getSubWithCampaign: func(ctx context.Context, id pgtype.UUID) (sqlc.GetSubmissionWithCampaignRow, error) {
			return sqlc.GetSubmissionWithCampaignRow{}, pgx.ErrNoRows
		},
	}
	svc := service.NewVerificationService(store)
	_, err := svc.ComputeEligibleViews(context.Background(), testCampaignID)
	if err != service.ErrSubmissionNotFound {
		t.Errorf("expected ErrSubmissionNotFound, got %v", err)
	}
}

func TestComputeEligibleViews_BelowMinFloor(t *testing.T) {
	store := &mockVerificationStore{
		getSubWithCampaign: func(ctx context.Context, id pgtype.UUID) (sqlc.GetSubmissionWithCampaignRow, error) {
			return testSubWithCampaign(1000), nil // min_views_per_clip = 1000
		},
		getInitialSnapshot: func(ctx context.Context, id pgtype.UUID) (sqlc.MetricSnapshot, error) {
			return testSnap(100), nil // initial: 100
		},
		getLatestSnapshot: func(ctx context.Context, id pgtype.UUID) (sqlc.MetricSnapshot, error) {
			return testSnap(500), nil // latest: 500, growth = 400, below 1000
		},
	}
	svc := service.NewVerificationService(store)
	result, err := svc.ComputeEligibleViews(context.Background(), testCampaignID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.HasSnapshots {
		t.Error("expected has_snapshots to be true")
	}
	if result.GrowthViews != 400 {
		t.Errorf("expected growth views 400, got %d", result.GrowthViews)
	}
	if result.EligibleViews != 0 {
		t.Errorf("expected eligible views 0 (below floor), got %d", result.EligibleViews)
	}
}

func TestComputeEligibleViews_AboveMinFloor(t *testing.T) {
	store := &mockVerificationStore{
		getSubWithCampaign: func(ctx context.Context, id pgtype.UUID) (sqlc.GetSubmissionWithCampaignRow, error) {
			return testSubWithCampaign(1000), nil // min_views_per_clip = 1000
		},
		getInitialSnapshot: func(ctx context.Context, id pgtype.UUID) (sqlc.MetricSnapshot, error) {
			return testSnap(100), nil // initial: 100
		},
		getLatestSnapshot: func(ctx context.Context, id pgtype.UUID) (sqlc.MetricSnapshot, error) {
			return testSnap(5000), nil // latest: 5000, growth = 4900
		},
	}
	svc := service.NewVerificationService(store)
	result, err := svc.ComputeEligibleViews(context.Background(), testCampaignID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.GrowthViews != 4900 {
		t.Errorf("expected growth views 4900, got %d", result.GrowthViews)
	}
	if result.EligibleViews != 4900 {
		t.Errorf("expected eligible views 4900, got %d", result.EligibleViews)
	}
}

func TestComputeEligibleViews_ViewDecrease(t *testing.T) {
	store := &mockVerificationStore{
		getSubWithCampaign: func(ctx context.Context, id pgtype.UUID) (sqlc.GetSubmissionWithCampaignRow, error) {
			return testSubWithCampaign(1000), nil
		},
		getInitialSnapshot: func(ctx context.Context, id pgtype.UUID) (sqlc.MetricSnapshot, error) {
			return testSnap(5000), nil // initial: 5000
		},
		getLatestSnapshot: func(ctx context.Context, id pgtype.UUID) (sqlc.MetricSnapshot, error) {
			return testSnap(2000), nil // latest: 2000, growth = -3000, clamped to 0
		},
	}
	svc := service.NewVerificationService(store)
	result, err := svc.ComputeEligibleViews(context.Background(), testCampaignID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.GrowthViews != 0 {
		t.Errorf("expected growth views 0 (views decreased), got %d", result.GrowthViews)
	}
	if result.EligibleViews != 0 {
		t.Errorf("expected eligible views 0, got %d", result.EligibleViews)
	}
}

func TestComputeEligibleViews_DefaultMinFloor(t *testing.T) {
	store := &mockVerificationStore{
		getSubWithCampaign: func(ctx context.Context, id pgtype.UUID) (sqlc.GetSubmissionWithCampaignRow, error) {
			return sqlc.GetSubmissionWithCampaignRow{
				ID:              testCampaignID,
				Status:          "approved",
				MinViewsPerClip: pgtype.Int4{Valid: false}, // unset, default 1000
				CpmRate:         100,
			}, nil
		},
		getInitialSnapshot: func(ctx context.Context, id pgtype.UUID) (sqlc.MetricSnapshot, error) {
			return testSnap(0), nil
		},
		getLatestSnapshot: func(ctx context.Context, id pgtype.UUID) (sqlc.MetricSnapshot, error) {
			return testSnap(999), nil // 999 < 1000 default
		},
	}
	svc := service.NewVerificationService(store)
	result, err := svc.ComputeEligibleViews(context.Background(), testCampaignID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.EligibleViews != 0 {
		t.Errorf("expected eligible views 0 (below default 1000), got %d", result.EligibleViews)
	}
}

func TestGetVerificationStatus_NoSnapshots(t *testing.T) {
	store := &mockVerificationStore{
		getSubWithCampaign: func(ctx context.Context, id pgtype.UUID) (sqlc.GetSubmissionWithCampaignRow, error) {
			return testSubWithCampaign(1000), nil
		},
		countSnapshots: func(ctx context.Context, id pgtype.UUID) (int64, error) {
			return 0, nil
		},
	}
	svc := service.NewVerificationService(store)
	status, err := svc.GetVerificationStatus(context.Background(), testCampaignID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.HasSnapshots {
		t.Error("expected has_snapshots to be false")
	}
	if status.SnapshotCount != 0 {
		t.Errorf("expected snapshot count 0, got %d", status.SnapshotCount)
	}
	if status.Status != "approved" {
		t.Errorf("expected status approved, got %s", status.Status)
	}
}

func TestGetVerificationStatus_WithSnapshots(t *testing.T) {
	store := &mockVerificationStore{
		getSubWithCampaign: func(ctx context.Context, id pgtype.UUID) (sqlc.GetSubmissionWithCampaignRow, error) {
			return testSubWithCampaign(1000), nil
		},
		countSnapshots: func(ctx context.Context, id pgtype.UUID) (int64, error) {
			return 5, nil
		},
		getInitialSnapshot: func(ctx context.Context, id pgtype.UUID) (sqlc.MetricSnapshot, error) {
			return testSnap(100), nil
		},
		getLatestSnapshot: func(ctx context.Context, id pgtype.UUID) (sqlc.MetricSnapshot, error) {
			return testSnap(5000), nil
		},
	}
	svc := service.NewVerificationService(store)
	status, err := svc.GetVerificationStatus(context.Background(), testCampaignID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !status.HasSnapshots {
		t.Error("expected has_snapshots to be true")
	}
	if status.SnapshotCount != 5 {
		t.Errorf("expected snapshot count 5, got %d", status.SnapshotCount)
	}
	if status.InitialViews != 100 {
		t.Errorf("expected initial views 100, got %d", status.InitialViews)
	}
	if status.LatestViews != 5000 {
		t.Errorf("expected latest views 5000, got %d", status.LatestViews)
	}
	if status.GrowthViews != 4900 {
		t.Errorf("expected growth views 4900, got %d", status.GrowthViews)
	}
	if status.EligibleViews != 4900 {
		t.Errorf("expected eligible views 4900, got %d", status.EligibleViews)
	}
}

func TestGetVerificationStatus_EngagementRate(t *testing.T) {
	store := &mockVerificationStore{
		getSubWithCampaign: func(ctx context.Context, id pgtype.UUID) (sqlc.GetSubmissionWithCampaignRow, error) {
			return testSubWithCampaign(100), nil // low min so eligible
		},
		countSnapshots: func(ctx context.Context, id pgtype.UUID) (int64, error) {
			return 3, nil
		},
		getInitialSnapshot: func(ctx context.Context, id pgtype.UUID) (sqlc.MetricSnapshot, error) {
			return testSnap(1000), nil
		},
		getLatestSnapshot: func(ctx context.Context, id pgtype.UUID) (sqlc.MetricSnapshot, error) {
			snap := testSnap(5000)
			snap.Likes = 500
			snap.Comments = 200
			snap.Shares = 100
			return snap, nil
		},
	}
	svc := service.NewVerificationService(store)
	status, err := svc.GetVerificationStatus(context.Background(), testCampaignID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// engagement = (500 + 200 + 100) / 5000 = 0.16
	expected := float64(800) / float64(5000)
	if status.EngagementRate != expected {
		t.Errorf("expected engagement rate %f, got %f", expected, status.EngagementRate)
	}
}

func TestGetVerificationStatus_SubmissionNotFound(t *testing.T) {
	store := &mockVerificationStore{
		getSubWithCampaign: func(ctx context.Context, id pgtype.UUID) (sqlc.GetSubmissionWithCampaignRow, error) {
			return sqlc.GetSubmissionWithCampaignRow{}, pgx.ErrNoRows
		},
	}
	svc := service.NewVerificationService(store)
	_, err := svc.GetVerificationStatus(context.Background(), testCampaignID)
	if err != service.ErrSubmissionNotFound {
		t.Errorf("expected ErrSubmissionNotFound, got %v", err)
	}
}

func TestGetSubmissionsNeedingVerification_DefaultLimit(t *testing.T) {
	store := &mockVerificationStore{
		getNeedingVerification: func(ctx context.Context, limit int32) ([]sqlc.GetSubmissionsNeedingVerificationRow, error) {
			if limit != 20 {
				t.Errorf("expected default limit 20, got %d", limit)
			}
			return nil, nil
		},
	}
	svc := service.NewVerificationService(store)
	_, err := svc.GetSubmissionsNeedingVerification(context.Background(), 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetSubmissionsNeedingVerification_CustomLimit(t *testing.T) {
	store := &mockVerificationStore{
		getNeedingVerification: func(ctx context.Context, limit int32) ([]sqlc.GetSubmissionsNeedingVerificationRow, error) {
			if limit != 50 {
				t.Errorf("expected limit 50, got %d", limit)
			}
			return []sqlc.GetSubmissionsNeedingVerificationRow{
				{
					ID:         testCampaignID,
					CampaignID: testCampaignID,
					ClipperID:  "clipper1",
					PostUrl:    "https://youtube.com/watch?v=1",
					Platform:   "youtube",
					Status:     "approved",
					CpmRate:    150,
				},
			}, nil
		},
	}
	svc := service.NewVerificationService(store)
	rows, err := svc.GetSubmissionsNeedingVerification(context.Background(), 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
}
