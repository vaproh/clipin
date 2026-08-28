package service_test

import (
	"context"
	"testing"

	sqlc "clipin/apps/api/internal/db/sqlc"
	"clipin/apps/api/internal/service"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// mockCampaignStore implements CampaignStore for testing.
type mockCampaignStore struct {
	getByID         func(ctx context.Context, id pgtype.UUID) (sqlc.Campaign, error)
	listFiltered    func(ctx context.Context, arg sqlc.ListCampaignsFilteredParams) ([]sqlc.Campaign, error)
	countFiltered   func(ctx context.Context, arg sqlc.CountCampaignsFilteredParams) (int64, error)
	listByOwner     func(ctx context.Context, ownerID string) ([]sqlc.Campaign, error)
	createCampaign  func(ctx context.Context, arg sqlc.CreateCampaignParams) (sqlc.Campaign, error)
	updateCampaign  func(ctx context.Context, arg sqlc.UpdateCampaignParams) (sqlc.Campaign, error)
	updateStatus    func(ctx context.Context, arg sqlc.UpdateCampaignStatusParams) (sqlc.Campaign, error)
}

func (m *mockCampaignStore) GetCampaignByID(ctx context.Context, id pgtype.UUID) (sqlc.Campaign, error) {
	if m.getByID != nil {
		return m.getByID(ctx, id)
	}
	return sqlc.Campaign{}, nil
}

func (m *mockCampaignStore) ListCampaignsFiltered(ctx context.Context, arg sqlc.ListCampaignsFilteredParams) ([]sqlc.Campaign, error) {
	if m.listFiltered != nil {
		return m.listFiltered(ctx, arg)
	}
	return nil, nil
}

func (m *mockCampaignStore) CountCampaignsFiltered(ctx context.Context, arg sqlc.CountCampaignsFilteredParams) (int64, error) {
	if m.countFiltered != nil {
		return m.countFiltered(ctx, arg)
	}
	return 0, nil
}

func (m *mockCampaignStore) ListCampaignsByOwner(ctx context.Context, ownerID string) ([]sqlc.Campaign, error) {
	if m.listByOwner != nil {
		return m.listByOwner(ctx, ownerID)
	}
	return nil, nil
}

func (m *mockCampaignStore) CreateCampaign(ctx context.Context, arg sqlc.CreateCampaignParams) (sqlc.Campaign, error) {
	if m.createCampaign != nil {
		return m.createCampaign(ctx, arg)
	}
	return sqlc.Campaign{}, nil
}

func (m *mockCampaignStore) UpdateCampaign(ctx context.Context, arg sqlc.UpdateCampaignParams) (sqlc.Campaign, error) {
	if m.updateCampaign != nil {
		return m.updateCampaign(ctx, arg)
	}
	return sqlc.Campaign{}, nil
}

func (m *mockCampaignStore) UpdateCampaignStatus(ctx context.Context, arg sqlc.UpdateCampaignStatusParams) (sqlc.Campaign, error) {
	if m.updateStatus != nil {
		return m.updateStatus(ctx, arg)
	}
	return sqlc.Campaign{}, nil
}

func testCampaign(id string, ownerID string, title string, status string) sqlc.Campaign {
	return sqlc.Campaign{
		ID:              pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true},
		OwnerID:         ownerID,
		Title:           title,
		Platform:        "youtube",
		Status:          status,
		CpmRate:         100,
		TotalBudget:     10000,
		RemainingBudget: 5000,
		PlatformFee:     1000,
		CreatedAt:       pgtype.Timestamptz{Valid: true},
		UpdatedAt:       pgtype.Timestamptz{Valid: true},
	}
}

