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

// campaignCacheEntry stores both campaigns and total count together.
type campaignCacheEntry struct {
	Campaigns []sqlc.Campaign `json:"campaigns"`
	Total     int64           `json:"total"`
}

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

// GetList returns cached campaigns and total count for a filter+page combination, or nil on miss.
func (cc *CampaignCache) GetList(ctx context.Context, platform string, maxCpm, minBudget int, search string, page int) ([]sqlc.Campaign, int64, error) {
	if cc.redis == nil {
		return nil, 0, nil
	}
	data, err := cc.redis.Get(ctx, cacheKey(platform, maxCpm, minBudget, search, page))
	if err != nil || len(data) == 0 {
		return nil, 0, nil // cache miss, not an error
	}
	var entry campaignCacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, 0, nil
	}
	return entry.Campaigns, entry.Total, nil
}

// SetList stores campaigns and total count for a filter+page combination.
func (cc *CampaignCache) SetList(ctx context.Context, platform string, maxCpm, minBudget int, search string, page int, campaigns []sqlc.Campaign, total int64) error {
	if cc.redis == nil {
		return nil
	}
	entry := campaignCacheEntry{Campaigns: campaigns, Total: total}
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	return cc.redis.Set(ctx, cacheKey(platform, maxCpm, minBudget, search, page), data, campaignCacheTTL)
}
