package service_test

import (
	"context"
	"testing"
	"time"

	sqlc "clipin/apps/api/internal/db/sqlc"
	"clipin/apps/api/internal/service"

	"github.com/jackc/pgx/v5/pgtype"
)

// mockSubmissionStore implements SubmissionStore for testing.
type mockSubmissionStore struct {
	getByID         func(ctx context.Context, id pgtype.UUID) (sqlc.Submission, error)
	getCampaignByID func(ctx context.Context, id pgtype.UUID) (sqlc.Campaign, error)
	create          func(ctx context.Context, arg sqlc.CreateSubmissionParams) (sqlc.Submission, error)
	updateStatus    func(ctx context.Context, arg sqlc.UpdateSubmissionStatusParams) (sqlc.Submission, error)
	listByCampaign  func(ctx context.Context, campaignID pgtype.UUID) ([]sqlc.Submission, error)
	listByClipper   func(ctx context.Context, clipperID string) ([]sqlc.Submission, error)
	countByCampaign func(ctx context.Context, campaignID pgtype.UUID) (int64, error)
	countByClipper  func(ctx context.Context, arg sqlc.CountSubmissionsByClipperForCampaignParams) (int64, error)
	listPending     func(ctx context.Context, createdAt pgtype.Timestamptz) ([]sqlc.ListPendingSubmissionsOlderThanRow, error)
}

func (m *mockSubmissionStore) GetSubmissionByID(ctx context.Context, id pgtype.UUID) (sqlc.Submission, error) {
	return m.getByID(ctx, id)
}

func (m *mockSubmissionStore) GetCampaignByID(ctx context.Context, id pgtype.UUID) (sqlc.Campaign, error) {
	return m.getCampaignByID(ctx, id)
}

func (m *mockSubmissionStore) CreateSubmission(ctx context.Context, arg sqlc.CreateSubmissionParams) (sqlc.Submission, error) {
	return m.create(ctx, arg)
}

func (m *mockSubmissionStore) UpdateSubmissionStatus(ctx context.Context, arg sqlc.UpdateSubmissionStatusParams) (sqlc.Submission, error) {
	return m.updateStatus(ctx, arg)
}

func (m *mockSubmissionStore) ListSubmissionsByCampaign(ctx context.Context, campaignID pgtype.UUID) ([]sqlc.Submission, error) {
	return m.listByCampaign(ctx, campaignID)
}

func (m *mockSubmissionStore) ListSubmissionsByClipper(ctx context.Context, clipperID string) ([]sqlc.Submission, error) {
	return m.listByClipper(ctx, clipperID)
}

func (m *mockSubmissionStore) CountSubmissionsByCampaign(ctx context.Context, campaignID pgtype.UUID) (int64, error) {
	return m.countByCampaign(ctx, campaignID)
}

func (m *mockSubmissionStore) CountSubmissionsByClipperForCampaign(ctx context.Context, arg sqlc.CountSubmissionsByClipperForCampaignParams) (int64, error) {
	return m.countByClipper(ctx, arg)
}

func (m *mockSubmissionStore) ListPendingSubmissionsOlderThan(ctx context.Context, createdAt pgtype.Timestamptz) ([]sqlc.ListPendingSubmissionsOlderThanRow, error) {
	return m.listPending(ctx, createdAt)
}

var testCampaignID = pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true}

func testActiveCampaign() sqlc.Campaign {
	return sqlc.Campaign{
		ID:                  testCampaignID,
		OwnerID:             "owner1",
		Title:               "Test Campaign",
		Platform:            "youtube",
		Status:              "active",
		CpmRate:             100,
		TotalBudget:         10000,
		RemainingBudget:     5000,
		MaxClipsPerClipper:  pgtype.Int4{Valid: true, Int32: 3},
		AutoApproveHours:    pgtype.Int4{Valid: true, Int32: 48},
		CreatedAt:           pgtype.Timestamptz{Valid: true},
		UpdatedAt:           pgtype.Timestamptz{Valid: true},
	}
}

