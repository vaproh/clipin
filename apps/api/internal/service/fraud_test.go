package service_test

import (
	"context"
	"testing"

	sqlc "clipin/apps/api/internal/db/sqlc"
	"clipin/apps/api/internal/service"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// mockFraudStore implements FraudStore for testing.
type mockFraudStore struct {
	createFlag    func(ctx context.Context, arg sqlc.CreateFraudFlagParams) (sqlc.FraudFlag, error)
	listOpen      func(ctx context.Context) ([]sqlc.FraudFlag, error)
	getByID       func(ctx context.Context, id pgtype.UUID) (sqlc.FraudFlag, error)
	updateStatus  func(ctx context.Context, arg sqlc.UpdateFraudFlagStatusParams) (sqlc.FraudFlag, error)
	countByUser   func(ctx context.Context, userID pgtype.Text) (int32, error)
	listHighRisk  func(ctx context.Context, minFlags int32) ([]sqlc.ListUsersWithManyFlagsRow, error)
}

func (m *mockFraudStore) CreateFraudFlag(ctx context.Context, arg sqlc.CreateFraudFlagParams) (sqlc.FraudFlag, error) {
	if m.createFlag != nil {
		return m.createFlag(ctx, arg)
	}
	return sqlc.FraudFlag{}, nil
}

func (m *mockFraudStore) ListOpenFraudFlags(ctx context.Context) ([]sqlc.FraudFlag, error) {
	if m.listOpen != nil {
		return m.listOpen(ctx)
	}
	return nil, nil
}

func (m *mockFraudStore) GetFraudFlagByID(ctx context.Context, id pgtype.UUID) (sqlc.FraudFlag, error) {
	if m.getByID != nil {
		return m.getByID(ctx, id)
	}
	return sqlc.FraudFlag{}, pgx.ErrNoRows
}

func (m *mockFraudStore) UpdateFraudFlagStatus(ctx context.Context, arg sqlc.UpdateFraudFlagStatusParams) (sqlc.FraudFlag, error) {
	if m.updateStatus != nil {
		return m.updateStatus(ctx, arg)
	}
	return sqlc.FraudFlag{}, nil
}

func (m *mockFraudStore) CountFraudFlagsByUser(ctx context.Context, userID pgtype.Text) (int32, error) {
	if m.countByUser != nil {
		return m.countByUser(ctx, userID)
	}
	return 0, nil
}

func (m *mockFraudStore) ListUsersWithManyFlags(ctx context.Context, minFlags int32) ([]sqlc.ListUsersWithManyFlagsRow, error) {
	if m.listHighRisk != nil {
		return m.listHighRisk(ctx, minFlags)
	}
	return nil, nil
}

var testFlagID = pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true}
var testSubmissionID = pgtype.UUID{Bytes: [16]byte{2, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true}

func testFlag() sqlc.FraudFlag {
	return sqlc.FraudFlag{
		ID:           testFlagID,
		FlagType:     "duplicate_url",
		Severity:     "high",
		Status:       "open",
		CreatedAt:    pgtype.Timestamptz{Valid: true},
		UpdatedAt:    pgtype.Timestamptz{Valid: true},
	}
}

func TestFraudFlagSubmission_HappyPath(t *testing.T) {
	store := &mockFraudStore{
		createFlag: func(ctx context.Context, arg sqlc.CreateFraudFlagParams) (sqlc.FraudFlag, error) {
			if arg.SubmissionID != testSubmissionID {
				t.Errorf("expected submission_id, got %+v", arg.SubmissionID)
			}
			if arg.FlagType != "suspicious_views" {
				t.Errorf("expected suspicious_views, got %q", arg.FlagType)
			}
			if arg.Severity != "medium" {
				t.Errorf("expected medium, got %q", arg.Severity)
			}
			return testFlag(), nil
		},
	}
	svc := service.NewFraudService(store)
	flag, err := svc.FlagSubmission(context.Background(), testSubmissionID, "suspicious_views", "medium", "unusual spike")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if flag.FlagType != "duplicate_url" {
		t.Errorf("expected duplicate_url, got %q", flag.FlagType)
	}
}

func TestFraudFlagUser_HappyPath(t *testing.T) {
	store := &mockFraudStore{
		createFlag: func(ctx context.Context, arg sqlc.CreateFraudFlagParams) (sqlc.FraudFlag, error) {
			if !arg.UserID.Valid || arg.UserID.String != "user_1" {
				t.Errorf("expected user_1, got %+v", arg.UserID)
			}
			return testFlag(), nil
		},
	}
	svc := service.NewFraudService(store)
	_, err := svc.FlagUser(context.Background(), "user_1", "bot_pattern", "high", "detected bot behavior")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFraudListOpenFlags(t *testing.T) {
	flags := []sqlc.FraudFlag{testFlag(), testFlag()}
	store := &mockFraudStore{
		listOpen: func(ctx context.Context) ([]sqlc.FraudFlag, error) {
			return flags, nil
		},
	}
	svc := service.NewFraudService(store)
	result, err := svc.ListOpenFlags(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 flags, got %d", len(result))
	}
}

func TestFraudResolveFlag_HappyPath(t *testing.T) {
	store := &mockFraudStore{
		getByID: func(ctx context.Context, id pgtype.UUID) (sqlc.FraudFlag, error) {
			return testFlag(), nil
		},
		updateStatus: func(ctx context.Context, arg sqlc.UpdateFraudFlagStatusParams) (sqlc.FraudFlag, error) {
			if arg.Status != "resolved" {
				t.Errorf("expected resolved, got %q", arg.Status)
			}
			if !arg.ResolvedBy.Valid || arg.ResolvedBy.String != "admin_1" {
				t.Errorf("expected admin_1 as resolver")
			}
			f := testFlag()
			f.Status = "resolved"
			f.ResolvedBy = arg.ResolvedBy
			f.Resolution = arg.Resolution
			return f, nil
		},
	}
	svc := service.NewFraudService(store)
	flag, err := svc.ResolveFlag(context.Background(), testFlagID, "admin_1", "investigated, false positive")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if flag.Status != "resolved" {
		t.Errorf("expected resolved, got %q", flag.Status)
	}
}

func TestFraudDismissFlag_HappyPath(t *testing.T) {
	store := &mockFraudStore{
		getByID: func(ctx context.Context, id pgtype.UUID) (sqlc.FraudFlag, error) {
			return testFlag(), nil
		},
		updateStatus: func(ctx context.Context, arg sqlc.UpdateFraudFlagStatusParams) (sqlc.FraudFlag, error) {
			if arg.Status != "dismissed" {
				t.Errorf("expected dismissed, got %q", arg.Status)
			}
			f := testFlag()
			f.Status = "dismissed"
			return f, nil
		},
	}
	svc := service.NewFraudService(store)
	_, err := svc.DismissFlag(context.Background(), testFlagID, "admin_1", "not applicable")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFraudResolveFlag_NotFound(t *testing.T) {
	store := &mockFraudStore{
		getByID: func(ctx context.Context, id pgtype.UUID) (sqlc.FraudFlag, error) {
			return sqlc.FraudFlag{}, pgx.ErrNoRows
		},
	}
	svc := service.NewFraudService(store)
	_, err := svc.ResolveFlag(context.Background(), testFlagID, "admin_1", "note")
	if err != service.ErrFlagNotFound {
		t.Fatalf("expected ErrFlagNotFound, got %v", err)
	}
}

func TestFraudDismissFlag_NotFound(t *testing.T) {
	store := &mockFraudStore{
		getByID: func(ctx context.Context, id pgtype.UUID) (sqlc.FraudFlag, error) {
			return sqlc.FraudFlag{}, pgx.ErrNoRows
		},
	}
	svc := service.NewFraudService(store)
	_, err := svc.DismissFlag(context.Background(), testFlagID, "admin_1", "note")
	if err != service.ErrFlagNotFound {
		t.Fatalf("expected ErrFlagNotFound, got %v", err)
	}
}

func TestFraudGetUserFlagCount(t *testing.T) {
	store := &mockFraudStore{
		countByUser: func(ctx context.Context, userID pgtype.Text) (int32, error) {
			if userID.String != "user_1" {
				t.Errorf("expected user_1, got %q", userID.String)
			}
			return 5, nil
		},
	}
	svc := service.NewFraudService(store)
	count, err := svc.GetUserFlagCount(context.Background(), "user_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 5 {
		t.Errorf("expected 5, got %d", count)
	}
}
