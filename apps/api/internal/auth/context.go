package auth

import (
	"context"

	sqlc "clipin/apps/api/internal/db/sqlc"
)

type contextKey string

const (
	claimsContextKey contextKey = "clipin.auth.claims"
	userContextKey   contextKey = "clipin.auth.user"
)

// ContextWithClaims returns a copy of ctx carrying the verified JWT claims.
func ContextWithClaims(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, claimsContextKey, claims)
}

// UserIDFromContext returns the Clerk user ID (sub claim) stored by the
// auth middleware, if present.
func UserIDFromContext(ctx context.Context) (string, bool) {
	claims, ok := ctx.Value(claimsContextKey).(*Claims)
	if !ok {
		return "", false
	}
	return claims.Sub, claims.Sub != ""
}

// UserContext wraps the persisted user record for downstream handlers.
type UserContext struct {
	User *sqlc.User
}

// ContextWithUser returns a copy of ctx carrying the persisted user record.
func ContextWithUser(ctx context.Context, user *sqlc.User) context.Context {
	return context.WithValue(ctx, userContextKey, &UserContext{User: user})
}

// UserFromContext returns the persisted user record stored by the session
// middleware, if present.
func UserFromContext(ctx context.Context) (*sqlc.User, bool) {
	uc, ok := ctx.Value(userContextKey).(*UserContext)
	if !ok || uc == nil || uc.User == nil {
		return nil, false
	}
	return uc.User, true
}
