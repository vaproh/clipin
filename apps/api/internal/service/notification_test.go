package service_test

import (
	"context"
	"testing"

	sqlc "clipin/apps/api/internal/db/sqlc"
	"clipin/apps/api/internal/service"

	"github.com/jackc/pgx/v5/pgtype"
)

type mockNotificationStore struct {
	create       func(ctx context.Context, arg sqlc.CreateNotificationParams) (sqlc.Notification, error)
	list         func(ctx context.Context, arg sqlc.ListNotificationsByUserParams) ([]sqlc.Notification, error)
	countUnread  func(ctx context.Context, userID string) (int64, error)
	markRead     func(ctx context.Context, arg sqlc.MarkNotificationReadParams) error
	markAllRead  func(ctx context.Context, userID string) error
	deleteNotif  func(ctx context.Context, arg sqlc.DeleteNotificationParams) error
}

func (m *mockNotificationStore) CreateNotification(ctx context.Context, arg sqlc.CreateNotificationParams) (sqlc.Notification, error) {
	if m.create != nil {
		return m.create(ctx, arg)
	}
	return sqlc.Notification{
		ID:     pgtype.UUID{Bytes: [16]byte{99}, Valid: true},
		UserID: arg.UserID,
		Type:   arg.Type,
		Title:  arg.Title,
		Body:   arg.Body,
		Link:   arg.Link,
	}, nil
}

func (m *mockNotificationStore) ListNotificationsByUser(ctx context.Context, arg sqlc.ListNotificationsByUserParams) ([]sqlc.Notification, error) {
	if m.list != nil {
		return m.list(ctx, arg)
	}
	return nil, nil
}

func (m *mockNotificationStore) CountUnreadNotifications(ctx context.Context, userID string) (int64, error) {
	if m.countUnread != nil {
		return m.countUnread(ctx, userID)
	}
	return 0, nil
}

func (m *mockNotificationStore) MarkNotificationRead(ctx context.Context, arg sqlc.MarkNotificationReadParams) error {
	if m.markRead != nil {
		return m.markRead(ctx, arg)
	}
	return nil
}

func (m *mockNotificationStore) MarkAllNotificationsRead(ctx context.Context, userID string) error {
	if m.markAllRead != nil {
		return m.markAllRead(ctx, userID)
	}
	return nil
}

func (m *mockNotificationStore) DeleteNotification(ctx context.Context, arg sqlc.DeleteNotificationParams) error {
	if m.deleteNotif != nil {
		return m.deleteNotif(ctx, arg)
	}
	return nil
}

func TestNotificationCreate(t *testing.T) {
	store := &mockNotificationStore{
		create: func(_ context.Context, arg sqlc.CreateNotificationParams) (sqlc.Notification, error) {
			if arg.UserID != "user1" {
				t.Errorf("expected user1, got %s", arg.UserID)
			}
			if arg.Type != "submission_approved" {
				t.Errorf("expected submission_approved, got %s", arg.Type)
			}
			if arg.Title != "Approved" {
				t.Errorf("expected Approved, got %s", arg.Title)
			}
			return sqlc.Notification{
				ID:     pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
				UserID: arg.UserID,
				Type:   arg.Type,
				Title:  arg.Title,
				Body:   arg.Body,
				Link:   arg.Link,
			}, nil
		},
	}
	svc := service.NewNotificationService(store)
	n, err := svc.Create(context.Background(), "user1", "submission_approved", "Approved", "Your clip was approved", "/campaigns/1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.UserID != "user1" {
		t.Errorf("expected user1, got %s", n.UserID)
	}
}

