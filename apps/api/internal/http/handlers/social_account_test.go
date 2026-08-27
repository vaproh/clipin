package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	sqlc "clipin/apps/api/internal/db/sqlc"
	"clipin/apps/api/internal/http/handlers"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type fakeSocialAccountService struct {
	listByUser func(ctx context.Context, userID string) ([]sqlc.SocialAccount, error)
	connect    func(ctx context.Context, userID, platform, platformUserID, platformUsername string) (*sqlc.SocialAccount, error)
	disconnect func(ctx context.Context, userID string, accountID pgtype.UUID) error
}

func (f *fakeSocialAccountService) ListByUser(ctx context.Context, userID string) ([]sqlc.SocialAccount, error) {
	if f.listByUser != nil {
		return f.listByUser(ctx, userID)
	}
	return nil, nil
}

func (f *fakeSocialAccountService) Connect(ctx context.Context, userID, platform, platformUserID, platformUsername string) (*sqlc.SocialAccount, error) {
	if f.connect != nil {
		return f.connect(ctx, userID, platform, platformUserID, platformUsername)
	}
	return &sqlc.SocialAccount{}, nil
}

func (f *fakeSocialAccountService) Disconnect(ctx context.Context, userID string, accountID pgtype.UUID) error {
	if f.disconnect != nil {
		return f.disconnect(ctx, userID, accountID)
	}
	return nil
}

func socialAccountRouter(svc handlers.SocialAccountServiceInterface) http.Handler {
	r := chi.NewRouter()
	api := humachi.New(r, huma.DefaultConfig("ClipIN API", "1.0.0"))
	handlers.RegisterSocialAccountHandlers(api, svc)
	return r
}

func TestListSocialAccounts(t *testing.T) {
	svc := &fakeSocialAccountService{
		listByUser: func(_ context.Context, userID string) ([]sqlc.SocialAccount, error) {
			return []sqlc.SocialAccount{
				{
					ID:               pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
					UserID:           "user_1",
					Platform:         "youtube",
					PlatformUserID:   "UC123",
					PlatformUsername: pgtype.Text{Valid: true, String: "TestChannel"},
				},
			}, nil
		},
	}
	handler := withUser(testUser())(socialAccountRouter(svc))
	rec := doRequest(t, handler, http.MethodGet, "/me/social-accounts", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var out struct {
		Accounts []struct {
			ID             string `json:"id"`
			Platform       string `json:"platform"`
			PlatformUserID string `json:"platform_user_id"`
		} `json:"accounts"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out.Accounts) != 1 {
		t.Fatalf("expected 1 account, got %d", len(out.Accounts))
	}
	if out.Accounts[0].Platform != "youtube" {
		t.Errorf("expected youtube, got %s", out.Accounts[0].Platform)
	}
}

func TestListSocialAccountsUnauthorized(t *testing.T) {
	rec := doRequest(t, socialAccountRouter(&fakeSocialAccountService{}), http.MethodGet, "/me/social-accounts", "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestConnectSocialAccount(t *testing.T) {
	svc := &fakeSocialAccountService{
		connect: func(_ context.Context, userID, platform, platformUserID, platformUsername string) (*sqlc.SocialAccount, error) {
			if platform != "youtube" {
				t.Errorf("expected youtube, got %s", platform)
			}
			return &sqlc.SocialAccount{
				ID:               pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
				UserID:           userID,
				Platform:         platform,
				PlatformUserID:   platformUserID,
				PlatformUsername: pgtype.Text{Valid: true, String: platformUsername},
			}, nil
		},
	}
	handler := withUser(testUser())(socialAccountRouter(svc))
	rec := doRequest(t, handler, http.MethodPost, "/me/social-accounts", `{"platform":"youtube","platform_user_id":"UC123","platform_username":"TestChannel"}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var out struct {
		Platform string `json:"platform"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Platform != "youtube" {
		t.Errorf("expected youtube, got %s", out.Platform)
	}
}

func TestConnectSocialAccountMissingPlatform(t *testing.T) {
	svc := &fakeSocialAccountService{}
	handler := withUser(testUser())(socialAccountRouter(svc))
	rec := doRequest(t, handler, http.MethodPost, "/me/social-accounts", `{"platform_user_id":"UC123"}`)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", rec.Code)
	}
}

func TestDisconnectSocialAccount(t *testing.T) {
	called := false
	svc := &fakeSocialAccountService{
		disconnect: func(_ context.Context, userID string, accountID pgtype.UUID) error {
			called = true
			if userID != "user_1" {
				t.Errorf("expected user_1, got %s", userID)
			}
			return nil
		},
	}
	handler := withUser(testUser())(socialAccountRouter(svc))
	rec := doRequest(t, handler, http.MethodDelete, "/me/social-accounts/0102030405060708090a0b0c0d0e0f10", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !called {
		t.Error("expected disconnect to be called")
	}
}

func TestDisconnectSocialAccountUnauthorized(t *testing.T) {
	rec := doRequest(t, socialAccountRouter(&fakeSocialAccountService{}), http.MethodDelete, "/me/social-accounts/0102030405060708090a0b0c0d0e0f10", "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
