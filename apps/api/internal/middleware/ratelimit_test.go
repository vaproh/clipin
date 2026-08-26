package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	"clipin/apps/api/internal/middleware"
)

func TestRateLimiter_AllowsUnderLimit(t *testing.T) {
	rl := middleware.NewRateLimiter()
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := rl.Middleware(10, func(r *http.Request) string { return "ip1" })(inner)

	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "ip1:1234"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i, rec.Code)
		}
	}
}

func TestRateLimiter_BlocksOverLimit(t *testing.T) {
	rl := middleware.NewRateLimiter()
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := rl.Middleware(5, func(r *http.Request) string { return "ip2" })(inner)

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "ip2:1234"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i, rec.Code)
		}
	}

	// 6th request should be rate limited
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "ip2:1234"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rec.Code)
	}
}

func TestRateLimiter_DifferentKeys(t *testing.T) {
	rl := middleware.NewRateLimiter()
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := rl.Middleware(2, middleware.IPKey)(inner)

	// Exhaust IP "10.0.0.1"
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.RemoteAddr = "10.0.0.1:1234"
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.RemoteAddr = "10.0.0.1:1234"
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	req3 := httptest.NewRequest(http.MethodGet, "/", nil)
	req3.RemoteAddr = "10.0.0.1:1234"
	rec3 := httptest.NewRecorder()
	handler.ServeHTTP(rec3, req3)

	// Should be blocked now
	if rec3.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rec3.Code)
	}

	// Different IP should still work
	req4 := httptest.NewRequest(http.MethodGet, "/", nil)
	req4.RemoteAddr = "10.0.0.2:1234"
	rec4 := httptest.NewRecorder()
	handler.ServeHTTP(rec4, req4)
	if rec4.Code != http.StatusOK {
		t.Fatalf("different key: expected 200, got %d", rec4.Code)
	}
}

func TestRateLimiter_SetsRetryAfter(t *testing.T) {
	rl := middleware.NewRateLimiter()
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := rl.Middleware(1, func(r *http.Request) string { return "ip3" })(inner)

	// First request passes
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.RemoteAddr = "ip3:1234"
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)

	// Second request blocked
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.RemoteAddr = "ip3:1234"
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	if rec2.Header().Get("Retry-After") != "60" {
		t.Errorf("expected Retry-After: 60, got %q", rec2.Header().Get("Retry-After"))
	}
}

func TestRateLimiter_ConcurrentAccess(t *testing.T) {
	rl := middleware.NewRateLimiter()
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := rl.Middleware(50, func(r *http.Request) string { return "ip4" })(inner)

	var wg sync.WaitGroup
	var blocked atomic.Int32

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = "ip4:1234"
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code == http.StatusTooManyRequests {
				blocked.Add(1)
			}
		}()
	}
	wg.Wait()

	// At least some should be blocked
	if blocked.Load() == 0 {
		t.Errorf("expected some requests to be blocked, but none were")
	}
}

func TestIPKey_XForwardedFor(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "1.2.3.4, 5.6.7.8")
	key := middleware.IPKey(req)
	if key != "1.2.3.4" {
		t.Errorf("expected 1.2.3.4, got %q", key)
	}
}

func TestIPKey_RemoteAddr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1:8080"
	key := middleware.IPKey(req)
	if key != "192.168.1.1" {
		t.Errorf("expected 192.168.1.1, got %q", key)
	}
}
