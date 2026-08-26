package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/url"
	"strings"
	"time"

	sqlc "clipin/apps/api/internal/db/sqlc"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// SubmissionStore is the persistence interface the submission service requires.
type SubmissionStore interface {
	GetCampaignByID(ctx context.Context, id pgtype.UUID) (sqlc.Campaign, error)
	GetSubmissionByID(ctx context.Context, id pgtype.UUID) (sqlc.Submission, error)
	CreateSubmission(ctx context.Context, arg sqlc.CreateSubmissionParams) (sqlc.Submission, error)
	UpdateSubmissionStatus(ctx context.Context, arg sqlc.UpdateSubmissionStatusParams) (sqlc.Submission, error)
	ListSubmissionsByCampaign(ctx context.Context, campaignID pgtype.UUID) ([]sqlc.Submission, error)
	ListSubmissionsByClipper(ctx context.Context, clipperID string) ([]sqlc.Submission, error)
	CountSubmissionsByCampaign(ctx context.Context, campaignID pgtype.UUID) (int64, error)
	CountSubmissionsByClipperForCampaign(ctx context.Context, arg sqlc.CountSubmissionsByClipperForCampaignParams) (int64, error)
	ListPendingSubmissionsOlderThan(ctx context.Context, createdAt pgtype.Timestamptz) ([]sqlc.ListPendingSubmissionsOlderThanRow, error)
}

// SubmissionService implements submission lifecycle business logic.
type SubmissionService struct {
	store  SubmissionStore
	ledger *LedgerService
}

// NewSubmissionService creates a new SubmissionService.
func NewSubmissionService(store SubmissionStore) *SubmissionService {
	return &SubmissionService{store: store}
}

// WithLedger attaches a ledger service for recording financial entries.
func (s *SubmissionService) WithLedger(ledger *LedgerService) {
	s.ledger = ledger
}

// Sentinel errors for submission operations.
var (
	ErrSubmissionNotFound   = fmt.Errorf("submission not found")
	ErrCampaignNotActive    = fmt.Errorf("campaign is not active")
	ErrDuplicateSubmission  = fmt.Errorf("you have already submitted this URL for this campaign")
	ErrClipperLimitExceeded = fmt.Errorf("you have reached the submission limit for this campaign")
	ErrPlatformMismatch     = fmt.Errorf("submission platform does not match campaign platform")
	ErrInvalidURL           = fmt.Errorf("invalid post URL")
	ErrSubmissionNotPending = fmt.Errorf("submission is not in pending status")
)

// Submit creates a new submission with validation.
func (s *SubmissionService) Submit(ctx context.Context, campaignID pgtype.UUID, clipperID, postURL, platform string) (*sqlc.Submission, error) {
	// Validate URL format.
	if err := validatePostURL(postURL); err != nil {
		return nil, err
	}

	// Validate platform.
	platform = strings.ToLower(strings.TrimSpace(platform))
	if platform == "" {
		return nil, &ValidationError{Errors: []string{"platform is required"}}
	}

	// Fetch campaign.
	campaign, err := s.store.GetCampaignByID(ctx, campaignID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrCampaignNotFound
		}
		return nil, fmt.Errorf("get campaign: %w", err)
	}

	// Campaign must be active or funded.
	if campaign.Status != "active" && campaign.Status != "funded" {
		return nil, ErrCampaignNotActive
	}

	// Platform must match campaign (unless campaign is multi).
	if campaign.Platform != "multi" && campaign.Platform != platform {
		return nil, ErrPlatformMismatch
	}

	// Check clipper submission limit.
	if campaign.MaxClipsPerClipper.Valid {
		count, err := s.store.CountSubmissionsByClipperForCampaign(ctx, sqlc.CountSubmissionsByClipperForCampaignParams{
			CampaignID: campaignID,
			ClipperID:  clipperID,
		})
		if err != nil {
			return nil, fmt.Errorf("count clipper submissions: %w", err)
		}
		if count >= int64(campaign.MaxClipsPerClipper.Int32) {
			return nil, ErrClipperLimitExceeded
		}
	}

	// Generate UUID.
	var id pgtype.UUID
	if _, err := rand.Read(id.Bytes[:]); err != nil {
		return nil, fmt.Errorf("generate uuid: %w", err)
	}
	id.Valid = true

	// Create submission. The unique index on (campaign_id, clipper_id, post_url)
	// catches duplicates at DB level.
	submission, err := s.store.CreateSubmission(ctx, sqlc.CreateSubmissionParams{
		ID:         id,
		CampaignID: campaignID,
		ClipperID:  clipperID,
		PostUrl:    postURL,
		Platform:   platform,
	})
	if err != nil {
		// Check for unique violation (pgcode 23505).
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return nil, ErrDuplicateSubmission
		}
		return nil, fmt.Errorf("create submission: %w", err)
	}

	return &submission, nil
}

