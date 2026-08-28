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

// AuditServiceInterface is the subset of AuditService the handlers need.
type AuditServiceInterface interface {
	Log(ctx context.Context, actorID, action, resourceType, resourceID string, details map[string]any, ipAddress string) error
	ListLogs(ctx context.Context, limit, offset int32) ([]sqlc.AuditLog, error)
	ListLogsForResource(ctx context.Context, resourceType, resourceID string) ([]sqlc.AuditLog, error)
}

// FraudServiceInterface is the subset of FraudService the handlers need.
type FraudServiceInterface interface {
	ListOpenFlags(ctx context.Context) ([]sqlc.FraudFlag, error)
	ResolveFlag(ctx context.Context, flagID pgtype.UUID, resolvedBy, resolution string) (*sqlc.FraudFlag, error)
	DismissFlag(ctx context.Context, flagID pgtype.UUID, resolvedBy, reason string) (*sqlc.FraudFlag, error)
	GetUserFlagCount(ctx context.Context, userID string) (int, error)
}

// AdminNotificationServiceInterface is the subset of NotificationService the admin handlers need.
type AdminNotificationServiceInterface interface {
	NotifyCampaignUpdate(ctx context.Context, ownerID, campaignTitle, message string) error
}

// AdminUserStore is the subset of the user persistence layer admin endpoints need.
type AdminUserStore interface {
	GetUserByID(ctx context.Context, id string) (sqlc.User, error)
	UpdateCampaignStatus(ctx context.Context, arg sqlc.UpdateCampaignStatusParams) (sqlc.Campaign, error)
	GetCampaignByID(ctx context.Context, id pgtype.UUID) (sqlc.Campaign, error)
	ListUsers(ctx context.Context, arg sqlc.ListUsersParams) ([]sqlc.User, error)
	CountUsers(ctx context.Context) (int32, error)
	ListSubmissionsByUser(ctx context.Context, clipperID string) ([]sqlc.Submission, error)
	ListPayoutsByUser(ctx context.Context, clipperID string) ([]sqlc.PayoutRequest, error)
	ListFraudFlagsByUser(ctx context.Context, userID pgtype.Text) ([]sqlc.FraudFlag, error)
}

// --- Admin User List ---

type adminListUsersInput struct {
	Page     int `query:"page" doc:"Page number (1-indexed)"`
	PageSize int `query:"page_size" doc:"Results per page (max 100)"`
}

type adminUserListItem struct {
	ID          string  `json:"id"`
	Email       string  `json:"email"`
	DisplayName *string `json:"display_name,omitempty"`
	Role        string  `json:"role"`
	CreatedAt   string  `json:"created_at"`
	FlagCount   int     `json:"flag_count"`
}

type adminListUsersOutput struct {
	Body struct {
		Users    []adminUserListItem `json:"users"`
		Total    int32               `json:"total"`
		Page     int                 `json:"page"`
		PageSize int                 `json:"page_size"`
	}
}

// --- Admin User Detail ---

type adminUserDetailInput struct {
	ID string `path:"id" doc:"User ID"`
}

type adminUserDetailOutput struct {
	Body struct {
		User        adminUserListItem      `json:"user"`
		Submissions []submissionListItem    `json:"submissions"`
		Payouts     []payoutListItem        `json:"payouts"`
		FraudFlags  []fraudFlagListItem     `json:"fraud_flags"`
	}
}

// --- Fraud Flag List ---

type adminListFraudFlagsOutput struct {
	Body struct {
		Flags []fraudFlagListItem `json:"flags"`
	}
}

type fraudFlagListItem struct {
	ID           string  `json:"id"`
	SubmissionID *string `json:"submission_id,omitempty"`
	UserID       *string `json:"user_id,omitempty"`
	FlagType     string  `json:"flag_type"`
	Severity     string  `json:"severity"`
	Description  *string `json:"description,omitempty"`
	Status       string  `json:"status"`
	ResolvedBy   *string `json:"resolved_by,omitempty"`
	Resolution   *string `json:"resolution,omitempty"`
	CreatedAt    string  `json:"created_at"`
	ResolvedAt   *string `json:"resolved_at,omitempty"`
}

