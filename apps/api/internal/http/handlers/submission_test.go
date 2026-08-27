package handlers_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	sqlc "clipin/apps/api/internal/db/sqlc"
	"clipin/apps/api/internal/http/handlers"
	"clipin/apps/api/internal/service"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// --- Mock SubmissionService ---

type mockSubmissionHandler struct {
	submit             func(ctx context.Context, campaignID pgtype.UUID, clipperID, postURL, platform string) (*sqlc.Submission, error)
	listByCampaign     func(ctx context.Context, campaignID pgtype.UUID) ([]sqlc.Submission, error)
	listByClipper      func(ctx context.Context, clipperID string) ([]sqlc.Submission, error)
	getByID            func(ctx context.Context, id pgtype.UUID) (*sqlc.Submission, error)
	approve            func(ctx context.Context, submissionID pgtype.UUID, ownerID string) (*sqlc.Submission, error)
	reject             func(ctx context.Context, submissionID pgtype.UUID, ownerID string, reason string) (*sqlc.Submission, error)
	batchApprove       func(ctx context.Context, submissionIDs []pgtype.UUID, ownerID string) (*service.BatchResult, error)
	batchReject        func(ctx context.Context, submissionIDs []pgtype.UUID, ownerID string, reason string) (*service.BatchResult, error)
	verifyOwnership    func(ctx context.Context, campaignID pgtype.UUID, ownerID string) error
}

func (m *mockSubmissionHandler) Submit(ctx context.Context, campaignID pgtype.UUID, clipperID, postURL, platform string) (*sqlc.Submission, error) {
	if m.submit != nil {
		return m.submit(ctx, campaignID, clipperID, postURL, platform)
	}
	return nil, nil
}

func (m *mockSubmissionHandler) ListByCampaign(ctx context.Context, campaignID pgtype.UUID) ([]sqlc.Submission, error) {
	if m.listByCampaign != nil {
		return m.listByCampaign(ctx, campaignID)
	}
	return nil, nil
}

func (m *mockSubmissionHandler) ListByClipper(ctx context.Context, clipperID string) ([]sqlc.Submission, error) {
	if m.listByClipper != nil {
		return m.listByClipper(ctx, clipperID)
	}
	return nil, nil
}

func (m *mockSubmissionHandler) GetByID(ctx context.Context, id pgtype.UUID) (*sqlc.Submission, error) {
	if m.getByID != nil {
		return m.getByID(ctx, id)
	}
	return nil, nil
}

func (m *mockSubmissionHandler) Approve(ctx context.Context, submissionID pgtype.UUID, ownerID string) (*sqlc.Submission, error) {
	if m.approve != nil {
		return m.approve(ctx, submissionID, ownerID)
	}
	return nil, nil
}

func (m *mockSubmissionHandler) Reject(ctx context.Context, submissionID pgtype.UUID, ownerID string, reason string) (*sqlc.Submission, error) {
	if m.reject != nil {
		return m.reject(ctx, submissionID, ownerID, reason)
	}
	return nil, nil
}

func (m *mockSubmissionHandler) BatchApprove(ctx context.Context, submissionIDs []pgtype.UUID, ownerID string) (*service.BatchResult, error) {
	if m.batchApprove != nil {
		return m.batchApprove(ctx, submissionIDs, ownerID)
	}
	return &service.BatchResult{}, nil
}

func (m *mockSubmissionHandler) BatchReject(ctx context.Context, submissionIDs []pgtype.UUID, ownerID string, reason string) (*service.BatchResult, error) {
	if m.batchReject != nil {
		return m.batchReject(ctx, submissionIDs, ownerID, reason)
	}
	return &service.BatchResult{}, nil
}

func (m *mockSubmissionHandler) VerifyCampaignOwnership(ctx context.Context, campaignID pgtype.UUID, ownerID string) error {
	if m.verifyOwnership != nil {
		return m.verifyOwnership(ctx, campaignID, ownerID)
	}
	return nil
}

