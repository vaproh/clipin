package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	sqlc "clipin/apps/api/internal/db/sqlc"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// UserStore is the subset of the user persistence layer the session
// middleware and user handlers need. *sqlc.Queries satisfies it; tests
// provide a fake.
type UserStore interface {
	GetUserByID(ctx context.Context, id string) (sqlc.User, error)
	CreateUser(ctx context.Context, arg sqlc.CreateUserParams) (sqlc.User, error)
	UpdateUser(ctx context.Context, arg sqlc.UpdateUserParams) (sqlc.User, error)
	UpdateUserRole(ctx context.Context, arg sqlc.UpdateUserRoleParams) (sqlc.User, error)
}

// SessionMiddleware runs after AuthMiddleware. It looks up the Clerk user in
// the database, creating the row on first login, and stores the persisted
// record in the context for downstream handlers.
func SessionMiddleware(store UserStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := UserIDFromContext(r.Context())
			// No verified identity (e.g. auth disabled in development):
			// pass through without a user so protected handlers can 401.
			if !ok {
				next.ServeHTTP(w, r)
				return
			}

			user, err := store.GetUserByID(r.Context(), userID)
			if err != nil {
				// Treat "not found" as a first login and create the user.
				// Any other error is a real database failure.
				if !isNotFound(err) {
					slog.Error("session: load user", "user_id", userID, "error", err)
					http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
					return
				}

				user, err = store.CreateUser(r.Context(), sqlc.CreateUserParams{
					ID:          userID,
					Email:       emailFromClaims(r.Context()),
					DisplayName: displayNameFromClaims(r.Context()),
					Role:        "clipper",
				})
				if err != nil {
					slog.Error("session: create user", "user_id", userID, "error", err)
					http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
					return
				}
			}

			next.ServeHTTP(w, r.WithContext(ContextWithUser(r.Context(), &user)))
		})
	}
}

func emailFromClaims(ctx context.Context) string {
	if claims, ok := ctx.Value(claimsContextKey).(*Claims); ok {
		return claims.Email
	}
	return ""
}

func displayNameFromClaims(ctx context.Context) pgtype.Text {
	var name pgtype.Text
	if claims, ok := ctx.Value(claimsContextKey).(*Claims); ok && claims.Name != "" {
		name.String = claims.Name
		name.Valid = true
	}
	return name
}

func isNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
