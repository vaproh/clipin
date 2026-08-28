package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	sqlc "clipin/apps/api/internal/db/sqlc"
	r "clipin/apps/api/internal/redis"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const notifUnreadCacheTTL = 15 * time.Second

// NotificationStore is the persistence interface the notification service requires.
type NotificationStore interface {
	CreateNotification(ctx context.Context, arg sqlc.CreateNotificationParams) (sqlc.Notification, error)
	ListNotificationsByUser(ctx context.Context, arg sqlc.ListNotificationsByUserParams) ([]sqlc.Notification, error)
	CountUnreadNotifications(ctx context.Context, userID string) (int64, error)
	MarkNotificationRead(ctx context.Context, arg sqlc.MarkNotificationReadParams) error
	MarkAllNotificationsRead(ctx context.Context, userID string) error
	DeleteNotification(ctx context.Context, arg sqlc.DeleteNotificationParams) error
}

// NotificationService implements in-app notification business logic.
type NotificationService struct {
	store NotificationStore
	redis *r.Client
}

// NewNotificationService creates a new NotificationService.
func NewNotificationService(store NotificationStore) *NotificationService {
	return &NotificationService{store: store}
}

// NewNotificationServiceWithCache creates a NotificationService with Redis caching.
func NewNotificationServiceWithCache(store NotificationStore, redis *r.Client) *NotificationService {
	return &NotificationService{store: store, redis: redis}
}

func notifUnreadKey(userID string) string {
	return fmt.Sprintf("notif:unread:%s", userID)
}

// Sentinel errors for notification operations.
var (
	ErrNotificationNotFound = fmt.Errorf("notification not found")
)

// Create creates a notification for a user.
func (s *NotificationService) Create(ctx context.Context, userID, notifType, title, body, link string) (*sqlc.Notification, error) {
	var bodyText, linkText pgtype.Text
	if body != "" {
		bodyText = pgtype.Text{Valid: true, String: body}
	}
	if link != "" {
		linkText = pgtype.Text{Valid: true, String: link}
	}

	n, err := s.store.CreateNotification(ctx, sqlc.CreateNotificationParams{
		UserID: userID,
		Type:   notifType,
		Title:  title,
		Body:   bodyText,
		Link:   linkText,
	})
	if err != nil {
		return nil, fmt.Errorf("create notification: %w", err)
	}
	s.invalidateUnreadCache(ctx, userID)
	return &n, nil
}

// List returns paginated notifications for a user.
func (s *NotificationService) List(ctx context.Context, userID string, limit, offset int32) ([]sqlc.Notification, error) {
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.store.ListNotificationsByUser(ctx, sqlc.ListNotificationsByUserParams{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	})
}

// UnreadCount returns the number of unread notifications for a user.
func (s *NotificationService) UnreadCount(ctx context.Context, userID string) (int64, error) {
	if s.redis != nil {
		if cached, err := s.redis.Get(ctx, notifUnreadKey(userID)); err == nil && len(cached) > 0 {
			if count, err := strconv.ParseInt(string(cached), 10, 64); err == nil {
				return count, nil
			}
		}
	}

	count, err := s.store.CountUnreadNotifications(ctx, userID)
	if err != nil {
		return 0, err
	}

	// Best-effort cache write.
	if s.redis != nil {
		_ = s.redis.Set(ctx, notifUnreadKey(userID), []byte(strconv.FormatInt(count, 10)), notifUnreadCacheTTL)
	}

	return count, nil
}

func (s *NotificationService) invalidateUnreadCache(ctx context.Context, userID string) {
	if s.redis != nil {
		_ = s.redis.Delete(ctx, notifUnreadKey(userID))
	}
}

// MarkRead marks a single notification as read, verifying ownership.
func (s *NotificationService) MarkRead(ctx context.Context, notificationID pgtype.UUID, userID string) error {
	err := s.store.MarkNotificationRead(ctx, sqlc.MarkNotificationReadParams{
		ID:     notificationID,
		UserID: userID,
	})
	if err == nil {
		s.invalidateUnreadCache(ctx, userID)
	}
	return err
}

// MarkAllRead marks all of a user's notifications as read.
func (s *NotificationService) MarkAllRead(ctx context.Context, userID string) error {
	err := s.store.MarkAllNotificationsRead(ctx, userID)
	if err == nil {
		s.invalidateUnreadCache(ctx, userID)
	}
	return err
}

// Delete deletes a notification, verifying ownership.
func (s *NotificationService) Delete(ctx context.Context, notificationID pgtype.UUID, userID string) error {
	return s.store.DeleteNotification(ctx, sqlc.DeleteNotificationParams{
		ID:     notificationID,
		UserID: userID,
	})
}

// --- Helper methods called from other services ---

// NotifySubmissionApproved sends a notification when a submission is approved.
func (s *NotificationService) NotifySubmissionApproved(ctx context.Context, clipperID, campaignTitle string) error {
	_, err := s.Create(ctx, clipperID, "submission_approved",
		"Submission Approved",
		fmt.Sprintf("Your clip for '%s' has been approved!", campaignTitle),
		"",
	)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("notify submission approved: %w", err)
	}
	return nil
}

// NotifySubmissionRejected sends a notification when a submission is rejected.
func (s *NotificationService) NotifySubmissionRejected(ctx context.Context, clipperID, campaignTitle, reason string) error {
	body := fmt.Sprintf("Your clip for '%s' was rejected.", campaignTitle)
	if reason != "" {
		body = fmt.Sprintf("Your clip for '%s' was rejected: %s", campaignTitle, reason)
	}
	_, err := s.Create(ctx, clipperID, "submission_rejected",
		"Submission Rejected",
		body,
		"",
	)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("notify submission rejected: %w", err)
	}
	return nil
}

// NotifyPayoutCompleted sends a notification when a payout completes.
func (s *NotificationService) NotifyPayoutCompleted(ctx context.Context, userID string, amount int64) error {
	_, err := s.Create(ctx, userID, "payout_completed",
		"Payout Completed",
		fmt.Sprintf("Your payout of %d paise has been processed.", amount),
		"",
	)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("notify payout completed: %w", err)
	}
	return nil
}

// NotifyCampaignUpdate sends a notification for campaign status changes.
func (s *NotificationService) NotifyCampaignUpdate(ctx context.Context, ownerID, campaignTitle, message string) error {
	_, err := s.Create(ctx, ownerID, "campaign_update",
		"Campaign Update",
		fmt.Sprintf("Campaign '%s': %s", campaignTitle, message),
		"",
	)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("notify campaign update: %w", err)
	}
	return nil
}

// NotifyAutoApproved sends a notification when a submission is auto-approved.
func (s *NotificationService) NotifyAutoApproved(ctx context.Context, clipperID, campaignTitle string) error {
	_, err := s.Create(ctx, clipperID, "submission_approved",
		"Submission Auto-Approved",
		fmt.Sprintf("Your clip for '%s' was auto-approved", campaignTitle),
		"",
	)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("notify auto-approved: %w", err)
	}
	return nil
}

// NotifyEarningsChanged sends a notification when a clip's eligible views/earnings change.
func (s *NotificationService) NotifyEarningsChanged(ctx context.Context, clipperID, campaignTitle string, eligibleViews int64) error {
	_, err := s.Create(ctx, clipperID, "system",
		"Earnings Updated",
		fmt.Sprintf("Your clip for '%s' now has %d eligible views", campaignTitle, eligibleViews),
		"",
	)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("notify earnings changed: %w", err)
	}
	return nil
}