func submissionRouter(svc handlers.SubmissionServiceInterface) http.Handler {
	r := chi.NewRouter()
	api := humachi.New(r, huma.DefaultConfig("ClipIN API", "1.0.0"))
	handlers.RegisterSubmissionHandlers(api, svc)
	return r
}

func submissionRouterAuth(svc handlers.SubmissionServiceInterface, user *sqlc.User) http.Handler {
	r := chi.NewRouter()
	api := humachi.New(r, huma.DefaultConfig("ClipIN API", "1.0.0"))
	handlers.RegisterSubmissionHandlers(api, svc)
	return withUser(user)(r)
}

func submissionRouterNoAuth(svc handlers.SubmissionServiceInterface) http.Handler {
	r := chi.NewRouter()
	api := humachi.New(r, huma.DefaultConfig("ClipIN API", "1.0.0"))
	handlers.RegisterSubmissionHandlers(api, svc)
	return r
}

var (
	subOwner  = &sqlc.User{ID: "owner_1", Email: "owner@test.com", Role: "owner"}
	subClipper = &sqlc.User{ID: "clipper_1", Email: "clipper@test.com", Role: "clipper"}
)

const (
	subCampaignIDHex = "0102030405060708090a0b0c0d0e0f10"
	subID1Hex        = "aabbccdd01020304aabbccdd01020304"
	subID2Hex        = "1122334455667788aabbccdd11223344"
	subID3Hex        = "deadbeef01020304deadbeef01020304"
)

// --- Batch Approve Tests ---

