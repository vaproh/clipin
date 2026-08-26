package payout_test

import (
	"context"
	"testing"

	"clipin/apps/api/internal/payout"
)

func TestRazorpayStub_CreateTransfer(t *testing.T) {
	stub := &payout.RazorpayStub{}
	resp, err := stub.CreateTransfer(context.Background(), payout.TransferRequest{
		Amount:    50000,
		Currency:  "INR",
		UPITarget: "test@upi",
		Reference: "ref-123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.TransferID == "" {
		t.Error("expected non-empty transfer ID")
	}
	if resp.Status != "completed" {
		t.Errorf("expected status 'completed', got %s", resp.Status)
	}
	if resp.Reference != "ref-123" {
		t.Errorf("expected reference 'ref-123', got %s", resp.Reference)
	}
}

func TestRazorpayStub_GetTransferStatus(t *testing.T) {
	stub := &payout.RazorpayStub{}
	resp, err := stub.GetTransferStatus(context.Background(), "pay_stub_abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.TransferID != "pay_stub_abc" {
		t.Errorf("expected transfer ID 'pay_stub_abc', got %s", resp.TransferID)
	}
	if resp.Status != "completed" {
		t.Errorf("expected status 'completed', got %s", resp.Status)
	}
}

func TestRazorpayStub_CreateTransfer_IDFormat(t *testing.T) {
	stub := &payout.RazorpayStub{}
	resp, err := stub.CreateTransfer(context.Background(), payout.TransferRequest{
		Amount:    100000,
		Currency:  "INR",
		UPITarget: "user@bank",
		Reference: "payout:clipper1:abc",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "pay_stub_payout:clipper1:abc"
	if resp.TransferID != expected {
		t.Errorf("expected transfer ID %s, got %s", expected, resp.TransferID)
	}
}

func TestRazorpayStub_CreateTransfer_VariousAmounts(t *testing.T) {
	stub := &payout.RazorpayStub{}
	amounts := []int64{50000, 100000, 500000, 1000000}
	for _, amt := range amounts {
		resp, err := stub.CreateTransfer(context.Background(), payout.TransferRequest{
			Amount:    amt,
			Currency:  "INR",
			UPITarget: "test@upi",
			Reference: "ref-test",
		})
		if err != nil {
			t.Fatalf("unexpected error for amount %d: %v", amt, err)
		}
		if resp.Status != "completed" {
			t.Errorf("expected status 'completed' for amount %d, got %s", amt, resp.Status)
		}
	}
}
