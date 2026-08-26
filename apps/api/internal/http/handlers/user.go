package handlers

import (
	"context"
	"errors"

	"clipin/apps/api/internal/auth"
	sqlc "clipin/apps/api/internal/db/sqlc"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// UserOutput is the user payload returned to clients.
type UserOutput struct {
	Body struct {
		ID          string  `json:"id"`
		Email       string  `json:"email"`
		DisplayName *string `json:"display_name,omitempty"`
		Role        string  `json:"role"`
		CreatedAt   string  `json:"created_at"`
	}
}

func toUserOutput(u sqlc.User) *UserOutput {
	out := &UserOutput{}
	out.Body.ID = u.ID
	out.Body.Email = u.Email
	out.Body.Role = u.Role
	out.Body.CreatedAt = u.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00")
	if u.DisplayName.Valid {
		name := u.DisplayName.String
		out.Body.DisplayName = &name
	}
	return out
}

func requireUser(ctx context.Context) (*sqlc.User, error) {
	user, ok := auth.UserFromContext(ctx)
	if !ok {
		return nil, huma.Error401Unauthorized("not authenticated")
	}
	return user, nil
}

// RegisterUserHandlers registers the /me endpoints on the authenticated API.
func RegisterUserHandlers(api huma.API, store auth.UserStore) {
	huma.Register(api, huma.Operation{
		OperationID: "get-me",
		Method:      "GET",
		Path:        "/me",
		Summary:     "Get current user",
		Tags:        []string{"Users"},
	}, func(ctx context.Context, input *struct{}) (*UserOutput, error) {
		user, err := requireUser(ctx)
		if err != nil {
			return nil, err
		}
		return toUserOutput(*user), nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "update-me",
		Method:      "PATCH",
		Path:        "/me",
		Summary:     "Update current user",
		Tags:        []string{"Users"},
	}, func(ctx context.Context, input *UpdateMeInput) (*UserOutput, error) {
		user, err := requireUser(ctx)
		if err != nil {
			return nil, err
		}

		updated, err := store.UpdateUser(ctx, sqlc.UpdateUserParams{
			ID:          user.ID,
			DisplayName: pgtype.Text{Valid: true, String: input.Body.DisplayName},
			Email:       user.Email,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, huma.Error401Unauthorized("user no longer exists")
			}
			return nil, huma.Error500InternalServerError("could not update user")
		}
		return toUserOutput(updated), nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "set-my-role",
		Method:      "POST",
		Path:        "/me/role",
		Summary:     "Set current user role",
		Tags:        []string{"Users"},
	}, func(ctx context.Context, input *SetRoleInput) (*UserOutput, error) {
		user, err := requireUser(ctx)
		if err != nil {
			return nil, err
		}
		if input.Body.Role != "clipper" && input.Body.Role != "owner" {
			return nil, huma.Error422UnprocessableEntity("role must be 'clipper' or 'owner'")
		}

		updated, err := store.UpdateUserRole(ctx, sqlc.UpdateUserRoleParams{
			ID:   user.ID,
			Role: input.Body.Role,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, huma.Error401Unauthorized("user no longer exists")
			}
			return nil, huma.Error500InternalServerError("could not update role")
		}
		return toUserOutput(updated), nil
	})
}

type UpdateMeInput struct {
	Body struct {
		DisplayName string `json:"display_name"`
	}
}

type SetRoleInput struct {
	Body struct {
		Role string `json:"role"`
	}
}