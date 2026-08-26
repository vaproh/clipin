package handlers

import (
	"context"
	"fmt"
	"time"

	sqlc "clipin/apps/api/internal/db/sqlc"
	"clipin/apps/api/internal/service"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgtype"
)

// VerificationServiceInterface is the subset of VerificationService the handlers need.
type VerificationServiceInterface interface {
	RecordSnapshot(ctx context.Context, submissionID pgtype.UUID, platform string, views, likes, comments, shares int64, capturedAt *time.Time) (*sqlc.MetricSnapshot, error)
	ComputeEligibleViews(ctx context.Context, submissionID pgtype.UUID) (*service.EligibleViewsResult, error)
	GetVerificationStatus(ctx context.Context, submissionID pgtype.UUID) (*service.VerificationStatus, error)
	GetSubmissionsNeedingVerification(ctx context.Context, limit int) ([]sqlc.GetSubmissionsNeedingVerificationRow, error)
}

// --- Internal snapshot ingestion ---

type createSnapshotInput struct {
	SubmissionID string `json:"submission_id" doc:"Submission UUID"`
	Platform     string `json:"platform" doc:"Platform where the clip is published"`
	Views        int64  `json:"views" doc:"Current view count"`
	Likes        int64  `json:"likes" doc:"Current like count"`
	Comments     int64  `json:"comments" doc:"Current comment count"`
	Shares       int64  `json:"shares" doc:"Current share count"`
	CapturedAt   string `json:"captured_at,omitempty" doc:"RFC3339 timestamp of when metrics were captured (default: now)"`
}

type snapshotOutput struct {
	Body struct {
		ID           string `json:"id"`
		SubmissionID string `json:"submission_id"`
		Platform     string `json:"platform"`
		Views        int64  `json:"views"`
		Likes        int64  `json:"likes"`
		Comments     int64  `json:"comments"`
		Shares       int64  `json:"shares"`
		CapturedAt   string `json:"captured_at"`
		CreatedAt    string `json:"created_at"`
	}
}

type verificationStatusOutput struct {
	Body struct {
		SubmissionID   string  `json:"submission_id"`
		Status         string  `json:"status"`
		HasSnapshots   bool    `json:"has_snapshots"`
		SnapshotCount  int64   `json:"snapshot_count"`
		InitialViews   int64   `json:"initial_views"`
		LatestViews    int64   `json:"latest_views"`
		GrowthViews    int64   `json:"growth_views"`
		EligibleViews  int64   `json:"eligible_views"`
		EngagementRate float64 `json:"engagement_rate"`
		LastCapturedAt *string `json:"last_captured_at,omitempty"`
	}
}

type needsVerificationListOutput struct {
	Body struct {
		Submissions []needsVerificationItem `json:"submissions"`
	}
}

type needsVerificationItem struct {
	SubmissionID    string `json:"submission_id"`
	CampaignID      string `json:"campaign_id"`
	ClipperID       string `json:"clipper_id"`
	PostURL         string `json:"post_url"`
	Platform        string `json:"platform"`
	Status          string `json:"status"`
	MinViewsPerClip int32  `json:"min_views_per_clip"`
	CPMRate         int32  `json:"cpm_rate"`
}

