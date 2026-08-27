//go:build integration

package integration

import (
	"context"
	"testing"

	sqlc "clipin/apps/api/internal/db/sqlc"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestGetUserPublicProfile(t *testing.T) {
	cleanupAll(t)

	_, err := testDB.Queries.CreateUser(context.Background(), sqlc.CreateUserParams{
		ID:          "profile_user1",
		Email:       "profile1@profile.com",
		DisplayName: pgtype.Text{String: "Profile User", Valid: true},
		Role:        "clipper",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	profile, err := testDB.Queries.GetUserPublicProfile(context.Background(), "profile_user1")
	if err != nil {
		t.Fatalf("GetUserPublicProfile: %v", err)
	}
	assertEqual(t, "ID", profile.ID, "profile_user1")
	assertEqual(t, "DisplayName", profile.DisplayName.String, "Profile User")
}

func TestGetClipperSubmissionStats(t *testing.T) {
	cleanupAll(t)

	ownerID := "cpstat_owner"
	clipperID := "cpstat_clipper"
	seedUser(t, ownerID, "cpstatowner@cp.com", "owner")
	seedUser(t, clipperID, "cpstatclipper@cp.com", "clipper")
	campaign := seedCampaign(t, ownerID, "CPStat Campaign", "youtube", 500000, 5000)

	// Create submissions with different statuses
	sub1 := seedSubmission(t, campaign.ID, clipperID, "https://yt.com/cp1", "youtube")
	sub2 := seedSubmission(t, campaign.ID, clipperID, "https://yt.com/cp2", "youtube")
	sub3 := seedSubmission(t, campaign.ID, clipperID, "https://yt.com/cp3", "youtube")

	// Approve sub1
	_, err := testDB.Queries.ApproveSubmission(context.Background(), sub1.ID)
	if err != nil {
		t.Fatalf("ApproveSubmission: %v", err)
	}

	// Reject sub2
	_, err = testDB.Queries.UpdateSubmissionStatus(context.Background(), sqlc.UpdateSubmissionStatusParams{
		ID:              sub2.ID,
		Status:          "rejected",
		RejectionReason: pgtype.Text{String: "low quality", Valid: true},
	})
	if err != nil {
		t.Fatalf("UpdateSubmissionStatus: %v", err)
	}

	stats, err := testDB.Queries.GetClipperSubmissionStats(context.Background(), clipperID)
	if err != nil {
		t.Fatalf("GetClipperSubmissionStats: %v", err)
	}
	assertIntEqual(t, "TotalSubmissions", stats.TotalSubmissions, 3)
	assertIntEqual(t, "ApprovedSubmissions", stats.ApprovedSubmissions, 1)
	assertIntEqual(t, "PendingSubmissions", stats.PendingSubmissions, 1)
	assertIntEqual(t, "RejectedSubmissions", stats.RejectedSubmissions, 1)
	_ = sub3
}

func TestGetClipperTotalViews(t *testing.T) {
	cleanupAll(t)

	ownerID := "cpview_owner"
	clipperID := "cpview_clipper"
	seedUser(t, ownerID, "cpviewowner@cp.com", "owner")
	seedUser(t, clipperID, "cpviewclipper@cp.com", "clipper")
	campaign := seedCampaign(t, ownerID, "CPView Campaign", "youtube", 500000, 5000)
	sub := seedSubmission(t, campaign.ID, clipperID, "https://yt.com/cpview", "youtube")

	_, err := testDB.Queries.CreateMetricSnapshot(context.Background(), sqlc.CreateMetricSnapshotParams{
		SubmissionID: sub.ID,
		Platform:     "youtube",
		Views:        5000,
		Likes:        100,
		Comments:     10,
		Shares:       5,
		CapturedAt:   pgtype.Timestamptz{Valid: true},
	})
	if err != nil {
		t.Fatalf("CreateMetricSnapshot: %v", err)
	}

	totalViews, err := testDB.Queries.GetClipperTotalViews(context.Background(), clipperID)
	if err != nil {
		t.Fatalf("GetClipperTotalViews: %v", err)
	}
	assertInt64Equal(t, "GetClipperTotalViews", totalViews, 5000)
}

func TestGetClipperTotalEarnings(t *testing.T) {
	cleanupAll(t)

	ownerID := "cpearn_owner"
	clipperID := "cpearn_clipper"
	seedUser(t, ownerID, "cpearnowner@cp.com", "owner")
	seedUser(t, clipperID, "cpearnclipper@cp.com", "clipper")
	campaign := seedCampaign(t, ownerID, "CPEarn Campaign", "youtube", 500000, 5000)

	seedLedgerEntry(t, "cp_e_idem1", "earning", campaign.ID, 25000, clipperID)
	seedLedgerEntry(t, "cp_e_idem2", "earning", campaign.ID, 35000, clipperID)
	seedLedgerEntry(t, "cp_e_idem3", "platform_fee", campaign.ID, 5000, "")

	clipperText := pgtype.Text{String: clipperID, Valid: true}
	totalEarnings, err := testDB.Queries.GetClipperTotalEarnings(context.Background(), clipperText)
	if err != nil {
		t.Fatalf("GetClipperTotalEarnings: %v", err)
	}
	assertIntEqual(t, "GetClipperTotalEarnings", totalEarnings, 60000)
}

func TestGetClipperCampaignCount(t *testing.T) {
	cleanupAll(t)

	ownerID := "cpcamp_owner"
	clipperID := "cpcamp_clipper"
	seedUser(t, ownerID, "cpcampowner@cp.com", "owner")
	seedUser(t, clipperID, "cpcampclipper@cp.com", "clipper")

	c1 := seedCampaign(t, ownerID, "CPCamp 1", "youtube", 500000, 5000)
	c2 := seedCampaign(t, ownerID, "CPCamp 2", "instagram", 300000, 3000)

	seedSubmission(t, c1.ID, clipperID, "https://yt.com/cpc1", "youtube")
	seedSubmission(t, c2.ID, clipperID, "https://ig.com/cpc1", "instagram")
	// Two submissions to same campaign - should count as 1
	seedSubmission(t, c1.ID, clipperID, "https://yt.com/cpc2", "youtube")

	count, err := testDB.Queries.GetClipperCampaignCount(context.Background(), clipperID)
	if err != nil {
		t.Fatalf("GetClipperCampaignCount: %v", err)
	}
	assertIntEqual(t, "GetClipperCampaignCount", count, 2)
}
