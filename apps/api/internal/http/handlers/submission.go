package handlers

import (
	"context"
	"errors"
	"fmt"

	sqlc "clipin/apps/api/internal/db/sqlc"
	"clipin/apps/api/internal/service"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgtype"
)

// SubmissionServiceInterface is the subset of SubmissionService the handlers need.
type SubmissionServiceInterface interface {
	Submit(ctx context.Context, campaignID pgtype.UUID, clipperID, postURL, platform string) (*sqlc.Submission, error)
	ListByCampaign(ctx context.Context, campaignID pgtype.UUID) ([]sqlc.Submission, error)
	ListByClipper(ctx context.Context, clipperID string) ([]sqlc.Submission, error)
	GetByID(ctx context.Context, id pgtype.UUID) (*sqlc.Submission, error)
	Approve(ctx context.Context, submissionID pgtype.UUID, ownerID string) (*sqlc.Submission, error)
	Reject(ctx context.Context, submissionID pgtype.UUID, ownerID string, reason string) (*sqlc.Submission, error)
	BatchApprove(ctx context.Context, submissionIDs []pgtype.UUID, ownerID string) (*service.BatchResult, error)
	BatchReject(ctx context.Context, submissionIDs []pgtype.UUID, ownerID string, reason string) (*service.BatchResult, error)
	VerifyCampaignOwnership(ctx context.Context, campaignID pgtype.UUID, ownerID string) error
}

