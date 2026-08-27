package auth_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"clipin/apps/api/internal/auth"
	sqlc "clipin/apps/api/internal/db/sqlc"

	"github.com/jackc/pgx/v5"
)

type fakeUserStore struct {
	users       map[string]sqlc.User
	getErr      error
	createErr   error
	createdUser *sqlc.CreateUserParams
}

func (f *fakeUserStore) GetUserByID(ctx context.Context, id string) (sqlc.User, error) {
	if f.getErr != nil {
		return sqlc.User{}, f.getErr
	}
	if u, ok := f.users[id]; ok {
		return u, nil
	}
	return sqlc.User{}, pgx.ErrNoRows
}

func (f *fakeUserStore) CreateUser(ctx context.Context, arg sqlc.CreateUserParams) (sqlc.User, error) {
	f.createdUser = &arg
	if f.createErr != nil {
		return sqlc.User{}, f.createErr
	}
	return sqlc.User{
		ID:          arg.ID,
		Email:       arg.Email,
		DisplayName: arg.DisplayName,
		Role:        arg.Role,
	}, nil
}

func (f *fakeUserStore) UpdateUser(ctx context.Context, arg sqlc.UpdateUserParams) (sqlc.User, error) {
	return sqlc.User{}, nil
}

func (f *fakeUserStore) UpdateUserRole(ctx context.Context, arg sqlc.UpdateUserRoleParams) (sqlc.User, error) {
	return sqlc.User{}, nil
}

func withSession(store auth.UserStore, claims *auth.Claims) *httptest.ResponseRecorder {
	handler := auth.SessionMiddleware(store, nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	if claims != nil {
		req = req.WithContext(auth.ContextWithClaims(req.Context(), claims))
	}
	handler.ServeHTTP(rec, req)

	return rec
}

func TestSessionMiddlewareLoadsExistingUser(t *testing.T) {
	store := &fakeUserStore{users: map[string]sqlc.User{
		"user_1": {ID: "user_1", Email: "a@b.com", Role: "clipper"},
	}}
	rec := withSession(store, &auth.Claims{Sub: "user_1"})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if store.createdUser != nil {
		t.Fatalf("expected no user creation, got %+v", store.createdUser)
	}
}

func TestSessionMiddlewareCreatesMissingUser(t *testing.T) {
	store := &fakeUserStore{users: map[string]sqlc.User{}}
	rec := withSession(store, &auth.Claims{Sub: "user_new", Email: "new@b.com", Name: "New Person"})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if store.createdUser == nil {
		t.Fatal("expected user creation")
	}
	if store.createdUser.ID != "user_new" {
		t.Errorf("expected id user_new, got %q", store.createdUser.ID)
	}
	if store.createdUser.Email != "new@b.com" {
		t.Errorf("expected email new@b.com, got %q", store.createdUser.Email)
	}
	if store.createdUser.Role != "clipper" {
		t.Errorf("expected role clipper, got %q", store.createdUser.Role)
	}
	if !store.createdUser.DisplayName.Valid || store.createdUser.DisplayName.String != "New Person" {
		t.Errorf("expected display name New Person, got %+v", store.createdUser.DisplayName)
	}
}

func TestSessionMiddlewareCreatesUserWithoutName(t *testing.T) {
	store := &fakeUserStore{users: map[string]sqlc.User{}}
	rec := withSession(store, &auth.Claims{Sub: "user_new", Email: "new@b.com"})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if store.createdUser == nil {
		t.Fatal("expected user creation")
	}
	if store.createdUser.DisplayName.Valid {
		t.Errorf("expected invalid display name, got %+v", store.createdUser.DisplayName)
	}
}

func TestSessionMiddlewarePassesThroughWithoutClaims(t *testing.T) {
	store := &fakeUserStore{users: map[string]sqlc.User{}}
	rec := withSession(store, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if store.createdUser != nil {
		t.Fatal("expected no user creation without claims")
	}
}

func TestSessionMiddlewareReturns500OnLoadError(t *testing.T) {
	store := &fakeUserStore{users: map[string]sqlc.User{}, getErr: errors.New("db down")}
	rec := withSession(store, &auth.Claims{Sub: "user_1"})

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestSessionMiddlewareReturns500OnCreateError(t *testing.T) {
	store := &fakeUserStore{users: map[string]sqlc.User{}, createErr: errors.New("db down")}
	rec := withSession(store, &auth.Claims{Sub: "user_new", Email: "new@b.com"})

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestSessionMiddlewareStoresUserInContext(t *testing.T) {
	store := &fakeUserStore{users: map[string]sqlc.User{
		"user_1": {ID: "user_1", Email: "a@b.com", Role: "owner"},
	}}

	var captured *sqlc.User
	handler := auth.SessionMiddleware(store, nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if u, ok := auth.UserFromContext(r.Context()); ok {
			captured = u
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(auth.ContextWithClaims(req.Context(), &auth.Claims{Sub: "user_1"}))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if captured == nil {
		t.Fatal("expected user in context")
	}
	if captured.Role != "owner" {
		t.Errorf("expected role owner, got %q", captured.Role)
	}
}