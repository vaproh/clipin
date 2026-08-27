//go:build integration

package integration

import (
	"context"
	"testing"

	sqlc "clipin/apps/api/internal/db/sqlc"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestCreateAndGetSocialAccount(t *testing.T) {
	cleanupAll(t)

	seedUser(t, "sa_user1", "sa1@example.com", "clipper")

	account, err := testDB.Queries.CreateSocialAccount(context.Background(), sqlc.CreateSocialAccountParams{
		UserID:         "sa_user1",
		Platform:       "youtube",
		PlatformUserID: "chan_abc123",
		PlatformUsername: pgtype.Text{String: "TestChannel", Valid: true},
		AccessToken:    pgtype.Text{String: "tok_abc", Valid: true},
	})
	if err != nil {
		t.Fatalf("CreateSocialAccount: %v", err)
	}

	accounts, err := testDB.Queries.ListSocialAccountsByUserID(context.Background(), "sa_user1")
	if err != nil {
		t.Fatalf("ListSocialAccountsByUserID: %v", err)
	}
	if len(accounts) != 1 {
		t.Fatalf("expected 1 account, got %d", len(accounts))
	}
	assertEqual(t, "Platform", accounts[0].Platform, "youtube")
	assertEqual(t, "PlatformUserID", accounts[0].PlatformUserID, "chan_abc123")
	_ = account
}

func TestDeleteSocialAccount(t *testing.T) {
	cleanupAll(t)

	seedUser(t, "sa_del1", "sadel@example.com", "clipper")

	account, err := testDB.Queries.CreateSocialAccount(context.Background(), sqlc.CreateSocialAccountParams{
		UserID:         "sa_del1",
		Platform:       "instagram",
		PlatformUserID: "ig_user123",
	})
	if err != nil {
		t.Fatalf("CreateSocialAccount: %v", err)
	}

	err = testDB.Queries.DeleteSocialAccount(context.Background(), sqlc.DeleteSocialAccountParams{
		ID:     account.ID,
		UserID: "sa_del1",
	})
	if err != nil {
		t.Fatalf("DeleteSocialAccount: %v", err)
	}

	accounts, err := testDB.Queries.ListSocialAccountsByUserID(context.Background(), "sa_del1")
	if err != nil {
		t.Fatalf("ListSocialAccountsByUserID after delete: %v", err)
	}
	if len(accounts) != 0 {
		t.Errorf("expected 0 accounts after delete, got %d", len(accounts))
	}
}

func TestUpsertSocialAccountInsert(t *testing.T) {
	cleanupAll(t)

	seedUser(t, "sa_upsert1", "upsert@example.com", "clipper")

	account, err := testDB.Queries.UpsertSocialAccount(context.Background(), sqlc.UpsertSocialAccountParams{
		UserID:         "sa_upsert1",
		Platform:       "youtube",
		PlatformUserID: "chan_new",
		PlatformUsername: pgtype.Text{String: "NewChannel", Valid: true},
		AccessToken:    pgtype.Text{String: "tok_new", Valid: true},
	})
	if err != nil {
		t.Fatalf("UpsertSocialAccount insert: %v", err)
	}

	assertEqual(t, "Platform", account.Platform, "youtube")
	assertEqual(t, "PlatformUserID", account.PlatformUserID, "chan_new")
	assertEqual(t, "PlatformUsername", account.PlatformUsername.String, "NewChannel")
}

func TestUpsertSocialAccountUpdate(t *testing.T) {
	cleanupAll(t)

	seedUser(t, "sa_upsert2", "upsert2@example.com", "clipper")

	// Insert
	_, err := testDB.Queries.UpsertSocialAccount(context.Background(), sqlc.UpsertSocialAccountParams{
		UserID:         "sa_upsert2",
		Platform:       "youtube",
		PlatformUserID: "chan_upd",
		PlatformUsername: pgtype.Text{String: "OldName", Valid: true},
		AccessToken:    pgtype.Text{String: "old_tok", Valid: true},
	})
	if err != nil {
		t.Fatalf("UpsertSocialAccount initial: %v", err)
	}

	// Update same platform+platform_user_id
	updated, err := testDB.Queries.UpsertSocialAccount(context.Background(), sqlc.UpsertSocialAccountParams{
		UserID:         "sa_upsert2",
		Platform:       "youtube",
		PlatformUserID: "chan_upd",
		PlatformUsername: pgtype.Text{String: "NewName", Valid: true},
		AccessToken:    pgtype.Text{String: "new_tok", Valid: true},
	})
	if err != nil {
		t.Fatalf("UpsertSocialAccount update: %v", err)
	}

	assertEqual(t, "PlatformUsername after upsert", updated.PlatformUsername.String, "NewName")
	assertEqual(t, "AccessToken after upsert", updated.AccessToken.String, "new_tok")
}

func TestDuplicatePlatformUserIDRejected(t *testing.T) {
	cleanupAll(t)

	seedUser(t, "sa_dup1", "dup1@sa.com", "clipper")
	seedUser(t, "sa_dup2", "dup2@sa.com", "clipper")

	_, err := testDB.Queries.CreateSocialAccount(context.Background(), sqlc.CreateSocialAccountParams{
		UserID:         "sa_dup1",
		Platform:       "youtube",
		PlatformUserID: "chan_dup",
	})
	if err != nil {
		t.Fatalf("first CreateSocialAccount: %v", err)
	}

	_, err = testDB.Queries.CreateSocialAccount(context.Background(), sqlc.CreateSocialAccountParams{
		UserID:         "sa_dup2",
		Platform:       "youtube",
		PlatformUserID: "chan_dup",
	})
	if err == nil {
		t.Error("expected error for duplicate platform_user_id, got nil")
	}
}

func TestListSocialAccountsMultiple(t *testing.T) {
	cleanupAll(t)

	seedUser(t, "sa_multi1", "multi1@sa.com", "clipper")

	for _, p := range []string{"youtube", "instagram", "tiktok"} {
		_, err := testDB.Queries.CreateSocialAccount(context.Background(), sqlc.CreateSocialAccountParams{
			UserID:         "sa_multi1",
			Platform:       p,
			PlatformUserID: p + "_id123",
		})
		if err != nil {
			t.Fatalf("CreateSocialAccount for %s: %v", p, err)
		}
	}

	accounts, err := testDB.Queries.ListSocialAccountsByUserID(context.Background(), "sa_multi1")
	if err != nil {
		t.Fatalf("ListSocialAccountsByUserID: %v", err)
	}
	if len(accounts) != 3 {
		t.Errorf("expected 3 accounts, got %d", len(accounts))
	}
}
