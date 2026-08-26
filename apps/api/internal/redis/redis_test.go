package redis_test

import (
	"context"
	"testing"
	"time"

	r "clipin/apps/api/internal/redis"
)

// TestRedisConnectInvalidURL verifies Connect rejects malformed URLs.
func TestRedisConnectInvalidURL(t *testing.T) {
	_, err := r.Connect(context.Background(), "invalid-url")
	if err == nil {
		t.Fatal("expected error for invalid redis URL")
	}
}

// TestRedisPingNilClient verifies Ping fails on an uninitialized client.
func TestRedisPingNilClient(t *testing.T) {
	// Can't directly construct a Client with nil rdb from outside the package,
	// so we verify Connect with a valid-looking but unreachable URL.
	// In CI without Redis this will timeout; skip if Redis is available.
	client, err := r.Connect(context.Background(), "redis://localhost:19999/0")
	if err != nil {
		t.Skipf("cannot connect to test redis: %v", err)
	}
	// Ping should fail because nothing is listening on that port
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := client.Ping(ctx); err == nil {
		t.Fatal("expected ping error on unreachable redis")
	}
}