func TestListPublic_DefaultPagination(t *testing.T) {
	store := &mockCampaignStore{
		listFiltered: func(ctx context.Context, arg sqlc.ListCampaignsFilteredParams) ([]sqlc.Campaign, error) {
			if arg.Limit != 20 {
				t.Errorf("expected default limit 20, got %d", arg.Limit)
			}
			if arg.Offset != 0 {
				t.Errorf("expected default offset 0, got %d", arg.Offset)
			}
			return []sqlc.Campaign{testCampaign("1", "owner1", "Campaign 1", "active")}, nil
		},
		countFiltered: func(ctx context.Context, arg sqlc.CountCampaignsFilteredParams) (int64, error) {
			return 1, nil
		},
	}
	svc := service.NewCampaignService(store, nil)
	result, err := svc.ListPublic(context.Background(), service.CampaignFilters{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Page != 1 {
		t.Errorf("expected page 1, got %d", result.Page)
	}
	if result.PageSize != 20 {
		t.Errorf("expected page_size 20, got %d", result.PageSize)
	}
	if result.Total != 1 {
		t.Errorf("expected total 1, got %d", result.Total)
	}
	if len(result.Campaigns) != 1 {
		t.Fatalf("expected 1 campaign, got %d", len(result.Campaigns))
	}
}

func TestListPublic_CustomPagination(t *testing.T) {
	store := &mockCampaignStore{
		listFiltered: func(ctx context.Context, arg sqlc.ListCampaignsFilteredParams) ([]sqlc.Campaign, error) {
			if arg.Limit != 10 {
				t.Errorf("expected limit 10, got %d", arg.Limit)
			}
			if arg.Offset != 20 {
				t.Errorf("expected offset 20 (page 3, size 10), got %d", arg.Offset)
			}
			return nil, nil
		},
		countFiltered: func(ctx context.Context, arg sqlc.CountCampaignsFilteredParams) (int64, error) {
			return 0, nil
		},
	}
	svc := service.NewCampaignService(store, nil)
	result, err := svc.ListPublic(context.Background(), service.CampaignFilters{Page: 3, PageSize: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Page != 3 {
		t.Errorf("expected page 3, got %d", result.Page)
	}
}

func TestListPublic_CapsPageSize(t *testing.T) {
	store := &mockCampaignStore{
		listFiltered: func(ctx context.Context, arg sqlc.ListCampaignsFilteredParams) ([]sqlc.Campaign, error) {
			if arg.Limit > 100 {
				t.Errorf("expected max limit 100, got %d", arg.Limit)
			}
			return nil, nil
		},
		countFiltered: func(ctx context.Context, arg sqlc.CountCampaignsFilteredParams) (int64, error) {
			return 0, nil
		},
	}
	svc := service.NewCampaignService(store, nil)
	result, err := svc.ListPublic(context.Background(), service.CampaignFilters{PageSize: 999})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PageSize != 20 {
		t.Errorf("expected page_size 20 (capped), got %d", result.PageSize)
	}
}

func TestListPublic_PassesFilters(t *testing.T) {
	store := &mockCampaignStore{
		listFiltered: func(ctx context.Context, arg sqlc.ListCampaignsFilteredParams) ([]sqlc.Campaign, error) {
			if arg.Column1 != "youtube" {
				t.Errorf("expected platform youtube, got %q", arg.Column1)
			}
			if arg.Column2 != 200 {
				t.Errorf("expected maxCPM 200, got %d", arg.Column2)
			}
			if arg.Column3 != 5000 {
				t.Errorf("expected minBudget 5000, got %d", arg.Column3)
			}
			return nil, nil
		},
		countFiltered: func(ctx context.Context, arg sqlc.CountCampaignsFilteredParams) (int64, error) {
			return 0, nil
		},
	}
	svc := service.NewCampaignService(store, nil)
	_, err := svc.ListPublic(context.Background(), service.CampaignFilters{
		Platform:  "youtube",
		MaxCPM:    200,
		MinBudget: 5000,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetByID_Found(t *testing.T) {
	c := testCampaign("1", "owner1", "My Campaign", "active")
	store := &mockCampaignStore{
		getByID: func(ctx context.Context, id pgtype.UUID) (sqlc.Campaign, error) {
			return c, nil
		},
	}
	svc := service.NewCampaignService(store, nil)
	result, err := svc.GetByID(context.Background(), c.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected campaign, got nil")
	}
	if result.Title != "My Campaign" {
		t.Errorf("expected title My Campaign, got %q", result.Title)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	store := &mockCampaignStore{
		getByID: func(ctx context.Context, id pgtype.UUID) (sqlc.Campaign, error) {
			return sqlc.Campaign{}, pgx.ErrNoRows
		},
	}
	svc := service.NewCampaignService(store, nil)
	result, err := svc.GetByID(context.Background(), pgtype.UUID{Bytes: [16]byte{1}, Valid: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil, got campaign")
	}
}

func TestListByOwner(t *testing.T) {
	store := &mockCampaignStore{
		listByOwner: func(ctx context.Context, ownerID string) ([]sqlc.Campaign, error) {
			if ownerID != "owner1" {
				t.Errorf("expected owner owner1, got %q", ownerID)
			}
			return []sqlc.Campaign{
				testCampaign("1", "owner1", "C1", "draft"),
				testCampaign("2", "owner1", "C2", "active"),
			}, nil
		},
	}
	svc := service.NewCampaignService(store, nil)
	result, err := svc.ListByOwner(context.Background(), "owner1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 campaigns, got %d", len(result))
	}
}

func TestUpdate_PersistsFields(t *testing.T) {
	c := testCampaign("1", "owner1", "Old Title", "draft")
	var capturedArg sqlc.UpdateCampaignParams
	store := &mockCampaignStore{
		getByID: func(ctx context.Context, id pgtype.UUID) (sqlc.Campaign, error) {
			return c, nil
		},
		updateCampaign: func(ctx context.Context, arg sqlc.UpdateCampaignParams) (sqlc.Campaign, error) {
			capturedArg = arg
			return sqlc.Campaign{
				ID:                  arg.ID,
				Title:               arg.Title,
				Description:         arg.Description,
				BriefUrl:            arg.BriefUrl,
				MaxClipsPerCampaign: arg.MaxClipsPerCampaign,
				MaxClipsPerClipper:  arg.MaxClipsPerClipper,
				MinViewsPerClip:     arg.MinViewsPerClip,
				AutoApproveHours:    arg.AutoApproveHours,
				EndsAt:              arg.EndsAt,
				Status:              c.Status,
			}, nil
		},
	}
	svc := service.NewCampaignService(store, nil)

	newTitle := "New Title"
	newDesc := "New description"
	result, err := svc.Update(context.Background(), "owner1", c.ID, &service.UpdateCampaignInput{
		Title:       &newTitle,
		Description: &newDesc,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedArg.Title != "New Title" {
		t.Errorf("expected title 'New Title', got %q", capturedArg.Title)
	}
	if !capturedArg.Description.Valid || capturedArg.Description.String != "New description" {
		t.Errorf("expected description 'New description', got %v", capturedArg.Description)
	}
	if result.Title != "New Title" {
		t.Errorf("expected returned title 'New Title', got %q", result.Title)
	}
}

func TestUpdate_NotFound(t *testing.T) {
	store := &mockCampaignStore{
		getByID: func(ctx context.Context, id pgtype.UUID) (sqlc.Campaign, error) {
			return sqlc.Campaign{}, pgx.ErrNoRows
		},
	}
	svc := service.NewCampaignService(store, nil)
	_, err := svc.Update(context.Background(), "owner1", pgtype.UUID{Bytes: [16]byte{1}, Valid: true}, &service.UpdateCampaignInput{})
	if err != service.ErrCampaignNotFound {
		t.Errorf("expected ErrCampaignNotFound, got %v", err)
	}
}

func TestUpdate_NotOwner(t *testing.T) {
	c := testCampaign("1", "owner1", "My Campaign", "draft")
	store := &mockCampaignStore{
		getByID: func(ctx context.Context, id pgtype.UUID) (sqlc.Campaign, error) {
			return c, nil
		},
	}
	svc := service.NewCampaignService(store, nil)
	_, err := svc.Update(context.Background(), "wrong-owner", c.ID, &service.UpdateCampaignInput{})
	if err != service.ErrNotOwner {
		t.Errorf("expected ErrNotOwner, got %v", err)
	}
}

func TestListPublic_SearchByTitle(t *testing.T) {
	store := &mockCampaignStore{
		listFiltered: func(ctx context.Context, arg sqlc.ListCampaignsFilteredParams) ([]sqlc.Campaign, error) {
			if arg.Column4 != "promo" {
				t.Errorf("expected search term 'promo', got %q", arg.Column4)
			}
			return []sqlc.Campaign{testCampaign("1", "owner1", "Summer Promo Campaign", "active")}, nil
		},
		countFiltered: func(ctx context.Context, arg sqlc.CountCampaignsFilteredParams) (int64, error) {
			if arg.Column4 != "promo" {
				t.Errorf("expected search term 'promo' in count, got %q", arg.Column4)
			}
			return 1, nil
		},
	}
	svc := service.NewCampaignService(store, nil)
	result, err := svc.ListPublic(context.Background(), service.CampaignFilters{Search: "promo"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("expected total 1, got %d", result.Total)
	}
	if len(result.Campaigns) != 1 {
		t.Fatalf("expected 1 campaign, got %d", len(result.Campaigns))
	}
	if result.Campaigns[0].Title != "Summer Promo Campaign" {
		t.Errorf("expected title 'Summer Promo Campaign', got %q", result.Campaigns[0].Title)
	}
}

func TestListPublic_SearchByDescription(t *testing.T) {
	store := &mockCampaignStore{
		listFiltered: func(ctx context.Context, arg sqlc.ListCampaignsFilteredParams) ([]sqlc.Campaign, error) {
			if arg.Column4 != "skincare" {
				t.Errorf("expected search term 'skincare', got %q", arg.Column4)
			}
			return []sqlc.Campaign{testCampaign("1", "owner1", "Beauty Campaign", "active")}, nil
		},
		countFiltered: func(ctx context.Context, arg sqlc.CountCampaignsFilteredParams) (int64, error) {
			return 1, nil
		},
	}
	svc := service.NewCampaignService(store, nil)
	result, err := svc.ListPublic(context.Background(), service.CampaignFilters{Search: "skincare"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Campaigns) != 1 {
		t.Fatalf("expected 1 campaign, got %d", len(result.Campaigns))
	}
}

func TestListPublic_SearchEmptyReturnsAll(t *testing.T) {
	store := &mockCampaignStore{
		listFiltered: func(ctx context.Context, arg sqlc.ListCampaignsFilteredParams) ([]sqlc.Campaign, error) {
			if arg.Column4 != "" {
				t.Errorf("expected empty search, got %q", arg.Column4)
			}
			return []sqlc.Campaign{
				testCampaign("1", "owner1", "Campaign A", "active"),
				testCampaign("2", "owner1", "Campaign B", "active"),
			}, nil
		},
		countFiltered: func(ctx context.Context, arg sqlc.CountCampaignsFilteredParams) (int64, error) {
			return 2, nil
		},
	}
	svc := service.NewCampaignService(store, nil)
	result, err := svc.ListPublic(context.Background(), service.CampaignFilters{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("expected total 2, got %d", result.Total)
	}
}

func TestListPublic_SearchWithFilters(t *testing.T) {
	store := &mockCampaignStore{
		listFiltered: func(ctx context.Context, arg sqlc.ListCampaignsFilteredParams) ([]sqlc.Campaign, error) {
			if arg.Column4 != "gaming" {
				t.Errorf("expected search 'gaming', got %q", arg.Column4)
			}
			if arg.Column1 != "youtube" {
				t.Errorf("expected platform 'youtube', got %q", arg.Column1)
			}
			if arg.Column2 != 200 {
				t.Errorf("expected maxCPM 200, got %d", arg.Column2)
			}
			if arg.Column3 != 5000 {
				t.Errorf("expected minBudget 5000, got %d", arg.Column3)
			}
			return []sqlc.Campaign{testCampaign("1", "owner1", "Gaming Promo", "active")}, nil
		},
		countFiltered: func(ctx context.Context, arg sqlc.CountCampaignsFilteredParams) (int64, error) {
			return 1, nil
		},
	}
	svc := service.NewCampaignService(store, nil)
	result, err := svc.ListPublic(context.Background(), service.CampaignFilters{
		Search:    "gaming",
		Platform:  "youtube",
		MaxCPM:    200,
		MinBudget: 5000,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("expected total 1, got %d", result.Total)
	}
}

func TestListPublic_SearchNoResults(t *testing.T) {
	store := &mockCampaignStore{
		listFiltered: func(ctx context.Context, arg sqlc.ListCampaignsFilteredParams) ([]sqlc.Campaign, error) {
			return nil, nil
		},
		countFiltered: func(ctx context.Context, arg sqlc.CountCampaignsFilteredParams) (int64, error) {
			return 0, nil
		},
	}
	svc := service.NewCampaignService(store, nil)
	result, err := svc.ListPublic(context.Background(), service.CampaignFilters{Search: "nonexistent"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 0 {
		t.Errorf("expected total 0, got %d", result.Total)
	}
	if len(result.Campaigns) != 0 {
		t.Errorf("expected 0 campaigns, got %d", len(result.Campaigns))
	}
}

func TestListPublic_SearchCaseInsensitive(t *testing.T) {
	store := &mockCampaignStore{
		listFiltered: func(ctx context.Context, arg sqlc.ListCampaignsFilteredParams) ([]sqlc.Campaign, error) {
			if arg.Column4 != "PROMO" {
				t.Errorf("expected search term 'PROMO', got %q", arg.Column4)
			}
			return []sqlc.Campaign{testCampaign("1", "owner1", "Summer Promo", "active")}, nil
		},
		countFiltered: func(ctx context.Context, arg sqlc.CountCampaignsFilteredParams) (int64, error) {
			return 1, nil
		},
	}
	svc := service.NewCampaignService(store, nil)
	result, err := svc.ListPublic(context.Background(), service.CampaignFilters{Search: "PROMO"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("expected total 1, got %d", result.Total)
	}
}

func TestListPublic_SearchPagination(t *testing.T) {
	store := &mockCampaignStore{
		listFiltered: func(ctx context.Context, arg sqlc.ListCampaignsFilteredParams) ([]sqlc.Campaign, error) {
			if arg.Column4 != "promo" {
				t.Errorf("expected search 'promo', got %q", arg.Column4)
			}
			if arg.Limit != 5 {
				t.Errorf("expected limit 5, got %d", arg.Limit)
			}
			if arg.Offset != 10 {
				t.Errorf("expected offset 10 (page 3, size 5), got %d", arg.Offset)
			}
			return nil, nil
		},
		countFiltered: func(ctx context.Context, arg sqlc.CountCampaignsFilteredParams) (int64, error) {
			return 0, nil
		},
	}
	svc := service.NewCampaignService(store, nil)
	result, err := svc.ListPublic(context.Background(), service.CampaignFilters{
		Search:   "promo",
		Page:     3,
		PageSize: 5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Page != 3 {
		t.Errorf("expected page 3, got %d", result.Page)
	}
}

func TestListPublic_SearchPartialMatch(t *testing.T) {
	store := &mockCampaignStore{
		listFiltered: func(ctx context.Context, arg sqlc.ListCampaignsFilteredParams) ([]sqlc.Campaign, error) {
			if arg.Column4 != "promo" {
				t.Errorf("expected search 'promo', got %q", arg.Column4)
			}
			return []sqlc.Campaign{
				testCampaign("1", "owner1", "Promo Campaign", "active"),
				testCampaign("2", "owner1", "Summer Promo", "active"),
				testCampaign("3", "owner1", "Promotional Video", "active"),
			}, nil
		},
		countFiltered: func(ctx context.Context, arg sqlc.CountCampaignsFilteredParams) (int64, error) {
			return 3, nil
		},
	}
	svc := service.NewCampaignService(store, nil)
	result, err := svc.ListPublic(context.Background(), service.CampaignFilters{Search: "promo"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 3 {
		t.Errorf("expected total 3, got %d", result.Total)
	}
}

// --- Notification wiring tests ---

func TestCreate_SendsNotification(t *testing.T) {
	store := &mockCampaignStore{
		createCampaign: func(ctx context.Context, arg sqlc.CreateCampaignParams) (sqlc.Campaign, error) {
			return sqlc.Campaign{
				ID:              arg.ID,
				OwnerID:         arg.OwnerID,
				Title:           arg.Title,
				Status:          "active",
				CpmRate:         arg.CpmRate,
				TotalBudget:     arg.TotalBudget,
				RemainingBudget: arg.TotalBudget,
				PlatformFee:     arg.PlatformFee,
			}, nil
		},
	}

	var notifCalled bool
	var notifOwner string
	var notifMsg string
	svc := service.NewCampaignService(store, nil)
	svc.WithNotifications(service.NewNotificationService(&notifStore{
		create: func(ctx context.Context, arg sqlc.CreateNotificationParams) (sqlc.Notification, error) {
			notifCalled = true
			notifOwner = arg.UserID
			notifMsg = arg.Body.String
			return sqlc.Notification{}, nil
		},
	}))

	title := "New Campaign"
	budget := int32(10000)
	cpm := int32(100)
	_, err := svc.Create(context.Background(), "owner1", &service.CreateCampaignInput{
		Title:       title,
		Platform:    "youtube",
		CpmRate:     cpm,
		TotalBudget: budget,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !notifCalled {
		t.Error("expected notification to be sent on campaign create")
	}
	if notifOwner != "owner1" {
		t.Errorf("expected owner1, got %s", notifOwner)
	}
	if notifMsg != "Campaign 'New Campaign': has been created" {
		t.Errorf("unexpected notification body: %q", notifMsg)
	}
}

func TestCreate_NoNotificationsWhenNil(t *testing.T) {
	store := &mockCampaignStore{
		createCampaign: func(ctx context.Context, arg sqlc.CreateCampaignParams) (sqlc.Campaign, error) {
			return sqlc.Campaign{
				ID:              arg.ID,
				OwnerID:         arg.OwnerID,
				Title:           arg.Title,
				Status:          "active",
				CpmRate:         arg.CpmRate,
				TotalBudget:     arg.TotalBudget,
				RemainingBudget: arg.TotalBudget,
				PlatformFee:     arg.PlatformFee,
			}, nil
		},
	}
	svc := service.NewCampaignService(store, nil)
	// No WithNotifications call; should not panic.
	title := "Test"
	budget := int32(10000)
	cpm := int32(100)
	_, err := svc.Create(context.Background(), "owner1", &service.CreateCampaignInput{
		Title:       title,
		Platform:    "youtube",
		CpmRate:     cpm,
		TotalBudget: budget,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func testCampaignForNotifs(id string, ownerID string, title string, status string, budget int32) sqlc.Campaign {
	return sqlc.Campaign{
		ID:              pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true},
		OwnerID:         ownerID,
		Title:           title,
		Platform:        "youtube",
		Status:          status,
		CpmRate:         100,
		TotalBudget:     budget,
		RemainingBudget: budget / 2,
		PlatformFee:     1000,
		CreatedAt:       pgtype.Timestamptz{Valid: true},
		UpdatedAt:       pgtype.Timestamptz{Valid: true},
	}
}

func TestPause_SendsNotification(t *testing.T) {
	c := testCampaignForNotifs("1", "owner1", "My Campaign", "active", 10000)
	store := &mockCampaignStore{
		getByID: func(ctx context.Context, id pgtype.UUID) (sqlc.Campaign, error) {
			return c, nil
		},
		updateStatus: func(ctx context.Context, arg sqlc.UpdateCampaignStatusParams) (sqlc.Campaign, error) {
			return sqlc.Campaign{ID: arg.ID, Status: arg.Status, Title: c.Title}, nil
		},
	}

	var notifBody string
	svc := service.NewCampaignService(store, nil)
	svc.WithNotifications(service.NewNotificationService(&notifStore{
		create: func(ctx context.Context, arg sqlc.CreateNotificationParams) (sqlc.Notification, error) {
			notifBody = arg.Body.String
			return sqlc.Notification{}, nil
		},
	}))

	_, err := svc.Pause(context.Background(), "owner1", c.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if notifBody != "Campaign 'My Campaign': has been paused" {
		t.Errorf("unexpected notification body: %q", notifBody)
	}
}

func TestResume_SendsNotification(t *testing.T) {
	c := testCampaignForNotifs("1", "owner1", "My Campaign", "paused", 10000)
	store := &mockCampaignStore{
		getByID: func(ctx context.Context, id pgtype.UUID) (sqlc.Campaign, error) {
			return c, nil
		},
		updateStatus: func(ctx context.Context, arg sqlc.UpdateCampaignStatusParams) (sqlc.Campaign, error) {
			return sqlc.Campaign{ID: arg.ID, Status: arg.Status, Title: c.Title}, nil
		},
	}

	var notifBody string
	svc := service.NewCampaignService(store, nil)
	svc.WithNotifications(service.NewNotificationService(&notifStore{
		create: func(ctx context.Context, arg sqlc.CreateNotificationParams) (sqlc.Notification, error) {
			notifBody = arg.Body.String
			return sqlc.Notification{}, nil
		},
	}))

	_, err := svc.Resume(context.Background(), "owner1", c.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if notifBody != "Campaign 'My Campaign': has been active" {
		t.Errorf("unexpected notification body: %q", notifBody)
	}
}

func TestCancel_SendsNotification(t *testing.T) {
	c := testCampaignForNotifs("1", "owner1", "My Campaign", "active", 10000)
	store := &mockCampaignStore{
		getByID: func(ctx context.Context, id pgtype.UUID) (sqlc.Campaign, error) {
			return c, nil
		},
		updateStatus: func(ctx context.Context, arg sqlc.UpdateCampaignStatusParams) (sqlc.Campaign, error) {
			return sqlc.Campaign{ID: arg.ID, Status: arg.Status, Title: c.Title}, nil
		},
	}

	var notifBody string
	svc := service.NewCampaignService(store, nil)
	svc.WithNotifications(service.NewNotificationService(&notifStore{
		create: func(ctx context.Context, arg sqlc.CreateNotificationParams) (sqlc.Notification, error) {
			notifBody = arg.Body.String
			return sqlc.Notification{}, nil
		},
	}))

	_, err := svc.Cancel(context.Background(), "owner1", c.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if notifBody != "Campaign 'My Campaign': has been cancelled. Remaining budget refunded." {
		t.Errorf("unexpected notification body: %q", notifBody)
	}
}

// notifStore is a minimal mock for NotificationStore used in campaign tests.
type notifStore struct {
	create func(ctx context.Context, arg sqlc.CreateNotificationParams) (sqlc.Notification, error)
}

func (n *notifStore) CreateNotification(ctx context.Context, arg sqlc.CreateNotificationParams) (sqlc.Notification, error) {
	if n.create != nil {
		return n.create(ctx, arg)
	}
	return sqlc.Notification{}, nil
}
func (n *notifStore) ListNotificationsByUser(_ context.Context, _ sqlc.ListNotificationsByUserParams) ([]sqlc.Notification, error) {
	return nil, nil
}
func (n *notifStore) CountUnreadNotifications(_ context.Context, _ string) (int64, error) { return 0, nil }
func (n *notifStore) MarkNotificationRead(_ context.Context, _ sqlc.MarkNotificationReadParams) error {
	return nil
}
func (n *notifStore) MarkAllNotificationsRead(_ context.Context, _ string) error { return nil }
func (n *notifStore) DeleteNotification(_ context.Context, _ sqlc.DeleteNotificationParams) error {
	return nil
}
