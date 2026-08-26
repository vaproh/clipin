package payout

import "context"

// TransferRequest describes a UPI payout to initiate.
type TransferRequest struct {
	Amount    int64
	Currency  string // "INR"
	UPITarget string // UPI ID
	Reference string // idempotency key
}

// TransferResponse describes the result of a payout transfer.
type TransferResponse struct {
	TransferID string
	Status     string // "pending" | "processing" | "completed" | "failed"
	Reference  string
}

// PayoutProvider abstracts the payment gateway for UPI transfers.
type PayoutProvider interface {
	CreateTransfer(ctx context.Context, req TransferRequest) (*TransferResponse, error)
	GetTransferStatus(ctx context.Context, transferID string) (*TransferResponse, error)
}
