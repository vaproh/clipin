package service_test

import (
	"context"
	"testing"

	sqlc "clipin/apps/api/internal/db/sqlc"
	"clipin/apps/api/internal/service"

	"github.com/jackc/pgx/v5/pgtype"
)

type mockSocialAccountStore struct {
	listByUser func(ctx context.Context, userID string) ([]sqlc.SocialAccount, error)
	upsert     func(ctx context.Context, arg sqlc.UpsertSocialAccountParams) (sqlc.SocialAccount, error)
	delete     func(ctx context.Context, arg sqlc.DeleteSocialAccountParams) error
}

func (m *mockSocialAccountStore) ListSocialAccountsByUserID(ctx context.Context, userID string) ([]sqlc.SocialAccount, error) {
	if m.listByUser != nil {
		return m.listByUser(ctx, userID)
	}
	return nil, nil
}

func (m *mockSocialAccountStore) UpsertSocialAccount(ctx context.Context, arg sqlc.UpsertSocialAccountParams) (sqlc.SocialAccount, error) {
	if m.upsert != nil {
		return m.upsert(ctx, arg)
	}
	return sqlc.SocialAccount{
		ID:             pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
		UserID:         arg.UserID,
		Platform:       arg.Platform,
		PlatformUserID: arg.PlatformUserID,
	}, nil
}

func (m *mockSocialAccountStore) DeleteSocialAccount(ctx context.Context, arg sqlc.DeleteSocialAccountParams) error {
	if m.delete != nil {
		return m.delete(ctx, arg)
	}
	return nil
}

func TestSocialAccountListByUser(t *testing.T) {
	expected := []sqlc.SocialAccount{
		{ID: pgtype.UUID{Bytes: [16]byte{1}, Valid: true}, Platform: "youtube", PlatformUserID: "UC123"},
	}
	store := &mockSocialAccountStore{
		listByUser: func(_ context.Context, userID string) ([]sqlc.SocialAccount, error) {
			if userID != "user1" {
				t.Errorf("expected user1, got %s", userID)
			}
			return expected, nil
		},
	}
	svc := service.NewSocialAccountService(store)
	accounts, err := svc.ListByUser(context.Background(), "user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(accounts) != 1 {
		t.Fatalf("expected 1, got %d", len(accounts))
	}
	if accounts[0].Platform != "youtube" {
		t.Errorf("expected youtube, got %s", accounts[0].Platform)
	}
}

func TestSocialAccountConnect(t *testing.T) {
	store := &mockSocialAccountStore{
		upsert: func(_ context.Context, arg sqlc.UpsertSocialAccountParams) (sqlc.SocialAccount, error) {
			if arg.UserID != "user1" {
				t.Errorf("expected user1, got %s", arg.UserID)
			}
			if arg.Platform != "youtube" {
				t.Errorf("expected youtube, got %s", arg.Platform)
			}
			if arg.PlatformUserID != "UC123" {
				t.Errorf("expected UC123, got %s", arg.PlatformUserID)
			}
			if !arg.PlatformUsername.Valid || arg.PlatformUsername.String != "TestChannel" {
				t.Errorf("expected TestChannel, got %+v", arg.PlatformUsername)
			}
			return sqlc.SocialAccount{
				ID:               pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
				UserID:           arg.UserID,
				Platform:         arg.Platform,
				PlatformUserID:   arg.PlatformUserID,
				PlatformUsername: arg.PlatformUsername,
			}, nil
		},
	}
	svc := service.NewSocialAccountService(store)
	account, err := svc.Connect(context.Background(), "user1", "youtube", "UC123", "TestChannel")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if account.Platform != "youtube" {
		t.Errorf("expected youtube, got %s", account.Platform)
	}
}

func TestSocialAccountConnectEmptyUsername(t *testing.T) {
	store := &mockSocialAccountStore{
		upsert: func(_ context.Context, arg sqlc.UpsertSocialAccountParams) (sqlc.SocialAccount, error) {
			if arg.PlatformUsername.Valid {
				t.Errorf("expected empty username, got %+v", arg.PlatformUsername)
			}
			return sqlc.SocialAccount{
				ID:             pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
				UserID:         arg.UserID,
				Platform:       arg.Platform,
				PlatformUserID: arg.PlatformUserID,
			}, nil
		},
	}
	svc := service.NewSocialAccountService(store)
	_, err := svc.Connect(context.Background(), "user1", "youtube", "UC123", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSocialAccountConnectRejectsTikTok(t *testing.T) {
	called := false
	store := &mockSocialAccountStore{
		upsert: func(_ context.Context, _ sqlc.UpsertSocialAccountParams) (sqlc.SocialAccount, error) {
			called = true
			return sqlc.SocialAccount{}, nil
		},
	}
	svc := service.NewSocialAccountService(store)
	_, err := svc.Connect(context.Background(), "user1", "tiktok", "123", "creator")
	if err != service.ErrInvalidPlatform {
		t.Fatalf("expected ErrInvalidPlatform, got %v", err)
	}
	if called {
		t.Fatal("did not expect TikTok account to be persisted")
	}
}

func TestSocialAccountDisconnect(t *testing.T) {
	accountID := pgtype.UUID{Bytes: [16]byte{42}, Valid: true}
	store := &mockSocialAccountStore{
		delete: func(_ context.Context, arg sqlc.DeleteSocialAccountParams) error {
			if arg.UserID != "user1" {
				t.Errorf("expected user1, got %s", arg.UserID)
			}
			if arg.ID != accountID {
				t.Errorf("expected account ID to match")
			}
			return nil
		},
	}
	svc := service.NewSocialAccountService(store)
	err := svc.Disconnect(context.Background(), "user1", accountID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