func TestNotificationCreateEmptyBody(t *testing.T) {
	store := &mockNotificationStore{
		create: func(_ context.Context, arg sqlc.CreateNotificationParams) (sqlc.Notification, error) {
			return sqlc.Notification{
				ID:     pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
				UserID: arg.UserID,
				Type:   arg.Type,
				Title:  arg.Title,
			}, nil
		},
	}
	svc := service.NewNotificationService(store)
	n, err := svc.Create(context.Background(), "user1", "system", "Hello", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.Title != "Hello" {
		t.Errorf("expected Hello, got %s", n.Title)
	}
}

func TestNotificationList(t *testing.T) {
	expected := []sqlc.Notification{
		{ID: pgtype.UUID{Bytes: [16]byte{1}, Valid: true}, Type: "submission_approved", Title: "Approved"},
	}
	store := &mockNotificationStore{
		list: func(_ context.Context, arg sqlc.ListNotificationsByUserParams) ([]sqlc.Notification, error) {
			if arg.UserID != "user1" {
				t.Errorf("expected user1, got %s", arg.UserID)
			}
			if arg.Limit != 10 {
				t.Errorf("expected limit 10, got %d", arg.Limit)
			}
			return expected, nil
		},
	}
	svc := service.NewNotificationService(store)
	notifs, err := svc.List(context.Background(), "user1", 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(notifs) != 1 {
		t.Fatalf("expected 1, got %d", len(notifs))
	}
	if notifs[0].Title != "Approved" {
		t.Errorf("expected Approved, got %s", notifs[0].Title)
	}
}

func TestNotificationListDefaultLimit(t *testing.T) {
	store := &mockNotificationStore{
		list: func(_ context.Context, arg sqlc.ListNotificationsByUserParams) ([]sqlc.Notification, error) {
			if arg.Limit != 20 {
				t.Errorf("expected default limit 20, got %d", arg.Limit)
			}
			return nil, nil
		},
	}
	svc := service.NewNotificationService(store)
	_, _ = svc.List(context.Background(), "user1", 0, 0)
}

func TestNotificationUnreadCount(t *testing.T) {
	store := &mockNotificationStore{
		countUnread: func(_ context.Context, userID string) (int64, error) {
			if userID != "user1" {
				t.Errorf("expected user1, got %s", userID)
			}
			return 5, nil
		},
	}
	svc := service.NewNotificationService(store)
	count, err := svc.UnreadCount(context.Background(), "user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 5 {
		t.Errorf("expected 5, got %d", count)
	}
}

func TestNotificationMarkRead(t *testing.T) {
	nID := pgtype.UUID{Bytes: [16]byte{42}, Valid: true}
	store := &mockNotificationStore{
		markRead: func(_ context.Context, arg sqlc.MarkNotificationReadParams) error {
			if arg.UserID != "user1" {
				t.Errorf("expected user1, got %s", arg.UserID)
			}
			if arg.ID != nID {
				t.Errorf("expected notification ID to match")
			}
			return nil
		},
	}
	svc := service.NewNotificationService(store)
	err := svc.MarkRead(context.Background(), nID, "user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNotificationMarkAllRead(t *testing.T) {
	store := &mockNotificationStore{
		markAllRead: func(_ context.Context, userID string) error {
			if userID != "user1" {
				t.Errorf("expected user1, got %s", userID)
			}
			return nil
		},
	}
	svc := service.NewNotificationService(store)
	err := svc.MarkAllRead(context.Background(), "user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNotificationDelete(t *testing.T) {
	nID := pgtype.UUID{Bytes: [16]byte{77}, Valid: true}
	store := &mockNotificationStore{
		deleteNotif: func(_ context.Context, arg sqlc.DeleteNotificationParams) error {
			if arg.UserID != "user1" {
				t.Errorf("expected user1, got %s", arg.UserID)
			}
			if arg.ID != nID {
				t.Errorf("expected notification ID to match")
			}
			return nil
		},
	}
	svc := service.NewNotificationService(store)
	err := svc.Delete(context.Background(), nID, "user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNotifySubmissionApproved(t *testing.T) {
	store := &mockNotificationStore{
		create: func(_ context.Context, arg sqlc.CreateNotificationParams) (sqlc.Notification, error) {
			if arg.Type != "submission_approved" {
				t.Errorf("expected submission_approved, got %s", arg.Type)
			}
			if arg.UserID != "clipper1" {
				t.Errorf("expected clipper1, got %s", arg.UserID)
			}
			return sqlc.Notification{}, nil
		},
	}
	svc := service.NewNotificationService(store)
	err := svc.NotifySubmissionApproved(context.Background(), "clipper1", "Test Campaign")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNotifySubmissionRejected(t *testing.T) {
	store := &mockNotificationStore{
		create: func(_ context.Context, arg sqlc.CreateNotificationParams) (sqlc.Notification, error) {
			if arg.Type != "submission_rejected" {
				t.Errorf("expected submission_rejected, got %s", arg.Type)
			}
			return sqlc.Notification{}, nil
		},
	}
	svc := service.NewNotificationService(store)
	err := svc.NotifySubmissionRejected(context.Background(), "clipper1", "Test Campaign", "low quality")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNotifyPayoutCompleted(t *testing.T) {
	store := &mockNotificationStore{
		create: func(_ context.Context, arg sqlc.CreateNotificationParams) (sqlc.Notification, error) {
			if arg.Type != "payout_completed" {
				t.Errorf("expected payout_completed, got %s", arg.Type)
			}
			return sqlc.Notification{}, nil
		},
	}
	svc := service.NewNotificationService(store)
	err := svc.NotifyPayoutCompleted(context.Background(), "user1", 50000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNotifyCampaignUpdate(t *testing.T) {
	store := &mockNotificationStore{
		create: func(_ context.Context, arg sqlc.CreateNotificationParams) (sqlc.Notification, error) {
			if arg.Type != "campaign_update" {
				t.Errorf("expected campaign_update, got %s", arg.Type)
			}
			return sqlc.Notification{}, nil
		},
	}
	svc := service.NewNotificationService(store)
	err := svc.NotifyCampaignUpdate(context.Background(), "owner1", "My Campaign", "was paused")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNotifyAutoApproved(t *testing.T) {
	store := &mockNotificationStore{
		create: func(_ context.Context, arg sqlc.CreateNotificationParams) (sqlc.Notification, error) {
			if arg.Type != "submission_approved" {
				t.Errorf("expected submission_approved, got %s", arg.Type)
			}
			if arg.UserID != "clipper1" {
				t.Errorf("expected clipper1, got %s", arg.UserID)
			}
			if arg.Title != "Submission Auto-Approved" {
				t.Errorf("expected 'Submission Auto-Approved', got %s", arg.Title)
			}
			return sqlc.Notification{}, nil
		},
	}
	svc := service.NewNotificationService(store)
	err := svc.NotifyAutoApproved(context.Background(), "clipper1", "Test Campaign")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNotifyEarningsChanged(t *testing.T) {
	store := &mockNotificationStore{
		create: func(_ context.Context, arg sqlc.CreateNotificationParams) (sqlc.Notification, error) {
			if arg.Type != "system" {
				t.Errorf("expected system, got %s", arg.Type)
			}
			if arg.UserID != "clipper1" {
				t.Errorf("expected clipper1, got %s", arg.UserID)
			}
			return sqlc.Notification{}, nil
		},
	}
	svc := service.NewNotificationService(store)
	err := svc.NotifyEarningsChanged(context.Background(), "clipper1", "Test Campaign", 5000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