type submissionListItem struct {
	ID              string  `json:"id"`
	CampaignID      string  `json:"campaign_id"`
	ClipperID       string  `json:"clipper_id"`
	PostURL         string  `json:"post_url"`
	Platform        string  `json:"platform"`
	Status          string  `json:"status"`
	RejectionReason *string `json:"rejection_reason,omitempty"`
	ApprovedAt      *string `json:"approved_at,omitempty"`
	AutoApprovedAt  *string `json:"auto_approved_at,omitempty"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

func toSubmissionListItem(s sqlc.Submission) submissionListItem {
	item := submissionListItem{
		ID:         fmt.Sprintf("%x", s.ID.Bytes),
		CampaignID: fmt.Sprintf("%x", s.CampaignID.Bytes),
		ClipperID:  s.ClipperID,
		PostURL:    s.PostUrl,
		Platform:   s.Platform,
		Status:     s.Status,
		CreatedAt:  s.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  s.UpdatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
	}
	if s.RejectionReason.Valid {
		item.RejectionReason = &s.RejectionReason.String
	}
	if s.ApprovedAt.Valid {
		t := s.ApprovedAt.Time.Format("2006-01-02T15:04:05Z07:00")
		item.ApprovedAt = &t
	}
	if s.AutoApprovedAt.Valid {
		t := s.AutoApprovedAt.Time.Format("2006-01-02T15:04:05Z07:00")
		item.AutoApprovedAt = &t
	}
	return item
}

// RegisterSubmissionHandlers registers submission endpoints on the authenticated API.
func RegisterSubmissionHandlers(api huma.API, svc SubmissionServiceInterface) {
	// POST /campaigns/{id}/submissions - submit a clip
	huma.Register(api, huma.Operation{
		OperationID: "create-submission",
		Method:      "POST",
		Path:        "/campaigns/{id}/submissions",
		Summary:     "Submit a clip",
		Description: "Submits a clip to a campaign. The user must be authenticated.",
		Tags:        []string{"Submissions"},
	}, func(ctx context.Context, input *struct {
		ID string `path:"id" doc:"Campaign UUID"`
		Body struct {
			PostURL  string `json:"post_url" doc:"URL of the published clip"`
			Platform string `json:"platform" doc:"Platform where the clip is published"`
		}
	}) (*submissionOutput, error) {
		user, err := requireUser(ctx)
		if err != nil {
			return nil, err
		}

		campaignID, err := parseCampaignID(input.ID)
		if err != nil {
			return nil, err
		}

		submission, err := svc.Submit(ctx, campaignID, user.ID, input.Body.PostURL, input.Body.Platform)
		if err != nil {
			if ve, ok := err.(*service.ValidationError); ok {
				return nil, huma.Error422UnprocessableEntity(ve.Error())
			}
			switch {
			case errors.Is(err, service.ErrCampaignNotFound):
				return nil, huma.Error404NotFound("campaign not found")
			case errors.Is(err, service.ErrCampaignNotActive):
				return nil, huma.Error409Conflict("campaign is not active")
			case errors.Is(err, service.ErrDuplicateSubmission):
				return nil, huma.Error409Conflict("duplicate submission: you have already submitted this URL")
			case errors.Is(err, service.ErrClipperLimitExceeded):
				return nil, huma.Error409Conflict("submission limit reached for this campaign")
			case errors.Is(err, service.ErrPlatformMismatch):
				return nil, huma.Error422UnprocessableEntity("platform does not match campaign")
			case errors.Is(err, service.ErrInvalidURL):
				return nil, huma.Error422UnprocessableEntity("invalid post URL")
			}
			return nil, huma.Error500InternalServerError("failed to create submission")
		}

		resp := &submissionOutput{}
		resp.Body = toSubmissionListItem(*submission)
		return resp, nil
	})

	// GET /campaigns/{id}/submissions - list submissions for a campaign (owner only)
	huma.Register(api, huma.Operation{
		OperationID: "list-campaign-submissions",
		Method:      "GET",
		Path:        "/campaigns/{id}/submissions",
		Summary:     "List campaign submissions",
		Description: "Returns all submissions for a campaign. Only the campaign owner may access.",
		Tags:        []string{"Submissions"},
	}, func(ctx context.Context, input *struct {
		ID string `path:"id" doc:"Campaign UUID"`
	}) (*submissionListOutput, error) {
		user, err := requireOwner(ctx)
		if err != nil {
			return nil, err
		}

		campaignID, err := parseCampaignID(input.ID)
		if err != nil {
			return nil, err
		}

		// Verify the user owns this campaign.
		if err := svc.VerifyCampaignOwnership(ctx, campaignID, user.ID); err != nil {
			if errors.Is(err, service.ErrCampaignNotFound) {
				return nil, huma.Error404NotFound("campaign not found")
			}
			if errors.Is(err, service.ErrNotOwner) {
				return nil, huma.Error403Forbidden("not campaign owner")
			}
			return nil, huma.Error500InternalServerError("failed to verify ownership")
		}

		submissions, err := svc.ListByCampaign(ctx, campaignID)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to list submissions")
		}

		resp := &submissionListOutput{}
		resp.Body.Submissions = make([]submissionListItem, 0, len(submissions))
		for _, s := range submissions {
			resp.Body.Submissions = append(resp.Body.Submissions, toSubmissionListItem(s))
		}
		return resp, nil
	})

	// GET /me/submissions - my submissions
	huma.Register(api, huma.Operation{
		OperationID: "list-my-submissions",
		Method:      "GET",
		Path:        "/me/submissions",
		Summary:     "List my submissions",
		Description: "Returns all submissions by the current user across all campaigns.",
		Tags:        []string{"Submissions"},
	}, func(ctx context.Context, input *struct{}) (*submissionListOutput, error) {
		user, err := requireUser(ctx)
		if err != nil {
			return nil, err
		}

		submissions, err := svc.ListByClipper(ctx, user.ID)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to list submissions")
		}

		resp := &submissionListOutput{}
		resp.Body.Submissions = make([]submissionListItem, 0, len(submissions))
		for _, s := range submissions {
			resp.Body.Submissions = append(resp.Body.Submissions, toSubmissionListItem(s))
		}
		return resp, nil
	})

	// POST /submissions/{id}/approve - approve a submission
	huma.Register(api, huma.Operation{
		OperationID: "approve-submission",
		Method:      "POST",
		Path:        "/submissions/{id}/approve",
		Summary:     "Approve a submission",
		Description: "Approves a pending submission. Only the campaign owner may approve.",
		Tags:        []string{"Submissions"},
	}, func(ctx context.Context, input *struct {
		ID string `path:"id" doc:"Submission UUID"`
	}) (*submissionOutput, error) {
		user, err := requireOwner(ctx)
		if err != nil {
			return nil, err
		}

		submissionID, err := parseSubmissionID(input.ID)
		if err != nil {
			return nil, err
		}

		submission, err := svc.Approve(ctx, submissionID, user.ID)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrSubmissionNotFound):
				return nil, huma.Error404NotFound("submission not found")
			case errors.Is(err, service.ErrNotOwner):
				return nil, huma.Error403Forbidden("not campaign owner")
			case errors.Is(err, service.ErrSubmissionNotPending):
				return nil, huma.Error409Conflict("submission is not pending")
			}
			return nil, huma.Error500InternalServerError("failed to approve submission")
		}

		resp := &submissionOutput{}
		resp.Body = toSubmissionListItem(*submission)
		return resp, nil
	})

	// POST /submissions/{id}/reject - reject a submission
	huma.Register(api, huma.Operation{
		OperationID: "reject-submission",
		Method:      "POST",
		Path:        "/submissions/{id}/reject",
		Summary:     "Reject a submission",
		Description: "Rejects a pending submission with an optional reason. Only the campaign owner may reject.",
		Tags:        []string{"Submissions"},
	}, func(ctx context.Context, input *struct {
		ID   string `path:"id" doc:"Submission UUID"`
		Body struct {
			Reason string `json:"reason" doc:"Optional rejection reason"`
		}
	}) (*submissionOutput, error) {
		user, err := requireOwner(ctx)
		if err != nil {
			return nil, err
		}

		submissionID, err := parseSubmissionID(input.ID)
		if err != nil {
			return nil, err
		}

		submission, err := svc.Reject(ctx, submissionID, user.ID, input.Body.Reason)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrSubmissionNotFound):
				return nil, huma.Error404NotFound("submission not found")
			case errors.Is(err, service.ErrNotOwner):
				return nil, huma.Error403Forbidden("not campaign owner")
			case errors.Is(err, service.ErrSubmissionNotPending):
				return nil, huma.Error409Conflict("submission is not pending")
			}
			return nil, huma.Error500InternalServerError("failed to reject submission")
		}

		resp := &submissionOutput{}
		resp.Body = toSubmissionListItem(*submission)
		return resp, nil
	})

	// POST /campaigns/{id}/submissions/batch-approve - batch approve submissions
	huma.Register(api, huma.Operation{
		OperationID: "batch-approve-submissions",
		Method:      "POST",
		Path:        "/campaigns/{id}/submissions/batch-approve",
		Summary:     "Batch approve submissions",
		Description: "Approves multiple pending submissions at once. Only the campaign owner may approve. Items are processed independently; some may fail while others succeed.",
		Tags:        []string{"Submissions"},
	}, func(ctx context.Context, input *struct {
		ID   string `path:"id" doc:"Campaign UUID"`
		Body struct {
			SubmissionIDs []string `json:"submission_ids" doc:"List of submission UUIDs to approve"`
		}
	}) (*batchResultOutput, error) {
		user, err := requireOwner(ctx)
		if err != nil {
			return nil, err
		}

		campaignID, err := parseCampaignID(input.ID)
		if err != nil {
			return nil, err
		}

		if err := svc.VerifyCampaignOwnership(ctx, campaignID, user.ID); err != nil {
			if errors.Is(err, service.ErrCampaignNotFound) {
				return nil, huma.Error404NotFound("campaign not found")
			}
			if errors.Is(err, service.ErrNotOwner) {
				return nil, huma.Error403Forbidden("not campaign owner")
			}
			return nil, huma.Error500InternalServerError("failed to verify ownership")
		}

		ids, err := parseSubmissionIDs(input.Body.SubmissionIDs)
		if err != nil {
			return nil, err
		}

		result, err := svc.BatchApprove(ctx, ids, user.ID)
		if err != nil {
			return nil, huma.Error500InternalServerError("batch approve failed")
		}

		resp := &batchResultOutput{}
		resp.Body = toBatchResult(*result)
		return resp, nil
	})

	// POST /campaigns/{id}/submissions/batch-reject - batch reject submissions
	huma.Register(api, huma.Operation{
		OperationID: "batch-reject-submissions",
		Method:      "POST",
		Path:        "/campaigns/{id}/submissions/batch-reject",
		Summary:     "Batch reject submissions",
		Description: "Rejects multiple pending submissions at once with an optional shared reason. Only the campaign owner may reject. Items are processed independently; some may fail while others succeed.",
		Tags:        []string{"Submissions"},
	}, func(ctx context.Context, input *struct {
		ID   string `path:"id" doc:"Campaign UUID"`
		Body struct {
			SubmissionIDs []string `json:"submission_ids" doc:"List of submission UUIDs to reject"`
			Reason        string   `json:"reason,omitempty" doc:"Optional shared rejection reason"`
		}
	}) (*batchResultOutput, error) {
		user, err := requireOwner(ctx)
		if err != nil {
			return nil, err
		}

		campaignID, err := parseCampaignID(input.ID)
		if err != nil {
			return nil, err
		}

		if err := svc.VerifyCampaignOwnership(ctx, campaignID, user.ID); err != nil {
			if errors.Is(err, service.ErrCampaignNotFound) {
				return nil, huma.Error404NotFound("campaign not found")
			}
			if errors.Is(err, service.ErrNotOwner) {
				return nil, huma.Error403Forbidden("not campaign owner")
			}
			return nil, huma.Error500InternalServerError("failed to verify ownership")
		}

		ids, err := parseSubmissionIDs(input.Body.SubmissionIDs)
		if err != nil {
			return nil, err
		}

		result, err := svc.BatchReject(ctx, ids, user.ID, input.Body.Reason)
		if err != nil {
			return nil, huma.Error500InternalServerError("batch reject failed")
		}

		resp := &batchResultOutput{}
		resp.Body = toBatchResult(*result)
		return resp, nil
	})
}

type submissionOutput struct {
	Body submissionListItem
}

type submissionListOutput struct {
	Body struct {
		Submissions []submissionListItem `json:"submissions"`
	}
}

type batchResultItemErr struct {
	ID     string `json:"id"`
	Reason string `json:"reason"`
}

type batchResultBody struct {
	Approved int                 `json:"approved"`
	Rejected int                 `json:"rejected,omitempty"`
	Failed   int                 `json:"failed"`
	Errors   []batchResultItemErr `json:"errors,omitempty"`
}

type batchResultOutput struct {
	Body batchResultBody
}

func toBatchResult(r service.BatchResult) batchResultBody {
	b := batchResultBody{
		Approved: r.Approved,
		Rejected: r.Rejected,
		Failed:   r.Failed,
	}
	if len(r.Errors) > 0 {
		b.Errors = make([]batchResultItemErr, len(r.Errors))
		for i, e := range r.Errors {
			b.Errors[i] = batchResultItemErr{ID: e.ID, Reason: e.Reason}
		}
	}
	return b
}

// parseSubmissionIDs converts a list of hex strings into pgtype.UUID values.
func parseSubmissionIDs(raw []string) ([]pgtype.UUID, error) {
	if len(raw) == 0 {
		return nil, huma.Error422UnprocessableEntity("submission_ids is required")
	}
	ids := make([]pgtype.UUID, len(raw))
	for i, s := range raw {
		var id pgtype.UUID
		if err := id.Scan(s); err != nil {
			return nil, huma.Error422UnprocessableEntity(fmt.Sprintf("invalid submission id at index %d", i))
		}
		ids[i] = id
	}
	return ids, nil
}

// parseSubmissionID is a small helper to parse a path param into pgtype.UUID.
func parseSubmissionID(raw string) (pgtype.UUID, error) {
	var id pgtype.UUID
	if err := id.Scan(raw); err != nil {
		return pgtype.UUID{}, huma.Error422UnprocessableEntity("invalid submission id")
	}
	return id, nil
}
