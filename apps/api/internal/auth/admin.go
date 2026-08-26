package auth

import (
	"net/http"
)

// AdminOnly returns middleware that restricts access to users with the "admin"
// role. Must be placed after SessionMiddleware so the user is in context.
func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := UserFromContext(r.Context())
		if !ok {
			http.Error(w, `{"error":"not authenticated"}`, http.StatusUnauthorized)
			return
		}
		if user.Role != "admin" {
			http.Error(w, `{"error":"admin role required"}`, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
