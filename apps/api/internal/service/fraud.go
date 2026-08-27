package service

import (
	"context"
	"fmt"

	sqlc "clipin/apps/api/internal/db/sqlc"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// FraudStore is the persistence interface the fraud service requires.
type FraudStore interface {
	CreateFraudFlag(ctx context.Context, arg sqlc.CreateFraudFlagParams) (sqlc.FraudFlag, error)
	ListOpenFraudFlags(ctx context.Context) ([]sqlc.FraudFlag, error)
	GetFraudFlagByID(ctx context.Context, id pgtype.UUID) (sqlc.FraudFlag, error)
	UpdateFraudFlagStatus(ctx context.Context, arg sqlc.UpdateFraudFlagStatusParams) (sqlc.FraudFlag, error)
	CountFraudFlagsByUser(ctx context.Context, userID pgtype.Text) (int32, error)
	ListUsersWithManyFlags(ctx context.Context, dollar_1 int32) ([]sqlc.ListUsersWithManyFlagsRow, error)
}

// FraudService implements fraud detection and flagging.
type FraudService struct {
	store FraudStore
}

// NewFraudService creates a new FraudService.
func NewFraudService(store FraudStore) *FraudService {
	return &FraudService{store: store}
}

// Sentinel errors for fraud operations.
var (
	ErrFlagNotFound = fmt.Errorf("fraud flag not found")
)

// FlagSubmission creates a fraud flag against a submission.
func (s *FraudService) FlagSubmission(ctx context.Context, submissionID pgtype.UUID, flagType, severity, description string) (*sqlc.FraudFlag, error) {
	flag, err := s.store.CreateFraudFlag(ctx, sqlc.CreateFraudFlagParams{
		SubmissionID: submissionID,
		FlagType:     flagType,
		Severity:     severity,
		Description:  pgtype.Text{Valid: description != "", String: description},
	})
	if err != nil {
		return nil, fmt.Errorf("create fraud flag: %w", err)
	}
	return &flag, nil
}

// FlagUser creates a fraud flag against a user (no submission association).
func (s *FraudService) FlagUser(ctx context.Context, userID, flagType, severity, description string) (*sqlc.FraudFlag, error) {
	flag, err := s.store.CreateFraudFlag(ctx, sqlc.CreateFraudFlagParams{
		UserID:      pgtype.Text{Valid: true, String: userID},
		FlagType:    flagType,
		Severity:    severity,
		Description: pgtype.Text{Valid: description != "", String: description},
	})
	if err != nil {
		return nil, fmt.Errorf("create fraud flag: %w", err)
	}
	return &flag, nil
}

// ListOpenFlags returns all open fraud flags, ordered by severity.
func (s *FraudService) ListOpenFlags(ctx context.Context) ([]sqlc.FraudFlag, error) {
	return s.store.ListOpenFraudFlags(ctx)
}

// ResolveFlag marks a flag as resolved with a resolution note.
func (s *FraudService) ResolveFlag(ctx context.Context, flagID pgtype.UUID, resolvedBy, resolution string) (*sqlc.FraudFlag, error) {
	flag, err := s.store.GetFraudFlagByID(ctx, flagID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrFlagNotFound
		}
		return nil, fmt.Errorf("get fraud flag: %w", err)
	}
	_ = flag // exists check

	updated, err := s.store.UpdateFraudFlagStatus(ctx, sqlc.UpdateFraudFlagStatusParams{
		ID:           flagID,
		Status:       "resolved",
		ResolvedBy:   pgtype.Text{Valid: true, String: resolvedBy},
		Resolution:   pgtype.Text{Valid: resolution != "", String: resolution},
	})
	if err != nil {
		return nil, fmt.Errorf("resolve fraud flag: %w", err)
	}
	return &updated, nil
}

// DismissFlag marks a flag as dismissed.
func (s *FraudService) DismissFlag(ctx context.Context, flagID pgtype.UUID, resolvedBy, reason string) (*sqlc.FraudFlag, error) {
	flag, err := s.store.GetFraudFlagByID(ctx, flagID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrFlagNotFound
		}
		return nil, fmt.Errorf("get fraud flag: %w", err)
	}
	_ = flag

	updated, err := s.store.UpdateFraudFlagStatus(ctx, sqlc.UpdateFraudFlagStatusParams{
		ID:           flagID,
		Status:       "dismissed",
		ResolvedBy:   pgtype.Text{Valid: true, String: resolvedBy},
		Resolution:   pgtype.Text{Valid: reason != "", String: reason},
	})
	if err != nil {
		return nil, fmt.Errorf("dismiss fraud flag: %w", err)
	}
	return &updated, nil
}

// GetUserFlagCount returns the number of open fraud flags for a user.
func (s *FraudService) GetUserFlagCount(ctx context.Context, userID string) (int, error) {
	count, err := s.store.CountFraudFlagsByUser(ctx, pgtype.Text{Valid: true, String: userID})
	if err != nil {
		return 0, fmt.Errorf("count fraud flags: %w", err)
	}
	return int(count), nil
}

// GetHighRiskUsers returns users with at least minFlags open flags.
func (s *FraudService) GetHighRiskUsers(ctx context.Context, minFlags int32) ([]sqlc.ListUsersWithManyFlagsRow, error) {
	return s.store.ListUsersWithManyFlags(ctx, minFlags)
}
