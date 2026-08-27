//go:build integration

package integration

import (
	"context"
	"testing"
)

func TestMigrationsApplied(t *testing.T) {
	var version int
	err := testDB.Pool.QueryRow(context.Background(),
		"SELECT COALESCE(MAX(CAST(version AS integer)), 0) FROM schema_migrations",
	).Scan(&version)
	if err != nil {
		t.Fatalf("query schema_migrations: %v", err)
	}
	if version < 9 {
		t.Errorf("expected migration version >= 9, got %d", version)
	}
}

func TestAllExpectedTablesExist(t *testing.T) {
	expectedTables := []string{
		"users", "social_accounts", "campaigns", "submissions",
		"metric_snapshots", "ledger_entries", "payout_requests",
		"audit_logs", "fraud_flags", "notifications", "campaign_templates",
	}
	for _, table := range expectedTables {
		var exists bool
		err := testDB.Pool.QueryRow(context.Background(),
			"SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = $1)", table,
		).Scan(&exists)
		if err != nil {
			t.Fatalf("check table %s: %v", table, err)
		}
		if !exists {
			t.Errorf("table %q does not exist", table)
		}
	}
}

func TestAllExpectedIndexesExist(t *testing.T) {
	expectedIndexes := []string{
		"idx_users_email",
		"idx_users_role",
		"idx_campaigns_status",
		"idx_campaigns_platform",
		"idx_submissions_status",
		"idx_ledger_entries_idempotency",
		"idx_payout_requests_status",
		"idx_notifications_user",
	}
	for _, idx := range expectedIndexes {
		var exists bool
		err := testDB.Pool.QueryRow(context.Background(),
			"SELECT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = $1)", idx,
		).Scan(&exists)
		if err != nil {
			t.Fatalf("check index %s: %v", idx, err)
		}
		if !exists {
			t.Errorf("index %q does not exist", idx)
		}
	}
}

func TestUsersTableColumns(t *testing.T) {
	expectedCols := map[string]bool{
		"id": false, "email": false, "display_name": false, "role": false,
		"created_at": false, "updated_at": false, "upi_id": false,
		"avatar_url": false, "bio": false,
	}
	rows, err := testDB.Pool.Query(context.Background(),
		"SELECT column_name FROM information_schema.columns WHERE table_name = 'users'")
	if err != nil {
		t.Fatalf("query columns: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var col string
		if err := rows.Scan(&col); err != nil {
			t.Fatalf("scan column: %v", err)
		}
		if _, ok := expectedCols[col]; ok {
			expectedCols[col] = true
		}
	}
	for col, found := range expectedCols {
		if !found {
			t.Errorf("expected column users.%s not found", col)
		}
	}
}
