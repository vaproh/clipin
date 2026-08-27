//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	sqlc "clipin/apps/api/internal/db/sqlc"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestCreateAndGetMetricSnapshot(t *testing.T) {
	cleanupAll(t)

	ownerID := "ver_owner1"
	clipperID := "ver_clipper1"
	seedUser(t, ownerID, "verowner1@ver.com", "owner")
	seedUser(t, clipperID, "verclipper1@ver.com", "clipper")
	campaign := seedCampaign(t, ownerID, "Ver Campaign", "youtube", 500000, 5000)
	sub := seedSubmission(t, campaign.ID, clipperID, "https://yt.com/ver1", "youtube")

	snap, err := testDB.Queries.CreateMetricSnapshot(context.Background(), sqlc.CreateMetricSnapshotParams{
		SubmissionID: sub.ID,
		Platform:     "youtube",
		Views:        10000,
		Likes:        500,
		Comments:     50,
		Shares:       25,
		CapturedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
	})
	if err != nil {
		t.Fatalf("CreateMetricSnapshot: %v", err)
	}

	fetched, err := testDB.Queries.GetLatestSnapshotForSubmission(context.Background(), sub.ID)
	if err != nil {
		t.Fatalf("GetLatestSnapshotForSubmission: %v", err)
	}
	assertInt64Equal(t, "Views", fetched.Views, 10000)
	assertInt64Equal(t, "Likes", fetched.Likes, 500)
	_ = snap
}

func TestGetLatestSnapshotForSubmission(t *testing.T) {
	cleanupAll(t)

	ownerID := "ver_latest_owner"
	clipperID := "ver_latest_clipper"
	seedUser(t, ownerID, "latestowner@ver.com", "owner")
	seedUser(t, clipperID, "latestclipper@ver.com", "clipper")
	campaign := seedCampaign(t, ownerID, "Latest Campaign", "youtube", 500000, 5000)
	sub := seedSubmission(t, campaign.ID, clipperID, "https://yt.com/latest", "youtube")

	now := time.Now()
	// First snapshot: 1000 views
	_, err := testDB.Queries.CreateMetricSnapshot(context.Background(), sqlc.CreateMetricSnapshotParams{
		SubmissionID: sub.ID,
		Platform:     "youtube",
		Views:        1000,
		CapturedAt:   pgtype.Timestamptz{Time: now.Add(-1 * time.Hour), Valid: true},
	})
	if err != nil {
		t.Fatalf("CreateMetricSnapshot 1: %v", err)
	}

	// Second snapshot: 5000 views (latest)
	_, err = testDB.Queries.CreateMetricSnapshot(context.Background(), sqlc.CreateMetricSnapshotParams{
		SubmissionID: sub.ID,
		Platform:     "youtube",
		Views:        5000,
		CapturedAt:   pgtype.Timestamptz{Time: now, Valid: true},
	})
	if err != nil {
		t.Fatalf("CreateMetricSnapshot 2: %v", err)
	}

	latest, err := testDB.Queries.GetLatestSnapshotForSubmission(context.Background(), sub.ID)
	if err != nil {
		t.Fatalf("GetLatestSnapshotForSubmission: %v", err)
	}
	assertInt64Equal(t, "Latest Views", latest.Views, 5000)
}

func TestGetInitialSnapshotForSubmission(t *testing.T) {
	cleanupAll(t)

	ownerID := "ver_init_owner"
	clipperID := "ver_init_clipper"
	seedUser(t, ownerID, "initowner@ver.com", "owner")
	seedUser(t, clipperID, "initclipper@ver.com", "clipper")
	campaign := seedCampaign(t, ownerID, "Init Campaign", "youtube", 500000, 5000)
	sub := seedSubmission(t, campaign.ID, clipperID, "https://yt.com/init", "youtube")

	now := time.Now()
	// First snapshot
	_, err := testDB.Queries.CreateMetricSnapshot(context.Background(), sqlc.CreateMetricSnapshotParams{
		SubmissionID: sub.ID,
		Platform:     "youtube",
		Views:        100,
		CapturedAt:   pgtype.Timestamptz{Time: now.Add(-2 * time.Hour), Valid: true},
	})
	if err != nil {
		t.Fatalf("CreateMetricSnapshot 1: %v", err)
	}

	// Second snapshot
	_, err = testDB.Queries.CreateMetricSnapshot(context.Background(), sqlc.CreateMetricSnapshotParams{
		SubmissionID: sub.ID,
		Platform:     "youtube",
		Views:        500,
		CapturedAt:   pgtype.Timestamptz{Time: now, Valid: true},
	})
	if err != nil {
		t.Fatalf("CreateMetricSnapshot 2: %v", err)
	}

	initial, err := testDB.Queries.GetInitialSnapshotForSubmission(context.Background(), sub.ID)
	if err != nil {
		t.Fatalf("GetInitialSnapshotForSubmission: %v", err)
	}
	assertInt64Equal(t, "Initial Views", initial.Views, 100)
}

