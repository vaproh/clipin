package handlers

import (
	"context"
	"crypto/rand"
	"fmt"

	sqlc "clipin/apps/api/internal/db/sqlc"
	"clipin/apps/api/internal/service"

	"github.com/danielgtaylor/huma/v2"
)

// PayoutServiceInterface is the subset of PayoutService the handlers need.
type PayoutServiceInterface interface {
	UpdateUPI(ctx context.Context, userID, upiID string) (*sqlc.User, error)
	RequestPayout(ctx context.Context, userID string, amount int64, idempotencyKey string) (*sqlc.PayoutRequest, error)
	ListMyPayouts(ctx context.Context, userID string) ([]sqlc.PayoutRequest, error)
}

// --- Request/response types ---

type updateUPIInput struct {
	Body struct {
		UpiID string `json:"upi_id" doc:"UPI ID (e.g. user@upi)"`
	}
}

type userWithUPIOutput struct {
	Body struct {
		ID          string  `json:"id"`
		Email       string  `json:"email"`
		DisplayName *string `json:"display_name,omitempty"`
		Role        string  `json:"role"`
		UpiID       *string `json:"upi_id,omitempty"`
		CreatedAt   string  `json:"created_at"`
	}
}

func toUserWithUPIOutput(u sqlc.User) *userWithUPIOutput {
	out := &userWithUPIOutput{}
	out.Body.ID = u.ID
	out.Body.Email = u.Email
	out.Body.Role = u.Role
	out.Body.CreatedAt = u.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00")
	if u.DisplayName.Valid {
		name := u.DisplayName.String
		out.Body.DisplayName = &name
	}
	if u.UpiID.Valid {
		upi := u.UpiID.String
		out.Body.UpiID = &upi
	}
	return out
}

type requestPayoutInput struct {
	Body struct {
		Amount int64 `json:"amount" doc:"Amount in paise (minimum 50000 = ₹500)"`
	}
}

type payoutRequestItem struct {
	ID             string  `json:"id"`
	ClipperID      string  `json:"clipper_id"`
	Amount         int32   `json:"amount"`
	UpiID          string  `json:"upi_id"`
	Status         string  `json:"status"`
	ProviderRef    *string `json:"provider_ref,omitempty"`
	FailureReason  *string `json:"failure_reason,omitempty"`
	IdempotencyKey string  `json:"idempotency_key"`
	CreatedAt      string  `json:"created_at"`
	ProcessedAt    *string `json:"processed_at,omitempty"`
}

func toPayoutRequestItem(p sqlc.PayoutRequest) payoutRequestItem {
	item := payoutRequestItem{
		ID:             fmt.Sprintf("%x", p.ID.Bytes),
		ClipperID:      p.ClipperID,
		Amount:         p.Amount,
		UpiID:          p.UpiID,
		Status:         p.Status,
		IdempotencyKey: p.IdempotencyKey,
	}
	if p.ProviderRef.Valid {
		item.ProviderRef = &p.ProviderRef.String
	}
	if p.FailureReason.Valid {
		item.FailureReason = &p.FailureReason.String
	}
	if p.CreatedAt.Valid {
		item.CreatedAt = p.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00")
	}
	if p.ProcessedAt.Valid {
		s := p.ProcessedAt.Time.Format("2006-01-02T15:04:05Z07:00")
		item.ProcessedAt = &s
	}
	return item
}

type requestPayoutOutput struct {
	Body payoutRequestItem
}

type listPayoutsOutput struct {
	Body struct {
		Payouts []payoutRequestItem `json:"payouts"`
	}
}

