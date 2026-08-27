package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	sqlc "clipin/apps/api/internal/db/sqlc"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// VerificationStore is the persistence interface the verification service requires.
type VerificationStore interface {
	GetSubmissionByID(ctx context.Context, id pgtype.UUID) (sqlc.Submission, error)
	GetCampaignByID(ctx context.Context, id pgtype.UUID) (sqlc.Campaign, error)
	CreateMetricSnapshot(ctx context.Context, arg sqlc.CreateMetricSnapshotParams) (sqlc.MetricSnapshot, error)
	GetLatestSnapshotForSubmission(ctx context.Context, submissionID pgtype.UUID) (sqlc.MetricSnapshot, error)
	GetInitialSnapshotForSubmission(ctx context.Context, submissionID pgtype.UUID) (sqlc.MetricSnapshot, error)
	ListSnapshotsBySubmission(ctx context.Context, submissionID pgtype.UUID) ([]sqlc.MetricSnapshot, error)
	CountSnapshotsBySubmission(ctx context.Context, submissionID pgtype.UUID) (int64, error)
	GetSubmissionWithCampaign(ctx context.Context, id pgtype.UUID) (sqlc.GetSubmissionWithCampaignRow, error)
	GetSubmissionsNeedingVerification(ctx context.Context, limit int32) ([]sqlc.GetSubmissionsNeedingVerificationRow, error)
}

// VerificationService implements snapshot ingestion and delta computation.
type VerificationService struct {
	store VerificationStore
}

// NewVerificationService creates a new VerificationService.
func NewVerificationService(store VerificationStore) *VerificationService {
	return &VerificationService{store: store}
}

// RecordSnapshot stores a new metric snapshot for a submission.
func (s *VerificationService) RecordSnapshot(ctx context.Context, submissionID pgtype.UUID, platform string, views, likes, comments, shares int64, capturedAt *time.Time) (*sqlc.MetricSnapshot, error) {
	ts := pgtype.Timestamptz{Valid: true, Time: time.Now()}
	if capturedAt != nil && !capturedAt.IsZero() {
		ts = pgtype.Timestamptz{Valid: true, Time: *capturedAt}
	}

	snap, err := s.store.CreateMetricSnapshot(ctx, sqlc.CreateMetricSnapshotParams{
		SubmissionID: submissionID,
		Platform:     platform,
		Views:        views,
		Likes:        likes,
		Comments:     comments,
		Shares:       shares,
		CapturedAt:   ts,
	})
	if err != nil {
		return nil, fmt.Errorf("create metric snapshot: %w", err)
	}
	return &snap, nil
}

// EligibleViewsResult contains the computed eligible views for a submission.
type EligibleViewsResult struct {
	SubmissionID  pgtype.UUID `json:"submission_id"`
	InitialViews  int64       `json:"initial_views"`
	LatestViews   int64       `json:"latest_views"`
	GrowthViews   int64       `json:"growth_views"`
	EligibleViews int64       `json:"eligible_views"`
	HasSnapshots  bool        `json:"has_snapshots"`
}

// ComputeEligibleViews computes eligible views for a submission:
// eligible = max(0, latest.views - initial.views), floored at min_views_per_clip.
func (s *VerificationService) ComputeEligibleViews(ctx context.Context, submissionID pgtype.UUID) (*EligibleViewsResult, error) {
	row, err := s.store.GetSubmissionWithCampaign(ctx, submissionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSubmissionNotFound
		}
		return nil, fmt.Errorf("get submission with campaign: %w", err)
	}

	result := &EligibleViewsResult{
		SubmissionID: submissionID,
		HasSnapshots: false,
	}

	// Get initial snapshot.
	initial, err := s.store.GetInitialSnapshotForSubmission(ctx, submissionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return result, nil
		}
		return nil, fmt.Errorf("get initial snapshot: %w", err)
	}

	// Get latest snapshot.
	latest, err := s.store.GetLatestSnapshotForSubmission(ctx, submissionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return result, nil
		}
		return nil, fmt.Errorf("get latest snapshot: %w", err)
	}

	result.HasSnapshots = true
	result.InitialViews = initial.Views
	result.LatestViews = latest.Views
	result.GrowthViews = int64(math.Max(0, float64(latest.Views-initial.Views)))

	// Apply min_views_per_clip floor.
	minViews := int64(1000) // default
	if row.MinViewsPerClip.Valid {
		minViews = int64(row.MinViewsPerClip.Int32)
	}
	if result.GrowthViews >= minViews {
		result.EligibleViews = result.GrowthViews
	}

	return result, nil
}

