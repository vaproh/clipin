package service_test

import (
	"context"
	"encoding/json"
	"testing"

	sqlc "clipin/apps/api/internal/db/sqlc"
	"clipin/apps/api/internal/service"

	"github.com/jackc/pgx/v5/pgtype"
)

// mockAuditStore implements AuditStore for testing.
type mockAuditStore struct {
	createLog      func(ctx context.Context, arg sqlc.CreateAuditLogParams) (sqlc.AuditLog, error)
	listLogs       func(ctx context.Context, arg sqlc.ListAuditLogsParams) ([]sqlc.AuditLog, error)
	listByResource func(ctx context.Context, arg sqlc.ListAuditLogsByResourceParams) ([]sqlc.AuditLog, error)
	listByActor    func(ctx context.Context, arg sqlc.ListAuditLogsByActorParams) ([]sqlc.AuditLog, error)
}

func (m *mockAuditStore) CreateAuditLog(ctx context.Context, arg sqlc.CreateAuditLogParams) (sqlc.AuditLog, error) {
	if m.createLog != nil {
		return m.createLog(ctx, arg)
	}
	return sqlc.AuditLog{}, nil
}

func (m *mockAuditStore) ListAuditLogs(ctx context.Context, arg sqlc.ListAuditLogsParams) ([]sqlc.AuditLog, error) {
	if m.listLogs != nil {
		return m.listLogs(ctx, arg)
	}
	return nil, nil
}

func (m *mockAuditStore) ListAuditLogsByResource(ctx context.Context, arg sqlc.ListAuditLogsByResourceParams) ([]sqlc.AuditLog, error) {
	if m.listByResource != nil {
		return m.listByResource(ctx, arg)
	}
	return nil, nil
}

func (m *mockAuditStore) ListAuditLogsByActor(ctx context.Context, arg sqlc.ListAuditLogsByActorParams) ([]sqlc.AuditLog, error) {
	if m.listByActor != nil {
		return m.listByActor(ctx, arg)
	}
	return nil, nil
}

func TestAuditLog_HappyPath(t *testing.T) {
	store := &mockAuditStore{
		createLog: func(ctx context.Context, arg sqlc.CreateAuditLogParams) (sqlc.AuditLog, error) {
			if arg.ActorID != "admin_1" {
				t.Errorf("expected actor admin_1, got %q", arg.ActorID)
			}
			if arg.Action != "campaign.create" {
				t.Errorf("expected action campaign.create, got %q", arg.Action)
			}
			if arg.ResourceType != "campaign" {
				t.Errorf("expected resource_type campaign, got %q", arg.ResourceType)
			}
			if arg.ResourceID != "camp123" {
				t.Errorf("expected resource_id camp123, got %q", arg.ResourceID)
			}
			if arg.IpAddress.Valid && arg.IpAddress.String != "127.0.0.1" {
				t.Errorf("expected ip 127.0.0.1, got %q", arg.IpAddress.String)
			}
			return sqlc.AuditLog{
				ID:           pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
				ActorID:      arg.ActorID,
				Action:       arg.Action,
				ResourceType: arg.ResourceType,
				ResourceID:   arg.ResourceID,
				Details:      arg.Details,
				IpAddress:    arg.IpAddress,
				CreatedAt:    pgtype.Timestamptz{Valid: true},
			}, nil
		},
	}
	svc := service.NewAuditService(store)

	details := map[string]any{"old_status": "draft", "new_status": "active"}
	err := svc.Log(context.Background(), "admin_1", "campaign.create", "campaign", "camp123", details, "127.0.0.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAuditLog_NilDetails(t *testing.T) {
	store := &mockAuditStore{
		createLog: func(ctx context.Context, arg sqlc.CreateAuditLogParams) (sqlc.AuditLog, error) {
			if arg.Details != nil {
				t.Errorf("expected nil details, got %v", arg.Details)
			}
			return sqlc.AuditLog{}, nil
		},
	}
	svc := service.NewAuditService(store)
	err := svc.Log(context.Background(), "admin_1", "user.view", "user", "u1", nil, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAuditListLogs_DefaultPagination(t *testing.T) {
	store := &mockAuditStore{
		listLogs: func(ctx context.Context, arg sqlc.ListAuditLogsParams) ([]sqlc.AuditLog, error) {
			if arg.Limit != 20 {
				t.Errorf("expected limit 20, got %d", arg.Limit)
			}
			if arg.Offset != 0 {
				t.Errorf("expected offset 0, got %d", arg.Offset)
			}
			return []sqlc.AuditLog{}, nil
		},
	}
	svc := service.NewAuditService(store)
	_, err := svc.ListLogs(context.Background(), 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAuditListLogs_CustomPagination(t *testing.T) {
	store := &mockAuditStore{
		listLogs: func(ctx context.Context, arg sqlc.ListAuditLogsParams) ([]sqlc.AuditLog, error) {
			if arg.Limit != 50 {
				t.Errorf("expected limit 50, got %d", arg.Limit)
			}
			if arg.Offset != 100 {
				t.Errorf("expected offset 100, got %d", arg.Offset)
			}
			return nil, nil
		},
	}
	svc := service.NewAuditService(store)
	_, err := svc.ListLogs(context.Background(), 50, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAuditListLogsForResource(t *testing.T) {
	store := &mockAuditStore{
		listByResource: func(ctx context.Context, arg sqlc.ListAuditLogsByResourceParams) ([]sqlc.AuditLog, error) {
			if arg.ResourceType != "campaign" {
				t.Errorf("expected campaign, got %q", arg.ResourceType)
			}
			if arg.ResourceID != "c1" {
				t.Errorf("expected c1, got %q", arg.ResourceID)
			}
			return []sqlc.AuditLog{}, nil
		},
	}
	svc := service.NewAuditService(store)
	_, err := svc.ListLogsForResource(context.Background(), "campaign", "c1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAuditLog_DetailsAreJSON(t *testing.T) {
	store := &mockAuditStore{
		createLog: func(ctx context.Context, arg sqlc.CreateAuditLogParams) (sqlc.AuditLog, error) {
			// Verify details are valid JSON.
			if arg.Details != nil {
				var parsed map[string]any
				if err := json.Unmarshal(arg.Details, &parsed); err != nil {
					t.Errorf("details not valid JSON: %v", err)
				}
				if parsed["key"] != "value" {
					t.Errorf("expected key=value, got %v", parsed["key"])
				}
			}
			return sqlc.AuditLog{}, nil
		},
	}
	svc := service.NewAuditService(store)
	err := svc.Log(context.Background(), "a1", "test.action", "resource", "r1", map[string]any{"key": "value"}, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
