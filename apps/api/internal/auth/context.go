package auth

import "context"

type contextKey string

const claimsContextKey contextKey = "clipin.auth.claims"

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