// ListByCampaign returns all submissions for a campaign.
func (s *SubmissionService) ListByCampaign(ctx context.Context, campaignID pgtype.UUID) ([]sqlc.Submission, error) {
	return s.store.ListSubmissionsByCampaign(ctx, campaignID)
}

// VerifyCampaignOwnership checks whether the given user owns the campaign.
func (s *SubmissionService) VerifyCampaignOwnership(ctx context.Context, campaignID pgtype.UUID, ownerID string) error {
	campaign, err := s.store.GetCampaignByID(ctx, campaignID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return ErrCampaignNotFound
		}
		return fmt.Errorf("get campaign: %w", err)
	}
	if campaign.OwnerID != ownerID {
		return ErrNotOwner
	}
	return nil
}

// ListByClipper returns all submissions by a clipper.
func (s *SubmissionService) ListByClipper(ctx context.Context, clipperID string) ([]sqlc.Submission, error) {
	return s.store.ListSubmissionsByClipper(ctx, clipperID)
}

// GetByID returns a single submission by ID.
func (s *SubmissionService) GetByID(ctx context.Context, id pgtype.UUID) (*sqlc.Submission, error) {
	submission, err := s.store.GetSubmissionByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get submission: %w", err)
	}
	return &submission, nil
}

// Approve approves a pending submission. The ownerID must match the campaign owner.
func (s *SubmissionService) Approve(ctx context.Context, submissionID pgtype.UUID, ownerID string) (*sqlc.Submission, error) {
	submission, err := s.store.GetSubmissionByID(ctx, submissionID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrSubmissionNotFound
		}
		return nil, fmt.Errorf("get submission: %w", err)
	}

	if submission.Status != "pending" {
		return nil, ErrSubmissionNotPending
	}

	// Verify ownership.
	campaign, err := s.store.GetCampaignByID(ctx, submission.CampaignID)
	if err != nil {
		return nil, fmt.Errorf("get campaign: %w", err)
	}
	if campaign.OwnerID != ownerID {
		return nil, ErrNotOwner
	}

	updated, err := s.store.UpdateSubmissionStatus(ctx, sqlc.UpdateSubmissionStatusParams{
		ID:     submissionID,
		Status: "approved",
	})
	if err != nil {
		return nil, fmt.Errorf("approve submission: %w", err)
	}
	return &updated, nil
}

// Reject rejects a pending submission with an optional reason.
func (s *SubmissionService) Reject(ctx context.Context, submissionID pgtype.UUID, ownerID string, reason string) (*sqlc.Submission, error) {
	submission, err := s.store.GetSubmissionByID(ctx, submissionID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrSubmissionNotFound
		}
		return nil, fmt.Errorf("get submission: %w", err)
	}

	if submission.Status != "pending" {
		return nil, ErrSubmissionNotPending
	}

	// Verify ownership.
	campaign, err := s.store.GetCampaignByID(ctx, submission.CampaignID)
	if err != nil {
		return nil, fmt.Errorf("get campaign: %w", err)
	}
	if campaign.OwnerID != ownerID {
		return nil, ErrNotOwner
	}

	var rejectionReason pgtype.Text
	if reason != "" {
		rejectionReason = pgtype.Text{Valid: true, String: reason}
	}

	updated, err := s.store.UpdateSubmissionStatus(ctx, sqlc.UpdateSubmissionStatusParams{
		ID:              submissionID,
		Status:          "rejected",
		RejectionReason: rejectionReason,
	})
	if err != nil {
		return nil, fmt.Errorf("reject submission: %w", err)
	}
	return &updated, nil
}

// AutoApprove finds pending submissions past their campaign's auto_approve_hours
// and approves them. Returns the count of auto-approved submissions.
func (s *SubmissionService) AutoApprove(ctx context.Context) (int, error) {
	// Get all pending submissions. created_at < now() is always true for existing rows.
	now := pgtype.Timestamptz{Valid: true, Time: time.Now()}
	rows, err := s.store.ListPendingSubmissionsOlderThan(ctx, now)
	if err != nil {
		return 0, fmt.Errorf("list pending submissions: %w", err)
	}

	count := 0
	for _, row := range rows {
		hours := int32(48) // default
		if row.AutoApproveHours.Valid {
			hours = row.AutoApproveHours.Int32
		}

		cutoff := row.CreatedAt.Time.Add(time.Duration(hours) * time.Hour)
		if time.Now().After(cutoff) {
			_, err := s.store.UpdateSubmissionStatus(ctx, sqlc.UpdateSubmissionStatusParams{
				ID:     row.ID,
				Status: "auto_approved",
			})
			if err != nil {
				return count, fmt.Errorf("auto-approve submission %s: %w", row.ID, err)
			}
			count++
		}
	}

	return count, nil
}

// validatePostURL checks that the URL is well-formed.
func validatePostURL(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ErrInvalidURL
	}
	u, err := url.ParseRequestURI(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return ErrInvalidURL
	}
	return nil
}
