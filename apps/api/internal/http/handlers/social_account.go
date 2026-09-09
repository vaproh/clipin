package handlers

import (
	"context"
	"errors"
	"fmt"

	sqlc "clipin/apps/api/internal/db/sqlc"
	"clipin/apps/api/internal/service"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgtype"
)

// SocialAccountServiceInterface is the subset of SocialAccountService the handlers need.
type SocialAccountServiceInterface interface {
	ListByUser(ctx context.Context, userID string) ([]sqlc.SocialAccount, error)
	Connect(ctx context.Context, userID, platform, platformUserID, platformUsername string) (*sqlc.SocialAccount, error)
	Disconnect(ctx context.Context, userID string, accountID pgtype.UUID) error
}

type socialAccountItem struct {
	ID               string  `json:"id"`
	Platform         string  `json:"platform"`
	PlatformUserID   string  `json:"platform_user_id"`
	PlatformUsername *string `json:"platform_username,omitempty"`
	CreatedAt        string  `json:"created_at"`
}

func toSocialAccountItem(a sqlc.SocialAccount) socialAccountItem {
	item := socialAccountItem{
		ID:             fmt.Sprintf("%x", a.ID.Bytes),
		Platform:       a.Platform,
		PlatformUserID: a.PlatformUserID,
	}
	if a.PlatformUsername.Valid {
		item.PlatformUsername = &a.PlatformUsername.String
	}
	if a.CreatedAt.Valid {
		item.CreatedAt = a.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00")
	}
	return item
}

// RegisterSocialAccountHandlers registers social account endpoints on the authenticated API.
func RegisterSocialAccountHandlers(api huma.API, svc SocialAccountServiceInterface) {
	// GET /me/social-accounts - list connected accounts
	huma.Register(api, huma.Operation{
		OperationID: "list-my-social-accounts",
		Method:      "GET",
		Path:        "/me/social-accounts",
		Summary:     "List connected social accounts",
		Description: "Returns all connected social accounts for the current user.",
		Tags:        []string{"Social Accounts"},
	}, func(ctx context.Context, input *struct{}) (*listSocialAccountsOutput, error) {
		user, err := requireUser(ctx)
		if err != nil {
			return nil, err
		}

		accounts, err := svc.ListByUser(ctx, user.ID)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to list social accounts")
		}

		resp := &listSocialAccountsOutput{}
		resp.Body.Accounts = make([]socialAccountItem, 0, len(accounts))
		for _, a := range accounts {
			resp.Body.Accounts = append(resp.Body.Accounts, toSocialAccountItem(a))
		}
		return resp, nil
	})

	// POST /me/social-accounts - connect account (stub)
	huma.Register(api, huma.Operation{
		OperationID: "connect-social-account",
		Method:      "POST",
		Path:        "/me/social-accounts",
		Summary:     "Connect a social account",
		Description: "Connects a social account (stub - no real OAuth).",
		Tags:        []string{"Social Accounts"},
	}, func(ctx context.Context, input *connectSocialAccountInput) (*socialAccountOutput, error) {
		user, err := requireUser(ctx)
		if err != nil {
			return nil, err
		}

		if input.Body.Platform == "" {
			return nil, huma.Error422UnprocessableEntity("platform is required")
		}
		if input.Body.PlatformUserID == "" {
			return nil, huma.Error422UnprocessableEntity("platform_user_id is required")
		}

		account, err := svc.Connect(ctx, user.ID, input.Body.Platform, input.Body.PlatformUserID, input.Body.PlatformUsername)
		if err != nil {
			if errors.Is(err, service.ErrInvalidPlatform) {
				return nil, huma.Error422UnprocessableEntity("platform must be youtube or instagram")
			}
			return nil, huma.Error500InternalServerError("failed to connect social account")
		}

		resp := &socialAccountOutput{}
		resp.Body = toSocialAccountItem(*account)
		return resp, nil
	})

	// DELETE /me/social-accounts/{id} - disconnect account
	huma.Register(api, huma.Operation{
		OperationID: "disconnect-social-account",
		Method:      "DELETE",
		Path:        "/me/social-accounts/{id}",
		Summary:     "Disconnect a social account",
		Description: "Disconnects and removes a social account.",
		Tags:        []string{"Social Accounts"},
	}, func(ctx context.Context, input *struct {
		ID string `path:"id" doc:"Social Account UUID"`
	}) (*socialAccountActionOutput, error) {
		user, err := requireUser(ctx)
		if err != nil {
			return nil, err
		}

		accountID, err := parseSocialAccountID(input.ID)
		if err != nil {
			return nil, err
		}

		if err := svc.Disconnect(ctx, user.ID, accountID); err != nil {
			return nil, huma.Error500InternalServerError("failed to disconnect social account")
		}

		return &socialAccountActionOutput{Body: struct {
			Status string `json:"status"`
		}{Status: "ok"}}, nil
	})
}

type connectSocialAccountInput struct {
	Body struct {
		Platform         string `json:"platform" doc:"Platform (youtube, instagram)"`
		PlatformUserID   string `json:"platform_user_id" doc:"Platform user/channel ID"`
		PlatformUsername string `json:"platform_username,omitempty" doc:"Platform username"`
	}
}

type listSocialAccountsOutput struct {
	Body struct {
		Accounts []socialAccountItem `json:"accounts"`
	}
}

type socialAccountOutput struct {
	Body socialAccountItem
}

type socialAccountActionOutput struct {
	Body struct {
		Status string `json:"status"`
	}
}

func parseSocialAccountID(raw string) (pgtype.UUID, error) {
	var id pgtype.UUID
	if err := id.Scan(raw); err != nil {
		return pgtype.UUID{}, huma.Error422UnprocessableEntity("invalid social account id")
	}
	return id, nil
}
