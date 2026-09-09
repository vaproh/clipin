package service

import (
	"context"
	"fmt"

	sqlc "clipin/apps/api/internal/db/sqlc"

	"github.com/jackc/pgx/v5/pgtype"
)

// SocialAccountStore is the persistence interface the social account service requires.
type SocialAccountStore interface {
	ListSocialAccountsByUserID(ctx context.Context, userID string) ([]sqlc.SocialAccount, error)
	UpsertSocialAccount(ctx context.Context, arg sqlc.UpsertSocialAccountParams) (sqlc.SocialAccount, error)
	DeleteSocialAccount(ctx context.Context, arg sqlc.DeleteSocialAccountParams) error
}

// SocialAccountService implements social account linking business logic.
type SocialAccountService struct {
	store SocialAccountStore
}

// NewSocialAccountService creates a new SocialAccountService.
func NewSocialAccountService(store SocialAccountStore) *SocialAccountService {
	return &SocialAccountService{store: store}
}

// Sentinel errors for social account operations.
var (
	ErrAccountNotFound = fmt.Errorf("social account not found")
	ErrInvalidPlatform = fmt.Errorf("social accounts support YouTube and Instagram only")
)

// ListByUser returns all social accounts for a user.
func (s *SocialAccountService) ListByUser(ctx context.Context, userID string) ([]sqlc.SocialAccount, error) {
	return s.store.ListSocialAccountsByUserID(ctx, userID)
}

// Connect upserts a social account for a user (stub, no real OAuth).
func (s *SocialAccountService) Connect(ctx context.Context, userID, platform, platformUserID, platformUsername string) (*sqlc.SocialAccount, error) {
	if platform != "youtube" && platform != "instagram" {
		return nil, ErrInvalidPlatform
	}
	var username pgtype.Text
	if platformUsername != "" {
		username = pgtype.Text{Valid: true, String: platformUsername}
	}

	account, err := s.store.UpsertSocialAccount(ctx, sqlc.UpsertSocialAccountParams{
		UserID:           userID,
		Platform:         platform,
		PlatformUserID:   platformUserID,
		PlatformUsername: username,
	})
	if err != nil {
		return nil, fmt.Errorf("upsert social account: %w", err)
	}
	return &account, nil
}

// Disconnect removes a social account, verifying ownership.
func (s *SocialAccountService) Disconnect(ctx context.Context, userID string, accountID pgtype.UUID) error {
	return s.store.DeleteSocialAccount(ctx, sqlc.DeleteSocialAccountParams{
		ID:     accountID,
		UserID: userID,
	})
}
