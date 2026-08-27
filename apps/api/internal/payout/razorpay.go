package payout

import (
	"context"
	"fmt"
	"log/slog"

	razorpay "github.com/razorpay/razorpay-go"
	"github.com/razorpay/razorpay-go/constants"
)

// RazorpayProvider implements PayoutProvider using the RazorpayX API.
type RazorpayProvider struct {
	client        *razorpay.Client
	accountNumber string
}

// NewRazorpayProvider creates a RazorpayProvider from API credentials.
func NewRazorpayProvider(keyID, keySecret, accountNumber string) *RazorpayProvider {
	client := razorpay.NewClient(keyID, keySecret)
	return &RazorpayProvider{
		client:        client,
		accountNumber: accountNumber,
	}
}

func (r *RazorpayProvider) CreateTransfer(ctx context.Context, req TransferRequest) (*TransferResponse, error) {
	slog.Info("razorpay payout initiated",
		"amount", req.Amount,
		"upi_target", RedactUPI(req.UPITarget),
	)

	// Step 1: Create a contact for the clipper.
	contact, err := r.client.Request.Post(
		"/"+constants.VERSION_V1+"/contacts",
		map[string]interface{}{
			"name":    req.Reference,
			"type":    "customer",
			"notes": map[string]interface{}{
				"clipin_ref": req.Reference,
			},
		},
		nil,
	)
	if err != nil {
		slog.Warn("razorpay contact creation failed", "error", err)
		return nil, fmt.Errorf("create contact: %w", err)
	}

	contactID, _ := contact["id"].(string)
	slog.Info("razorpay contact created", "contact_id", contactID)

	// Step 2: Create a fund account for the UPI.
	fundAccount, err := r.client.FundAccount.Create(
		map[string]interface{}{
			"contact_id":   contactID,
			"account_type": "vpa",
			"vpa": map[string]interface{}{
				"address": req.UPITarget,
			},
		},
		nil,
	)
	if err != nil {
		slog.Warn("razorpay fund account creation failed", "error", err)
		return nil, fmt.Errorf("create fund account: %w", err)
	}

	fundAccountID, _ := fundAccount["id"].(string)
	slog.Info("razorpay fund account created", "fund_account_id", fundAccountID)

	// Step 3: Create the payout.
	payoutData := map[string]interface{}{
		"fund_account_id": fundAccountID,
		"amount":          req.Amount,
		"currency":        "INR",
		"mode":            "UPI",
		"purpose":         "payout",
		"queue_if_low_balance": true,
		"reference_id":    req.Reference,
		"narration":       "ClipIN payout",
	}
	if r.accountNumber != "" {
		payoutData["account_number"] = r.accountNumber
	}

	payout, err := r.client.Request.Post(
		"/"+constants.VERSION_V1+"/payouts",
		payoutData,
		nil,
	)
	if err != nil {
		slog.Warn("razorpay payout creation failed", "error", err)
		return nil, fmt.Errorf("create payout: %w", err)
	}

	payoutID, _ := payout["id"].(string)
	status := MapPayoutStatus(payout["status"])

	slog.Info("razorpay payout created",
		"payout_id", payoutID,
		"status", status,
	)

	return &TransferResponse{
		TransferID: payoutID,
		Status:     status,
		Reference:  req.Reference,
	}, nil
}

func (r *RazorpayProvider) GetTransferStatus(ctx context.Context, transferID string) (*TransferResponse, error) {
	payout, err := r.client.Payout.Fetch(transferID, nil, nil)
	if err != nil {
		slog.Warn("razorpay payout fetch failed", "error", err)
		return nil, fmt.Errorf("fetch payout: %w", err)
	}

	status := MapPayoutStatus(payout["status"])
	reference, _ := payout["reference_id"].(string)

	return &TransferResponse{
		TransferID: transferID,
		Status:     status,
		Reference:  reference,
	}, nil
}

// MapPayoutStatus maps Razorpay payout status to our internal status.
func MapPayoutStatus(razorpayStatus interface{}) string {
	status, _ := razorpayStatus.(string)
	switch status {
	case "processed":
		return "completed"
	case "processing":
		return "processing"
	case "pending":
		return "pending"
	case "reversed", "failed":
		return "failed"
	default:
		return "pending"
	}
}

// RedactUPI returns a masked version of a UPI ID for safe logging.
func RedactUPI(upi string) string {
	if len(upi) <= 4 {
		return "****"
	}
	return upi[:2] + "****" + upi[len(upi)-2:]
}

// NewRazorpayProviderOrStub returns a real provider when keys are set, or a stub for dev.
func NewRazorpayProviderOrStub(keyID, keySecret, accountNumber string) PayoutProvider {
	if keyID == "" || keySecret == "" {
		slog.Info("razorpay keys not configured, using stub provider")
		return &RazorpayStub{}
	}
	slog.Info("razorpay provider initialized", "key_id_prefix", keyID[:min(8, len(keyID))]+"...")
	return NewRazorpayProvider(keyID, keySecret, accountNumber)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