// RegisterInternalVerificationHandlers registers internal endpoints for the
// verifier service (API key auth, not Clerk JWT).
func RegisterInternalVerificationHandlers(api huma.API, svc VerificationServiceInterface) {
	// POST /internal/snapshots - record a metric snapshot
	huma.Register(api, huma.Operation{
		OperationID: "create-snapshot",
		Method:      "POST",
		Path:        "/internal/snapshots",
		Summary:     "Record a metric snapshot",
		Description: "Internal endpoint for the verifier service to record metrics for a submission.",
		Tags:        []string{"Verification"},
	}, func(ctx context.Context, input *struct {
		Body createSnapshotInput
	}) (*snapshotOutput, error) {
		submissionID, err := parseSubmissionID(input.Body.SubmissionID)
		if err != nil {
			return nil, err
		}

		if input.Body.Platform == "" {
			return nil, huma.Error422UnprocessableEntity("platform is required")
		}
		if input.Body.Views < 0 {
			return nil, huma.Error422UnprocessableEntity("views must be non-negative")
		}
		if input.Body.Likes < 0 || input.Body.Comments < 0 || input.Body.Shares < 0 {
			return nil, huma.Error422UnprocessableEntity("likes, comments, and shares must be non-negative")
		}

		var capturedAt *time.Time
		if input.Body.CapturedAt != "" {
			t, err := time.Parse(time.RFC3339, input.Body.CapturedAt)
			if err != nil {
				return nil, huma.Error422UnprocessableEntity("captured_at must be RFC3339 format")
			}
			capturedAt = &t
		}

		snap, err := svc.RecordSnapshot(ctx, submissionID, input.Body.Platform,
			input.Body.Views, input.Body.Likes, input.Body.Comments, input.Body.Shares, capturedAt)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to record snapshot")
		}

		resp := &snapshotOutput{}
		resp.Body.ID = fmt.Sprintf("%x", snap.ID.Bytes)
		resp.Body.SubmissionID = fmt.Sprintf("%x", snap.SubmissionID.Bytes)
		resp.Body.Platform = snap.Platform
		resp.Body.Views = snap.Views
		resp.Body.Likes = snap.Likes
		resp.Body.Comments = snap.Comments
		resp.Body.Shares = snap.Shares
		resp.Body.CapturedAt = snap.CapturedAt.Time.Format("2006-01-02T15:04:05Z07:00")
		resp.Body.CreatedAt = snap.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00")
		return resp, nil
	})

	// POST /internal/snapshots/list - list submissions needing verification
	huma.Register(api, huma.Operation{
		OperationID: "list-needs-verification",
		Method:      "POST",
		Path:        "/internal/snapshots/list",
		Summary:     "List submissions needing verification",
		Description: "Internal endpoint for the verifier to discover submissions that need fresh snapshots.",
		Tags:        []string{"Verification"},
	}, func(ctx context.Context, input *struct {
		Body struct {
			Limit int `json:"limit" doc:"Max results (1-100, default 20)"`
		}
	}) (*needsVerificationListOutput, error) {
		limit := input.Body.Limit
		if limit < 1 || limit > 100 {
			limit = 20
		}

		rows, err := svc.GetSubmissionsNeedingVerification(ctx, limit)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to list submissions")
		}

		resp := &needsVerificationListOutput{}
		resp.Body.Submissions = make([]needsVerificationItem, 0, len(rows))
		for _, r := range rows {
			resp.Body.Submissions = append(resp.Body.Submissions, needsVerificationItem{
				SubmissionID:    fmt.Sprintf("%x", r.ID.Bytes),
				CampaignID:      fmt.Sprintf("%x", r.CampaignID.Bytes),
				ClipperID:       r.ClipperID,
				PostURL:         r.PostUrl,
				Platform:        r.Platform,
				Status:          r.Status,
				MinViewsPerClip: r.MinViewsPerClip.Int32,
				CPMRate:         r.CpmRate,
			})
		}
		return resp, nil
	})
}

// RegisterVerificationStatusHandlers registers authenticated verification
// endpoints (Clerk JWT required).
func RegisterVerificationStatusHandlers(api huma.API, svc VerificationServiceInterface) {
	// GET /submissions/{id}/verification - get verification status
	huma.Register(api, huma.Operation{
		OperationID: "get-verification-status",
		Method:      "GET",
		Path:        "/submissions/{id}/verification",
		Summary:     "Get verification status",
		Description: "Returns the verification status for a submission. Requires authentication.",
		Tags:        []string{"Verification"},
	}, func(ctx context.Context, input *struct {
		ID string `path:"id" doc:"Submission UUID"`
	}) (*verificationStatusOutput, error) {
		// Require authentication.
		if _, err := requireUser(ctx); err != nil {
			return nil, err
		}

		submissionID, err := parseSubmissionID(input.ID)
		if err != nil {
			return nil, err
		}

		status, err := svc.GetVerificationStatus(ctx, submissionID)
		if err != nil {
			if err == service.ErrSubmissionNotFound {
				return nil, huma.Error404NotFound("submission not found")
			}
			return nil, huma.Error500InternalServerError("failed to get verification status")
		}

		resp := &verificationStatusOutput{}
		resp.Body.SubmissionID = fmt.Sprintf("%x", status.SubmissionID.Bytes)
		resp.Body.Status = status.Status
		resp.Body.HasSnapshots = status.HasSnapshots
		resp.Body.SnapshotCount = status.SnapshotCount
		resp.Body.InitialViews = status.InitialViews
		resp.Body.LatestViews = status.LatestViews
		resp.Body.GrowthViews = status.GrowthViews
		resp.Body.EligibleViews = status.EligibleViews
		resp.Body.EngagementRate = status.EngagementRate
		resp.Body.LastCapturedAt = status.LastCapturedAt
		return resp, nil
	})
}
