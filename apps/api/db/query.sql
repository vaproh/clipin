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

-- name: CreateMetricSnapshot :one
INSERT INTO metric_snapshots (submission_id, platform, views, likes, comments, shares, captured_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetLatestSnapshotForSubmission :one
SELECT * FROM metric_snapshots
WHERE submission_id = $1
ORDER BY captured_at DESC
LIMIT 1;

-- name: GetInitialSnapshotForSubmission :one
SELECT * FROM metric_snapshots
WHERE submission_id = $1
ORDER BY captured_at ASC
LIMIT 1;

-- name: ListSnapshotsBySubmission :many
SELECT * FROM metric_snapshots
WHERE submission_id = $1
ORDER BY captured_at ASC;

-- name: CountSnapshotsBySubmission :one
SELECT COUNT(*) FROM metric_snapshots
WHERE submission_id = $1;

-- name: GetSubmissionsNeedingVerification :many
SELECT s.*, c.min_views_per_clip, c.cpm_rate, c.owner_id
FROM submissions s
JOIN campaigns c ON s.campaign_id = c.id
WHERE s.status IN ('approved', 'auto_approved')
  AND s.id NOT IN (
    SELECT submission_id FROM metric_snapshots
    WHERE captured_at > NOW() - INTERVAL '1 hour'
  )
LIMIT $1;

-- name: GetSubmissionWithCampaign :one
SELECT s.*, c.min_views_per_clip, c.cpm_rate, c.owner_id
FROM submissions s
JOIN campaigns c ON s.campaign_id = c.id
WHERE s.id = $1;

-- name: CreateLedgerEntry :one
INSERT INTO ledger_entries (idempotency_key, entry_type, campaign_id, submission_id, clipper_id, amount, description, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (idempotency_key) DO NOTHING
RETURNING *;

-- name: GetLedgerEntryByIdempotencyKey :one
SELECT * FROM ledger_entries WHERE idempotency_key = $1;

-- name: ListLedgerEntriesByCampaign :many
SELECT * FROM ledger_entries WHERE campaign_id = $1 ORDER BY created_at ASC;

-- name: ListLedgerEntriesByClipper :many
SELECT * FROM ledger_entries WHERE clipper_id = $1 ORDER BY created_at ASC;

-- name: SumEarningsByClipper :one
SELECT COALESCE(SUM(amount), 0)::bigint as total FROM ledger_entries
WHERE clipper_id = $1 AND entry_type = 'earning';

-- name: SumEarningsByClipperForCampaign :one
SELECT COALESCE(SUM(amount), 0)::bigint as total FROM ledger_entries
WHERE clipper_id = $1 AND campaign_id = $2 AND entry_type = 'earning';

-- name: SumFeesByCampaign :one
SELECT COALESCE(SUM(amount), 0)::bigint as total FROM ledger_entries
WHERE campaign_id = $1 AND entry_type = 'platform_fee';

-- name: SumSpendByCampaign :one
SELECT COALESCE(SUM(amount), 0)::bigint as total FROM ledger_entries
WHERE campaign_id = $1 AND entry_type = 'earning';

-- Payout queries

-- name: UpdateUserUPI :one
UPDATE users SET upi_id = $2, updated_at = NOW() WHERE id = $1 RETURNING *;

-- name: CreatePayoutRequest :one
INSERT INTO payout_requests (id, clipper_id, amount, upi_id, idempotency_key)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (idempotency_key) DO NOTHING
RETURNING *;

-- name: GetPayoutRequestByID :one
SELECT * FROM payout_requests WHERE id = $1;

-- name: ListPayoutRequestsByClipper :many
SELECT * FROM payout_requests WHERE clipper_id = $1 ORDER BY created_at DESC;

-- name: UpdatePayoutRequestStatus :one
UPDATE payout_requests
SET status = $2, provider_ref = $3, failure_reason = $4,
    processed_at = CASE WHEN $2 IN ('completed', 'failed') THEN NOW() ELSE processed_at END,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: SumPendingPayoutsByClipper :one
SELECT COALESCE(SUM(amount), 0)::bigint as total FROM payout_requests
WHERE clipper_id = $1 AND status IN ('pending', 'processing');

-- name: ListPendingPayoutRequests :many
SELECT * FROM payout_requests WHERE status = 'pending' ORDER BY created_at ASC;

-- Audit Logs

-- name: CreateAuditLog :one
INSERT INTO audit_logs (actor_id, action, resource_type, resource_id, details, ip_address)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListAuditLogs :many
SELECT * FROM audit_logs
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListAuditLogsByResource :many
SELECT * FROM audit_logs
WHERE resource_type = $1 AND resource_id = $2
ORDER BY created_at DESC;

-- name: ListAuditLogsByActor :many
SELECT * FROM audit_logs
WHERE actor_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- Fraud Flags

-- name: CreateFraudFlag :one
INSERT INTO fraud_flags (submission_id, user_id, flag_type, severity, description)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListOpenFraudFlags :many
SELECT * FROM fraud_flags
WHERE status = 'open'
ORDER BY
  CASE severity WHEN 'critical' THEN 0 WHEN 'high' THEN 1 WHEN 'medium' THEN 2 ELSE 3 END,
  created_at ASC;

-- name: GetFraudFlagByID :one
SELECT * FROM fraud_flags WHERE id = $1;

-- name: UpdateFraudFlagStatus :one
UPDATE fraud_flags
SET status = $2, resolved_by = $3, resolution = $4,
    resolved_at = CASE WHEN $2 IN ('resolved', 'dismissed') THEN NOW() ELSE resolved_at END,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: CountFraudFlagsByUser :one
SELECT COUNT(*)::int FROM fraud_flags WHERE user_id = $1 AND status = 'open';

-- name: ListUsersWithManyFlags :many
SELECT u.*, COUNT(f.id)::int as flag_count
FROM users u
JOIN fraud_flags f ON f.user_id = u.id
WHERE f.status = 'open'
GROUP BY u.id
HAVING COUNT(f.id) >= $1
ORDER BY COUNT(f.id) DESC;

-- Admin user queries

-- name: ListUsers :many
SELECT * FROM users
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT COUNT(*)::int FROM users;

-- name: ListSubmissionsByUser :many
SELECT * FROM submissions
WHERE clipper_id = $1
ORDER BY created_at DESC;

-- name: ListPayoutsByUser :many
SELECT * FROM payout_requests
WHERE clipper_id = $1
ORDER BY created_at DESC;

-- name: ListFraudFlagsByUser :many
SELECT * FROM fraud_flags
WHERE user_id = $1
ORDER BY created_at DESC;
