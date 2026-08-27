//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	sqlc "clipin/apps/api/internal/db/sqlc"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestCreateAndGetCampaign(t *testing.T) {
	cleanupAll(t)

	seedUser(t, "camp_owner1", "owner1@camp.com", "owner")
	campaign := seedCampaign(t, "camp_owner1", "Test Campaign", "youtube", 1000000, 5000)

	fetched, err := testDB.Queries.GetCampaignByID(context.Background(), campaign.ID)
	if err != nil {
		t.Fatalf("GetCampaignByID: %v", err)
	}
	assertEqual(t, "Title", fetched.Title, "Test Campaign")
	assertEqual(t, "Platform", fetched.Platform, "youtube")
	assertEqual(t, "Status", fetched.Status, "active")
	assertIntEqual(t, "CpmRate", fetched.CpmRate, 5000)
	assertIntEqual(t, "TotalBudget", fetched.TotalBudget, 1000000)
	assertIntEqual(t, "RemainingBudget", fetched.RemainingBudget, 1000000)
}

func TestListActiveCampaigns(t *testing.T) {
	cleanupAll(t)

	ownerID := "camp_active_owner"
	seedUser(t, ownerID, "activeowner@camp.com", "owner")

	// Create active campaign
	seedCampaign(t, ownerID, "Active Camp", "youtube", 500000, 3000)

	// Create draft campaign (should not appear)
	_, err := testDB.Queries.CreateCampaign(context.Background(), sqlc.CreateCampaignParams{
		ID:              newUUID(),
		OwnerID:         ownerID,
		Title:           "Draft Camp",
		Platform:        "youtube",
		Status:          "draft",
		CpmRate:         2000,
		TotalBudget:     200000,
		RemainingBudget: 200000,
		PlatformFee:     20000,
	})
	if err != nil {
		t.Fatalf("CreateCampaign draft: %v", err)
	}

	campaigns, err := testDB.Queries.ListActiveCampaigns(context.Background())
	if err != nil {
		t.Fatalf("ListActiveCampaigns: %v", err)
	}
	if len(campaigns) != 1 {
		t.Fatalf("expected 1 active campaign, got %d", len(campaigns))
	}
	assertEqual(t, "Title", campaigns[0].Title, "Active Camp")
}

func TestListActiveCampaignsExpiredExcluded(t *testing.T) {
	cleanupAll(t)

	ownerID := "camp_expire_owner"
	seedUser(t, ownerID, "expireowner@camp.com", "owner")

	// Campaign with ends_at in the past
	_, err := testDB.Queries.CreateCampaign(context.Background(), sqlc.CreateCampaignParams{
		ID:              newUUID(),
		OwnerID:         ownerID,
		Title:           "Expired Camp",
		Platform:        "youtube",
		Status:          "active",
		CpmRate:         3000,
		TotalBudget:     100000,
		RemainingBudget: 100000,
		PlatformFee:     10000,
		EndsAt:          pgtype.Timestamptz{Time: time.Now().Add(-24 * time.Hour), Valid: true},
	})
	if err != nil {
		t.Fatalf("CreateCampaign expired: %v", err)
	}

	campaigns, err := testDB.Queries.ListActiveCampaigns(context.Background())
	if err != nil {
		t.Fatalf("ListActiveCampaigns: %v", err)
	}
	if len(campaigns) != 0 {
		t.Errorf("expected 0 active campaigns (expired), got %d", len(campaigns))
	}
}

func TestListCampaignsFiltered(t *testing.T) {
	cleanupAll(t)

	ownerID := "camp_filt_owner"
	seedUser(t, ownerID, "filtowner@camp.com", "owner")

	seedCampaign(t, ownerID, "YouTube Campaign", "youtube", 500000, 5000)
	seedCampaign(t, ownerID, "Instagram Campaign", "instagram", 300000, 3000)

	// Filter by platform
	campaigns, err := testDB.Queries.ListCampaignsFiltered(context.Background(), sqlc.ListCampaignsFilteredParams{
		Column1: "youtube",
		Limit:   10,
		Offset:  0,
	})
	if err != nil {
		t.Fatalf("ListCampaignsFiltered: %v", err)
	}
	if len(campaigns) != 1 {
		t.Fatalf("expected 1 youtube campaign, got %d", len(campaigns))
	}
	assertEqual(t, "Platform", campaigns[0].Platform, "youtube")
}

func TestListCampaignsFilteredBySearch(t *testing.T) {
	cleanupAll(t)

	ownerID := "camp_search_owner"
	seedUser(t, ownerID, "searchowner@camp.com", "owner")

	seedCampaign(t, ownerID, "Summer Sale", "youtube", 500000, 5000)
	seedCampaign(t, ownerID, "Winter Promo", "youtube", 300000, 3000)

	campaigns, err := testDB.Queries.ListCampaignsFiltered(context.Background(), sqlc.ListCampaignsFilteredParams{
		Column1: "",
		Column4: "Summer",
		Limit:   10,
		Offset:  0,
	})
	if err != nil {
		t.Fatalf("ListCampaignsFiltered search: %v", err)
	}
	if len(campaigns) != 1 {
		t.Fatalf("expected 1 campaign matching 'Summer', got %d", len(campaigns))
	}
	assertEqual(t, "Title", campaigns[0].Title, "Summer Sale")
}

