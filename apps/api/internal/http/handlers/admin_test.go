package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	sqlc "clipin/apps/api/internal/db/sqlc"
	"clipin/apps/api/internal/http/handlers"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// --- Mock stores ---

type fakeAdminUserStore struct {
	listUsers     func(ctx context.Context, arg sqlc.ListUsersParams) ([]sqlc.User, error)
	countUsers    func(ctx context.Context) (int32, error)
	getUserByID   func(ctx context.Context, id string) (sqlc.User, error)
	listSubs      func(ctx context.Context, clipperID string) ([]sqlc.Submission, error)
	listPayouts   func(ctx context.Context, clipperID string) ([]sqlc.PayoutRequest, error)
	listFlags     func(ctx context.Context, userID pgtype.Text) ([]sqlc.FraudFlag, error)
	getCampaign   func(ctx context.Context, id pgtype.UUID) (sqlc.Campaign, error)
	updateStatus  func(ctx context.Context, arg sqlc.UpdateCampaignStatusParams) (sqlc.Campaign, error)
}

func (f *fakeAdminUserStore) ListUsers(ctx context.Context, arg sqlc.ListUsersParams) ([]sqlc.User, error) {
	if f.listUsers != nil {
		return f.listUsers(ctx, arg)
	}
	return nil, nil
}

func (f *fakeAdminUserStore) CountUsers(ctx context.Context) (int32, error) {
	if f.countUsers != nil {
		return f.countUsers(ctx)
	}
	return 0, nil
}

func (f *fakeAdminUserStore) GetUserByID(ctx context.Context, id string) (sqlc.User, error) {
	if f.getUserByID != nil {
		return f.getUserByID(ctx, id)
	}
	return sqlc.User{}, nil
}

func (f *fakeAdminUserStore) ListSubmissionsByUser(ctx context.Context, clipperID string) ([]sqlc.Submission, error) {
	if f.listSubs != nil {
		return f.listSubs(ctx, clipperID)
	}
	return nil, nil
}

func (f *fakeAdminUserStore) ListPayoutsByUser(ctx context.Context, clipperID string) ([]sqlc.PayoutRequest, error) {
	if f.listPayouts != nil {
		return f.listPayouts(ctx, clipperID)
	}
	return nil, nil
}

func (f *fakeAdminUserStore) ListFraudFlagsByUser(ctx context.Context, userID pgtype.Text) ([]sqlc.FraudFlag, error) {
	if f.listFlags != nil {
		return f.listFlags(ctx, userID)
	}
	return nil, nil
}

func (f *fakeAdminUserStore) GetCampaignByID(ctx context.Context, id pgtype.UUID) (sqlc.Campaign, error) {
	if f.getCampaign != nil {
		return f.getCampaign(ctx, id)
	}
	return sqlc.Campaign{}, nil
}

func (f *fakeAdminUserStore) UpdateCampaignStatus(ctx context.Context, arg sqlc.UpdateCampaignStatusParams) (sqlc.Campaign, error) {
	if f.updateStatus != nil {
		return f.updateStatus(ctx, arg)
	}
	return sqlc.Campaign{}, nil
}

type fakeAuditSvc struct {
	log func(ctx context.Context, actorID, action, resourceType, resourceID string, details map[string]any, ipAddress string) error
}

func (f *fakeAuditSvc) Log(ctx context.Context, actorID, action, resourceType, resourceID string, details map[string]any, ipAddress string) error {
	if f.log != nil {
		return f.log(ctx, actorID, action, resourceType, resourceID, details, ipAddress)
	}
	return nil
}

func (f *fakeAuditSvc) ListLogs(ctx context.Context, limit, offset int32) ([]sqlc.AuditLog, error) {
	return nil, nil
}

func (f *fakeAuditSvc) ListLogsForResource(ctx context.Context, resourceType, resourceID string) ([]sqlc.AuditLog, error) {
	return nil, nil
}

type fakeFraudSvc struct {
	listOpen   func(ctx context.Context) ([]sqlc.FraudFlag, error)
	resolve    func(ctx context.Context, flagID pgtype.UUID, resolvedBy, resolution string) (*sqlc.FraudFlag, error)
	dismiss    func(ctx context.Context, flagID pgtype.UUID, resolvedBy, reason string) (*sqlc.FraudFlag, error)
	flagCount  func(ctx context.Context, userID string) (int, error)
}