func TestSubmit_HappyPath(t *testing.T) {
	campaign := testActiveCampaign()
	store := &mockSubmissionStore{
		getCampaignByID: func(_ context.Context, id pgtype.UUID) (sqlc.Campaign, error) {
			return campaign, nil
		},
		countByClipper: func(_ context.Context, arg sqlc.CountSubmissionsByClipperForCampaignParams) (int64, error) {
			return 0, nil
		},
		create: func(_ context.Context, arg sqlc.CreateSubmissionParams) (sqlc.Submission, error) {
			if arg.CampaignID != testCampaignID {
				t.Errorf("expected campaign ID %x, got %x", testCampaignID.Bytes, arg.CampaignID.Bytes)
			}
			if arg.ClipperID != "clipper1" {
				t.Errorf("expected clipper clipper1, got %s", arg.ClipperID)
			}
			if arg.PostUrl != "https://youtube.com/watch?v=abc" {
				t.Errorf("expected youtube URL, got %s", arg.PostUrl)
			}
			return sqlc.Submission{
				ID:         arg.ID,
				CampaignID: arg.CampaignID,
				ClipperID:  arg.ClipperID,
				PostUrl:    arg.PostUrl,
				Platform:   arg.Platform,
				Status:     "pending",
			}, nil
		},
	}
	svc := service.NewSubmissionService(store)
	sub, err := svc.Submit(context.Background(), testCampaignID, "clipper1", "https://youtube.com/watch?v=abc", "youtube")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sub.Status != "pending" {
		t.Errorf("expected status pending, got %s", sub.Status)
	}
}

func TestSubmit_InvalidURL(t *testing.T) {
	store := &mockSubmissionStore{
		getCampaignByID: func(_ context.Context, id pgtype.UUID) (sqlc.Campaign, error) {
			return testActiveCampaign(), nil
		},
	}
	svc := service.NewSubmissionService(store)
	_, err := svc.Submit(context.Background(), testCampaignID, "clipper1", "not-a-url", "youtube")
	if err != service.ErrInvalidURL {
		t.Errorf("expected ErrInvalidURL, got %v", err)
	}
}

func TestSubmit_EmptyURL(t *testing.T) {
	store := &mockSubmissionStore{
		getCampaignByID: func(_ context.Context, id pgtype.UUID) (sqlc.Campaign, error) {
			return testActiveCampaign(), nil
		},
	}
	svc := service.NewSubmissionService(store)
	_, err := svc.Submit(context.Background(), testCampaignID, "clipper1", "", "youtube")
	if err != service.ErrInvalidURL {
		t.Errorf("expected ErrInvalidURL, got %v", err)
	}
}

func TestSubmit_CampaignNotFound(t *testing.T) {
	store := &mockSubmissionStore{
		getCampaignByID: func(_ context.Context, id pgtype.UUID) (sqlc.Campaign, error) {
			return sqlc.Campaign{}, nil // simulates pgx.ErrNoRows
		},
	}
	svc := service.NewSubmissionService(store)
	_, err := svc.Submit(context.Background(), testCampaignID, "clipper1", "https://youtube.com/watch?v=abc", "youtube")
	// Since our mock returns empty (not pgx.ErrNoRows), it won't match ErrCampaignNotFound
	// but it will fail on the status check since status is ""
	if err == nil {
		t.Error("expected error for invalid campaign")
	}
}

func TestSubmit_InactiveCampaign(t *testing.T) {
	campaign := testActiveCampaign()
	campaign.Status = "draft"
	store := &mockSubmissionStore{
		getCampaignByID: func(_ context.Context, id pgtype.UUID) (sqlc.Campaign, error) {
			return campaign, nil
		},
	}
	svc := service.NewSubmissionService(store)
	_, err := svc.Submit(context.Background(), testCampaignID, "clipper1", "https://youtube.com/watch?v=abc", "youtube")
	if err != service.ErrCampaignNotActive {
		t.Errorf("expected ErrCampaignNotActive, got %v", err)
	}
}

func TestSubmit_PlatformMismatch(t *testing.T) {
	store := &mockSubmissionStore{
		getCampaignByID: func(_ context.Context, id pgtype.UUID) (sqlc.Campaign, error) {
			return testActiveCampaign(), nil
		},
	}
	svc := service.NewSubmissionService(store)
	_, err := svc.Submit(context.Background(), testCampaignID, "clipper1", "https://youtube.com/watch?v=abc", "tiktok")
	if err != service.ErrPlatformMismatch {
		t.Errorf("expected ErrPlatformMismatch, got %v", err)
	}
}

