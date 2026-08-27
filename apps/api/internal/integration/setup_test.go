//go:build integration

package integration

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"testing"

	"clipin/apps/api/internal/db"
	sqlc "clipin/apps/api/internal/db/sqlc"

	"github.com/jackc/pgx/v5/pgtype"
)

var testDB *db.Database

func TestMain(m *testing.M) {
	ctx := context.Background()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://clipin:clipin_dev_pass@localhost:5433/clipin_dev?sslmode=disable"
	}

	var err error
	testDB, err = db.Connect(ctx, databaseURL)
	if err != nil {
		panic("failed to connect to test database: " + err.Error())
	}

	code := m.Run()
	testDB.Close()
	os.Exit(code)
}

// cleanupAll deletes data in reverse dependency order so FKs don't block.
// Skips tables that don't exist (e.g. if migrations are partial).
func cleanupAll(t *testing.T) {
	t.Helper()
	tables := []string{
		"audit_logs", "fraud_flags", "notifications",
		"ledger_entries", "payout_requests", "metric_snapshots",
		"submissions", "campaign_templates", "social_accounts",
		"campaigns", "users",
	}
	for _, table := range tables {
		_, err := testDB.Pool.Exec(context.Background(), "DELETE FROM "+table)
		if err != nil {
			// Skip tables that don't exist
			t.Logf("cleanup %s (skipped): %v", table, err)
		}
	}
}

// newUUID generates a random pgtype.UUID.
func newUUID() pgtype.UUID {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return pgtype.UUID{Bytes: b, Valid: true}
}

// seedUser inserts a user and returns it.
func seedUser(t *testing.T, id, email, role string) sqlc.User {
	t.Helper()
	user, err := testDB.Queries.CreateUser(context.Background(), sqlc.CreateUserParams{
		ID:          id,
		Email:       email,
		DisplayName: pgtype.Text{String: "Test User", Valid: true},
		Role:        role,
	})
	if err != nil {
		t.Fatalf("seedUser: %v", err)
	}
	return user
}

// seedCampaign inserts a campaign and returns it.
func seedCampaign(t *testing.T, ownerID, title, platform string, budget, cpm int32) sqlc.Campaign {
	t.Helper()
	campaign, err := testDB.Queries.CreateCampaign(context.Background(), sqlc.CreateCampaignParams{
		ID:              newUUID(),
		OwnerID:         ownerID,
		Title:           title,
		Platform:        platform,
		Status:          "active",
		CpmRate:         cpm,
		TotalBudget:     budget,
		RemainingBudget: budget,
		PlatformFee:     budget / 10,
		MaxClipsPerCampaign: pgtype.Int4{Int32: 100, Valid: true},
		MaxClipsPerClipper:  pgtype.Int4{Int32: 3, Valid: true},
		MinViewsPerClip:     pgtype.Int4{Int32: 1000, Valid: true},
		AutoApproveHours:    pgtype.Int4{Int32: 48, Valid: true},
	})
	if err != nil {
		t.Fatalf("seedCampaign: %v", err)
	}
	return campaign
}

// seedSubmission inserts a submission and returns it.
func seedSubmission(t *testing.T, campaignID pgtype.UUID, clipperID, postURL, platform string) sqlc.Submission {
	t.Helper()
	sub, err := testDB.Queries.CreateSubmission(context.Background(), sqlc.CreateSubmissionParams{
		ID:         newUUID(),
		CampaignID: campaignID,
		ClipperID:  clipperID,
		PostUrl:    postURL,
		Platform:   platform,
	})
	if err != nil {
		t.Fatalf("seedSubmission: %v", err)
	}
	return sub
}

// seedLedgerEntry inserts a ledger entry and returns it.
func seedLedgerEntry(t *testing.T, idempotencyKey, entryType string, campaignID pgtype.UUID, amount int32, clipperID string) sqlc.LedgerEntry {
	t.Helper()
	var clipper pgtype.Text
	if clipperID != "" {
		clipper = pgtype.Text{String: clipperID, Valid: true}
	}
	entry, err := testDB.Queries.CreateLedgerEntry(context.Background(), sqlc.CreateLedgerEntryParams{
		IdempotencyKey: idempotencyKey,
		EntryType:      entryType,
		CampaignID:     campaignID,
		ClipperID:      clipper,
		Amount:         amount,
		Description:    pgtype.Text{String: "test", Valid: true},
	})
	if err != nil {
		t.Fatalf("seedLedgerEntry: %v", err)
	}
	return entry
}

func assertEqual(t *testing.T, label, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got %q, want %q", label, got, want)
	}
}

func assertIntEqual(t *testing.T, label string, got, want int32) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got %d, want %d", label, got, want)
	}
}

func assertInt64Equal(t *testing.T, label string, got, want int64) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got %d, want %d", label, got, want)
	}
}

func formatUUID(u pgtype.UUID) string {
	return fmt.Sprintf("%x-%x-%x-%x-%x", u.Bytes[0:4], u.Bytes[4:6], u.Bytes[6:8], u.Bytes[8:10], u.Bytes[10:])
}