// RegisterPayoutHandlers registers payout endpoints on the authenticated API.
func RegisterPayoutHandlers(api huma.API, svc PayoutServiceInterface) {
	// PATCH /me/upi - Update UPI ID
	huma.Register(api, huma.Operation{
		OperationID: "update-my-upi",
		Method:      "PATCH",
		Path:        "/me/upi",
		Summary:     "Update UPI ID",
		Description: "Sets the UPI ID for receiving payouts.",
		Tags:        []string{"Payouts"},
	}, func(ctx context.Context, input *updateUPIInput) (*userWithUPIOutput, error) {
		user, err := requireUser(ctx)
		if err != nil {
			return nil, err
		}
		if input.Body.UpiID == "" {
			return nil, huma.Error422UnprocessableEntity("upi_id is required")
		}
		updated, err := svc.UpdateUPI(ctx, user.ID, input.Body.UpiID)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to update UPI ID")
		}
		return toUserWithUPIOutput(*updated), nil
	})

	// POST /me/payouts - Request payout
	huma.Register(api, huma.Operation{
		OperationID: "request-payout",
		Method:      "POST",
		Path:        "/me/payouts",
		Summary:     "Request payout",
		Description: "Creates a payout request. Requires UPI ID to be set and sufficient available balance.",
		Tags:        []string{"Payouts"},
	}, func(ctx context.Context, input *requestPayoutInput) (*requestPayoutOutput, error) {
		user, err := requireUser(ctx)
		if err != nil {
			return nil, err
		}
		if input.Body.Amount <= 0 {
			return nil, huma.Error422UnprocessableEntity("amount must be positive")
		}

		// Generate idempotency key.
		var idemBytes [16]byte
		rand.Read(idemBytes[:])
		idempotencyKey := fmt.Sprintf("payout:%s:%d:%x", user.ID, input.Body.Amount, idemBytes)

		pr, err := svc.RequestPayout(ctx, user.ID, input.Body.Amount, idempotencyKey)
		if err != nil {
			switch err {
			case service.ErrNoUPIID:
				return nil, huma.Error422UnprocessableEntity("UPI ID is not set. Update via PATCH /me/upi first.")
			case service.ErrBelowThreshold:
				return nil, huma.Error422UnprocessableEntity("payout amount must be at least ₹500 (50000 paise)")
			case service.ErrInsufficientFunds:
				return nil, huma.Error422UnprocessableEntity("insufficient available balance")
			default:
				return nil, huma.Error500InternalServerError("failed to create payout request")
			}
		}
		item := toPayoutRequestItem(*pr)
		resp := &requestPayoutOutput{Body: item}
		return resp, nil
	})

	// GET /me/payouts - List my payouts
	huma.Register(api, huma.Operation{
		OperationID: "list-my-payouts",
		Method:      "GET",
		Path:        "/me/payouts",
		Summary:     "List my payouts",
		Description: "Returns all payout requests for the current user.",
		Tags:        []string{"Payouts"},
	}, func(ctx context.Context, input *struct{}) (*listPayoutsOutput, error) {
		user, err := requireUser(ctx)
		if err != nil {
			return nil, err
		}
		payouts, err := svc.ListMyPayouts(ctx, user.ID)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to list payouts")
		}
		resp := &listPayoutsOutput{}
		resp.Body.Payouts = make([]payoutRequestItem, 0, len(payouts))
		for _, p := range payouts {
			resp.Body.Payouts = append(resp.Body.Payouts, toPayoutRequestItem(p))
		}
		return resp, nil
	})
}

// --- Razorpay webhook ---

type razorpayWebhookInput struct {
	Headers struct {
		XRazorpaySignature string `header:"X-Razorpay-Signature" doc:"Razorpay webhook signature"`
		XRazorpayEvent     string `header:"X-Razorpay-Event" doc:"Razorpay event type"`
	}
}

type razorpayWebhookOutput struct {
	Body struct {
		Status string `json:"status"`
	}
}

// RegisterWebhookHandlers registers webhook endpoints on the public API.
func RegisterWebhookHandlers(api huma.API) {
	// POST /webhooks/razorpay - Razorpay webhook handler
	huma.Register(api, huma.Operation{
		OperationID: "razorpay-webhook",
		Method:      "POST",
		Path:        "/webhooks/razorpay",
		Summary:     "Razorpay webhook",
		Description: "Receives Razorpay webhook events. Signature verification is structured but disabled until keys are configured.",
		Tags:        []string{"Webhooks"},
	}, func(ctx context.Context, input *razorpayWebhookInput) (*razorpayWebhookOutput, error) {
		// TODO: Verify signature with RAZORPAY_WEBHOOK_SECRET when keys exist.
		// For now, log the event and acknowledge.
		_ = input.Headers.XRazorpaySignature
		_ = input.Headers.XRazorpayEvent

		resp := &razorpayWebhookOutput{}
		resp.Body.Status = "ok"
		return resp, nil
	})
}