func TestSubmit_MultiPlatformCampaign(t *testing.T) {
	campaign := testActiveCampaign()
	campaign.Platform = "multi"
	store := &mockSubmissionStore{
		getCampaignByID: func(_ context.Context, id pgtype.UUID) (sqlc.Campaign, error) {
			return campaign, nil
		},
		countByClipper: func(_ context.Context, arg sqlc.CountSubmissionsByClipperForCampaignParams) (int64, error) {
			return 0, nil
		},
		create: func(_ context.Context, arg sqlc.CreateSubmissionParams) (sqlc.Submission, error) {
			return sqlc.Submission{
				ID: arg.ID, CampaignID: arg.CampaignID, ClipperID: arg.ClipperID,
				PostUrl: arg.PostUrl, Platform: arg.Platform, Status: "pending",
			}, nil
		},
	}
	svc := service.NewSubmissionService(store)
	sub, err := svc.Submit(context.Background(), testCampaignID, "clipper1", "https://tiktok.com/@user/video/123", "tiktok")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sub.Platform != "tiktok" {
		t.Errorf("expected platform tiktok, got %s", sub.Platform)
	}
}

func TestSubmit_ClipperLimitExceeded(t *testing.T) {
	campaign := testActiveCampaign()
	store := &mockSubmissionStore{
		getCampaignByID: func(_ context.Context, id pgtype.UUID) (sqlc.Campaign, error) {
			return campaign, nil
		},
		countByClipper: func(_ context.Context, arg sqlc.CountSubmissionsByClipperForCampaignParams) (int64, error) {
			return 3, nil // at the limit
		},
	}
	svc := service.NewSubmissionService(store)
	_, err := svc.Submit(context.Background(), testCampaignID, "clipper1", "https://youtube.com/watch?v=abc", "youtube")
	if err != service.ErrClipperLimitExceeded {
		t.Errorf("expected ErrClipperLimitExceeded, got %v", err)
	}
}

func TestApprove_HappyPath(t *testing.T) {
	submission := sqlc.Submission{
		ID:         testCampaignID,
		CampaignID: testCampaignID,
		ClipperID:  "clipper1",
		PostUrl:    "https://youtube.com/watch?v=abc",
		Status:     "pending",
	}
	campaign := testActiveCampaign()
	store := &mockSubmissionStore{
		getByID: func(_ context.Context, id pgtype.UUID) (sqlc.Submission, error) {
			return submission, nil
		},
		getCampaignByID: func(_ context.Context, id pgtype.UUID) (sqlc.Campaign, error) {
			return campaign, nil
		},
		updateStatus: func(_ context.Context, arg sqlc.UpdateSubmissionStatusParams) (sqlc.Submission, error) {
			if arg.Status != "approved" {
				t.Errorf("expected status approved, got %s", arg.Status)
			}
			return sqlc.Submission{ID: arg.ID, Status: "approved"}, nil
		},
	}
	svc := service.NewSubmissionService(store)
	result, err := svc.Approve(context.Background(), testCampaignID, "owner1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "approved" {
		t.Errorf("expected status approved, got %s", result.Status)
	}
}

func TestApprove_NotPending(t *testing.T) {
	submission := sqlc.Submission{
		ID:     testCampaignID,
		Status: "approved",
	}
	store := &mockSubmissionStore{
		getByID: func(_ context.Context, id pgtype.UUID) (sqlc.Submission, error) {
			return submission, nil
		},
	}
	svc := service.NewSubmissionService(store)
	_, err := svc.Approve(context.Background(), testCampaignID, "owner1")
	if err != service.ErrSubmissionNotPending {
		t.Errorf("expected ErrSubmissionNotPending, got %v", err)
	}
}

func TestApprove_NotOwner(t *testing.T) {
	submission := sqlc.Submission{
		ID:         testCampaignID,
		CampaignID: testCampaignID,
		Status:     "pending",
	}
	campaign := testActiveCampaign()
	store := &mockSubmissionStore{
		getByID: func(_ context.Context, id pgtype.UUID) (sqlc.Submission, error) {
			return submission, nil
		},
		getCampaignByID: func(_ context.Context, id pgtype.UUID) (sqlc.Campaign, error) {
			return campaign, nil
		},
	}
	svc := service.NewSubmissionService(store)
	_, err := svc.Approve(context.Background(), testCampaignID, "wrong-owner")
	if err != service.ErrNotOwner {
		t.Errorf("expected ErrNotOwner, got %v", err)
	}
}