func TestListSnapshotsBySubmission(t *testing.T) {
	cleanupAll(t)

	ownerID := "ver_list_owner"
	clipperID := "ver_list_clipper"
	seedUser(t, ownerID, "listowner@ver.com", "owner")
	seedUser(t, clipperID, "listclipper@ver.com", "clipper")
	campaign := seedCampaign(t, ownerID, "List Campaign", "youtube", 500000, 5000)
	sub := seedSubmission(t, campaign.ID, clipperID, "https://yt.com/list", "youtube")

	now := time.Now()
	for i := 0; i < 3; i++ {
		_, err := testDB.Queries.CreateMetricSnapshot(context.Background(), sqlc.CreateMetricSnapshotParams{
			SubmissionID: sub.ID,
			Platform:     "youtube",
			Views:        int64((i + 1) * 1000),
			CapturedAt:   pgtype.Timestamptz{Time: now.Add(time.Duration(i) * time.Hour), Valid: true},
		})
		if err != nil {
			t.Fatalf("CreateMetricSnapshot %d: %v", i, err)
		}
	}

	snaps, err := testDB.Queries.ListSnapshotsBySubmission(context.Background(), sub.ID)
	if err != nil {
		t.Fatalf("ListSnapshotsBySubmission: %v", err)
	}
	if len(snaps) != 3 {
		t.Fatalf("expected 3 snapshots, got %d", len(snaps))
	}
	// Verify ordering (ASC)
	assertInt64Equal(t, "First snapshot views", snaps[0].Views, 1000)
	assertInt64Equal(t, "Last snapshot views", snaps[2].Views, 3000)
}

func TestCountSnapshotsBySubmission(t *testing.T) {
	cleanupAll(t)

	ownerID := "ver_cnt_owner"
	clipperID := "ver_cnt_clipper"
	seedUser(t, ownerID, "cntowner@ver.com", "owner")
	seedUser(t, clipperID, "cntclipper@ver.com", "clipper")
	campaign := seedCampaign(t, ownerID, "Cnt Campaign", "youtube", 500000, 5000)
	sub := seedSubmission(t, campaign.ID, clipperID, "https://yt.com/cnt", "youtube")

	for i := 0; i < 5; i++ {
		_, err := testDB.Queries.CreateMetricSnapshot(context.Background(), sqlc.CreateMetricSnapshotParams{
			SubmissionID: sub.ID,
			Platform:     "youtube",
			Views:        int64((i + 1) * 100),
			CapturedAt:   pgtype.Timestamptz{Time: time.Now().Add(time.Duration(i) * time.Minute), Valid: true},
		})
		if err != nil {
			t.Fatalf("CreateMetricSnapshot %d: %v", i, err)
		}
	}

	count, err := testDB.Queries.CountSnapshotsBySubmission(context.Background(), sub.ID)
	if err != nil {
		t.Fatalf("CountSnapshotsBySubmission: %v", err)
	}
	assertInt64Equal(t, "CountSnapshotsBySubmission", count, 5)
}

func TestGetSubmissionsNeedingVerification(t *testing.T) {
	cleanupAll(t)

	ownerID := "ver_need_owner"
	clipperID := "ver_need_clipper"
	seedUser(t, ownerID, "needowner@ver.com", "owner")
	seedUser(t, clipperID, "needclipper@ver.com", "clipper")
	campaign := seedCampaign(t, ownerID, "Need Campaign", "youtube", 500000, 5000)

	// Approved submission without recent snapshot
	sub := seedSubmission(t, campaign.ID, clipperID, "https://yt.com/need", "youtube")
	_, err := testDB.Queries.ApproveSubmission(context.Background(), sub.ID)
	if err != nil {
		t.Fatalf("ApproveSubmission: %v", err)
	}

	needing, err := testDB.Queries.GetSubmissionsNeedingVerification(context.Background(), 10)
	if err != nil {
		t.Fatalf("GetSubmissionsNeedingVerification: %v", err)
	}
	if len(needing) != 1 {
		t.Fatalf("expected 1 submission needing verification, got %d", len(needing))
	}
	assertEqual(t, "SubmissionID", formatUUID(needing[0].ID), formatUUID(sub.ID))
}
