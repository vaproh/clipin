//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	sqlc "clipin/apps/api/internal/db/sqlc"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestCreateAndGetSubmission(t *testing.T) {
	cleanupAll(t)

	ownerID := "sub_owner1"
	clipperID := "sub_clipper1"
	seedUser(t, ownerID, "subowner1@sub.com", "owner")
	seedUser(t, clipperID, "subclipper1@sub.com", "clipper")
	campaign := seedCampaign(t, ownerID, "Sub Campaign", "youtube", 500000, 5000)

	sub := seedSubmission(t, campaign.ID, clipperID, "https://youtube.com/watch?v=abc", "youtube")

	fetched, err := testDB.Queries.GetSubmissionByID(context.Background(), sub.ID)
	if err != nil {
		t.Fatalf("GetSubmissionByID: %v", err)
	}
	assertEqual(t, "ClipperID", fetched.ClipperID, clipperID)
	assertEqual(t, "PostUrl", fetched.PostUrl, "https://youtube.com/watch?v=abc")
	assertEqual(t, "Status", fetched.Status, "pending")
}

func TestListSubmissionsByCampaign(t *testing.T) {
	cleanupAll(t)

	ownerID := "sub_list_owner"
	clipperID := "sub_list_clipper"
	seedUser(t, ownerID, "listowner@sub.com", "owner")
	seedUser(t, clipperID, "listclipper@sub.com", "clipper")
	campaign := seedCampaign(t, ownerID, "List Campaign", "youtube", 500000, 5000)

	for i := 0; i < 3; i++ {
		seedSubmission(t, campaign.ID, clipperID, "https://youtube.com/watch?v=v"+string(rune('a'+i)), "youtube")
	}

	subs, err := testDB.Queries.ListSubmissionsByCampaign(context.Background(), campaign.ID)
	if err != nil {
		t.Fatalf("ListSubmissionsByCampaign: %v", err)
	}
	if len(subs) != 3 {
		t.Errorf("expected 3 submissions, got %d", len(subs))
	}
}

func TestListSubmissionsByClipper(t *testing.T) {
	cleanupAll(t)

	ownerID := "sub_clip_owner"
	clipperID := "sub_clip_clipper"
	seedUser(t, ownerID, "clipowner@sub.com", "owner")
	seedUser(t, clipperID, "clipclipper@sub.com", "clipper")
	c1 := seedCampaign(t, ownerID, "Campaign 1", "youtube", 500000, 5000)
	c2 := seedCampaign(t, ownerID, "Campaign 2", "instagram", 300000, 3000)

	seedSubmission(t, c1.ID, clipperID, "https://yt.com/1", "youtube")
	seedSubmission(t, c2.ID, clipperID, "https://ig.com/1", "instagram")

	subs, err := testDB.Queries.ListSubmissionsByClipper(context.Background(), clipperID)
	if err != nil {
		t.Fatalf("ListSubmissionsByClipper: %v", err)
	}
	if len(subs) != 2 {
		t.Errorf("expected 2 submissions, got %d", len(subs))
	}
}

func TestUpdateSubmissionStatus(t *testing.T) {
	cleanupAll(t)

	ownerID := "sub_upd_owner"
	clipperID := "sub_upd_clipper"
	seedUser(t, ownerID, "updowner@sub.com", "owner")
	seedUser(t, clipperID, "updclipper@sub.com", "clipper")
	campaign := seedCampaign(t, ownerID, "Upd Campaign", "youtube", 500000, 5000)
	sub := seedSubmission(t, campaign.ID, clipperID, "https://yt.com/upd", "youtube")

	updated, err := testDB.Queries.UpdateSubmissionStatus(context.Background(), sqlc.UpdateSubmissionStatusParams{
		ID:     sub.ID,
		Status: "approved",
	})
	if err != nil {
		t.Fatalf("UpdateSubmissionStatus: %v", err)
	}
	assertEqual(t, "Status", updated.Status, "approved")
	if !updated.ApprovedAt.Valid {
		t.Error("expected ApprovedAt to be set")
	}
}

