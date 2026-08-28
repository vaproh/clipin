package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	sqlc "clipin/apps/api/internal/db/sqlc"
	"clipin/apps/api/internal/service"

	"github.com/jackc/pgx/v5/pgtype"
)

// --- mock reputation store ---

type mockReputationStore struct {
	getTierStats func(ctx context.Context, id string) (sqlc.GetClipperTierStatsRow, error)
	getProfile   func(ctx context.Context, id string) (sqlc.GetUserPublicProfileRow, error)
}

func (m *mockReputationStore) GetClipperTierStats(ctx context.Context, id string) (sqlc.GetClipperTierStatsRow, error) {
	return m.getTierStats(ctx, id)
}

func (m *mockReputationStore) GetUserPublicProfile(ctx context.Context, id string) (sqlc.GetUserPublicProfileRow, error) {
	return m.getProfile(ctx, id)
}

func testProfile() sqlc.GetUserPublicProfileRow {
	return sqlc.GetUserPublicProfileRow{
		ID:          "clipper1",
		DisplayName: pgtype.Text{String: "Test Clipper", Valid: true},
		CreatedAt:   pgtype.Timestamptz{Valid: true, Time: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
	}
}

// --- ComputeTier tests ---

func TestComputeTier_Bronze(t *testing.T) {
	tier := service.ComputeTier(0)
	if tier.Level != service.TierBronze {
		t.Errorf("expected TierBronze, got %d", tier.Level)
	}
	if tier.Name != "bronze" {
		t.Errorf("expected name bronze, got %s", tier.Name)
	}
	if tier.Label != "New Clipper" {
		t.Errorf("expected label New Clipper, got %s", tier.Label)
	}
	if tier.MinEarnings != 0 {
		t.Errorf("expected min_earnings 0, got %d", tier.MinEarnings)
	}
	if tier.NextTier == nil || *tier.NextTier != "silver" {
		t.Errorf("expected next_tier silver, got %v", tier.NextTier)
	}
}

func TestComputeTier_BronzeBelowThreshold(t *testing.T) {
	tier := service.ComputeTier(99999)
	if tier.Level != service.TierBronze {
		t.Errorf("expected TierBronze at 99999, got %d", tier.Level)
	}
}

func TestComputeTier_Silver(t *testing.T) {
	tier := service.ComputeTier(100000)
	if tier.Level != service.TierSilver {
		t.Errorf("expected TierSilver at 100000, got %d", tier.Level)
	}
	if tier.Name != "silver" {
		t.Errorf("expected name silver, got %s", tier.Name)
	}
	if tier.Label != "Active Clipper" {
		t.Errorf("expected label Active Clipper, got %s", tier.Label)
	}
	if tier.NextTier == nil || *tier.NextTier != "gold" {
		t.Errorf("expected next_tier gold, got %v", tier.NextTier)
	}
}

func TestComputeTier_SilverJustAbove(t *testing.T) {
	tier := service.ComputeTier(100001)
	if tier.Level != service.TierSilver {
		t.Errorf("expected TierSilver at 100001, got %d", tier.Level)
	}
}

func TestComputeTier_Gold(t *testing.T) {
	tier := service.ComputeTier(1000000)
	if tier.Level != service.TierGold {
		t.Errorf("expected TierGold at 1000000, got %d", tier.Level)
	}
	if tier.Name != "gold" {
		t.Errorf("expected name gold, got %s", tier.Name)
	}
	if tier.Label != "Top Clipper" {
		t.Errorf("expected label Top Clipper, got %s", tier.Label)
	}
	if tier.NextTier == nil || *tier.NextTier != "platinum" {
		t.Errorf("expected next_tier platinum, got %v", tier.NextTier)
	}
}

func TestComputeTier_GoldBelowPlatinum(t *testing.T) {
	tier := service.ComputeTier(4999999)
	if tier.Level != service.TierGold {
		t.Errorf("expected TierGold at 4999999, got %d", tier.Level)
	}
}

func TestComputeTier_Platinum(t *testing.T) {
	tier := service.ComputeTier(5000000)
	if tier.Level != service.TierPlatinum {
		t.Errorf("expected TierPlatinum at 5000000, got %d", tier.Level)
	}
	if tier.Name != "platinum" {
		t.Errorf("expected name platinum, got %s", tier.Name)
	}
	if tier.Label != "Elite Clipper" {
		t.Errorf("expected label Elite Clipper, got %s", tier.Label)
	}
	if tier.NextTier != nil {
		t.Errorf("expected no next_tier for platinum, got %v", tier.NextTier)
	}
}

func TestComputeTier_PlatinumExceeds(t *testing.T) {
	tier := service.ComputeTier(10000000)
	if tier.Level != service.TierPlatinum {
		t.Errorf("expected TierPlatinum at 10000000, got %d", tier.Level)
	}
}

func TestComputeTier_NegativeEarnings(t *testing.T) {
	// Negative earnings should still give Bronze.
	tier := service.ComputeTier(-100)
	if tier.Level != service.TierBronze {
		t.Errorf("expected TierBronze for negative earnings, got %d", tier.Level)
	}
}

// --- GetReputation tests ---

func TestGetReputation_BronzeClipper(t *testing.T) {
	store := &mockReputationStore{
		getTierStats: func(ctx context.Context, id string) (sqlc.GetClipperTierStatsRow, error) {
			return sqlc.GetClipperTierStatsRow{
				TotalEarnings:         0,
				TotalSubmissions:      5,
				CampaignsParticipated: 2,
			}, nil
		},
		getProfile: func(ctx context.Context, id string) (sqlc.GetUserPublicProfileRow, error) {
			return testProfile(), nil
		},
	}
	svc := service.NewReputationService(store)
	rep, err := svc.GetReputation(context.Background(), "clipper1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rep.Tier.Level != service.TierBronze {
		t.Errorf("expected TierBronze, got %d", rep.Tier.Level)
	}
	if rep.TotalEarnings != 0 {
		t.Errorf("expected total_earnings 0, got %d", rep.TotalEarnings)
	}
	if rep.TotalSubmissions != 5 {
		t.Errorf("expected total_submissions 5, got %d", rep.TotalSubmissions)
	}
	if rep.CampaignsParticipated != 2 {
		t.Errorf("expected campaigns_participated 2, got %d", rep.CampaignsParticipated)
	}
}

func TestGetReputation_GoldClipper(t *testing.T) {
	store := &mockReputationStore{
		getTierStats: func(ctx context.Context, id string) (sqlc.GetClipperTierStatsRow, error) {
			return sqlc.GetClipperTierStatsRow{
				TotalEarnings:         2500000,
				TotalSubmissions:      50,
				CampaignsParticipated: 12,
			}, nil
		},
		getProfile: func(ctx context.Context, id string) (sqlc.GetUserPublicProfileRow, error) {
			return testProfile(), nil
		},
	}
	svc := service.NewReputationService(store)
	rep, err := svc.GetReputation(context.Background(), "clipper1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rep.Tier.Level != service.TierGold {
		t.Errorf("expected TierGold, got %d", rep.Tier.Level)
	}
	if rep.TotalEarnings != 2500000 {
		t.Errorf("expected total_earnings 2500000, got %d", rep.TotalEarnings)
	}
}

func TestGetReputation_StoreError(t *testing.T) {
	store := &mockReputationStore{
		getTierStats: func(ctx context.Context, id string) (sqlc.GetClipperTierStatsRow, error) {
			return sqlc.GetClipperTierStatsRow{}, fmt.Errorf("db error")
		},
	}
	svc := service.NewReputationService(store)
	_, err := svc.GetReputation(context.Background(), "clipper1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetReputation_ProfileError(t *testing.T) {
	store := &mockReputationStore{
		getTierStats: func(ctx context.Context, id string) (sqlc.GetClipperTierStatsRow, error) {
			return sqlc.GetClipperTierStatsRow{TotalEarnings: 50000}, nil
		},
		getProfile: func(ctx context.Context, id string) (sqlc.GetUserPublicProfileRow, error) {
			return sqlc.GetUserPublicProfileRow{}, fmt.Errorf("profile fetch error")
		},
	}
	svc := service.NewReputationService(store)
	_, err := svc.GetReputation(context.Background(), "clipper1")
	if err == nil {
		t.Fatal("expected error")
	}
}