// VerificationStatus is the full verification state for a submission.
type VerificationStatus struct {
	SubmissionID   pgtype.UUID `json:"submission_id"`
	Status         string      `json:"status"`
	HasSnapshots   bool        `json:"has_snapshots"`
	SnapshotCount  int64       `json:"snapshot_count"`
	InitialViews   int64       `json:"initial_views"`
	LatestViews    int64       `json:"latest_views"`
	GrowthViews    int64       `json:"growth_views"`
	EligibleViews  int64       `json:"eligible_views"`
	EngagementRate float64     `json:"engagement_rate"`
	LastCapturedAt *string     `json:"last_captured_at,omitempty"`
}

// GetVerificationStatus returns the current verification state for a submission.
func (s *VerificationService) GetVerificationStatus(ctx context.Context, submissionID pgtype.UUID) (*VerificationStatus, error) {
	row, err := s.store.GetSubmissionWithCampaign(ctx, submissionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSubmissionNotFound
		}
		return nil, fmt.Errorf("get submission with campaign: %w", err)
	}

	count, err := s.store.CountSnapshotsBySubmission(ctx, submissionID)
	if err != nil {
		return nil, fmt.Errorf("count snapshots: %w", err)
	}

	status := &VerificationStatus{
		SubmissionID:  submissionID,
		Status:        row.Status,
		HasSnapshots:  count > 0,
		SnapshotCount: count,
	}

	if count == 0 {
		return status, nil
	}

	// Get initial and latest for delta computation.
	initial, err := s.store.GetInitialSnapshotForSubmission(ctx, submissionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return status, nil
		}
		return nil, fmt.Errorf("get initial snapshot: %w", err)
	}

	latest, err := s.store.GetLatestSnapshotForSubmission(ctx, submissionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return status, nil
		}
		return nil, fmt.Errorf("get latest snapshot: %w", err)
	}

	status.InitialViews = initial.Views
	status.LatestViews = latest.Views
	status.GrowthViews = int64(math.Max(0, float64(latest.Views-initial.Views)))

	// Compute eligible views with min floor.
	minViews := int64(1000)
	if row.MinViewsPerClip.Valid {
		minViews = int64(row.MinViewsPerClip.Int32)
	}
	if status.GrowthViews >= minViews {
		status.EligibleViews = status.GrowthViews
	}

	// Engagement rate from latest snapshot: (likes + comments + shares) / views.
	if latest.Views > 0 {
		status.EngagementRate = float64(latest.Likes+latest.Comments+latest.Shares) / float64(latest.Views)
	}

	if latest.CapturedAt.Valid {
		t := latest.CapturedAt.Time.Format("2006-01-02T15:04:05Z07:00")
		status.LastCapturedAt = &t
	}

	return status, nil
}

// GetSubmissionsNeedingVerification returns approved submissions that have no
// recent snapshots, for the verifier to poll.
func (s *VerificationService) GetSubmissionsNeedingVerification(ctx context.Context, limit int) ([]sqlc.GetSubmissionsNeedingVerificationRow, error) {
	if limit < 1 || limit > 100 {
		limit = 20
	}

	return s.store.GetSubmissionsNeedingVerification(ctx, int32(limit))
}
