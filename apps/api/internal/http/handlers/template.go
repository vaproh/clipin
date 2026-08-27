package handlers

import (
	"context"
	"fmt"

	sqlc "clipin/apps/api/internal/db/sqlc"
	"clipin/apps/api/internal/service"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgtype"
)

// TemplateServiceInterface is the subset of TemplateService the handlers need.
type TemplateServiceInterface interface {
	ListTemplates(ctx context.Context) ([]sqlc.CampaignTemplate, error)
	GetTemplateByID(ctx context.Context, id pgtype.UUID) (*sqlc.CampaignTemplate, error)
	CreateTemplate(ctx context.Context, name, platform string, cpmRate, totalBudget int32, maxClipsPerClipper, minViewsPerClip, autoApproveHours int32, descriptionTemplate string) (*sqlc.CampaignTemplate, error)
	UpdateTemplate(ctx context.Context, id pgtype.UUID, name, platform string, cpmRate, totalBudget int32, maxClipsPerClipper, minViewsPerClip, autoApproveHours int32, descriptionTemplate string) (*sqlc.CampaignTemplate, error)
	DeleteTemplate(ctx context.Context, id pgtype.UUID) error
}

type templateItem struct {
	ID                  string  `json:"id"`
	Name                string  `json:"name"`
	Platform            string  `json:"platform"`
	CpmRate             int32   `json:"cpm_rate"`
	TotalBudget         int32   `json:"total_budget"`
	MaxClipsPerClipper  *int32  `json:"max_clips_per_clipper,omitempty"`
	MinViewsPerClip     *int32  `json:"min_views_per_clip,omitempty"`
	AutoApproveHours    *int32  `json:"auto_approve_hours,omitempty"`
	DescriptionTemplate *string `json:"description_template,omitempty"`
	CreatedAt           string  `json:"created_at"`
}

func toTemplateItem(t sqlc.CampaignTemplate) templateItem {
	item := templateItem{
		ID:          fmt.Sprintf("%x", t.ID.Bytes),
		Name:        t.Name,
		Platform:    t.Platform,
		CpmRate:     t.CpmRate,
		TotalBudget: t.TotalBudget,
	}
	if t.MaxClipsPerClipper.Valid {
		item.MaxClipsPerClipper = &t.MaxClipsPerClipper.Int32
	}
	if t.MinViewsPerClip.Valid {
		item.MinViewsPerClip = &t.MinViewsPerClip.Int32
	}
	if t.AutoApproveHours.Valid {
		item.AutoApproveHours = &t.AutoApproveHours.Int32
	}
	if t.DescriptionTemplate.Valid {
		item.DescriptionTemplate = &t.DescriptionTemplate.String
	}
	if t.CreatedAt.Valid {
		item.CreatedAt = t.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00")
	}
	return item
}

