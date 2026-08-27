//go:build integration

package integration

import (
	"context"
	"testing"

	sqlc "clipin/apps/api/internal/db/sqlc"
)

func TestGetLeaderboardByEarnings(t *testing.T) {
	cleanupAll(t)

	ownerID := "lb_owner"
	seedUser(t, ownerID, "lbowner@lb.com", "owner")
	campaign := seedCampaign(t, ownerID, "LB Campaign", "youtube", 1000000, 5000)

	// Create two clippers with earnings
	for i, cid := range []string{"lb_clip1", "lb_clip2"} {
		seedUser(t, cid, cid+"@lb.com", "clipper")
		sub := seedSubmission(t, campaign.ID, cid, "https://yt.com/lb"+string(rune('a'+i)), "youtube")
		_, _ = testDB.Queries.ApproveSubmission(context.Background(), sub.ID)
		seedLedgerEntry(t, "lb_idem_"+cid, "earning", campaign.ID, int32((i+1)*50000), cid)
	}

	leaderboard, err := testDB.Queries.GetLeaderboardByEarnings(context.Background(), sqlc.GetLeaderboardByEarningsParams{
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("GetLeaderboardByEarnings: %v", err)
	}
	if len(leaderboard) != 2 {
		t.Fatalf("expected 2 entries in leaderboard, got %d", len(leaderboard))
	}
	// lb_clip2 should be first (100000 > 50000)
	assertEqual(t, "Top earner", leaderboard[0].ID, "lb_clip2")
	assertIntEqual(t, "Top earnings", leaderboard[0].TotalEarnings, 100000)
}

func TestGetLeaderboardBySubmissions(t *testing.T) {
	cleanupAll(t)

	ownerID := "lb_sub_owner"
	seedUser(t, ownerID, "lbsubowner@lb.com", "owner")
	campaign := seedCampaign(t, ownerID, "LB Sub Campaign", "youtube", 1000000, 5000)

	// Clipper with 3 submissions
	seedUser(t, "lb_sub_multi", "lbsubmulti@lb.com", "clipper")
	for i := 0; i < 3; i++ {
		seedSubmission(t, campaign.ID, "lb_sub_multi", "https://yt.com/lbsm"+string(rune('a'+i)), "youtube")
	}

	// Clipper with 1 submission
	seedUser(t, "lb_sub_single", "lbsubsingle@lb.com", "clipper")
	seedSubmission(t, campaign.ID, "lb_sub_single", "https://yt.com lbss", "youtube")

	leaderboard, err := testDB.Queries.GetLeaderboardBySubmissions(context.Background(), sqlc.GetLeaderboardBySubmissionsParams{
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("GetLeaderboardBySubmissions: %v", err)
	}
	if len(leaderboard) != 2 {
		t.Fatalf("expected 2 entries in leaderboard, got %d", len(leaderboard))
	}
	assertEqual(t, "Top by submissions", leaderboard[0].ID, "lb_sub_multi")
	assertIntEqual(t, "Top submission count", leaderboard[0].TotalSubmissions, 3)
}

func TestLeaderboardPagination(t *testing.T) {
	cleanupAll(t)

	ownerID := "lb_page_owner"
	seedUser(t, ownerID, "lbpageowner@lb.com", "owner")
	campaign := seedCampaign(t, ownerID, "LB Page Campaign", "youtube", 1000000, 5000)

	// Create 5 clippers with submissions
	for i := 0; i < 5; i++ {
		cid := "lb_page_clip" + string(rune('0'+i))
		seedUser(t, cid, cid+"@lb.com", "clipper")
		seedSubmission(t, campaign.ID, cid, "https://yt.com/lbp"+string(rune('a'+i)), "youtube")
	}

	// Page 1
	page1, err := testDB.Queries.GetLeaderboardBySubmissions(context.Background(), sqlc.GetLeaderboardBySubmissionsParams{
		Limit:  2,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("GetLeaderboardBySubmissions page 1: %v", err)
	}
	if len(page1) != 2 {
		t.Errorf("expected 2 entries on page 1, got %d", len(page1))
	}

	// Page 2
	page2, err := testDB.Queries.GetLeaderboardBySubmissions(context.Background(), sqlc.GetLeaderboardBySubmissionsParams{
		Limit:  2,
		Offset: 2,
	})
	if err != nil {
		t.Fatalf("GetLeaderboardBySubmissions page 2: %v", err)
	}
	if len(page2) != 2 {
		t.Errorf("expected 2 entries on page 2, got %d", len(page2))
	}
}

func TestLeaderboardEmpty(t *testing.T) {
	cleanupAll(t)

	leaderboard, err := testDB.Queries.GetLeaderboardByEarnings(context.Background(), sqlc.GetLeaderboardByEarningsParams{
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("GetLeaderboardByEarnings empty: %v", err)
	}
	if len(leaderboard) != 0 {
		t.Errorf("expected empty leaderboard, got %d entries", len(leaderboard))
	}
}

func TestLeaderboardWithNoEarnings(t *testing.T) {
	cleanupAll(t)

	// Create a clipper with submissions but no ledger entries
	ownerID := "lb_noearn_owner"
	seedUser(t, ownerID, "lbnoearnowner@lb.com", "owner")
	campaign := seedCampaign(t, ownerID, "LB No Earn", "youtube", 500000, 5000)

	seedUser(t, "lb_noearn_clip", "lbnoearnclip@lb.com", "clipper")
	seedSubmission(t, campaign.ID, "lb_noearn_clip", "https://yt.com/lbne", "youtube")

	leaderboard, err := testDB.Queries.GetLeaderboardByEarnings(context.Background(), sqlc.GetLeaderboardByEarningsParams{
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("GetLeaderboardByEarnings no earnings: %v", err)
	}
	// Clipper has no ledger entries, so the JOIN should produce 0 rows
	if len(leaderboard) != 0 {
		t.Errorf("expected empty leaderboard (no earnings), got %d entries", len(leaderboard))
	}
}