func TestUpdateCampaign(t *testing.T) {
	cleanupAll(t)

	ownerID := "camp_upd_owner"
	seedUser(t, ownerID, "updowner@camp.com", "owner")
	campaign := seedCampaign(t, ownerID, "Original Title", "youtube", 500000, 5000)

	updated, err := testDB.Queries.UpdateCampaign(context.Background(), sqlc.UpdateCampaignParams{
		ID:              campaign.ID,
		Title:           "Updated Title",
		Description:     pgtype.Text{String: "New description", Valid: true},
		MaxClipsPerCampaign: pgtype.Int4{Int32: 50, Valid: true},
		MaxClipsPerClipper:  pgtype.Int4{Int32: 5, Valid: true},
		MinViewsPerClip:     pgtype.Int4{Int32: 2000, Valid: true},
		AutoApproveHours:    pgtype.Int4{Int32: 24, Valid: true},
	})
	if err != nil {
		t.Fatalf("UpdateCampaign: %v", err)
	}
	assertEqual(t, "Title", updated.Title, "Updated Title")
	assertIntEqual(t, "MaxClipsPerCampaign", updated.MaxClipsPerCampaign.Int32, 50)
}

func TestUpdateCampaignStatus(t *testing.T) {
	cleanupAll(t)

	ownerID := "camp_stat_owner"
	seedUser(t, ownerID, "statowner@camp.com", "owner")
	campaign := seedCampaign(t, ownerID, "Status Camp", "youtube", 500000, 5000)

	transitions := []string{"funded", "active", "paused", "completed"}
	for _, status := range transitions {
		updated, err := testDB.Queries.UpdateCampaignStatus(context.Background(), sqlc.UpdateCampaignStatusParams{
			ID:     campaign.ID,
			Status: status,
		})
		if err != nil {
			t.Fatalf("UpdateCampaignStatus to %s: %v", status, err)
		}
		assertEqual(t, "Status after transition to "+status, updated.Status, status)
	}
}

func TestDeductCampaignBudget(t *testing.T) {
	cleanupAll(t)

	ownerID := "camp_deduct_owner"
	seedUser(t, ownerID, "deductowner@camp.com", "owner")
	campaign := seedCampaign(t, ownerID, "Deduct Camp", "youtube", 1000000, 5000)

	deducted, err := testDB.Queries.DeductCampaignBudget(context.Background(), sqlc.DeductCampaignBudgetParams{
		ID:              campaign.ID,
		RemainingBudget: 100000,
	})
	if err != nil {
		t.Fatalf("DeductCampaignBudget: %v", err)
	}
	assertIntEqual(t, "RemainingBudget after deduct", deducted.RemainingBudget, 900000)
}

func TestDeductCampaignBudgetInsufficient(t *testing.T) {
	cleanupAll(t)

	ownerID := "camp_dedins_owner"
	seedUser(t, ownerID, "dedinsowner@camp.com", "owner")
	campaign := seedCampaign(t, ownerID, "Deduct Insufficient", "youtube", 100000, 5000)

	_, err := testDB.Queries.DeductCampaignBudget(context.Background(), sqlc.DeductCampaignBudgetParams{
		ID:              campaign.ID,
		RemainingBudget: 200000, // More than available
	})
	if err == nil {
		t.Error("expected error for insufficient budget, got nil")
	}
}

func TestDeductCampaignBudgetIdempotent(t *testing.T) {
	cleanupAll(t)

	ownerID := "camp_dedid_owner"
	seedUser(t, ownerID, "dedidowner@camp.com", "owner")
	campaign := seedCampaign(t, ownerID, "Deduct Idempotent", "youtube", 500000, 5000)

	_, err := testDB.Queries.DeductCampaignBudget(context.Background(), sqlc.DeductCampaignBudgetParams{
		ID:              campaign.ID,
		RemainingBudget: 100000,
	})
	if err != nil {
		t.Fatalf("first deduct: %v", err)
	}

	deducted, err := testDB.Queries.DeductCampaignBudget(context.Background(), sqlc.DeductCampaignBudgetParams{
		ID:              campaign.ID,
		RemainingBudget: 100000,
	})
	if err != nil {
		t.Fatalf("second deduct: %v", err)
	}
	assertIntEqual(t, "RemainingBudget after double deduct", deducted.RemainingBudget, 300000)
}

func TestListCampaignsByOwner(t *testing.T) {
	cleanupAll(t)

	ownerID := "camp_ownlist_owner"
	seedUser(t, ownerID, "ownlistowner@camp.com", "owner")

	seedCampaign(t, ownerID, "Camp 1", "youtube", 100000, 5000)
	seedCampaign(t, ownerID, "Camp 2", "instagram", 200000, 3000)

	camps, err := testDB.Queries.ListCampaignsByOwner(context.Background(), ownerID)
	if err != nil {
		t.Fatalf("ListCampaignsByOwner: %v", err)
	}
	if len(camps) != 2 {
		t.Errorf("expected 2 campaigns for owner, got %d", len(camps))
	}
}

func TestUpdateCampaignBudget(t *testing.T) {
	cleanupAll(t)

	ownerID := "camp_budupd_owner"
	seedUser(t, ownerID, "budupdowner@camp.com", "owner")
	campaign := seedCampaign(t, ownerID, "Budget Update", "youtube", 500000, 5000)

	updated, err := testDB.Queries.UpdateCampaignBudget(context.Background(), sqlc.UpdateCampaignBudgetParams{
		ID:              campaign.ID,
		RemainingBudget: 250000,
	})
	if err != nil {
		t.Fatalf("UpdateCampaignBudget: %v", err)
	}
	assertIntEqual(t, "RemainingBudget", updated.RemainingBudget, 250000)
}
