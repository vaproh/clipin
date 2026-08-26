-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: CreateUser :one
INSERT INTO users (id, email, display_name, role)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateUser :one
UPDATE users
SET display_name = COALESCE($2, display_name),
    email = COALESCE($3, email),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateUserRole :one
UPDATE users
SET role = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ListSocialAccountsByUserID :many
SELECT * FROM social_accounts
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: CreateSocialAccount :one
INSERT INTO social_accounts (user_id, platform, platform_user_id, platform_username, access_token, refresh_token, token_expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: DeleteSocialAccount :exec
DELETE FROM social_accounts WHERE id = $1 AND user_id = $2;

-- name: GetCampaignByID :one
SELECT * FROM campaigns WHERE id = $1;

-- name: ListActiveCampaigns :many
SELECT * FROM campaigns
WHERE status IN ('active', 'funded')
  AND (ends_at IS NULL OR ends_at > NOW())
ORDER BY created_at DESC;

-- name: ListCampaignsFiltered :many
SELECT * FROM campaigns
WHERE status IN ('active', 'funded')
  AND (ends_at IS NULL OR ends_at > NOW())
  AND ($1::text = '' OR platform = $1)
  AND ($2::int = 0 OR cpm_rate <= $2)
  AND ($3::int = 0 OR remaining_budget >= $3)
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- name: CountCampaignsFiltered :one
SELECT COUNT(*) FROM campaigns
WHERE status IN ('active', 'funded')
  AND (ends_at IS NULL OR ends_at > NOW())
  AND ($1::text = '' OR platform = $1)
  AND ($2::int = 0 OR cpm_rate <= $2)
  AND ($3::int = 0 OR remaining_budget >= $3);

-- name: ListCampaignsByOwner :many
SELECT * FROM campaigns
WHERE owner_id = $1
ORDER BY created_at DESC;

-- name: CreateCampaign :one
INSERT INTO campaigns (
    id, owner_id, title, description, brief_url, platform, status,
    cpm_rate, total_budget, remaining_budget, platform_fee, escrow_id,
    max_clips_per_campaign, max_clips_per_clipper, min_views_per_clip,
    auto_approve_hours, starts_at, ends_at
)
VALUES (
    $1, $2, $3, $4, $5, $6, $7,
    $8, $9, $10, $11, $12,
    $13, $14, $15, $16, $17, $18
)
RETURNING *;

-- name: UpdateCampaignStatus :one
UPDATE campaigns
SET status = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateCampaignBudget :one
UPDATE campaigns
SET remaining_budget = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: GetSubmissionByID :one
SELECT * FROM submissions WHERE id = $1;

-- name: ListSubmissionsByCampaign :many
SELECT * FROM submissions
WHERE campaign_id = $1
ORDER BY created_at DESC;

-- name: ListSubmissionsByClipper :many
SELECT * FROM submissions
WHERE clipper_id = $1
ORDER BY created_at DESC;

-- name: CreateSubmission :one
INSERT INTO submissions (id, campaign_id, clipper_id, post_url, platform, platform_post_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateSubmissionStatus :one
UPDATE submissions
SET status = $2,
    rejection_reason = $3,
    approved_at = CASE WHEN $2 = 'approved' THEN NOW() ELSE approved_at END,
    auto_approved_at = CASE WHEN $2 = 'auto_approved' THEN NOW() ELSE auto_approved_at END,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: CountSubmissionsByCampaign :one
SELECT COUNT(*) FROM submissions WHERE campaign_id = $1;

-- name: CountSubmissionsByClipperForCampaign :one
SELECT COUNT(*) FROM submissions WHERE campaign_id = $1 AND clipper_id = $2;

-- name: ListPendingSubmissionsOlderThan :many
SELECT s.*, c.auto_approve_hours, c.owner_id
FROM submissions s
JOIN campaigns c ON c.id = s.campaign_id
WHERE s.status = 'pending'
  AND s.created_at < $1
ORDER BY s.created_at ASC;
