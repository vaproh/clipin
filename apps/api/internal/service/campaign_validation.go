package service

import (
	"fmt"
	"strings"
	"time"
)

// validPlatforms is the set of allowed platform values.
var validPlatforms = map[string]bool{
	"youtube":  true,
	"instagram": true,
	"tiktok":   true,
	"multi":    true,
}

// CreateCampaignInput holds the fields needed to create a campaign.
type CreateCampaignInput struct {
	Title               string
	Description         *string
	BriefURL            *string
	Platform            string
	CpmRate             int32
	TotalBudget         int32
	MaxClipsPerCampaign *int32
	MaxClipsPerClipper  *int32
	MinViewsPerClip     *int32
	AutoApproveHours    *int32
	StartsAt            *string
	EndsAt              *string
}

// UpdateCampaignInput holds optional fields for updating a campaign.
type UpdateCampaignInput struct {
	Title               *string
	Description         *string
	BriefURL            *string
	MaxClipsPerCampaign *int32
	MaxClipsPerClipper  *int32
	MinViewsPerClip     *int32
	AutoApproveHours    *int32
	EndsAt              *string
}

// ValidationError collects all validation failures.
type ValidationError struct {
	Errors []string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed: %s", strings.Join(e.Errors, "; "))
}

func (e *ValidationError) Add(msg string) {
	e.Errors = append(e.Errors, msg)
}

func (e *ValidationError) HasErrors() bool {
	return len(e.Errors) > 0
}

// ValidateCreateCampaign validates all fields for campaign creation.
func ValidateCreateCampaign(in *CreateCampaignInput) error {
	var ve ValidationError

	// title: required, 1-200 chars
	title := strings.TrimSpace(in.Title)
	if title == "" {
		ve.Add("title is required")
	} else if len(title) > 200 {
		ve.Add("title must be at most 200 characters")
	}

	// platform: must be one of the valid set
	if !validPlatforms[in.Platform] {
		ve.Add("platform must be one of: youtube, instagram, tiktok, multi")
	}

	// cpm_rate: > 0, max 100000 paise (₹1000 per 1k views)
	if in.CpmRate <= 0 {
		ve.Add("cpm_rate must be greater than 0")
	} else if in.CpmRate > 100000 {
		ve.Add("cpm_rate must not exceed 100000 paise (₹1000 per 1k views)")
	}

	// total_budget: > 0, min 10000 paise (₹100), max 100000000 paise (₹10 lakh)
	if in.TotalBudget <= 0 {
		ve.Add("total_budget must be greater than 0")
	} else if in.TotalBudget < 10000 {
		ve.Add("total_budget must be at least 10000 paise (₹100)")
	} else if in.TotalBudget > 100000000 {
		ve.Add("total_budget must not exceed 100000000 paise (₹10,00,000)")
	}

	// max_clips_per_clipper: default 3 if nil, max 50
	if in.MaxClipsPerClipper != nil {
		if *in.MaxClipsPerClipper < 1 {
			ve.Add("max_clips_per_clipper must be at least 1")
		} else if *in.MaxClipsPerClipper > 50 {
			ve.Add("max_clips_per_clipper must not exceed 50")
		}
	}

	// min_views_per_clip: default 1000 if nil, min 100
	if in.MinViewsPerClip != nil {
		if *in.MinViewsPerClip < 100 {
			ve.Add("min_views_per_clip must be at least 100")
		}
	}

	// auto_approve_hours: default 48 if nil, range 1-720
	if in.AutoApproveHours != nil {
		if *in.AutoApproveHours < 1 || *in.AutoApproveHours > 720 {
			ve.Add("auto_approve_hours must be between 1 and 720")
		}
	}

	// starts_at: if provided, must be valid RFC3339
	if in.StartsAt != nil {
		_, err := time.Parse(time.RFC3339, *in.StartsAt)
		if err != nil {
			ve.Add("starts_at must be a valid RFC3339 timestamp")
		}
	}

	// ends_at: if provided, must be after now
	if in.EndsAt != nil {
		t, err := time.Parse(time.RFC3339, *in.EndsAt)
		if err != nil {
			ve.Add("ends_at must be a valid RFC3339 timestamp")
		} else if t.Before(time.Now()) {
			ve.Add("ends_at must be in the future")
		}
	}

	if ve.HasErrors() {
		return &ve
	}
	return nil
}

// ValidateUpdateCampaign validates fields for a campaign update.
func ValidateUpdateCampaign(in *UpdateCampaignInput) error {
	var ve ValidationError

	if in.Title != nil {
		title := strings.TrimSpace(*in.Title)
		if title == "" {
			ve.Add("title cannot be empty")
		} else if len(title) > 200 {
			ve.Add("title must be at most 200 characters")
		}
	}

	if in.MaxClipsPerClipper != nil && *in.MaxClipsPerClipper > 50 {
		ve.Add("max_clips_per_clipper must not exceed 50")
	}

	if in.MinViewsPerClip != nil && *in.MinViewsPerClip < 100 {
		ve.Add("min_views_per_clip must be at least 100")
	}

	if in.AutoApproveHours != nil && (*in.AutoApproveHours < 1 || *in.AutoApproveHours > 720) {
		ve.Add("auto_approve_hours must be between 1 and 720")
	}

	if in.EndsAt != nil {
		t, err := time.Parse(time.RFC3339, *in.EndsAt)
		if err != nil {
			ve.Add("ends_at must be a valid RFC3339 timestamp")
		} else if t.Before(time.Now()) {
			ve.Add("ends_at must be in the future")
		}
	}

	if ve.HasErrors() {
		return &ve
	}
	return nil
}

// PlatformFee calculates the 10% platform fee for a given budget.
func PlatformFee(totalBudget int32) int32 {
	return totalBudget * 10 / 100
}

// DefaultMaxClipsPerClipper returns the value or the default (3).
func DefaultMaxClipsPerClipper(v *int32) int32 {
	if v != nil {
		return *v
	}
	return 3
}

// DefaultMinViewsPerClip returns the value or the default (1000).
func DefaultMinViewsPerClip(v *int32) int32 {
	if v != nil {
		return *v
	}
	return 1000
}

// DefaultAutoApproveHours returns the value or the default (48).
func DefaultAutoApproveHours(v *int32) int32 {
	if v != nil {
		return *v
	}
	return 48
}