func TestBatchApprove_Unauthorized(t *testing.T) {
	svc := &mockSubmissionHandler{}
	router := submissionRouterNoAuth(svc)
	body := fmt.Sprintf(`{"submission_ids":["%s"]}`, subID1Hex)
	rec := doRequestJSON(t, router, http.MethodPost, "/campaigns/"+subCampaignIDHex+"/submissions/batch-approve", body)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestBatchApprove_Forbidden_Clipper(t *testing.T) {
	svc := &mockSubmissionHandler{}
	router := submissionRouterAuth(svc, subClipper)
	body := fmt.Sprintf(`{"submission_ids":["%s"]}`, subID1Hex)
	rec := doRequestJSON(t, router, http.MethodPost, "/campaigns/"+subCampaignIDHex+"/submissions/batch-approve", body)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestBatchApprove_CampaignNotFound(t *testing.T) {
	svc := &mockSubmissionHandler{
		verifyOwnership: func(_ context.Context, _ pgtype.UUID, _ string) error {
			return service.ErrCampaignNotFound
		},
	}
	router := submissionRouterAuth(svc, subOwner)
	body := fmt.Sprintf(`{"submission_ids":["%s"]}`, subID1Hex)
	rec := doRequestJSON(t, router, http.MethodPost, "/campaigns/"+subCampaignIDHex+"/submissions/batch-approve", body)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestBatchApprove_EmptyIDs(t *testing.T) {
	svc := &mockSubmissionHandler{
		verifyOwnership: func(_ context.Context, _ pgtype.UUID, _ string) error {
			return nil
		},
	}
	router := submissionRouterAuth(svc, subOwner)
	body := `{"submission_ids":[]}`
	rec := doRequestJSON(t, router, http.MethodPost, "/campaigns/"+subCampaignIDHex+"/submissions/batch-approve", body)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestBatchApprove_InvalidID(t *testing.T) {
	svc := &mockSubmissionHandler{
		verifyOwnership: func(_ context.Context, _ pgtype.UUID, _ string) error {
			return nil
		},
	}
	router := submissionRouterAuth(svc, subOwner)
	body := `{"submission_ids":["not-a-uuid"]}`
	rec := doRequestJSON(t, router, http.MethodPost, "/campaigns/"+subCampaignIDHex+"/submissions/batch-approve", body)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestBatchApprove_HappyPath_AllApproved(t *testing.T) {
	svc := &mockSubmissionHandler{
		verifyOwnership: func(_ context.Context, _ pgtype.UUID, _ string) error {
			return nil
		},
		batchApprove: func(_ context.Context, ids []pgtype.UUID, _ string) (*service.BatchResult, error) {
			return &service.BatchResult{Approved: len(ids), Failed: 0}, nil
		},
	}
	router := submissionRouterAuth(svc, subOwner)
	body := fmt.Sprintf(`{"submission_ids":["%s","%s"]}`, subID1Hex, subID2Hex)
	rec := doRequestJSON(t, router, http.MethodPost, "/campaigns/"+subCampaignIDHex+"/submissions/batch-approve", body)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var out struct {
		Approved int `json:"approved"`
		Failed   int `json:"failed"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Approved != 2 {
		t.Errorf("expected 2 approved, got %d", out.Approved)
	}
	if out.Failed != 0 {
		t.Errorf("expected 0 failed, got %d", out.Failed)
	}
}

func TestBatchApprove_PartialFailure(t *testing.T) {
	svc := &mockSubmissionHandler{
		verifyOwnership: func(_ context.Context, _ pgtype.UUID, _ string) error {
			return nil
		},
		batchApprove: func(_ context.Context, ids []pgtype.UUID, _ string) (*service.BatchResult, error) {
			return &service.BatchResult{
				Approved: 1,
				Failed:   1,
				Errors: []service.BatchItemErr{
					{ID: subID2Hex, Reason: "submission not found"},
				},
			}, nil
		},
	}
	router := submissionRouterAuth(svc, subOwner)
	body := fmt.Sprintf(`{"submission_ids":["%s","%s"]}`, subID1Hex, subID2Hex)
	rec := doRequestJSON(t, router, http.MethodPost, "/campaigns/"+subCampaignIDHex+"/submissions/batch-approve", body)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var out struct {
		Approved int `json:"approved"`
		Failed   int `json:"failed"`
		Errors   []struct {
			ID     string `json:"id"`
			Reason string `json:"reason"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Approved != 1 {
		t.Errorf("expected 1 approved, got %d", out.Approved)
	}
	if out.Failed != 1 {
		t.Errorf("expected 1 failed, got %d", out.Failed)
	}
	if len(out.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(out.Errors))
	}
	if out.Errors[0].ID != subID2Hex {
		t.Errorf("expected error ID %s, got %s", subID2Hex, out.Errors[0].ID)
	}
}

func TestBatchApprove_AlreadyApproved(t *testing.T) {
	svc := &mockSubmissionHandler{
		verifyOwnership: func(_ context.Context, _ pgtype.UUID, _ string) error {
			return nil
		},
		batchApprove: func(_ context.Context, ids []pgtype.UUID, _ string) (*service.BatchResult, error) {
			return &service.BatchResult{
				Approved: 0,
				Failed:   1,
				Errors: []service.BatchItemErr{
					{ID: subID1Hex, Reason: "submission is not in pending status"},
				},
			}, nil
		},
	}
	router := submissionRouterAuth(svc, subOwner)
	body := fmt.Sprintf(`{"submission_ids":["%s"]}`, subID1Hex)
	rec := doRequestJSON(t, router, http.MethodPost, "/campaigns/"+subCampaignIDHex+"/submissions/batch-approve", body)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var out struct {
		Approved int `json:"approved"`
		Failed   int `json:"failed"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Approved != 0 {
		t.Errorf("expected 0 approved, got %d", out.Approved)
	}
	if out.Failed != 1 {
		t.Errorf("expected 1 failed, got %d", out.Failed)
	}
}

// --- Batch Reject Tests ---

func TestBatchReject_Unauthorized(t *testing.T) {
	svc := &mockSubmissionHandler{}
	router := submissionRouterNoAuth(svc)
	body := fmt.Sprintf(`{"submission_ids":["%s"],"reason":"bad"}`, subID1Hex)
	rec := doRequestJSON(t, router, http.MethodPost, "/campaigns/"+subCampaignIDHex+"/submissions/batch-reject", body)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestBatchReject_Forbidden_Clipper(t *testing.T) {
	svc := &mockSubmissionHandler{}
	router := submissionRouterAuth(svc, subClipper)
	body := fmt.Sprintf(`{"submission_ids":["%s"],"reason":"bad"}`, subID1Hex)
	rec := doRequestJSON(t, router, http.MethodPost, "/campaigns/"+subCampaignIDHex+"/submissions/batch-reject", body)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestBatchReject_HappyPath_WithReason(t *testing.T) {
	svc := &mockSubmissionHandler{
		verifyOwnership: func(_ context.Context, _ pgtype.UUID, _ string) error {
			return nil
		},
		batchReject: func(_ context.Context, ids []pgtype.UUID, _ string, reason string) (*service.BatchResult, error) {
			if reason != "low quality" {
				t.Errorf("expected reason 'low quality', got %q", reason)
			}
			return &service.BatchResult{Rejected: len(ids), Failed: 0}, nil
		},
	}
	router := submissionRouterAuth(svc, subOwner)
	body := fmt.Sprintf(`{"submission_ids":["%s","%s"],"reason":"low quality"}`, subID1Hex, subID2Hex)
	rec := doRequestJSON(t, router, http.MethodPost, "/campaigns/"+subCampaignIDHex+"/submissions/batch-reject", body)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var out struct {
		Rejected int `json:"rejected"`
		Failed   int `json:"failed"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Rejected != 2 {
		t.Errorf("expected 2 rejected, got %d", out.Rejected)
	}
	if out.Failed != 0 {
		t.Errorf("expected 0 failed, got %d", out.Failed)
	}
}

func TestBatchReject_PartialFailure(t *testing.T) {
	svc := &mockSubmissionHandler{
		verifyOwnership: func(_ context.Context, _ pgtype.UUID, _ string) error {
			return nil
		},
		batchReject: func(_ context.Context, ids []pgtype.UUID, _ string, _ string) (*service.BatchResult, error) {
			return &service.BatchResult{
				Rejected: 1,
				Failed:   1,
				Errors: []service.BatchItemErr{
					{ID: subID3Hex, Reason: "submission not found"},
				},
			}, nil
		},
	}
	router := submissionRouterAuth(svc, subOwner)
	body := fmt.Sprintf(`{"submission_ids":["%s","%s"]}`, subID1Hex, subID3Hex)
	rec := doRequestJSON(t, router, http.MethodPost, "/campaigns/"+subCampaignIDHex+"/submissions/batch-reject", body)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var out struct {
		Rejected int `json:"rejected"`
		Failed   int `json:"failed"`
		Errors   []struct {
			ID     string `json:"id"`
			Reason string `json:"reason"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Rejected != 1 {
		t.Errorf("expected 1 rejected, got %d", out.Rejected)
	}
	if out.Failed != 1 {
		t.Errorf("expected 1 failed, got %d", out.Failed)
	}
	if len(out.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(out.Errors))
	}
}

func TestBatchReject_CampaignNotFound(t *testing.T) {
	svc := &mockSubmissionHandler{
		verifyOwnership: func(_ context.Context, _ pgtype.UUID, _ string) error {
			return service.ErrCampaignNotFound
		},
	}
	router := submissionRouterAuth(svc, subOwner)
	body := fmt.Sprintf(`{"submission_ids":["%s"]}`, subID1Hex)
	rec := doRequestJSON(t, router, http.MethodPost, "/campaigns/"+subCampaignIDHex+"/submissions/batch-reject", body)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestBatchReject_EmptyIDs(t *testing.T) {
	svc := &mockSubmissionHandler{
		verifyOwnership: func(_ context.Context, _ pgtype.UUID, _ string) error {
			return nil
		},
	}
	router := submissionRouterAuth(svc, subOwner)
	body := `{"submission_ids":[]}`
	rec := doRequestJSON(t, router, http.MethodPost, "/campaigns/"+subCampaignIDHex+"/submissions/batch-reject", body)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d body=%s", rec.Code, rec.Body.String())
	}
}
