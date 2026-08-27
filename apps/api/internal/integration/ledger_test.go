//go:build integration

package integration

import (
	"context"
	"errors"
	"testing"

	sqlc "clipin/apps/api/internal/db/sqlc"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestCreateLedgerEntry(t *testing.T) {
	cleanupAll(t)

	ownerID := "led_owner1"
	clipperID := "led_clipper1"
	seedUser(t, ownerID, "ledowner1@led.com", "owner")
	seedUser(t, clipperID, "ledclipper1@led.com", "clipper")
	campaign := seedCampaign(t, ownerID, "Led Campaign", "youtube", 500000, 5000)

	entry := seedLedgerEntry(t, "idem_001", "earning", campaign.ID, 10000, clipperID)

	fetched, err := testDB.Queries.GetLedgerEntryByIdempotencyKey(context.Background(), "idem_001")
	if err != nil {
		t.Fatalf("GetLedgerEntryByIdempotencyKey: %v", err)
	}
	assertEqual(t, "IdempotencyKey", fetched.IdempotencyKey, "idem_001")
	assertEqual(t, "EntryType", fetched.EntryType, "earning")
	assertIntEqual(t, "Amount", fetched.Amount, 10000)
	_ = entry
}

func TestCreateLedgerEntryDuplicateIdempotencyKey(t *testing.T) {
	cleanupAll(t)

	ownerID := "led_dup_owner"
	clipperID := "led_dup_clipper"
	seedUser(t, ownerID, "leddupowner@led.com", "owner")
	seedUser(t, clipperID, "leddupclipper@led.com", "clipper")
	campaign := seedCampaign(t, ownerID, "Led Dup", "youtube", 500000, 5000)

	// First insert
	_, err := testDB.Queries.CreateLedgerEntry(context.Background(), sqlc.CreateLedgerEntryParams{
		IdempotencyKey: "idem_dup",
		EntryType:      "earning",
		CampaignID:     campaign.ID,
		Amount:         5000,
	})
	if err != nil {
		t.Fatalf("first CreateLedgerEntry: %v", err)
	}

	// Second insert with same idempotency_key - should DO NOTHING and return no rows
	_, err = testDB.Queries.CreateLedgerEntry(context.Background(), sqlc.CreateLedgerEntryParams{
		IdempotencyKey: "idem_dup",
		EntryType:      "earning",
		CampaignID:     campaign.ID,
		Amount:         99999, // Different amount
	})
	if err == nil {
		t.Fatal("expected pgx.ErrNoRows for duplicate idempotency key, got nil")
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("expected pgx.ErrNoRows, got %v", err)
	}

	// Verify original amount unchanged
	fetched, err := testDB.Queries.GetLedgerEntryByIdempotencyKey(context.Background(), "idem_dup")
	if err != nil {
		t.Fatalf("GetLedgerEntryByIdempotencyKey: %v", err)
	}
	assertIntEqual(t, "Amount after duplicate", fetched.Amount, 5000)
}

func TestSumEarningsByClipper(t *testing.T) {
	cleanupAll(t)

	ownerID := "led_sum_owner"
	clipperID := "led_sum_clipper"
	seedUser(t, ownerID, "ledsumowner@led.com", "owner")
	seedUser(t, clipperID, "ledsumclipper@led.com", "clipper")
	campaign := seedCampaign(t, ownerID, "Led Sum", "youtube", 500000, 5000)

	seedLedgerEntry(t, "idem_sum1", "earning", campaign.ID, 10000, clipperID)
	seedLedgerEntry(t, "idem_sum2", "earning", campaign.ID, 20000, clipperID)
	seedLedgerEntry(t, "idem_sum3", "platform_fee", campaign.ID, 5000, "") // Different type

	clipperText := pgtype.Text{String: clipperID, Valid: true}
	total, err := testDB.Queries.SumEarningsByClipper(context.Background(), clipperText)
	if err != nil {
		t.Fatalf("SumEarningsByClipper: %v", err)
	}
	assertInt64Equal(t, "SumEarningsByClipper", total, 30000)
}

func TestSumEarningsByClipperForCampaign(t *testing.T) {
	cleanupAll(t)

	ownerID := "led_sumcamp_owner"
	clipperID := "led_sumcamp_clipper"
	seedUser(t, ownerID, "ledsumcampowner@led.com", "owner")
	seedUser(t, clipperID, "ledsumcampclipper@led.com", "clipper")
	c1 := seedCampaign(t, ownerID, "Campaign 1", "youtube", 500000, 5000)
	c2 := seedCampaign(t, ownerID, "Campaign 2", "youtube", 300000, 3000)

	seedLedgerEntry(t, "idem_sc1", "earning", c1.ID, 15000, clipperID)
	seedLedgerEntry(t, "idem_sc2", "earning", c2.ID, 25000, clipperID)

	clipperText := pgtype.Text{String: clipperID, Valid: true}
	total, err := testDB.Queries.SumEarningsByClipperForCampaign(context.Background(), sqlc.SumEarningsByClipperForCampaignParams{
		ClipperID:  clipperText,
		CampaignID: c1.ID,
	})
	if err != nil {
		t.Fatalf("SumEarningsByClipperForCampaign: %v", err)
	}
	assertInt64Equal(t, "SumEarningsByClipperForCampaign c1", total, 15000)
}

func TestSumFeesByCampaign(t *testing.T) {
	cleanupAll(t)

	ownerID := "led_fee_owner"
	clipperID := "led_fee_clipper"
	seedUser(t, ownerID, "ledfeeowner@led.com", "owner")
	seedUser(t, clipperID, "ledfeeclipper@led.com", "clipper")
	campaign := seedCampaign(t, ownerID, "Fee Campaign", "youtube", 500000, 5000)

	seedLedgerEntry(t, "idem_fee1", "platform_fee", campaign.ID, 50000, "")
	seedLedgerEntry(t, "idem_fee2", "platform_fee", campaign.ID, 30000, "")
	seedLedgerEntry(t, "idem_fee3", "earning", campaign.ID, 10000, clipperID) // Different type

	total, err := testDB.Queries.SumFeesByCampaign(context.Background(), campaign.ID)
	if err != nil {
		t.Fatalf("SumFeesByCampaign: %v", err)
	}
	assertInt64Equal(t, "SumFeesByCampaign", total, 80000)
}

func TestSumSpendByCampaign(t *testing.T) {
	cleanupAll(t)

	ownerID := "led_spend_owner"
	clipperID := "led_spend_clipper"
	seedUser(t, ownerID, "ledspendowner@led.com", "owner")
	seedUser(t, clipperID, "ledspendclipper@led.com", "clipper")
	campaign := seedCampaign(t, ownerID, "Spend Campaign", "youtube", 500000, 5000)

	seedLedgerEntry(t, "idem_sp1", "earning", campaign.ID, 20000, clipperID)
	seedLedgerEntry(t, "idem_sp2", "earning", campaign.ID, 30000, clipperID)
	seedLedgerEntry(t, "idem_sp3", "platform_fee", campaign.ID, 10000, "") // Different type

	total, err := testDB.Queries.SumSpendByCampaign(context.Background(), campaign.ID)
	if err != nil {
		t.Fatalf("SumSpendByCampaign: %v", err)
	}
	assertInt64Equal(t, "SumSpendByCampaign", total, 50000)
}

func TestListLedgerEntriesByCampaign(t *testing.T) {
	cleanupAll(t)

	ownerID := "led_list_owner"
	clipperID := "led_list_clipper"
	seedUser(t, ownerID, "ledlistowner@led.com", "owner")
	seedUser(t, clipperID, "ledlistclipper@led.com", "clipper")
	campaign := seedCampaign(t, ownerID, "List Campaign", "youtube", 500000, 5000)

	seedLedgerEntry(t, "idem_lst1", "earning", campaign.ID, 10000, clipperID)
	seedLedgerEntry(t, "idem_lst2", "platform_fee", campaign.ID, 5000, "")
	seedLedgerEntry(t, "idem_lst3", "earning", campaign.ID, 20000, clipperID)

	entries, err := testDB.Queries.ListLedgerEntriesByCampaign(context.Background(), campaign.ID)
	if err != nil {
		t.Fatalf("ListLedgerEntriesByCampaign: %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(entries))
	}
}

func TestSumEarningsByClipperNoEntries(t *testing.T) {
	cleanupAll(t)

	clipperText := pgtype.Text{String: "nonexistent_clipper", Valid: true}
	total, err := testDB.Queries.SumEarningsByClipper(context.Background(), clipperText)
	if err != nil {
		t.Fatalf("SumEarningsByClipper: %v", err)
	}
	assertInt64Equal(t, "SumEarningsByClipper empty", total, 0)
}
