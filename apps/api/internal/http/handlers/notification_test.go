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

type fakeNotificationService struct {
	create      func(ctx context.Context, userID, notifType, title, body, link string) (*sqlc.Notification, error)
	list        func(ctx context.Context, userID string, limit, offset int32) ([]sqlc.Notification, error)
	unreadCount func(ctx context.Context, userID string) (int64, error)
	markRead    func(ctx context.Context, notificationID pgtype.UUID, userID string) error
	markAllRead func(ctx context.Context, userID string) error
	delete      func(ctx context.Context, notificationID pgtype.UUID, userID string) error
}

func (f *fakeNotificationService) Create(ctx context.Context, userID, notifType, title, body, link string) (*sqlc.Notification, error) {
	if f.create != nil {
		return f.create(ctx, userID, notifType, title, body, link)
	}
	return &sqlc.Notification{}, nil
}

func (f *fakeNotificationService) List(ctx context.Context, userID string, limit, offset int32) ([]sqlc.Notification, error) {
	if f.list != nil {
		return f.list(ctx, userID, limit, offset)
	}
	return nil, nil
}

func (f *fakeNotificationService) UnreadCount(ctx context.Context, userID string) (int64, error) {
	if f.unreadCount != nil {
		return f.unreadCount(ctx, userID)
	}
	return 0, nil
}

func (f *fakeNotificationService) MarkRead(ctx context.Context, notificationID pgtype.UUID, userID string) error {
	if f.markRead != nil {
		return f.markRead(ctx, notificationID, userID)
	}
	return nil
}

func (f *fakeNotificationService) MarkAllRead(ctx context.Context, userID string) error {
	if f.markAllRead != nil {
		return f.markAllRead(ctx, userID)
	}
	return nil
}

func (f *fakeNotificationService) Delete(ctx context.Context, notificationID pgtype.UUID, userID string) error {
	if f.delete != nil {
		return f.delete(ctx, notificationID, userID)
	}
	return nil
}

func notifRouter(svc handlers.NotificationServiceInterface) http.Handler {
	r := chi.NewRouter()
	api := humachi.New(r, huma.DefaultConfig("ClipIN API", "1.0.0"))
	handlers.RegisterNotificationHandlers(api, svc)
	return r
}

func TestListNotifications(t *testing.T) {
	svc := &fakeNotificationService{
		list: func(_ context.Context, userID string, limit, offset int32) ([]sqlc.Notification, error) {
			return []sqlc.Notification{
				{
					ID:     pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
					UserID: "user_1",
					Type:   "submission_approved",
					Title:  "Approved",
					IsRead: false,
				},
			}, nil
		},
	}
	handler := withUser(testUser())(notifRouter(svc))
	rec := doRequest(t, handler, http.MethodGet, "/me/notifications", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var out struct {
		Notifications []struct {
			ID    string `json:"id"`
			Type  string `json:"type"`
			Title string `json:"title"`
		} `json:"notifications"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out.Notifications) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(out.Notifications))
	}
	if out.Notifications[0].Title != "Approved" {
		t.Errorf("expected Approved, got %s", out.Notifications[0].Title)
	}
}

func TestListNotificationsUnauthorized(t *testing.T) {
	rec := doRequest(t, notifRouter(&fakeNotificationService{}), http.MethodGet, "/me/notifications", "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestUnreadCount(t *testing.T) {
	svc := &fakeNotificationService{
		unreadCount: func(_ context.Context, userID string) (int64, error) {
			return 3, nil
		},
	}
	handler := withUser(testUser())(notifRouter(svc))
	rec := doRequest(t, handler, http.MethodGet, "/me/notifications/unread-count", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var out struct {
		Count int64 `json:"count"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Count != 3 {
		t.Errorf("expected 3, got %d", out.Count)
	}
}

func TestMarkAllNotificationsRead(t *testing.T) {
	called := false
	svc := &fakeNotificationService{
		markAllRead: func(_ context.Context, userID string) error {
			called = true
			return nil
		},
	}
	handler := withUser(testUser())(notifRouter(svc))
	rec := doRequest(t, handler, http.MethodPost, "/me/notifications/read-all", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !called {
		t.Error("expected markAllRead to be called")
	}
}

func TestMarkAllNotificationsReadUnauthorized(t *testing.T) {
	rec := doRequest(t, notifRouter(&fakeNotificationService{}), http.MethodPost, "/me/notifications/read-all", "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