func (f *fakeFraudSvc) ListOpenFlags(ctx context.Context) ([]sqlc.FraudFlag, error) {
	if f.listOpen != nil {
		return f.listOpen(ctx)
	}
	return nil, nil
}

func (f *fakeFraudSvc) ResolveFlag(ctx context.Context, flagID pgtype.UUID, resolvedBy, resolution string) (*sqlc.FraudFlag, error) {
	if f.resolve != nil {
		return f.resolve(ctx, flagID, resolvedBy, resolution)
	}
	return nil, nil
}

func (f *fakeFraudSvc) DismissFlag(ctx context.Context, flagID pgtype.UUID, resolvedBy, reason string) (*sqlc.FraudFlag, error) {
	if f.dismiss != nil {
		return f.dismiss(ctx, flagID, resolvedBy, reason)
	}
	return nil, nil
}

func (f *fakeFraudSvc) GetUserFlagCount(ctx context.Context, userID string) (int, error) {
	if f.flagCount != nil {
		return f.flagCount(ctx, userID)
	}
	return 0, nil
}

// --- Helpers ---

var adminUser1 = &sqlc.User{ID: "admin_1", Email: "admin@test.com", Role: "admin"}
var regularUser = &sqlc.User{ID: "user_1", Email: "user@test.com", Role: "clipper"}

func adminRouter(userStore handlers.AdminUserStore, auditSvc handlers.AuditServiceInterface, fraudSvc handlers.FraudServiceInterface, user *sqlc.User) http.Handler {
	r := chi.NewRouter()
	api := humachi.New(r, huma.DefaultConfig("ClipIN API", "1.0.0"))
	handlers.RegisterAdminHandlers(api, userStore, auditSvc, fraudSvc)
	return withUser(user)(r)
}

func adminRouterNoAuth(userStore handlers.AdminUserStore, auditSvc handlers.AuditServiceInterface, fraudSvc handlers.FraudServiceInterface) http.Handler {
	r := chi.NewRouter()
	api := humachi.New(r, huma.DefaultConfig("ClipIN API", "1.0.0"))
	handlers.RegisterAdminHandlers(api, userStore, auditSvc, fraudSvc)
	return r
}

// --- Tests ---