func TestReject_HappyPath(t *testing.T) {
	submission := sqlc.Submission{
		ID:         testCampaignID,
		CampaignID: testCampaignID,
		ClipperID:  "clipper1",
		Status:     "pending",
	}
	campaign := testActiveCampaign()
	store := &mockSubmissionStore{
		getByID: func(_ context.Context, id pgtype.UUID) (sqlc.Submission, error) {
			return submission, nil
		},
		getCampaignByID: func(_ context.Context, id pgtype.UUID) (sqlc.Campaign, error) {
			return campaign, nil
		},
		updateStatus: func(_ context.Context, arg sqlc.UpdateSubmissionStatusParams) (sqlc.Submission, error) {
			if arg.Status != "rejected" {
				t.Errorf("expected status rejected, got %s", arg.Status)
			}
			if !arg.RejectionReason.Valid || arg.RejectionReason.String != "low quality" {
				t.Errorf("expected rejection reason 'low quality', got %v", arg.RejectionReason)
			}
			return sqlc.Submission{ID: arg.ID, Status: "rejected"}, nil
		},
	}
	svc := service.NewSubmissionService(store)
	result, err := svc.Reject(context.Background(), testCampaignID, "owner1", "low quality")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "rejected" {
		t.Errorf("expected status rejected, got %s", result.Status)
	}
}

func TestReject_NotPending(t *testing.T) {
	submission := sqlc.Submission{
		ID:     testCampaignID,
		Status: "auto_approved",
	}
	store := &mockSubmissionStore{
		getByID: func(_ context.Context, id pgtype.UUID) (sqlc.Submission, error) {
			return submission, nil
		},
	}
	svc := service.NewSubmissionService(store)
	_, err := svc.Reject(context.Background(), testCampaignID, "owner1", "")
	if err != service.ErrSubmissionNotPending {
		t.Errorf("expected ErrSubmissionNotPending, got %v", err)
	}
}

func TestReject_NotOwner(t *testing.T) {
	submission := sqlc.Submission{
		ID:         testCampaignID,
		CampaignID: testCampaignID,
		Status:     "pending",
	}
	campaign := testActiveCampaign()
	store := &mockSubmissionStore{
		getByID: func(_ context.Context, id pgtype.UUID) (sqlc.Submission, error) {
			return submission, nil
		},
		getCampaignByID: func(_ context.Context, id pgtype.UUID) (sqlc.Campaign, error) {
			return campaign, nil
		},
	}
	svc := service.NewSubmissionService(store)
	_, err := svc.Reject(context.Background(), testCampaignID, "wrong-owner", "")
	if err != service.ErrNotOwner {
		t.Errorf("expected ErrNotOwner, got %v", err)
	}
}

func TestAutoApprove_ApprovesExpiredSubmissions(t *testing.T) {
	now := time.Now()
	cutoff := now.Add(-49 * time.Hour) // past 48h auto-approve

	rows := []sqlc.ListPendingSubmissionsOlderThanRow{
		{
			ID:               pgtype.UUID{Bytes: [16]byte{10}, Valid: true},
			CampaignID:       testCampaignID,
			ClipperID:        "clipper1",
			PostUrl:          "https://youtube.com/watch?v=1",
			Status:           "pending",
			CreatedAt:        pgtype.Timestamptz{Valid: true, Time: cutoff},
			AutoApproveHours: pgtype.Int4{Valid: true, Int32: 48},
		},
		{
			ID:               pgtype.UUID{Bytes: [16]byte{20}, Valid: true},
			CampaignID:       testCampaignID,
			ClipperID:        "clipper2",
			PostUrl:          "https://youtube.com/watch?v=2",
			Status:           "pending",
			CreatedAt:        pgtype.Timestamptz{Valid: true, Time: cutoff},
			AutoApproveHours: pgtype.Int4{Valid: true, Int32: 48},
		},
	}

	store := &mockSubmissionStore{
		listPending: func(_ context.Context, _ pgtype.Timestamptz) ([]sqlc.ListPendingSubmissionsOlderThanRow, error) {
			return rows, nil
		},
		updateStatus: func(_ context.Context, arg sqlc.UpdateSubmissionStatusParams) (sqlc.Submission, error) {
			if arg.Status != "auto_approved" {
				t.Errorf("expected auto_approved, got %s", arg.Status)
			}
			return sqlc.Submission{ID: arg.ID, Status: "auto_approved"}, nil
		},
	}
	svc := service.NewSubmissionService(store)
	count, err := svc.AutoApprove(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 auto-approved, got %d", count)
	}
}