func toFraudFlagListItem(f sqlc.FraudFlag) fraudFlagListItem {
	item := fraudFlagListItem{
		ID:        fmt.Sprintf("%x", f.ID.Bytes),
		FlagType:  f.FlagType,
		Severity:  f.Severity,
		Status:    f.Status,
		CreatedAt: f.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
	}
	if f.SubmissionID.Valid {
		s := fmt.Sprintf("%x", f.SubmissionID.Bytes)
		item.SubmissionID = &s
	}
	if f.UserID.Valid {
		item.UserID = &f.UserID.String
	}
	if f.Description.Valid {
		item.Description = &f.Description.String
	}
	if f.ResolvedBy.Valid {
		item.ResolvedBy = &f.ResolvedBy.String
	}
	if f.Resolution.Valid {
		item.Resolution = &f.Resolution.String
	}
	if f.ResolvedAt.Valid {
		t := f.ResolvedAt.Time.Format("2006-01-02T15:04:05Z07:00")
		item.ResolvedAt = &t
	}
	return item
}

// --- Fraud Flag Resolve/Dismiss ---

type adminResolveFlagInput struct {
	ID   string `path:"id" doc:"Fraud flag UUID"`
	Body struct {
		Resolution string `json:"resolution" doc:"Resolution notes"`
	}
}

type adminDismissFlagInput struct {
	ID   string `path:"id" doc:"Fraud flag UUID"`
	Body struct {
		Reason string `json:"reason" doc:"Dismissal reason"`
	}
}

type fraudFlagOutput struct {
	Body fraudFlagListItem
}

// --- Admin Audit Logs ---

type adminListAuditLogsInput struct {
	Page     int `query:"page" doc:"Page number (1-indexed)"`
	PageSize int `query:"page_size" doc:"Results per page (max 100)"`
}

type auditLogListItem struct {
	ID           string  `json:"id"`
	ActorID      string  `json:"actor_id"`
	Action       string  `json:"action"`
	ResourceType string  `json:"resource_type"`
	ResourceID   string  `json:"resource_id"`
	Details      any     `json:"details,omitempty"`
	IPAddress    *string `json:"ip_address,omitempty"`
	CreatedAt    string  `json:"created_at"`
}

type adminListAuditLogsOutput struct {
	Body struct {
		Logs     []auditLogListItem `json:"logs"`
		Page     int                `json:"page"`
		PageSize int                `json:"page_size"`
	}
}

// --- Admin Campaign Override ---

type adminOverrideCampaignInput struct {
	ID   string `path:"id" doc:"Campaign UUID"`
	Body struct {
		Status string `json:"status" doc:"New campaign status"`
	}
}

type campaignOverrideOutput struct {
	Body campaignListItem
}

// --- Payout list item (for user detail) ---

type payoutListItem struct {
	ID             string  `json:"id"`
	ClipperID      string  `json:"clipper_id"`
	Amount         int32   `json:"amount"`
	Status         string  `json:"status"`
	IdempotencyKey string  `json:"idempotency_key"`
	CreatedAt      string  `json:"created_at"`
	ProviderRef    *string `json:"provider_ref,omitempty"`
	FailureReason  *string `json:"failure_reason,omitempty"`
}

func toPayoutListItem(p sqlc.PayoutRequest) payoutListItem {
	item := payoutListItem{
		ID:             fmt.Sprintf("%x", p.ID.Bytes),
		ClipperID:      p.ClipperID,
		Amount:         p.Amount,
		Status:         p.Status,
		IdempotencyKey: p.IdempotencyKey,
		CreatedAt:      p.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
	}
	if p.ProviderRef.Valid {
		item.ProviderRef = &p.ProviderRef.String
	}
	if p.FailureReason.Valid {
		item.FailureReason = &p.FailureReason.String
	}
	return item
}

