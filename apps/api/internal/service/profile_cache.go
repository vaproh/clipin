package service

import (
	"context"
	"fmt"
	"time"

	r "clipin/apps/api/internal/redis"
)

const clipperProfileCacheTTL = 60 * time.Second

// ClipperProfileCache wraps Redis for clipper profile response caching.
type ClipperProfileCache struct {
	redis *r.Client
}

func NewClipperProfileCache(redis *r.Client) *ClipperProfileCache {
	return &ClipperProfileCache{redis: redis}
}

func clipperProfileKey(id string) string {
	return fmt.Sprintf("clipper:profile:%s", id)
}

// Get returns cached profile JSON bytes, or nil on miss.
func (c *ClipperProfileCache) Get(ctx context.Context, id string) ([]byte, error) {
	if c.redis == nil {
		return nil, nil
	}
	return c.redis.Get(ctx, clipperProfileKey(id))
}

// Set stores profile JSON bytes.
func (c *ClipperProfileCache) Set(ctx context.Context, id string, data []byte) error {
	if c.redis == nil {
		return nil
	}
	return c.redis.Set(ctx, clipperProfileKey(id), data, clipperProfileCacheTTL)
}
