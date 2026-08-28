package handlers

import (
	"context"
	"database/sql"
	"errors"

	"clipin/apps/api/internal/auth"
	"clipin/apps/api/internal/service"

	"github.com/danielgtaylor/huma/v2"
)

// ReputationServiceInterface is what the handler needs from the reputation service.
type ReputationServiceInterface interface {
	GetReputation(ctx context.Context, userID string) (*service.ClipperReputation, error)
}

// RegisterReputationHandlers registers reputation endpoints.
func RegisterReputationHandlers(api huma.API, svc ReputationServiceInterface) {
	// GET /me/reputation (authenticated)
	huma.Register(api, huma.Operation{
		OperationID: "get-my-reputation",
		Method:      "GET",
		Path:        "/me/reputation",
		Summary:     "Get my reputation",
		Description: "Returns the authenticated clipper's reputation tier and stats.",
		Tags:        []string{"Reputation"},
	}, func(ctx context.Context, input *struct{}) (*reputationOutput, error) {
		userID := getUserID(ctx)
		if userID == "" {
			return nil, huma.Error401Unauthorized("authentication required")
		}

		rep, err := svc.GetReputation(ctx, userID)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to get reputation")
		}

		return toReputationOutput(rep), nil
	})

	// GET /clippers/{id}/reputation (public)
	huma.Register(api, huma.Operation{
		OperationID: "get-clipper-reputation",
		Method:      "GET",
		Path:        "/clippers/{id}/reputation",
		Summary:     "Get clipper reputation",
		Description: "Returns a public clipper's reputation tier and stats.",
		Tags:        []string{"Reputation", "Clippers"},
	}, func(ctx context.Context, input *struct {
		ID string `path:"id" doc:"Clipper user ID"`
	}) (*reputationOutput, error) {
		rep, err := svc.GetReputation(ctx, input.ID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, huma.Error404NotFound("clipper not found")
			}
			return nil, huma.Error500InternalServerError("failed to get reputation")
		}

		return toReputationOutput(rep), nil
	})
}

type tierOutput struct {
	Name            string  `json:"name"`
	Label           string  `json:"label"`
	MinEarnings     int64   `json:"min_earnings"`
	NextTier        *string `json:"next_tier,omitempty"`
	NextMinEarnings *int64  `json:"next_min_earnings,omitempty"`
}

type reputationOutput struct {
	Body struct {
		UserID                string      `json:"user_id"`
		Tier                  tierOutput  `json:"tier"`
		TotalEarnings         int64       `json:"total_earnings"`
		TotalSubmissions      int         `json:"total_submissions"`
		CampaignsParticipated int         `json:"campaigns_participated"`
		MemberSince           string      `json:"member_since"`
	}
}

func toReputationOutput(rep *service.ClipperReputation) *reputationOutput {
	out := &reputationOutput{}
	out.Body.UserID = rep.UserID
	out.Body.Tier = tierOutput{
		Name:            rep.Tier.Name,
		Label:           rep.Tier.Label,
		MinEarnings:     rep.Tier.MinEarnings,
		NextTier:        rep.Tier.NextTier,
		NextMinEarnings: rep.Tier.NextMin,
	}
	out.Body.TotalEarnings = rep.TotalEarnings
	out.Body.TotalSubmissions = rep.TotalSubmissions
	out.Body.CampaignsParticipated = rep.CampaignsParticipated
	out.Body.MemberSince = rep.MemberSince
	return out
}

// getUserID extracts the user ID from context. This is set by auth middleware.
func getUserID(ctx context.Context) string {
	if uid, ok := auth.UserIDFromContext(ctx); ok {
		return uid
	}
	return ""
}