// requireAdmin checks the user is authenticated and has the "admin" role.
func requireAdmin(ctx context.Context) (*sqlc.User, error) {
	user, err := requireUser(ctx)
	if err != nil {
		return nil, err
	}
	if user.Role != "admin" {
		return nil, huma.Error403Forbidden("admin role required")
	}
	return user, nil
}

// RegisterAdminHandlers registers all admin-only endpoints.
// These must be mounted on the authenticated router group with AdminOnly middleware.
func RegisterAdminHandlers(
	api huma.API,
	userStore AdminUserStore,
	auditSvc AuditServiceInterface,
	fraudSvc FraudServiceInterface,
	notifSvc AdminNotificationServiceInterface,
) {
	// GET /admin/users - list users
	huma.Register(api, huma.Operation{
		OperationID: "admin-list-users",
		Method:      "GET",
		Path:        "/admin/users",
		Summary:     "List users (admin)",
		Description: "Returns a paginated list of all users.",
		Tags:        []string{"Admin"},
	}, func(ctx context.Context, input *adminListUsersInput) (*adminListUsersOutput, error) {
		if _, err := requireAdmin(ctx); err != nil {
			return nil, err
		}

		page := input.Page
		if page < 1 {
			page = 1
		}
		pageSize := input.PageSize
		if pageSize < 1 || pageSize > 100 {
			pageSize = 20
		}
		offset := int32((page - 1) * pageSize)

		users, err := userStore.ListUsers(ctx, sqlc.ListUsersParams{
			Limit:  int32(pageSize),
			Offset: offset,
		})
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to list users")
		}

		total, err := userStore.CountUsers(ctx)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to count users")
		}

		resp := &adminListUsersOutput{}
		resp.Body.Total = total
		resp.Body.Page = page
		resp.Body.PageSize = pageSize
		resp.Body.Users = make([]adminUserListItem, 0, len(users))
		for _, u := range users {
			item := adminUserListItem{
				ID:        u.ID,
				Email:     u.Email,
				Role:      u.Role,
				CreatedAt: u.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
			}
			if u.DisplayName.Valid {
				item.DisplayName = &u.DisplayName.String
			}
			resp.Body.Users = append(resp.Body.Users, item)
		}
		return resp, nil
	})

	// GET /admin/users/{id} - user detail
	huma.Register(api, huma.Operation{
		OperationID: "admin-get-user",
		Method:      "GET",
		Path:        "/admin/users/{id}",
		Summary:     "Get user detail (admin)",
		Description: "Returns full user profile with submissions, payouts, and fraud flags.",
		Tags:        []string{"Admin"},
	}, func(ctx context.Context, input *adminUserDetailInput) (*adminUserDetailOutput, error) {
		if _, err := requireAdmin(ctx); err != nil {
			return nil, err
		}

		user, err := userStore.GetUserByID(ctx, input.ID)
		if err != nil {
			return nil, huma.Error404NotFound("user not found")
		}

		item := adminUserListItem{
			ID:        user.ID,
			Email:     user.Email,
			Role:      user.Role,
			CreatedAt: user.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		}
		if user.DisplayName.Valid {
			item.DisplayName = &user.DisplayName.String
		}

		// Fetch related data.
		submissions, err := userStore.ListSubmissionsByUser(ctx, user.ID)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to list user submissions")
		}
		payouts, err := userStore.ListPayoutsByUser(ctx, user.ID)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to list user payouts")
		}
		flags, err := userStore.ListFraudFlagsByUser(ctx, pgtype.Text{Valid: true, String: user.ID})
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to list user fraud flags")
		}

		resp := &adminUserDetailOutput{}
		resp.Body.User = item

		resp.Body.Submissions = make([]submissionListItem, 0, len(submissions))
		for _, s := range submissions {
			resp.Body.Submissions = append(resp.Body.Submissions, toSubmissionListItem(s))
		}

		resp.Body.Payouts = make([]payoutListItem, 0, len(payouts))
		for _, p := range payouts {
			resp.Body.Payouts = append(resp.Body.Payouts, toPayoutListItem(p))
		}

		resp.Body.FraudFlags = make([]fraudFlagListItem, 0, len(flags))
		for _, f := range flags {
			resp.Body.FraudFlags = append(resp.Body.FraudFlags, toFraudFlagListItem(f))
		}

		return resp, nil
	})

	// GET /admin/fraud-flags - list open fraud flags
	huma.Register(api, huma.Operation{
		OperationID: "admin-list-fraud-flags",
		Method:      "GET",
		Path:        "/admin/fraud-flags",
		Summary:     "List open fraud flags (admin)",
		Description: "Returns all open fraud flags ordered by severity.",
		Tags:        []string{"Admin"},
	}, func(ctx context.Context, input *struct{}) (*adminListFraudFlagsOutput, error) {
		if _, err := requireAdmin(ctx); err != nil {
			return nil, err
		}

		flags, err := fraudSvc.ListOpenFlags(ctx)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to list fraud flags")
		}

		resp := &adminListFraudFlagsOutput{}
		resp.Body.Flags = make([]fraudFlagListItem, 0, len(flags))
		for _, f := range flags {
			resp.Body.Flags = append(resp.Body.Flags, toFraudFlagListItem(f))
		}
		return resp, nil
	})

	// POST /admin/fraud-flags/{id}/resolve
	huma.Register(api, huma.Operation{
		OperationID: "admin-resolve-fraud-flag",
		Method:      "POST",
		Path:        "/admin/fraud-flags/{id}/resolve",
		Summary:     "Resolve a fraud flag (admin)",
		Description: "Marks a fraud flag as resolved.",
		Tags:        []string{"Admin"},
	}, func(ctx context.Context, input *adminResolveFlagInput) (*fraudFlagOutput, error) {
		admin, err := requireAdmin(ctx)
		if err != nil {
			return nil, err
		}

		flagID, err := parseFlagID(input.ID)
		if err != nil {
			return nil, err
		}

		flag, err := fraudSvc.ResolveFlag(ctx, flagID, admin.ID, input.Body.Resolution)
		if err != nil {
			if errors.Is(err, service.ErrFlagNotFound) {
				return nil, huma.Error404NotFound("fraud flag not found")
			}
			return nil, huma.Error500InternalServerError("failed to resolve flag")
		}

		resp := &fraudFlagOutput{}
		resp.Body = toFraudFlagListItem(*flag)
		return resp, nil
	})

	// POST /admin/fraud-flags/{id}/dismiss
	huma.Register(api, huma.Operation{
		OperationID: "admin-dismiss-fraud-flag",
		Method:      "POST",
		Path:        "/admin/fraud-flags/{id}/dismiss",
		Summary:     "Dismiss a fraud flag (admin)",
		Description: "Marks a fraud flag as dismissed.",
		Tags:        []string{"Admin"},
	}, func(ctx context.Context, input *adminDismissFlagInput) (*fraudFlagOutput, error) {
		admin, err := requireAdmin(ctx)
		if err != nil {
			return nil, err
		}

		flagID, err := parseFlagID(input.ID)
		if err != nil {
			return nil, err
		}

		flag, err := fraudSvc.DismissFlag(ctx, flagID, admin.ID, input.Body.Reason)
		if err != nil {
			if errors.Is(err, service.ErrFlagNotFound) {
				return nil, huma.Error404NotFound("fraud flag not found")
			}
			return nil, huma.Error500InternalServerError("failed to dismiss flag")
		}

		resp := &fraudFlagOutput{}
		resp.Body = toFraudFlagListItem(*flag)
		return resp, nil
	})

	// GET /admin/audit-logs - list audit logs
	huma.Register(api, huma.Operation{
		OperationID: "admin-list-audit-logs",
		Method:      "GET",
		Path:        "/admin/audit-logs",
		Summary:     "List audit logs (admin)",
		Description: "Returns a paginated list of all audit logs.",
		Tags:        []string{"Admin"},
	}, func(ctx context.Context, input *adminListAuditLogsInput) (*adminListAuditLogsOutput, error) {
		if _, err := requireAdmin(ctx); err != nil {
			return nil, err
		}

		page := input.Page
		if page < 1 {
			page = 1
		}
		pageSize := input.PageSize
		if pageSize < 1 || pageSize > 100 {
			pageSize = 20
		}
		offset := int32((page - 1) * pageSize)

		logs, err := auditSvc.ListLogs(ctx, int32(pageSize), offset)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to list audit logs")
		}

		resp := &adminListAuditLogsOutput{}
		resp.Body.Page = page
		resp.Body.PageSize = pageSize
		resp.Body.Logs = make([]auditLogListItem, 0, len(logs))
		for _, l := range logs {
			item := auditLogListItem{
				ID:           fmt.Sprintf("%x", l.ID.Bytes),
				ActorID:      l.ActorID,
				Action:       l.Action,
				ResourceType: l.ResourceType,
				ResourceID:   l.ResourceID,
				CreatedAt:    l.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
			}
			if l.IpAddress.Valid {
				item.IPAddress = &l.IpAddress.String
			}
			resp.Body.Logs = append(resp.Body.Logs, item)
		}
		return resp, nil
	})

	// POST /admin/campaigns/{id}/override - admin override campaign status
	huma.Register(api, huma.Operation{
		OperationID: "admin-override-campaign",
		Method:      "POST",
		Path:        "/admin/campaigns/{id}/override",
		Summary:     "Override campaign status (admin)",
		Description: "Admin override to force a campaign to a new status.",
		Tags:        []string{"Admin"},
	}, func(ctx context.Context, input *adminOverrideCampaignInput) (*campaignOverrideOutput, error) {
		admin, err := requireAdmin(ctx)
		if err != nil {
			return nil, err
		}

		campaignID, err := parseCampaignID(input.ID)
		if err != nil {
			return nil, err
		}

		validStatuses := map[string]bool{
			"active": true, "paused": true, "cancelled": true, "draft": true, "funded": true,
		}
		if !validStatuses[input.Body.Status] {
			return nil, huma.Error422UnprocessableEntity("invalid status value")
		}

		// Fetch campaign to verify it exists.
		campaign, err := userStore.GetCampaignByID(ctx, campaignID)
		if err != nil {
			return nil, huma.Error404NotFound("campaign not found")
		}

		details := map[string]any{
			"old_status": campaign.Status,
			"new_status": input.Body.Status,
		}

		updated, err := userStore.UpdateCampaignStatus(ctx, sqlc.UpdateCampaignStatusParams{
			ID:     campaignID,
			Status: input.Body.Status,
		})
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to override campaign status")
		}

		// Log the admin override (best-effort).
		if auditSvc != nil {
			_ = auditSvc.Log(ctx, admin.ID, "admin.campaign.override", "campaign",
				fmt.Sprintf("%x", campaignID.Bytes), details, "")
		}

		// Best-effort notification to campaign owner.
		if notifSvc != nil {
			title := campaign.Title
			if title == "" {
				title = "your campaign"
			}
			_ = notifSvc.NotifyCampaignUpdate(ctx, campaign.OwnerID, title,
				fmt.Sprintf("was set to '%s' by an admin", input.Body.Status))
		}

		resp := &campaignOverrideOutput{}
		resp.Body = toCampaignListItem(updated)
		return resp, nil
	})
}

func parseFlagID(raw string) (pgtype.UUID, error) {
	var id pgtype.UUID
	if err := id.Scan(raw); err != nil {
		return pgtype.UUID{}, huma.Error422UnprocessableEntity("invalid fraud flag id")
	}
	return id, nil
}