// RegisterTemplateHandlers registers campaign template endpoints.
func RegisterTemplateHandlers(api huma.API, svc TemplateServiceInterface) {
	// GET /templates - public listing
	huma.Register(api, huma.Operation{
		OperationID: "list-templates",
		Method:      "GET",
		Path:        "/templates",
		Summary:     "List campaign templates",
		Description: "Returns all available campaign templates.",
		Tags:        []string{"Templates"},
	}, func(ctx context.Context, input *struct{}) (*listTemplatesOutput, error) {
		templates, err := svc.ListTemplates(ctx)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to list templates")
		}

		resp := &listTemplatesOutput{}
		resp.Body.Templates = make([]templateItem, 0, len(templates))
		for _, t := range templates {
			resp.Body.Templates = append(resp.Body.Templates, toTemplateItem(t))
		}
		return resp, nil
	})

	// POST /admin/templates - admin-only create
	huma.Register(api, huma.Operation{
		OperationID: "create-template",
		Method:      "POST",
		Path:        "/admin/templates",
		Summary:     "Create a campaign template (admin)",
		Description: "Creates a new campaign template. Admin only.",
		Tags:        []string{"Templates"},
	}, func(ctx context.Context, input *createTemplateInput) (*templateOutput, error) {
		if _, err := requireAdmin(ctx); err != nil {
			return nil, err
		}

		if input.Body.Name == "" {
			return nil, huma.Error422UnprocessableEntity("name is required")
		}
		if input.Body.CpmRate <= 0 {
			return nil, huma.Error422UnprocessableEntity("cpm_rate must be positive")
		}
		if input.Body.TotalBudget <= 0 {
			return nil, huma.Error422UnprocessableEntity("total_budget must be positive")
		}

		var maxClips, minViews, autoApprove int32
		if input.Body.MaxClipsPerClipper != nil {
			maxClips = *input.Body.MaxClipsPerClipper
		} else {
			maxClips = 3
		}
		if input.Body.MinViewsPerClip != nil {
			minViews = *input.Body.MinViewsPerClip
		} else {
			minViews = 1000
		}
		if input.Body.AutoApproveHours != nil {
			autoApprove = *input.Body.AutoApproveHours
		} else {
			autoApprove = 48
		}
		var descTpl string
		if input.Body.DescriptionTemplate != nil {
			descTpl = *input.Body.DescriptionTemplate
		}

		template, err := svc.CreateTemplate(ctx,
			input.Body.Name,
			input.Body.Platform,
			input.Body.CpmRate,
			input.Body.TotalBudget,
			maxClips,
			minViews,
			autoApprove,
			descTpl,
		)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to create template")
		}

		resp := &templateOutput{}
		resp.Body = toTemplateItem(*template)
		return resp, nil
	})

	// PATCH /admin/templates/{id} - admin-only update
	huma.Register(api, huma.Operation{
		OperationID: "update-template",
		Method:      "PATCH",
		Path:        "/admin/templates/{id}",
		Summary:     "Update a campaign template (admin)",
		Description: "Updates an existing campaign template. Admin only.",
		Tags:        []string{"Templates"},
	}, func(ctx context.Context, input *updateTemplateInput) (*templateOutput, error) {
		if _, err := requireAdmin(ctx); err != nil {
			return nil, err
		}

		id, err := parseTemplateID(input.ID)
		if err != nil {
			return nil, huma.Error422UnprocessableEntity("invalid template id")
		}

		if input.Body.Name == "" {
			return nil, huma.Error422UnprocessableEntity("name is required")
		}
		if input.Body.CpmRate <= 0 {
			return nil, huma.Error422UnprocessableEntity("cpm_rate must be positive")
		}
		if input.Body.TotalBudget <= 0 {
			return nil, huma.Error422UnprocessableEntity("total_budget must be positive")
		}

		var maxClips, minViews, autoApprove int32
		if input.Body.MaxClipsPerClipper != nil {
			maxClips = *input.Body.MaxClipsPerClipper
		}
		if input.Body.MinViewsPerClip != nil {
			minViews = *input.Body.MinViewsPerClip
		}
		if input.Body.AutoApproveHours != nil {
			autoApprove = *input.Body.AutoApproveHours
		}
		var descTpl string
		if input.Body.DescriptionTemplate != nil {
			descTpl = *input.Body.DescriptionTemplate
		}

		template, err := svc.UpdateTemplate(ctx, id,
			input.Body.Name,
			input.Body.Platform,
			input.Body.CpmRate,
			input.Body.TotalBudget,
			maxClips,
			minViews,
			autoApprove,
			descTpl,
		)
		if err != nil {
			if err == service.ErrTemplateNotFound {
				return nil, huma.Error404NotFound("template not found")
			}
			return nil, huma.Error500InternalServerError("failed to update template")
		}

		resp := &templateOutput{}
		resp.Body = toTemplateItem(*template)
		return resp, nil
	})

	// DELETE /admin/templates/{id} - admin-only delete
	huma.Register(api, huma.Operation{
		OperationID: "delete-template",
		Method:      "DELETE",
		Path:        "/admin/templates/{id}",
		Summary:     "Delete a campaign template (admin)",
		Description: "Deletes a campaign template. Admin only.",
		Tags:        []string{"Templates"},
	}, func(ctx context.Context, input *deleteTemplateInput) (*struct{}, error) {
		if _, err := requireAdmin(ctx); err != nil {
			return nil, err
		}

		id, err := parseTemplateID(input.ID)
		if err != nil {
			return nil, huma.Error422UnprocessableEntity("invalid template id")
		}

		if err := svc.DeleteTemplate(ctx, id); err != nil {
			return nil, huma.Error500InternalServerError("failed to delete template")
		}

		return &struct{}{}, nil
	})
}

type createTemplateInput struct {
	Body struct {
		Name                string `json:"name" doc:"Template name"`
		Platform            string `json:"platform" doc:"Platform (youtube, instagram, tiktok, multi)"`
		CpmRate             int32  `json:"cpm_rate" doc:"CPM rate in paise"`
		TotalBudget         int32  `json:"total_budget" doc:"Total budget in paise"`
		MaxClipsPerClipper  *int32 `json:"max_clips_per_clipper,omitempty" doc:"Max clips per clipper"`
		MinViewsPerClip     *int32 `json:"min_views_per_clip,omitempty" doc:"Min views per clip"`
		AutoApproveHours    *int32 `json:"auto_approve_hours,omitempty" doc:"Auto-approve after hours"`
		DescriptionTemplate *string `json:"description_template,omitempty" doc:"Description template text"`
	}
}

type listTemplatesOutput struct {
	Body struct {
		Templates []templateItem `json:"templates"`
	}
}

type templateOutput struct {
	Body templateItem
}

type updateTemplateInput struct {
	ID   string `path:"id" doc:"Template ID"`
	Body struct {
		Name                string `json:"name" doc:"Template name"`
		Platform            string `json:"platform" doc:"Platform (youtube, instagram, tiktok, multi)"`
		CpmRate             int32  `json:"cpm_rate" doc:"CPM rate in paise"`
		TotalBudget         int32  `json:"total_budget" doc:"Total budget in paise"`
		MaxClipsPerClipper  *int32 `json:"max_clips_per_clipper,omitempty" doc:"Max clips per clipper"`
		MinViewsPerClip     *int32 `json:"min_views_per_clip,omitempty" doc:"Min views per clip"`
		AutoApproveHours    *int32 `json:"auto_approve_hours,omitempty" doc:"Auto-approve after hours"`
		DescriptionTemplate *string `json:"description_template,omitempty" doc:"Description template text"`
	}
}

type deleteTemplateInput struct {
	ID string `path:"id" doc:"Template ID"`
}

func parseTemplateID(raw string) (pgtype.UUID, error) {
	var id pgtype.UUID
	if err := id.Scan(raw); err != nil {
		return pgtype.UUID{}, fmt.Errorf("invalid template id")
	}
	return id, nil
}