func TestApproveSubmissionConditionalWhere(t *testing.T) {
	cleanupAll(t)

	ownerID := "sub_appr_owner"
	clipperID := "sub_appr_clipper"
	seedUser(t, ownerID, "approwner@sub.com", "owner")
	seedUser(t, clipperID, "apprclipper@sub.com", "clipper")
	campaign := seedCampaign(t, ownerID, "Appr Campaign", "youtube", 500000, 5000)
	sub := seedSubmission(t, campaign.ID, clipperID, "https://yt.com/appr", "youtube")

	rows, err := testDB.Queries.ApproveSubmission(context.Background(), sub.ID)
	if err != nil {
		t.Fatalf("ApproveSubmission: %v", err)
	}
	if rows != 1 {
		t.Errorf("expected 1 row affected, got %d", rows)
	}

	fetched, err := testDB.Queries.GetSubmissionByID(context.Background(), sub.ID)
	if err != nil {
		t.Fatalf("GetSubmissionByID: %v", err)
	}
	assertEqual(t, "Status", fetched.Status, "approved")
}

func TestApproveSubmissionAlreadyApproved(t *testing.T) {
	cleanupAll(t)

	ownerID := "sub_appr2_owner"
	clipperID := "sub_appr2_clipper"
	seedUser(t, ownerID, "appr2owner@sub.com", "owner")
	seedUser(t, clipperID, "appr2clipper@sub.com", "clipper")
	campaign := seedCampaign(t, ownerID, "Appr2 Campaign", "youtube", 500000, 5000)
	sub := seedSubmission(t, campaign.ID, clipperID, "https://yt.com/appr2", "youtube")

	// Approve once
	_, err := testDB.Queries.ApproveSubmission(context.Background(), sub.ID)
	if err != nil {
		t.Fatalf("first ApproveSubmission: %v", err)
	}

	// Try to approve again - conditional WHERE should affect 0 rows
	rows, err := testDB.Queries.ApproveSubmission(context.Background(), sub.ID)
	if err != nil {
		t.Fatalf("second ApproveSubmission: %v", err)
	}
	if rows != 0 {
		t.Errorf("expected 0 rows affected (already approved), got %d", rows)
	}
}

func TestAutoApproveSubmission(t *testing.T) {
	cleanupAll(t)

	ownerID := "sub_auto_owner"
	clipperID := "sub_auto_clipper"
	seedUser(t, ownerID, "autoowner@sub.com", "owner")
	seedUser(t, clipperID, "autoclipper@sub.com", "clipper")
	campaign := seedCampaign(t, ownerID, "Auto Campaign", "youtube", 500000, 5000)
	sub := seedSubmission(t, campaign.ID, clipperID, "https://yt.com/auto", "youtube")

	rows, err := testDB.Queries.AutoApproveSubmission(context.Background(), sub.ID)
	if err != nil {
		t.Fatalf("AutoApproveSubmission: %v", err)
	}
	if rows != 1 {
		t.Errorf("expected 1 row affected, got %d", rows)
	}

	fetched, err := testDB.Queries.GetSubmissionByID(context.Background(), sub.ID)
	if err != nil {
		t.Fatalf("GetSubmissionByID: %v", err)
	}
	assertEqual(t, "Status", fetched.Status, "auto_approved")
	if !fetched.AutoApprovedAt.Valid {
		t.Error("expected AutoApprovedAt to be set")
	}
}

func TestDuplicateSubmissionURLRejected(t *testing.T) {
	cleanupAll(t)

	ownerID := "sub_dup_owner"
	clipperID := "sub_dup_clipper"
	seedUser(t, ownerID, "dupowner@sub.com", "owner")
	seedUser(t, clipperID, "dupclipper@sub.com", "clipper")
	campaign := seedCampaign(t, ownerID, "Dup Campaign", "youtube", 500000, 5000)

	_, err := testDB.Queries.CreateSubmission(context.Background(), sqlc.CreateSubmissionParams{
		ID:         newUUID(),
		CampaignID: campaign.ID,
		ClipperID:  clipperID,
		PostUrl:    "https://yt.com/dup",
		Platform:   "youtube",
	})
	if err != nil {
		t.Fatalf("first CreateSubmission: %v", err)
	}

	_, err = testDB.Queries.CreateSubmission(context.Background(), sqlc.CreateSubmissionParams{
		ID:         newUUID(),
		CampaignID: campaign.ID,
		ClipperID:  clipperID,
		PostUrl:    "https://yt.com/dup",
		Platform:   "youtube",
	})
	if err == nil {
		t.Error("expected error for duplicate URL, got nil")
	}
}

