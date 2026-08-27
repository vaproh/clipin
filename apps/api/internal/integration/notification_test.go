//go:build integration

package integration

import (
	"context"
	"testing"

	sqlc "clipin/apps/api/internal/db/sqlc"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestCreateNotification(t *testing.T) {
	cleanupAll(t)

	seedUser(t, "notif_user1", "notif1@notif.com", "clipper")

	notif, err := testDB.Queries.CreateNotification(context.Background(), sqlc.CreateNotificationParams{
		UserID: "notif_user1",
		Type:   "submission_approved",
		Title:  "Your clip was approved!",
		Body:   pgtype.Text{String: "Great work!", Valid: true},
	})
	if err != nil {
		t.Fatalf("CreateNotification: %v", err)
	}
	assertEqual(t, "Type", notif.Type, "submission_approved")
	assertEqual(t, "Title", notif.Title, "Your clip was approved!")
	if notif.IsRead {
		t.Error("expected IsRead to be false")
	}
}

func TestListNotificationsByUser(t *testing.T) {
	cleanupAll(t)

	seedUser(t, "notif_list_user", "notiflist@notif.com", "clipper")

	for i := 0; i < 5; i++ {
		_, err := testDB.Queries.CreateNotification(context.Background(), sqlc.CreateNotificationParams{
			UserID: "notif_list_user",
			Type:   "system",
			Title:  "Notification " + string(rune('A'+i)),
		})
		if err != nil {
			t.Fatalf("CreateNotification %d: %v", i, err)
		}
	}

	notifs, err := testDB.Queries.ListNotificationsByUser(context.Background(), sqlc.ListNotificationsByUserParams{
		UserID: "notif_list_user",
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("ListNotificationsByUser: %v", err)
	}
	if len(notifs) != 5 {
		t.Errorf("expected 5 notifications, got %d", len(notifs))
	}
}

func TestCountUnreadNotifications(t *testing.T) {
	cleanupAll(t)

	seedUser(t, "notif_unread_user", "notifunread@notif.com", "clipper")

	// Create 3 notifications
	for i := 0; i < 3; i++ {
		_, err := testDB.Queries.CreateNotification(context.Background(), sqlc.CreateNotificationParams{
			UserID: "notif_unread_user",
			Type:   "system",
			Title:  "Unread " + string(rune('A'+i)),
		})
		if err != nil {
			t.Fatalf("CreateNotification %d: %v", i, err)
		}
	}

	count, err := testDB.Queries.CountUnreadNotifications(context.Background(), "notif_unread_user")
	if err != nil {
		t.Fatalf("CountUnreadNotifications: %v", err)
	}
	assertInt64Equal(t, "CountUnreadNotifications", count, 3)
}

func TestMarkNotificationRead(t *testing.T) {
	cleanupAll(t)

	seedUser(t, "notif_mark_user", "notifmark@notif.com", "clipper")

	notif, err := testDB.Queries.CreateNotification(context.Background(), sqlc.CreateNotificationParams{
		UserID: "notif_mark_user",
		Type:   "system",
		Title:  "Mark me",
	})
	if err != nil {
		t.Fatalf("CreateNotification: %v", err)
	}

	err = testDB.Queries.MarkNotificationRead(context.Background(), sqlc.MarkNotificationReadParams{
		ID:     notif.ID,
		UserID: "notif_mark_user",
	})
	if err != nil {
		t.Fatalf("MarkNotificationRead: %v", err)
	}

	count, err := testDB.Queries.CountUnreadNotifications(context.Background(), "notif_mark_user")
	if err != nil {
		t.Fatalf("CountUnreadNotifications after mark: %v", err)
	}
	assertInt64Equal(t, "CountUnreadNotifications after mark", count, 0)
}

func TestMarkAllNotificationsRead(t *testing.T) {
	cleanupAll(t)

	seedUser(t, "notif_markall_user", "notifmarkall@notif.com", "clipper")

	for i := 0; i < 5; i++ {
		_, err := testDB.Queries.CreateNotification(context.Background(), sqlc.CreateNotificationParams{
			UserID: "notif_markall_user",
			Type:   "system",
			Title:  "Bulk " + string(rune('A'+i)),
		})
		if err != nil {
			t.Fatalf("CreateNotification %d: %v", i, err)
		}
	}

	err := testDB.Queries.MarkAllNotificationsRead(context.Background(), "notif_markall_user")
	if err != nil {
		t.Fatalf("MarkAllNotificationsRead: %v", err)
	}

	count, err := testDB.Queries.CountUnreadNotifications(context.Background(), "notif_markall_user")
	if err != nil {
		t.Fatalf("CountUnreadNotifications after mark all: %v", err)
	}
	assertInt64Equal(t, "CountUnreadNotifications after mark all", count, 0)
}

func TestDeleteNotification(t *testing.T) {
	cleanupAll(t)

	seedUser(t, "notif_del_user", "notifdel@notif.com", "clipper")

	notif, err := testDB.Queries.CreateNotification(context.Background(), sqlc.CreateNotificationParams{
		UserID: "notif_del_user",
		Type:   "system",
		Title:  "Delete me",
	})
	if err != nil {
		t.Fatalf("CreateNotification: %v", err)
	}

	err = testDB.Queries.DeleteNotification(context.Background(), sqlc.DeleteNotificationParams{
		ID:     notif.ID,
		UserID: "notif_del_user",
	})
	if err != nil {
		t.Fatalf("DeleteNotification: %v", err)
	}

	count, err := testDB.Queries.CountUnreadNotifications(context.Background(), "notif_del_user")
	if err != nil {
		t.Fatalf("CountUnreadNotifications after delete: %v", err)
	}
	assertInt64Equal(t, "CountUnreadNotifications after delete", count, 0)
}

func TestListNotificationsPagination(t *testing.T) {
	cleanupAll(t)

	seedUser(t, "notif_page_user", "notifpage@notif.com", "clipper")

	for i := 0; i < 10; i++ {
		_, err := testDB.Queries.CreateNotification(context.Background(), sqlc.CreateNotificationParams{
			UserID: "notif_page_user",
			Type:   "system",
			Title:  "Page " + string(rune('A'+i)),
		})
		if err != nil {
			t.Fatalf("CreateNotification %d: %v", i, err)
		}
	}

	// First page
	page1, err := testDB.Queries.ListNotificationsByUser(context.Background(), sqlc.ListNotificationsByUserParams{
		UserID: "notif_page_user",
		Limit:  3,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("ListNotificationsByUser page 1: %v", err)
	}
	if len(page1) != 3 {
		t.Errorf("expected 3 notifications on page 1, got %d", len(page1))
	}

	// Second page
	page2, err := testDB.Queries.ListNotificationsByUser(context.Background(), sqlc.ListNotificationsByUserParams{
		UserID: "notif_page_user",
		Limit:  3,
		Offset: 3,
	})
	if err != nil {
		t.Fatalf("ListNotificationsByUser page 2: %v", err)
	}
	if len(page2) != 3 {
		t.Errorf("expected 3 notifications on page 2, got %d", len(page2))
	}
}
