package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	sqlc "clipin/apps/api/internal/db/sqlc"
	r "clipin/apps/api/internal/redis"
)

const campaignCacheTTL = 30 * time.Second

// CampaignCache wraps Redis with typed cache operations for campaign listings.
type CampaignCache struct {
	redis *r.Client
}

func NewCampaignCache(redis *r.Client) *CampaignCache {
	return &CampaignCache{redis: redis}
}

func cacheKey(platform string, maxCpm int, minBudget int, search string, page int) string {
	return fmt.Sprintf("campaigns:list:%s:%d:%d:%s:%d", platform, maxCpm, minBudget, search, page)
}

// GetList returns cached campaigns for a filter+page combination, or nil on miss.
func (cc *CampaignCache) GetList(ctx context.Context, platform string, maxCpm, minBudget int, search string, page int) ([]sqlc.Campaign, error) {
	if cc.redis == nil {
		return nil, nil
	}
	data, err := cc.redis.Get(ctx, cacheKey(platform, maxCpm, minBudget, search, page))
	if err != nil || len(data) == 0 {
		return nil, nil // cache miss, not an error
	}
	var campaigns []sqlc.Campaign
	if err := json.Unmarshal(data, &campaigns); err != nil {
		return nil, nil
	}
	return campaigns, nil
}

// SetList stores campaigns for a filter+page combination.
func (cc *CampaignCache) SetList(ctx context.Context, platform string, maxCpm, minBudget int, search string, page int, campaigns []sqlc.Campaign) error {
	if cc.redis == nil {
		return nil
	}
	data, err := json.Marshal(campaigns)
	if err != nil {
		return err
	}
	return cc.redis.Set(ctx, cacheKey(platform, maxCpm, minBudget, search, page), data, campaignCacheTTL)
}
