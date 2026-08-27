//go:build integration

package integration

import (
	"context"
	"errors"
	"testing"

	sqlc "clipin/apps/api/internal/db/sqlc"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestCreateAndGetUser(t *testing.T) {
	cleanupAll(t)

	user, err := testDB.Queries.CreateUser(context.Background(), sqlc.CreateUserParams{
		ID:          "user_test123",
		Email:       "test@example.com",
		DisplayName: pgtype.Text{String: "Test User", Valid: true},
		Role:        "clipper",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	fetched, err := testDB.Queries.GetUserByID(context.Background(), "user_test123")
	if err != nil {
		t.Fatalf("GetUserByID: %v", err)
	}

	assertEqual(t, "ID", fetched.ID, user.ID)
	assertEqual(t, "Email", fetched.Email, "test@example.com")
	assertEqual(t, "Role", fetched.Role, "clipper")
	if !fetched.DisplayName.Valid || fetched.DisplayName.String != "Test User" {
		t.Errorf("DisplayName: got %v, want 'Test User'", fetched.DisplayName)
	}
}

func TestGetUserByEmail(t *testing.T) {
	cleanupAll(t)

	_, err := testDB.Queries.CreateUser(context.Background(), sqlc.CreateUserParams{
		ID:    "userByEmail1",
		Email: "byemail@example.com",
		Role:  "clipper",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	fetched, err := testDB.Queries.GetUserByEmail(context.Background(), "byemail@example.com")
	if err != nil {
		t.Fatalf("GetUserByEmail: %v", err)
	}
	assertEqual(t, "ID", fetched.ID, "userByEmail1")
	assertEqual(t, "Email", fetched.Email, "byemail@example.com")
}

func TestGetUserByEmailNotFound(t *testing.T) {
	cleanupAll(t)

	_, err := testDB.Queries.GetUserByEmail(context.Background(), "nonexistent@example.com")
	if err == nil {
		t.Error("expected error for nonexistent email, got nil")
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("expected pgx.ErrNoRows, got %v", err)
	}
}

func TestUpdateUser(t *testing.T) {
	cleanupAll(t)

	_, err := testDB.Queries.CreateUser(context.Background(), sqlc.CreateUserParams{
		ID:          "user_upd1",
		Email:       "old@example.com",
		DisplayName: pgtype.Text{String: "Old Name", Valid: true},
		Role:        "clipper",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	updated, err := testDB.Queries.UpdateUser(context.Background(), sqlc.UpdateUserParams{
		ID:          "user_upd1",
		DisplayName: pgtype.Text{String: "New Name", Valid: true},
		Email:       "new@example.com",
	})
	if err != nil {
		t.Fatalf("UpdateUser: %v", err)
	}

	assertEqual(t, "DisplayName", updated.DisplayName.String, "New Name")
	assertEqual(t, "Email", updated.Email, "new@example.com")
}

func TestUpdateUserPartial(t *testing.T) {
	cleanupAll(t)

	_, err := testDB.Queries.CreateUser(context.Background(), sqlc.CreateUserParams{
		ID:          "user_part1",
		Email:       "keep@example.com",
		DisplayName: pgtype.Text{String: "Keep Name", Valid: true},
		Role:        "clipper",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	// Update only email, leave display_name NULL (COALESCE preserves original)
	updated, err := testDB.Queries.UpdateUser(context.Background(), sqlc.UpdateUserParams{
		ID:          "user_part1",
		DisplayName: pgtype.Text{},
		Email:       "changed@example.com",
	})
	if err != nil {
		t.Fatalf("UpdateUser partial: %v", err)
	}

	assertEqual(t, "DisplayName", updated.DisplayName.String, "Keep Name")
	assertEqual(t, "Email", updated.Email, "changed@example.com")
}

func TestUpdateUserRole(t *testing.T) {
	cleanupAll(t)

	_, err := testDB.Queries.CreateUser(context.Background(), sqlc.CreateUserParams{
		ID:    "user_role1",
		Email: "role@example.com",
		Role:  "clipper",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	roles := []string{"owner", "admin", "clipper"}
	current := "clipper"
	for _, role := range roles {
		updated, err := testDB.Queries.UpdateUserRole(context.Background(), sqlc.UpdateUserRoleParams{
			ID:   "user_role1",
			Role: role,
		})
		if err != nil {
			t.Fatalf("UpdateUserRole to %s: %v", role, err)
		}
		assertEqual(t, "Role after "+current+"->"+role, updated.Role, role)
		current = role
	}
}

func TestDuplicateUserIDRejected(t *testing.T) {
	cleanupAll(t)

	params := sqlc.CreateUserParams{
		ID:    "user_dup1",
		Email: "dup1@example.com",
		Role:  "clipper",
	}
	_, err := testDB.Queries.CreateUser(context.Background(), params)
	if err != nil {
		t.Fatalf("first CreateUser: %v", err)
	}

	params.Email = "dup2@example.com"
	_, err = testDB.Queries.CreateUser(context.Background(), params)
	if err == nil {
		t.Error("expected error for duplicate user ID, got nil")
	}
}

func TestCreateUserNullDisplayName(t *testing.T) {
	cleanupAll(t)

	user, err := testDB.Queries.CreateUser(context.Background(), sqlc.CreateUserParams{
		ID:    "user_null1",
		Email: "nullname@example.com",
		Role:  "clipper",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	if user.DisplayName.Valid {
		t.Errorf("expected DisplayName to be NULL, got %q", user.DisplayName.String)
	}
}

func TestGetUserByIDNotFound(t *testing.T) {
	cleanupAll(t)

	_, err := testDB.Queries.GetUserByID(context.Background(), "nonexistent_id")
	if err == nil {
		t.Error("expected error for nonexistent ID, got nil")
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("expected pgx.ErrNoRows, got %v", err)
	}
}
