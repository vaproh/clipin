package payout

import (
	"context"
	"fmt"
	"log"
)

// RazorpayStub is a stub implementation of PayoutProvider.
// It logs the transfer request and returns a fake "completed" response.
// Replace with the real Razorpay SDK when the account is available.
type RazorpayStub struct{}

func (r *RazorpayStub) CreateTransfer(ctx context.Context, req TransferRequest) (*TransferResponse, error) {
	log.Printf("[STUB] Payout: %d paise to %s (ref: %s)", req.Amount, req.UPITarget, req.Reference)
	return &TransferResponse{
		TransferID: fmt.Sprintf("pay_stub_%s", req.Reference),
		Status:     "completed",
		Reference:  req.Reference,
	}, nil
}

func (r *RazorpayStub) GetTransferStatus(ctx context.Context, transferID string) (*TransferResponse, error) {
	log.Printf("[STUB] GetTransferStatus: %s", transferID)
	return &TransferResponse{
		TransferID: transferID,
		Status:     "completed",
	}, nil
}