func TestAdminListUsers_Unauthorized(t *testing.T) {
	router := adminRouterNoAuth(&fakeAdminUserStore{}, &fakeAuditSvc{}, &fakeFraudSvc{})
	rec := doRequest(t, router, http.MethodGet, "/admin/users", "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAdminListUsers_Forbidden_NonAdmin(t *testing.T) {
	router := adminRouter(&fakeAdminUserStore{}, &fakeAuditSvc{}, &fakeFraudSvc{}, regularUser)
	rec := doRequest(t, router, http.MethodGet, "/admin/users", "")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestAdminListUsers_HappyPath(t *testing.T) {
	store := &fakeAdminUserStore{
		listUsers: func(ctx context.Context, arg sqlc.ListUsersParams) ([]sqlc.User, error) {
			return []sqlc.User{
				{ID: "u1", Email: "a@test.com", Role: "clipper", CreatedAt: pgtype.Timestamptz{Valid: true}},
			}, nil
		},
		countUsers: func(ctx context.Context) (int32, error) {
			return 1, nil
		},
	}
	router := adminRouter(store, &fakeAuditSvc{}, &fakeFraudSvc{}, adminUser1)
	rec := doRequest(t, router, http.MethodGet, "/admin/users", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var out struct {
		Users []struct {
			ID    string `json:"id"`
			Email string `json:"email"`
		} `json:"users"`
		Total int32 `json:"total"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out.Users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(out.Users))
	}
	if out.Total != 1 {
		t.Errorf("expected total 1, got %d", out.Total)
	}
}

func TestAdminGetUser_NotFound(t *testing.T) {
	store := &fakeAdminUserStore{
		getUserByID: func(ctx context.Context, id string) (sqlc.User, error) {
			return sqlc.User{}, pgx.ErrNoRows
		},
	}
	router := adminRouter(store, &fakeAuditSvc{}, &fakeFraudSvc{}, adminUser1)
	rec := doRequest(t, router, http.MethodGet, "/admin/users/nonexistent", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestAdminGetUser_HappyPath(t *testing.T) {
	store := &fakeAdminUserStore{
		getUserByID: func(ctx context.Context, id string) (sqlc.User, error) {
			return sqlc.User{ID: id, Email: "test@test.com", Role: "clipper", CreatedAt: pgtype.Timestamptz{Valid: true}}, nil
		},
	}
	router := adminRouter(store, &fakeAuditSvc{}, &fakeFraudSvc{}, adminUser1)
	rec := doRequest(t, router, http.MethodGet, "/admin/users/user_1", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var out struct {
		User struct {
			ID string `json:"id"`
		} `json:"user"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.User.ID != "user_1" {
		t.Errorf("expected user_1, got %q", out.User.ID)
	}
}

func TestAdminListFraudFlags_HappyPath(t *testing.T) {
	fraud := &fakeFraudSvc{
		listOpen: func(ctx context.Context) ([]sqlc.FraudFlag, error) {
			return []sqlc.FraudFlag{
				{
					ID:        pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
					FlagType:  "duplicate_url",
					Severity:  "high",
					Status:    "open",
					CreatedAt: pgtype.Timestamptz{Valid: true},
				},
			}, nil
		},
	}
	router := adminRouter(&fakeAdminUserStore{}, &fakeAuditSvc{}, fraud, adminUser1)
	rec := doRequest(t, router, http.MethodGet, "/admin/fraud-flags", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var out struct {
		Flags []struct {
			FlagType string `json:"flag_type"`
			Severity string `json:"severity"`
		} `json:"flags"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out.Flags) != 1 {
		t.Fatalf("expected 1 flag, got %d", len(out.Flags))
	}
	if out.Flags[0].FlagType != "duplicate_url" {
		t.Errorf("expected duplicate_url, got %q", out.Flags[0].FlagType)
	}
}

func TestAdminResolveFlag_HappyPath(t *testing.T) {
	fraud := &fakeFraudSvc{
		resolve: func(ctx context.Context, flagID pgtype.UUID, resolvedBy, resolution string) (*sqlc.FraudFlag, error) {
			return &sqlc.FraudFlag{
				ID:       flagID,
				Status:   "resolved",
				Severity: "high",
				FlagType: "duplicate_url",
			}, nil
		},
	}
	router := adminRouter(&fakeAdminUserStore{}, &fakeAuditSvc{}, fraud, adminUser1)
	body := `{"resolution":"investigated and confirmed"}`
	rec := doRequestJSON(t, router, http.MethodPost, "/admin/fraud-flags/01020304-0506-0708-090a-0b0c0d0e0f10/resolve", body)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAdminDismissFlag_HappyPath(t *testing.T) {
	fraud := &fakeFraudSvc{
		dismiss: func(ctx context.Context, flagID pgtype.UUID, resolvedBy, reason string) (*sqlc.FraudFlag, error) {
			return &sqlc.FraudFlag{
				ID:       flagID,
				Status:   "dismissed",
				Severity: "low",
				FlagType: "manual_review",
			}, nil
		},
	}
	router := adminRouter(&fakeAdminUserStore{}, &fakeAuditSvc{}, fraud, adminUser1)
	body := `{"reason":"not applicable"}`
	rec := doRequestJSON(t, router, http.MethodPost, "/admin/fraud-flags/01020304-0506-0708-090a-0b0c0d0e0f10/dismiss", body)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAdminAuditLogs_HappyPath(t *testing.T) {
	router := adminRouter(&fakeAdminUserStore{}, &fakeAuditSvc{}, &fakeFraudSvc{}, adminUser1)
	rec := doRequest(t, router, http.MethodGet, "/admin/audit-logs", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAdminOverrideCampaign_Forbidden_NonAdmin(t *testing.T) {
	store := &fakeAdminUserStore{}
	router := adminRouter(store, &fakeAuditSvc{}, &fakeFraudSvc{}, regularUser)
	body := `{"status":"paused"}`
	rec := doRequestJSON(t, router, http.MethodPost, "/admin/campaigns/0102030405060708090a0b0c0d0e0f10/override", body)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestAdminOverrideCampaign_InvalidStatus(t *testing.T) {
	store := &fakeAdminUserStore{
		getCampaign: func(ctx context.Context, id pgtype.UUID) (sqlc.Campaign, error) {
			return sqlc.Campaign{ID: id, Status: "active"}, nil
		},
	}
	router := adminRouter(store, &fakeAuditSvc{}, &fakeFraudSvc{}, adminUser1)
	body := `{"status":"invalid_status"}`
	rec := doRequestJSON(t, router, http.MethodPost, "/admin/campaigns/0102030405060708090a0b0c0d0e0f10/override", body)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAdminOverrideCampaign_HappyPath(t *testing.T) {
	store := &fakeAdminUserStore{
		getCampaign: func(ctx context.Context, id pgtype.UUID) (sqlc.Campaign, error) {
			return sqlc.Campaign{ID: id, Status: "active", CpmRate: 100, TotalBudget: 5000, RemainingBudget: 5000, PlatformFee: 500}, nil
		},
		updateStatus: func(ctx context.Context, arg sqlc.UpdateCampaignStatusParams) (sqlc.Campaign, error) {
			return sqlc.Campaign{ID: arg.ID, Status: arg.Status, CpmRate: 100, TotalBudget: 5000, RemainingBudget: 5000, PlatformFee: 500, CreatedAt: pgtype.Timestamptz{Valid: true}, UpdatedAt: pgtype.Timestamptz{Valid: true}}, nil
		},
	}
	audit := &fakeAuditSvc{
		log: func(ctx context.Context, actorID, action, resourceType, resourceID string, details map[string]any, ipAddress string) error {
			if action != "admin.campaign.override" {
				t.Errorf("expected admin.campaign.override, got %q", action)
			}
			return nil
		},
	}
	router := adminRouter(store, audit, &fakeFraudSvc{}, adminUser1)
	body := `{"status":"paused"}`
	rec := doRequestJSON(t, router, http.MethodPost, "/admin/campaigns/0102030405060708090a0b0c0d0e0f10/override", body)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAdminListUsers_DefaultPagination(t *testing.T) {
	store := &fakeAdminUserStore{
		listUsers: func(ctx context.Context, arg sqlc.ListUsersParams) ([]sqlc.User, error) {
			if arg.Limit != 20 {
				t.Errorf("expected limit 20, got %d", arg.Limit)
			}
			if arg.Offset != 0 {
				t.Errorf("expected offset 0, got %d", arg.Offset)
			}
			return nil, nil
		},
		countUsers: func(ctx context.Context) (int32, error) {
			return 0, nil
		},
	}
	router := adminRouter(store, &fakeAuditSvc{}, &fakeFraudSvc{}, adminUser1)
	rec := doRequest(t, router, http.MethodGet, "/admin/users?page=0&page_size=0", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestAdminListUsers_CustomPagination(t *testing.T) {
	store := &fakeAdminUserStore{
		listUsers: func(ctx context.Context, arg sqlc.ListUsersParams) ([]sqlc.User, error) {
			if arg.Limit != 50 {
				t.Errorf("expected limit 50, got %d", arg.Limit)
			}
			if arg.Offset != 100 {
				t.Errorf("expected offset 100, got %d", arg.Offset)
			}
			return nil, nil
		},
		countUsers: func(ctx context.Context) (int32, error) {
			return 0, nil
		},
	}
	router := adminRouter(store, &fakeAuditSvc{}, &fakeFraudSvc{}, adminUser1)
	rec := doRequest(t, router, http.MethodGet, "/admin/users?page=3&page_size=50", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestAdminResolveFlag_Forbidden_NonAdmin(t *testing.T) {
	fraud := &fakeFraudSvc{}
	router := adminRouter(&fakeAdminUserStore{}, &fakeAuditSvc{}, fraud, regularUser)
	body := `{"resolution":"ok"}`
	rec := doRequestJSON(t, router, http.MethodPost, "/admin/fraud-flags/01020304-0506-0708-090a-0b0c0d0e0f10/resolve", body)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestAdminDismissFlag_Forbidden_NonAdmin(t *testing.T) {
	fraud := &fakeFraudSvc{}
	router := adminRouter(&fakeAdminUserStore{}, &fakeAuditSvc{}, fraud, regularUser)
	body := `{"reason":"no"}`
	rec := doRequestJSON(t, router, http.MethodPost, "/admin/fraud-flags/01020304-0506-0708-090a-0b0c0d0e0f10/dismiss", body)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}
