package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"clipin/apps/api/internal/auth"
	sqlc "clipin/apps/api/internal/db/sqlc"
	"clipin/apps/api/internal/http/handlers"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type fakeUserStore struct {
	updateUser     func(ctx context.Context, arg sqlc.UpdateUserParams) (sqlc.User, error)
	updateUserRole func(ctx context.Context, arg sqlc.UpdateUserRoleParams) (sqlc.User, error)
}

func (f *fakeUserStore) GetUserByID(ctx context.Context, id string) (sqlc.User, error) {
	return sqlc.User{}, nil
}

func (f *fakeUserStore) CreateUser(ctx context.Context, arg sqlc.CreateUserParams) (sqlc.User, error) {
	return sqlc.User{}, nil
}

func (f *fakeUserStore) UpdateUser(ctx context.Context, arg sqlc.UpdateUserParams) (sqlc.User, error) {
	if f.updateUser != nil {
		return f.updateUser(ctx, arg)
	}
	return sqlc.User{}, nil
}

func (f *fakeUserStore) UpdateUserRole(ctx context.Context, arg sqlc.UpdateUserRoleParams) (sqlc.User, error) {
	if f.updateUserRole != nil {
		return f.updateUserRole(ctx, arg)
	}
	return sqlc.User{}, nil
}

// withUser wraps the handler so a user is present in context, as the session
// middleware would.
func withUser(user *sqlc.User) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(auth.ContextWithUser(r.Context(), user)))
		})
	}
}

func userRouter(store auth.UserStore) http.Handler {
	r := chi.NewRouter()
	api := humachi.New(r, huma.DefaultConfig("ClipIN API", "1.0.0"))
	handlers.RegisterUserHandlers(api, store)
	return r
}

func doRequest(t *testing.T, handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, bytes.NewBufferString(body))
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func testUser() *sqlc.User {
	return &sqlc.User{
		ID:          "user_1",
		Email:       "clipper@example.com",
		DisplayName: pgtype.Text{String: "Clipper One", Valid: true},
		Role:        "clipper",
		CreatedAt:   pgtype.Timestamptz{Valid: true},
	}
}

func TestGetMe(t *testing.T) {
	handler := withUser(testUser())(userRouter(&fakeUserStore{}))
	rec := doRequest(t, handler, http.MethodGet, "/me", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var out struct {
		ID          string `json:"id"`
		Email       string `json:"email"`
		DisplayName string `json:"display_name"`
		Role        string `json:"role"`
		CreatedAt   string `json:"created_at"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if out.ID != "user_1" {
		t.Errorf("expected id user_1, got %q", out.ID)
	}
	if out.Email != "clipper@example.com" {
		t.Errorf("expected email clipper@example.com, got %q", out.Email)
	}
	if out.DisplayName != "Clipper One" {
		t.Errorf("expected display name Clipper One, got %q", out.DisplayName)
	}
	if out.Role != "clipper" {
		t.Errorf("expected role clipper, got %q", out.Role)
	}
}

func TestGetMeUnauthorized(t *testing.T) {
	rec := doRequest(t, userRouter(&fakeUserStore{}), http.MethodGet, "/me", "")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestPatchMeUpdatesDisplayName(t *testing.T) {
	store := &fakeUserStore{
		updateUser: func(ctx context.Context, arg sqlc.UpdateUserParams) (sqlc.User, error) {
			if arg.ID != "user_1" {
				t.Errorf("expected id user_1, got %q", arg.ID)
			}
			if !arg.DisplayName.Valid || arg.DisplayName.String != "New Name" {
				t.Errorf("expected display name New Name, got %+v", arg.DisplayName)
			}
			return sqlc.User{
				ID:          arg.ID,
				Email:       "clipper@example.com",
				DisplayName: arg.DisplayName,
				Role:        "clipper",
				CreatedAt:   pgtype.Timestamptz{Valid: true},
			}, nil
		},
	}
	handler := withUser(testUser())(userRouter(store))
	rec := doRequest(t, handler, http.MethodPatch, "/me", `{"display_name":"New Name"}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var out struct {
		DisplayName string `json:"display_name"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if out.DisplayName != "New Name" {
		t.Errorf("expected display name New Name, got %q", out.DisplayName)
	}
}

func TestPatchMeUnauthorized(t *testing.T) {
	rec := doRequest(t, userRouter(&fakeUserStore{}), http.MethodPatch, "/me", `{"display_name":"x"}`)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestPatchMeStoreError(t *testing.T) {
	store := &fakeUserStore{
		updateUser: func(ctx context.Context, arg sqlc.UpdateUserParams) (sqlc.User, error) {
			return sqlc.User{}, &storeError{}
		},
	}
	handler := withUser(testUser())(userRouter(store))
	rec := doRequest(t, handler, http.MethodPatch, "/me", `{"display_name":"x"}`)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

type storeError struct{}

func (e *storeError) Error() string { return "store failure" }

func TestPostRoleSetsOwner(t *testing.T) {
	store := &fakeUserStore{
		updateUserRole: func(ctx context.Context, arg sqlc.UpdateUserRoleParams) (sqlc.User, error) {
			if arg.ID != "user_1" {
				t.Errorf("expected id user_1, got %q", arg.ID)
			}
			if arg.Role != "owner" {
				t.Errorf("expected role owner, got %q", arg.Role)
			}
			return sqlc.User{
				ID:          arg.ID,
				Email:       "clipper@example.com",
				DisplayName: pgtype.Text{String: "Clipper One", Valid: true},
				Role:        arg.Role,
				CreatedAt:   pgtype.Timestamptz{Valid: true},
			}, nil
		},
	}
	handler := withUser(testUser())(userRouter(store))
	rec := doRequest(t, handler, http.MethodPost, "/me/role", `{"role":"owner"}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var out struct {
		Role string `json:"role"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if out.Role != "owner" {
		t.Errorf("expected role owner, got %q", out.Role)
	}
}

func TestPostRoleRejectsAdmin(t *testing.T) {
	store := &fakeUserStore{}
	handler := withUser(testUser())(userRouter(store))
	rec := doRequest(t, handler, http.MethodPost, "/me/role", `{"role":"admin"}`)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", rec.Code)
	}
}

func TestPostRoleRejectsUnknownRole(t *testing.T) {
	store := &fakeUserStore{}
	handler := withUser(testUser())(userRouter(store))
	rec := doRequest(t, handler, http.MethodPost, "/me/role", `{"role":"superadmin"}`)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", rec.Code)
	}
}

func TestPostRoleUnauthorized(t *testing.T) {
	rec := doRequest(t, userRouter(&fakeUserStore{}), http.MethodPost, "/me/role", `{"role":"owner"}`)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}