func TestCountSubmissionsByCampaign(t *testing.T) {
	cleanupAll(t)

	ownerID := "sub_cnt_owner"
	clipperID := "sub_cnt_clipper"
	seedUser(t, ownerID, "cntowner@sub.com", "owner")
	seedUser(t, clipperID, "cntclipper@sub.com", "clipper")
	campaign := seedCampaign(t, ownerID, "Count Campaign", "youtube", 500000, 5000)

	for i := 0; i < 4; i++ {
		seedSubmission(t, campaign.ID, clipperID, "https://yt.com/cnt"+string(rune('a'+i)), "youtube")
	}

	count, err := testDB.Queries.CountSubmissionsByCampaign(context.Background(), campaign.ID)
	if err != nil {
		t.Fatalf("CountSubmissionsByCampaign: %v", err)
	}
	assertInt64Equal(t, "CountSubmissionsByCampaign", count, 4)
}

func TestCountSubmissionsByClipperForCampaign(t *testing.T) {
	cleanupAll(t)

	ownerID := "sub_ccnt_owner"
	clipper1 := "sub_ccnt_clipper1"
	clipper2 := "sub_ccnt_clipper2"
	seedUser(t, ownerID, "ccntowner@sub.com", "owner")
	seedUser(t, clipper1, "ccntclipper1@sub.com", "clipper")
	seedUser(t, clipper2, "ccntclipper2@sub.com", "clipper")
	campaign := seedCampaign(t, ownerID, "CCnt Campaign", "youtube", 500000, 5000)

	seedSubmission(t, campaign.ID, clipper1, "https://yt.com/cc1", "youtube")
	seedSubmission(t, campaign.ID, clipper1, "https://yt.com/cc2", "youtube")
	seedSubmission(t, campaign.ID, clipper2, "https://yt.com/cc3", "youtube")

	count, err := testDB.Queries.CountSubmissionsByClipperForCampaign(context.Background(), sqlc.CountSubmissionsByClipperForCampaignParams{
		CampaignID: campaign.ID,
		ClipperID:  clipper1,
	})
	if err != nil {
		t.Fatalf("CountSubmissionsByClipperForCampaign: %v", err)
	}
	assertInt64Equal(t, "CountSubmissionsByClipperForCampaign clipper1", count, 2)
}

func TestListPendingSubmissionsOlderThan(t *testing.T) {
	cleanupAll(t)

	ownerID := "sub_pend_owner"
	clipperID := "sub_pend_clipper"
	seedUser(t, ownerID, "pendowner@sub.com", "owner")
	seedUser(t, clipperID, "pendclipper@sub.com", "clipper")
	campaign := seedCampaign(t, ownerID, "Pending Campaign", "youtube", 500000, 5000)

	seedSubmission(t, campaign.ID, clipperID, "https://yt.com/pend1", "youtube")

	// Query for submissions older than future time (should find the one we just created)
	futureTime := time.Now().Add(1 * time.Hour)
	subs, err := testDB.Queries.ListPendingSubmissionsOlderThan(context.Background(), pgtype.Timestamptz{Time: futureTime, Valid: true})
	if err != nil {
		t.Fatalf("ListPendingSubmissionsOlderThan: %v", err)
	}
	if len(subs) != 1 {
		t.Fatalf("expected 1 pending submission, got %d", len(subs))
	}
	assertEqual(t, "ClipperID", subs[0].ClipperID, clipperID)
}

func TestListPendingSubmissionsOlderThanNone(t *testing.T) {
	cleanupAll(t)

	ownerID := "sub_pend2_owner"
	clipperID := "sub_pend2_clipper"
	seedUser(t, ownerID, "pend2owner@sub.com", "owner")
	seedUser(t, clipperID, "pend2clipper@sub.com", "clipper")
	campaign := seedCampaign(t, ownerID, "Pending2 Campaign", "youtube", 500000, 5000)
	sub := seedSubmission(t, campaign.ID, clipperID, "https://yt.com/pend2", "youtube")

	// Approve it first
	_, err := testDB.Queries.ApproveSubmission(context.Background(), sub.ID)
	if err != nil {
		t.Fatalf("ApproveSubmission: %v", err)
	}

	// Now query for pending older than future - should find 0
	futureTime := time.Now().Add(1 * time.Hour)
	subs, err := testDB.Queries.ListPendingSubmissionsOlderThan(context.Background(), pgtype.Timestamptz{Time: futureTime, Valid: true})
	if err != nil {
		t.Fatalf("ListPendingSubmissionsOlderThan: %v", err)
	}
	if len(subs) != 0 {
		t.Errorf("expected 0 pending submissions (all approved), got %d", len(subs))
	}
}
