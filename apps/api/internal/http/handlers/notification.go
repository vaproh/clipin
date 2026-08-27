package handlers

import (
	"context"
	"fmt"

	sqlc "clipin/apps/api/internal/db/sqlc"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgtype"
)

// NotificationServiceInterface is the subset of NotificationService the handlers need.
type NotificationServiceInterface interface {
	Create(ctx context.Context, userID, notifType, title, body, link string) (*sqlc.Notification, error)
	List(ctx context.Context, userID string, limit, offset int32) ([]sqlc.Notification, error)
	UnreadCount(ctx context.Context, userID string) (int64, error)
	MarkRead(ctx context.Context, notificationID pgtype.UUID, userID string) error
	MarkAllRead(ctx context.Context, userID string) error
	Delete(ctx context.Context, notificationID pgtype.UUID, userID string) error
}

type notificationItem struct {
	ID        string  `json:"id"`
	Type      string  `json:"type"`
	Title     string  `json:"title"`
	Body      *string `json:"body,omitempty"`
	Link      *string `json:"link,omitempty"`
	IsRead    bool    `json:"is_read"`
	CreatedAt string  `json:"created_at"`
}

func toNotificationItem(n sqlc.Notification) notificationItem {
	item := notificationItem{
		ID:     fmt.Sprintf("%x", n.ID.Bytes),
		Type:   n.Type,
		Title:  n.Title,
		IsRead: n.IsRead,
	}
	if n.Body.Valid {
		item.Body = &n.Body.String
	}
	if n.Link.Valid {
		item.Link = &n.Link.String
	}
	if n.CreatedAt.Valid {
		item.CreatedAt = n.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00")
	}
	return item
}

// RegisterNotificationHandlers registers notification endpoints on the authenticated API.
func RegisterNotificationHandlers(api huma.API, svc NotificationServiceInterface) {
	// GET /me/notifications - paginated list
	huma.Register(api, huma.Operation{
		OperationID: "list-my-notifications",
		Method:      "GET",
		Path:        "/me/notifications",
		Summary:     "List my notifications",
		Description: "Returns a paginated list of the current user's notifications.",
		Tags:        []string{"Notifications"},
	}, func(ctx context.Context, input *struct {
		Page     int `query:"page" doc:"Page number (1-indexed)"`
		PageSize int `query:"page_size" doc:"Results per page (max 100)"`
	}) (*listNotificationsOutput, error) {
		user, err := requireUser(ctx)
		if err != nil {
			return nil, err
		}

		page := input.Page
		if page < 1 {
			page = 1
		}
		pageSize := input.PageSize
		if pageSize < 1 || pageSize > 100 {
			pageSize = 20
		}
		offset := int32((page - 1) * pageSize)

		notifs, err := svc.List(ctx, user.ID, int32(pageSize), offset)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to list notifications")
		}

		resp := &listNotificationsOutput{}
		resp.Body.Notifications = make([]notificationItem, 0, len(notifs))
		for _, n := range notifs {
			resp.Body.Notifications = append(resp.Body.Notifications, toNotificationItem(n))
		}
		return resp, nil
	})

	// GET /me/notifications/unread-count - count only
	huma.Register(api, huma.Operation{
		OperationID: "unread-notification-count",
		Method:      "GET",
		Path:        "/me/notifications/unread-count",
		Summary:     "Unread notification count",
		Description: "Returns the count of unread notifications for the current user.",
		Tags:        []string{"Notifications"},
	}, func(ctx context.Context, input *struct{}) (*unreadCountOutput, error) {
		user, err := requireUser(ctx)
		if err != nil {
			return nil, err
		}

		count, err := svc.UnreadCount(ctx, user.ID)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to count notifications")
		}

		resp := &unreadCountOutput{}
		resp.Body.Count = count
		return resp, nil
	})

	// POST /me/notifications/{id}/read - mark one read
	huma.Register(api, huma.Operation{
		OperationID: "mark-notification-read",
		Method:      "POST",
		Path:        "/me/notifications/{id}/read",
		Summary:     "Mark notification as read",
		Description: "Marks a single notification as read.",
		Tags:        []string{"Notifications"},
	}, func(ctx context.Context, input *struct {
		ID string `path:"id" doc:"Notification UUID"`
	}) (*notificationActionOutput, error) {
		user, err := requireUser(ctx)
		if err != nil {
			return nil, err
		}

		nID, err := parseNotificationID(input.ID)
		if err != nil {
			return nil, err
		}

		if err := svc.MarkRead(ctx, nID, user.ID); err != nil {
			return nil, huma.Error500InternalServerError("failed to mark notification as read")
		}

		return &notificationActionOutput{Body: struct {
			Status string `json:"status"`
		}{Status: "ok"}}, nil
	})

	// POST /me/notifications/read-all - mark all read
	huma.Register(api, huma.Operation{
		OperationID: "mark-all-notifications-read",
		Method:      "POST",
		Path:        "/me/notifications/read-all",
		Summary:     "Mark all notifications as read",
		Description: "Marks all of the current user's notifications as read.",
		Tags:        []string{"Notifications"},
	}, func(ctx context.Context, input *struct{}) (*notificationActionOutput, error) {
		user, err := requireUser(ctx)
		if err != nil {
			return nil, err
		}

		if err := svc.MarkAllRead(ctx, user.ID); err != nil {
			return nil, huma.Error500InternalServerError("failed to mark all notifications as read")
		}

		return &notificationActionOutput{Body: struct {
			Status string `json:"status"`
		}{Status: "ok"}}, nil
	})

	// DELETE /me/notifications/{id} - delete one
	huma.Register(api, huma.Operation{
		OperationID: "delete-notification",
		Method:      "DELETE",
		Path:        "/me/notifications/{id}",
		Summary:     "Delete a notification",
		Description: "Deletes a single notification.",
		Tags:        []string{"Notifications"},
	}, func(ctx context.Context, input *struct {
		ID string `path:"id" doc:"Notification UUID"`
	}) (*notificationActionOutput, error) {
		user, err := requireUser(ctx)
		if err != nil {
			return nil, err
		}

		nID, err := parseNotificationID(input.ID)
		if err != nil {
			return nil, err
		}

		if err := svc.Delete(ctx, nID, user.ID); err != nil {
			return nil, huma.Error500InternalServerError("failed to delete notification")
		}

		return &notificationActionOutput{Body: struct {
			Status string `json:"status"`
		}{Status: "ok"}}, nil
	})
}

type listNotificationsOutput struct {
	Body struct {
		Notifications []notificationItem `json:"notifications"`
	}
}

type unreadCountOutput struct {
	Body struct {
		Count int64 `json:"count"`
	}
}

type notificationActionOutput struct {
	Body struct {
		Status string `json:"status"`
	}
}

func parseNotificationID(raw string) (pgtype.UUID, error) {
	var id pgtype.UUID
	if err := id.Scan(raw); err != nil {
		return pgtype.UUID{}, huma.Error422UnprocessableEntity("invalid notification id")
	}
	return id, nil
}