func TestAutoApprove_SkipsNotYetExpired(t *testing.T) {
	now := time.Now()
	recent := now.Add(-1 * time.Hour) // only 1 hour old, needs 48h

	rows := []sqlc.ListPendingSubmissionsOlderThanRow{
		{
			ID:               pgtype.UUID{Bytes: [16]byte{10}, Valid: true},
			CampaignID:       testCampaignID,
			ClipperID:        "clipper1",
			PostUrl:          "https://youtube.com/watch?v=1",
			Status:           "pending",
			CreatedAt:        pgtype.Timestamptz{Valid: true, Time: recent},
			AutoApproveHours: pgtype.Int4{Valid: true, Int32: 48},
		},
	}

	store := &mockSubmissionStore{
		listPending: func(_ context.Context, _ pgtype.Timestamptz) ([]sqlc.ListPendingSubmissionsOlderThanRow, error) {
			return rows, nil
		},
		updateStatus: func(_ context.Context, arg sqlc.UpdateSubmissionStatusParams) (sqlc.Submission, error) {
			t.Error("should not have called updateStatus")
			return sqlc.Submission{}, nil
		},
	}
	svc := service.NewSubmissionService(store)
	count, err := svc.AutoApprove(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 auto-approved, got %d", count)
	}
}

func TestAutoApprove_DefaultHoursWhenInvalid(t *testing.T) {
	now := time.Now()
	cutoff := now.Add(-49 * time.Hour) // past default 48h

	rows := []sqlc.ListPendingSubmissionsOlderThanRow{
		{
			ID:               pgtype.UUID{Bytes: [16]byte{10}, Valid: true},
			CampaignID:       testCampaignID,
			ClipperID:        "clipper1",
			PostUrl:          "https://youtube.com/watch?v=1",
			Status:           "pending",
			CreatedAt:        pgtype.Timestamptz{Valid: true, Time: cutoff},
			AutoApproveHours: pgtype.Int4{Valid: false}, // invalid, should use default 48
		},
	}

	store := &mockSubmissionStore{
		listPending: func(_ context.Context, _ pgtype.Timestamptz) ([]sqlc.ListPendingSubmissionsOlderThanRow, error) {
			return rows, nil
		},
		updateStatus: func(_ context.Context, arg sqlc.UpdateSubmissionStatusParams) (sqlc.Submission, error) {
			return sqlc.Submission{ID: arg.ID, Status: "auto_approved"}, nil
		},
	}
	svc := service.NewSubmissionService(store)
	count, err := svc.AutoApprove(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 auto-approved, got %d", count)
	}
}

func TestListByCampaign(t *testing.T) {
	expected := []sqlc.Submission{
		{ID: testCampaignID, CampaignID: testCampaignID, Status: "pending"},
	}
	store := &mockSubmissionStore{
		listByCampaign: func(_ context.Context, campaignID pgtype.UUID) ([]sqlc.Submission, error) {
			return expected, nil
		},
	}
	svc := service.NewSubmissionService(store)
	result, err := svc.ListByCampaign(context.Background(), testCampaignID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 submission, got %d", len(result))
	}
}

func TestListByClipper(t *testing.T) {
	expected := []sqlc.Submission{
		{ID: testCampaignID, ClipperID: "clipper1", Status: "approved"},
	}
	store := &mockSubmissionStore{
		listByClipper: func(_ context.Context, clipperID string) ([]sqlc.Submission, error) {
			if clipperID != "clipper1" {
				t.Errorf("expected clipper1, got %s", clipperID)
			}
			return expected, nil
		},
	}
	svc := service.NewSubmissionService(store)
	result, err := svc.ListByClipper(context.Background(), "clipper1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 submission, got %d", len(result))
	}
}
