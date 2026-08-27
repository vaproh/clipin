package service

import (
	"context"
	"fmt"
	"time"

	r "clipin/apps/api/internal/redis"
)

const leaderboardCacheTTL = 60 * time.Second

// LeaderboardCache wraps Redis for leaderboard response caching.
type LeaderboardCache struct {
	redis *r.Client
}

func NewLeaderboardCache(redis *r.Client) *LeaderboardCache {
	return &LeaderboardCache{redis: redis}
}

func leaderboardKey(sort string, limit, offset int) string {
	return fmt.Sprintf("leaderboard:%s:%d:%d", sort, limit, offset)
}

// Get returns cached leaderboard JSON bytes, or nil on miss.
func (c *LeaderboardCache) Get(ctx context.Context, sort string, limit, offset int) ([]byte, error) {
	if c.redis == nil {
		return nil, nil
	}
	return c.redis.Get(ctx, leaderboardKey(sort, limit, offset))
}

// Set stores leaderboard JSON bytes.
func (c *LeaderboardCache) Set(ctx context.Context, sort string, limit, offset int, data []byte) error {
	if c.redis == nil {
		return nil
	}
	return c.redis.Set(ctx, leaderboardKey(sort, limit, offset), data, leaderboardCacheTTL)
}
