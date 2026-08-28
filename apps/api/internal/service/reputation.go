package service

import (
	"context"
	"fmt"

	sqlc "clipin/apps/api/internal/db/sqlc"
)

// TierLevel represents the numeric level of a reputation tier.
type TierLevel int

const (
	TierBronze   TierLevel = 0
	TierSilver   TierLevel = 1
	TierGold     TierLevel = 2
	TierPlatinum TierLevel = 3
)

// Tier thresholds in paise.
const (
	SilverThreshold   int64 = 100000  // ₹1,000
	GoldThreshold     int64 = 1000000 // ₹10,000
	PlatinumThreshold int64 = 5000000 // ₹50,000
)

// TierInfo describes a single reputation tier.
type TierInfo struct {
	Level       TierLevel `json:"level"`
	Name        string    `json:"name"`
	Label       string    `json:"label"`
	MinEarnings int64     `json:"min_earnings"`
	NextTier    *string   `json:"next_tier,omitempty"`
	NextMin     *int64    `json:"next_min_earnings,omitempty"`
}

// ClipperReputation is the full reputation response for a clipper.
type ClipperReputation struct {
	UserID                string   `json:"user_id"`
	Tier                  TierInfo `json:"tier"`
	TotalEarnings         int64    `json:"total_earnings"`
	TotalSubmissions      int      `json:"total_submissions"`
	CampaignsParticipated int      `json:"campaigns_participated"`
	MemberSince           string   `json:"member_since"`
}

// ReputationStore is the persistence interface the reputation service needs.
type ReputationStore interface {
	GetClipperTierStats(ctx context.Context, id string) (sqlc.GetClipperTierStatsRow, error)
	GetUserPublicProfile(ctx context.Context, id string) (sqlc.GetUserPublicProfileRow, error)
}

// ReputationService computes clipper tiers from earnings.
type ReputationService struct {
	store ReputationStore
}

// NewReputationService creates a new ReputationService.
func NewReputationService(store ReputationStore) *ReputationService {
	return &ReputationService{store: store}
}

// ComputeTier is a pure function that maps cumulative earnings to a tier.
func ComputeTier(earnings int64) TierInfo {
	switch {
	case earnings >= PlatinumThreshold:
		return TierInfo{
			Level:       TierPlatinum,
			Name:        "platinum",
			Label:       "Elite Clipper",
			MinEarnings: PlatinumThreshold,
		}
	case earnings >= GoldThreshold:
		return TierInfo{
			Level:       TierGold,
			Name:        "gold",
			Label:       "Top Clipper",
			MinEarnings: GoldThreshold,
			NextTier:    strPtr("platinum"),
			NextMin:     int64Ptr(PlatinumThreshold),
		}
	case earnings >= SilverThreshold:
		return TierInfo{
			Level:       TierSilver,
			Name:        "silver",
			Label:       "Active Clipper",
			MinEarnings: SilverThreshold,
			NextTier:    strPtr("gold"),
			NextMin:     int64Ptr(GoldThreshold),
		}
	default:
		return TierInfo{
			Level:       TierBronze,
			Name:        "bronze",
			Label:       "New Clipper",
			MinEarnings: 0,
			NextTier:    strPtr("silver"),
			NextMin:     int64Ptr(SilverThreshold),
		}
	}
}

// GetReputation fetches tier stats and returns the full reputation for a clipper.
func (s *ReputationService) GetReputation(ctx context.Context, userID string) (*ClipperReputation, error) {
	stats, err := s.store.GetClipperTierStats(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get tier stats: %w", err)
	}

	profile, err := s.store.GetUserPublicProfile(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user profile: %w", err)
	}

	tier := ComputeTier(stats.TotalEarnings)

	return &ClipperReputation{
		UserID:                userID,
		Tier:                  tier,
		TotalEarnings:         stats.TotalEarnings,
		TotalSubmissions:      int(stats.TotalSubmissions),
		CampaignsParticipated: int(stats.CampaignsParticipated),
		MemberSince:           profile.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

func strPtr(s string) *string     { return &s }
func int64Ptr(v int64) *int64     { return &v }
