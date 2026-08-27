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

func TestCreateAndGetPayoutRequest(t *testing.T) {
	cleanupAll(t)

	clipperID := "pay_clipper1"
	seedUser(t, clipperID, "payclipper1@pay.com", "clipper")

	payout, err := testDB.Queries.CreatePayoutRequest(context.Background(), sqlc.CreatePayoutRequestParams{
		ID:             newUUID(),
		ClipperID:      clipperID,
		Amount:         100000,
		UpiID:          "user@upi",
		IdempotencyKey: "pay_idem_001",
	})
	if err != nil {
		t.Fatalf("CreatePayoutRequest: %v", err)
	}
	assertEqual(t, "Status", payout.Status, "pending")
	assertEqual(t, "UpiID", payout.UpiID, "user@upi")
	assertIntEqual(t, "Amount", payout.Amount, 100000)
}

func TestCreatePayoutRequestDuplicateIdempotencyKey(t *testing.T) {
	cleanupAll(t)

	clipperID := "pay_dup_clipper"
	seedUser(t, clipperID, "paydupclipper@pay.com", "clipper")

	_, err := testDB.Queries.CreatePayoutRequest(context.Background(), sqlc.CreatePayoutRequestParams{
		ID:             newUUID(),
		ClipperID:      clipperID,
		Amount:         100000,
		UpiID:          "user@upi",
		IdempotencyKey: "pay_idem_dup",
	})
	if err != nil {
		t.Fatalf("first CreatePayoutRequest: %v", err)
	}

	// Second insert with same idempotency key - should DO NOTHING and return no rows
	_, err = testDB.Queries.CreatePayoutRequest(context.Background(), sqlc.CreatePayoutRequestParams{
		ID:             newUUID(),
		ClipperID:      clipperID,
		Amount:         99999,
		UpiID:          "different@upi",
		IdempotencyKey: "pay_idem_dup",
	})
	if err == nil {
		t.Fatal("expected pgx.ErrNoRows for duplicate idempotency key, got nil")
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("expected pgx.ErrNoRows, got %v", err)
	}
}

func TestSumPendingPayoutsByClipper(t *testing.T) {
	cleanupAll(t)

	clipperID := "pay_sum_clipper"
	seedUser(t, clipperID, "paysumclipper@pay.com", "clipper")

	// Create two pending payouts
	_, err := testDB.Queries.CreatePayoutRequest(context.Background(), sqlc.CreatePayoutRequestParams{
		ID:             newUUID(),
		ClipperID:      clipperID,
		Amount:         50000,
		UpiID:          "user@upi",
		IdempotencyKey: "pay_idem_s1",
	})
	if err != nil {
		t.Fatalf("CreatePayoutRequest 1: %v", err)
	}

	_, err = testDB.Queries.CreatePayoutRequest(context.Background(), sqlc.CreatePayoutRequestParams{
		ID:             newUUID(),
		ClipperID:      clipperID,
		Amount:         75000,
		UpiID:          "user@upi",
		IdempotencyKey: "pay_idem_s2",
	})
	if err != nil {
		t.Fatalf("CreatePayoutRequest 2: %v", err)
	}

	total, err := testDB.Queries.SumPendingPayoutsByClipper(context.Background(), clipperID)
	if err != nil {
		t.Fatalf("SumPendingPayoutsByClipper: %v", err)
	}
	assertInt64Equal(t, "SumPendingPayoutsByClipper", total, 125000)
}

func TestUpdatePayoutRequestStatus(t *testing.T) {
	cleanupAll(t)

	clipperID := "pay_upd_clipper"
	seedUser(t, clipperID, "payupdclipper@pay.com", "clipper")

	payout, err := testDB.Queries.CreatePayoutRequest(context.Background(), sqlc.CreatePayoutRequestParams{
		ID:             newUUID(),
		ClipperID:      clipperID,
		Amount:         100000,
		UpiID:          "user@upi",
		IdempotencyKey: "pay_idem_upd",
	})
	if err != nil {
		t.Fatalf("CreatePayoutRequest: %v", err)
	}

	updated, err := testDB.Queries.UpdatePayoutRequestStatus(context.Background(), sqlc.UpdatePayoutRequestStatusParams{
		ID:          payout.ID,
		Status:      "completed",
		ProviderRef: pgtype.Text{String: "razorpay_ref_123", Valid: true},
	})
	if err != nil {
		t.Fatalf("UpdatePayoutRequestStatus: %v", err)
	}
	assertEqual(t, "Status", updated.Status, "completed")
	if !updated.ProcessedAt.Valid {
		t.Error("expected ProcessedAt to be set for completed status")
	}
}

func TestUpdateUserUPI(t *testing.T) {
	cleanupAll(t)

	clipperID := "pay_upi_clipper"
	seedUser(t, clipperID, "payupiclipper@pay.com", "clipper")

	updated, err := testDB.Queries.UpdateUserUPI(context.Background(), sqlc.UpdateUserUPIParams{
		ID:    clipperID,
		UpiID: pgtype.Text{String: "new@upi", Valid: true},
	})
	if err != nil {
		t.Fatalf("UpdateUserUPI: %v", err)
	}
	if !updated.UpiID.Valid || updated.UpiID.String != "new@upi" {
		t.Errorf("UPI ID: got %v, want 'new@upi'", updated.UpiID)
	}
}

func TestListPayoutRequestsByClipper(t *testing.T) {
	cleanupAll(t)

	clipperID := "pay_list_clipper"
	seedUser(t, clipperID, "paylistclipper@pay.com", "clipper")

	for i := 0; i < 3; i++ {
		_, err := testDB.Queries.CreatePayoutRequest(context.Background(), sqlc.CreatePayoutRequestParams{
			ID:             newUUID(),
			ClipperID:      clipperID,
			Amount:         int32((i + 1) * 50000),
			UpiID:          "user@upi",
			IdempotencyKey: "pay_idem_lst" + string(rune('a'+i)),
		})
		if err != nil {
			t.Fatalf("CreatePayoutRequest %d: %v", i, err)
		}
	}

	payouts, err := testDB.Queries.ListPayoutRequestsByClipper(context.Background(), clipperID)
	if err != nil {
		t.Fatalf("ListPayoutRequestsByClipper: %v", err)
	}
	if len(payouts) != 3 {
		t.Errorf("expected 3 payout requests, got %d", len(payouts))
	}
}

func TestUpdatePayoutRequestStatusToFailed(t *testing.T) {
	cleanupAll(t)

	clipperID := "pay_fail_clipper"
	seedUser(t, clipperID, "payfailclipper@pay.com", "clipper")

	payout, err := testDB.Queries.CreatePayoutRequest(context.Background(), sqlc.CreatePayoutRequestParams{
		ID:             newUUID(),
		ClipperID:      clipperID,
		Amount:         100000,
		UpiID:          "user@upi",
		IdempotencyKey: "pay_idem_fail",
	})
	if err != nil {
		t.Fatalf("CreatePayoutRequest: %v", err)
	}

	updated, err := testDB.Queries.UpdatePayoutRequestStatus(context.Background(), sqlc.UpdatePayoutRequestStatusParams{
		ID:            payout.ID,
		Status:        "failed",
		FailureReason: pgtype.Text{String: "insufficient balance", Valid: true},
	})
	if err != nil {
		t.Fatalf("UpdatePayoutRequestStatus: %v", err)
	}
	assertEqual(t, "Status", updated.Status, "failed")
	if !updated.ProcessedAt.Valid {
		t.Error("expected ProcessedAt to be set for failed status")
	}
}
