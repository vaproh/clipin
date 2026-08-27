package service

import (
	"context"
	"fmt"
	"time"

	r "clipin/apps/api/internal/redis"
)

const campaignAnalyticsCacheTTL = 30 * time.Second

// CampaignAnalyticsCache wraps Redis for campaign analytics response caching.
type CampaignAnalyticsCache struct {
	redis *r.Client
}

func NewCampaignAnalyticsCache(redis *r.Client) *CampaignAnalyticsCache {
	return &CampaignAnalyticsCache{redis: redis}
}

func campaignAnalyticsKey(campaignID string) string {
	return fmt.Sprintf("campaign:analytics:%s", campaignID)
}

// Get returns cached analytics JSON bytes, or nil on miss.
func (c *CampaignAnalyticsCache) Get(ctx context.Context, campaignID string) ([]byte, error) {
	if c.redis == nil {
		return nil, nil
	}
	return c.redis.Get(ctx, campaignAnalyticsKey(campaignID))
}

// Set stores analytics JSON bytes.
func (c *CampaignAnalyticsCache) Set(ctx context.Context, campaignID string, data []byte) error {
	if c.redis == nil {
		return nil
	}
	return c.redis.Set(ctx, campaignAnalyticsKey(campaignID), data, campaignAnalyticsCacheTTL)
}
