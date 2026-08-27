//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"clipin/apps/api/internal/redis"
)

func TestRedisConnectAndPing(t *testing.T) {
	redisURL := "redis://localhost:6380"
	client, err := redis.Connect(context.Background(), redisURL)
	if err != nil {
		t.Skipf("Redis not available at %s: %v", redisURL, err)
	}
	defer client.Close()

	if err := client.Ping(context.Background()); err != nil {
		t.Skipf("Redis not reachable: %v", err)
	}
}

func TestRedisSetGetDelete(t *testing.T) {
	redisURL := "redis://localhost:6380"
	client, err := redis.Connect(context.Background(), redisURL)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}
	defer client.Close()

	if err := client.Ping(context.Background()); err != nil {
		t.Skipf("Redis not reachable: %v", err)
	}

	key := "integration_test:setget:" + time.Now().Format("20060102150405.000000000")
	val := []byte("hello-integration")

	// Set
	if err := client.Set(context.Background(), key, val, 30*time.Second); err != nil {
		t.Fatalf("Redis Set: %v", err)
	}

	// Get
	got, err := client.Get(context.Background(), key)
	if err != nil {
		t.Fatalf("Redis Get: %v", err)
	}
	if string(got) != string(val) {
		t.Errorf("Redis Get: got %q, want %q", got, val)
	}

	// Delete
	if err := client.Delete(context.Background(), key); err != nil {
		t.Fatalf("Redis Delete: %v", err)
	}

	// Verify deleted
	_, err = client.Get(context.Background(), key)
	if err == nil {
		t.Error("expected error after delete, got nil")
	}
}

func TestRedisTTLExpiry(t *testing.T) {
	redisURL := "redis://localhost:6380"
	client, err := redis.Connect(context.Background(), redisURL)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}
	defer client.Close()

	if err := client.Ping(context.Background()); err != nil {
		t.Skipf("Redis not reachable: %v", err)
	}

	key := "integration_test:ttl:" + time.Now().Format("20060102150405.000000000")
	val := []byte("expire-me")

	// Set with very short TTL
	if err := client.Set(context.Background(), key, val, 1*time.Second); err != nil {
		t.Fatalf("Redis Set: %v", err)
	}

	// Verify exists immediately
	got, err := client.Get(context.Background(), key)
	if err != nil {
		t.Fatalf("Redis Get immediately after Set: %v", err)
	}
	if string(got) != string(val) {
		t.Errorf("Redis Get: got %q, want %q", got, val)
	}

	// Wait for expiry
	time.Sleep(1500 * time.Millisecond)

	// Verify expired
	_, err = client.Get(context.Background(), key)
	if err == nil {
		t.Error("expected error after TTL expiry, got nil")
	}
}

func TestRedisNilClientBehavior(t *testing.T) {
	// Creating a client with an invalid URL should error
	_, err := redis.Connect(context.Background(), "invalid://not-a-url")
	if err == nil {
		t.Error("expected error for invalid URL, got nil")
	}
}
