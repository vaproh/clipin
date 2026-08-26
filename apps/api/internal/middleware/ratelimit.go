package middleware

import (
	"net/http"
	"sync"
	"time"
)

// RateLimiter is a simple in-memory sliding-window rate limiter.
// Suitable for single-instance deployments.
type RateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
}

type bucket struct {
	tokens   int
	refillAt time.Time
}

// NewRateLimiter creates a new in-memory rate limiter.
func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{buckets: make(map[string]*bucket)}
	// Background cleanup every 5 minutes.
	go rl.cleanup()
	return rl
}

const refillInterval = time.Minute

func (rl *RateLimiter) allow(key string, limit int) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, ok := rl.buckets[key]
	if !ok || now.After(b.refillAt) {
		rl.buckets[key] = &bucket{tokens: limit - 1, refillAt: now.Add(refillInterval)}
		return true
	}
	if b.tokens <= 0 {
		return false
	}
	b.tokens--
	return true
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for k, b := range rl.buckets {
			if now.After(b.refillAt.Add(refillInterval)) {
				delete(rl.buckets, k)
			}
		}
		rl.mu.Unlock()
	}
}

// Middleware returns HTTP middleware that applies the given per-key rate limit.
// keyFunc extracts the rate-limit key from the request (e.g., IP or user ID).
func (rl *RateLimiter) Middleware(limit int, keyFunc func(r *http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := keyFunc(r)
			if !rl.allow(key, limit) {
				w.Header().Set("Retry-After", "60")
				http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// IPKey returns the client IP from RemoteAddr or X-Forwarded-For.
func IPKey(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the chain.
		for i := 0; i < len(xff); i++ {
			if xff[i] == ',' {
				return xff[:i]
			}
		}
		return xff
	}
	// Strip port from RemoteAddr.
	addr := r.RemoteAddr
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			return addr[:i]
		}
	}
	return addr
}
