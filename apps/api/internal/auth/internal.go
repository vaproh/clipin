package auth

import (
	"crypto/subtle"
	"net/http"
)

// InternalAuthMiddleware returns middleware that verifies the X-Verifier-Key
// header against the configured API key. Requests pass through untouched when
// the key is empty (development mode).
func InternalAuthMiddleware(apiKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if apiKey == "" {
				next.ServeHTTP(w, r)
				return
			}

			provided := r.Header.Get("X-Verifier-Key")
			if provided == "" {
				http.Error(w, `{"error":"missing X-Verifier-Key header"}`, http.StatusUnauthorized)
				return
			}

			if subtle.ConstantTimeCompare([]byte(provided), []byte(apiKey)) != 1 {
				http.Error(w, `{"error":"invalid verifier key"}`, http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
