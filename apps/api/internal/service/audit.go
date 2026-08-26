package service

import (
	"context"
	"encoding/json"
	"fmt"

	sqlc "clipin/apps/api/internal/db/sqlc"

	"github.com/jackc/pgx/v5/pgtype"
)

// AuditStore is the persistence interface the audit service requires.
type AuditStore interface {
	CreateAuditLog(ctx context.Context, arg sqlc.CreateAuditLogParams) (sqlc.AuditLog, error)
	ListAuditLogs(ctx context.Context, arg sqlc.ListAuditLogsParams) ([]sqlc.AuditLog, error)
	ListAuditLogsByResource(ctx context.Context, arg sqlc.ListAuditLogsByResourceParams) ([]sqlc.AuditLog, error)
	ListAuditLogsByActor(ctx context.Context, arg sqlc.ListAuditLogsByActorParams) ([]sqlc.AuditLog, error)
}

// AuditService implements append-only audit logging.
type AuditService struct {
	store AuditStore
}

// NewAuditService creates a new AuditService.
func NewAuditService(store AuditStore) *AuditService {
	return &AuditService{store: store}
}

// Log records an audit event. Details is serialized to JSON if non-nil.
func (s *AuditService) Log(ctx context.Context, actorID, action, resourceType, resourceID string, details map[string]any, ipAddress string) error {
	var detailsJSON []byte
	if details != nil {
		var err error
		detailsJSON, err = json.Marshal(details)
		if err != nil {
			return fmt.Errorf("marshal details: %w", err)
		}
	}

	var ip pgtype.Text
	if ipAddress != "" {
		ip = pgtype.Text{Valid: true, String: ipAddress}
	}

	_, err := s.store.CreateAuditLog(ctx, sqlc.CreateAuditLogParams{
		ActorID:      actorID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Details:      detailsJSON,
		IpAddress:    ip,
	})
	if err != nil {
		return fmt.Errorf("create audit log: %w", err)
	}
	return nil
}

// ListLogs returns the most recent audit logs, paginated.
func (s *AuditService) ListLogs(ctx context.Context, limit, offset int32) ([]sqlc.AuditLog, error) {
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.store.ListAuditLogs(ctx, sqlc.ListAuditLogsParams{
		Limit:  limit,
		Offset: offset,
	})
}

// ListLogsForResource returns all audit logs for a given resource.
func (s *AuditService) ListLogsForResource(ctx context.Context, resourceType, resourceID string) ([]sqlc.AuditLog, error) {
	return s.store.ListAuditLogsByResource(ctx, sqlc.ListAuditLogsByResourceParams{
		ResourceType: resourceType,
		ResourceID:   resourceID,
	})
}

// ListLogsByActor returns audit logs for a specific actor, paginated.
func (s *AuditService) ListLogsByActor(ctx context.Context, actorID string, limit, offset int32) ([]sqlc.AuditLog, error) {
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.store.ListAuditLogsByActor(ctx, sqlc.ListAuditLogsByActorParams{
		ActorID: actorID,
		Limit:   limit,
		Offset:  offset,
	})
}
