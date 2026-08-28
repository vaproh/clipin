import { useMutation, useQueries, useQuery, useQueryClient } from '@tanstack/vue-query'
import type { Ref } from 'vue'

export type UserRole = 'clipper' | 'owner' | 'admin'

export interface UserProfile {
  id: string
  email: string
  display_name: string
  role: UserRole
  upi_id: string | null
  created_at: string
  updated_at: string
}

export interface Campaign {
  id: string
  owner_id: string
  title: string
  description: string | null
  brief_url: string | null
  platform: 'youtube' | 'instagram' | 'tiktok' | 'multi'
  status: 'draft' | 'funded' | 'active' | 'paused' | 'completed' | 'cancelled'
  cpm_rate: number
  total_budget: number
  remaining_budget: number
  platform_fee: number
  max_clips_per_campaign: number | null
  max_clips_per_clipper: number
  min_views_per_clip: number
  auto_approve_hours: number
  starts_at: string | null
  ends_at: string | null
  created_at: string
  updated_at: string
}

export interface CampaignListResponse {
  campaigns: Campaign[]
  total: number
  page: number
  page_size: number
}

export interface CampaignFilters {
  q?: string
  platform?: string
  max_cpm?: number
  min_budget?: number
  page?: number
  page_size?: number
}

/**
 * Base API client. Reads the API base URL from runtime config and attaches
 * the Clerk session token as a Bearer header on every request.
 */
export function useApi() {
  const config = useRuntimeConfig()
  const { getToken } = useAuth()

  async function fetchApi<T>(path: string, options: RequestInit = {}): Promise<T> {
    const token = getToken.value ? await getToken.value() : null
    const headers = new Headers(options.headers)
    headers.set('Content-Type', 'application/json')
    if (token) headers.set('Authorization', `Bearer ${token}`)

    const res = await fetch(`${config.public.apiBase}${path}`, { ...options, headers })
    if (!res.ok) {
      const body = await res.json().catch(() => null) as Record<string, unknown> | null
      const serverMsg = typeof body?.error === 'string' ? body.error : typeof body?.message === 'string' ? body.message : null
      const message = res.status >= 500
        ? 'Something went wrong. Please try again.'
        : (serverMsg || `Request failed (${res.status})`)
      throw new Error(message)
    }
    return res.status === 204 ? (undefined as T) : (res.json() as Promise<T>)
  }

  return { fetchApi }
}

/** Fetches the authenticated user profile from GET /me. */
export function useUserQuery() {
  const { fetchApi } = useApi()
  const { userId } = useAuth()

  return useQuery({
    queryKey: ['me'],
    queryFn: () => fetchApi<UserProfile>('/me'),
    enabled: () => !!userId.value,
    staleTime: 30_000,
  })
}

/** Updates the user role via POST /me/role and refreshes the cached profile. */
export function useSetRole() {
  const queryClient = useQueryClient()
  const { fetchApi } = useApi()

  return useMutation({
    mutationFn: (role: UserRole) =>
      fetchApi<UserProfile>('/me/role', {
        method: 'POST',
        body: JSON.stringify({ role }),
      }),
    onSuccess: (user) => {
      queryClient.setQueryData(['me'], user)
    },
  })
}

/** Fetches a paginated list of campaigns with optional filters. */
export function useCampaigns(filters: Ref<CampaignFilters>) {
  const { fetchApi } = useApi()

  return useQuery({
    queryKey: ['campaigns', filters],
    queryFn: () => {
      const params = new URLSearchParams()
      if (filters.value.q) params.set('q', filters.value.q)
      if (filters.value.platform) params.set('platform', filters.value.platform)
      if (filters.value.max_cpm) params.set('max_cpm', String(filters.value.max_cpm))
      if (filters.value.min_budget) params.set('min_budget', String(filters.value.min_budget))
      params.set('page', String(filters.value.page ?? 1))
      params.set('page_size', String(filters.value.page_size ?? 20))
      return fetchApi<CampaignListResponse>(`/campaigns?${params.toString()}`)
    },
  })
}

/** Fetches a single campaign by ID. */
export function useCampaign(id: Ref<string>) {
  const { fetchApi } = useApi()

  return useQuery({
    queryKey: ['campaign', id],
    queryFn: () => fetchApi<Campaign>(`/campaigns/${id.value}`),
    enabled: () => !!id.value,
  })
}

/** Fetches campaigns owned by the authenticated user. */
export function useMyCampaigns() {
  const { fetchApi } = useApi()

  return useQuery({
    queryKey: ['campaigns', 'mine'],
    queryFn: () => fetchApi<CampaignListResponse>('/me/campaigns'),
  })
}

/** Owner stats from GET /me/campaigns/stats. */
export interface OwnerStats {
  total_campaigns: number
  active_campaigns: number
  total_budget: number
  total_remaining: number
}

export function useOwnerStats() {
  const { fetchApi } = useApi()

  return useQuery({
    queryKey: ['campaigns', 'stats'],
    queryFn: () => fetchApi<OwnerStats>('/me/campaigns/stats'),
  })
}

/** Body fields for creating/updating a campaign. */
export interface CreateCampaignBody {
  title: string
  description?: string
  brief_url?: string
  platform: 'youtube' | 'instagram' | 'tiktok' | 'multi'
  cpm_rate: number
  total_budget: number
  max_clips_per_campaign?: number
  max_clips_per_clipper?: number
  min_views_per_clip?: number
  auto_approve_hours?: number
  starts_at?: string
  ends_at?: string
}

export function useCreateCampaign() {
  const queryClient = useQueryClient()
  const { fetchApi } = useApi()

  return useMutation({
    mutationFn: (body: CreateCampaignBody) =>
      fetchApi<Campaign>('/campaigns', {
        method: 'POST',
        body: JSON.stringify(body),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['campaigns'] })
    },
  })
}

export function useUpdateCampaign(id: Ref<string>) {
  const queryClient = useQueryClient()
  const { fetchApi } = useApi()

  return useMutation({
    mutationFn: (body: Partial<CreateCampaignBody>) =>
      fetchApi<Campaign>(`/campaigns/${id.value}`, {
        method: 'PATCH',
        body: JSON.stringify(body),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['campaigns'] })
      queryClient.invalidateQueries({ queryKey: ['campaign', id] })
    },
  })
}

function useCampaignAction(id: Ref<string>, action: string) {
  const queryClient = useQueryClient()
  const { fetchApi } = useApi()

  return useMutation({
    mutationFn: () =>
      fetchApi<Campaign>(`/campaigns/${id.value}/${action}`, {
        method: 'POST',
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['campaigns'] })
      queryClient.invalidateQueries({ queryKey: ['campaign', id] })
    },
  })
}

export function usePauseCampaign(id: Ref<string>) {
  return useCampaignAction(id, 'pause')
}

export function useResumeCampaign(id: Ref<string>) {
  return useCampaignAction(id, 'resume')
}

export function useCancelCampaign(id: Ref<string>) {
  return useCampaignAction(id, 'cancel')
}

// ---------------------------------------------------------------------------
// Verification
// ---------------------------------------------------------------------------

export interface VerificationStatus {
  has_snapshots: boolean
  current_views: number
  eligible_views: number
  snapshot_count: number
  last_snapshot_at: string | null
}

/** Fetch verification status for a single submission. */
export function useVerificationStatus(submissionId: Ref<string>) {
  const { fetchApi } = useApi()

  return useQuery({
    queryKey: ['verification', submissionId],
    queryFn: () => fetchApi<VerificationStatus>(`/submissions/${submissionId.value}/verification`),
    enabled: () => !!submissionId.value,
    staleTime: 30_000,
  })
}

export interface AggregateVerification {
  total_eligible_views: number
  total_current_views: number
  verified_count: number
  tracked_count: number
}

/** Fetch verification for multiple submissions and compute aggregates. */
export function useAggregateVerification(submissionIds: Ref<string[]>) {
  const { fetchApi } = useApi()

  return useQuery({
    queryKey: ['verification', 'aggregate', submissionIds],
    queryFn: async (): Promise<AggregateVerification> => {
      const ids = submissionIds.value
      if (ids.length === 0) return { total_eligible_views: 0, total_current_views: 0, verified_count: 0, tracked_count: 0 }

      const results = await Promise.all(
        ids.map((id) => fetchApi<VerificationStatus>(`/submissions/${id}/verification`))
      )

      return results.reduce(
        (acc, r) => {
          acc.total_eligible_views += r.eligible_views
          acc.total_current_views += r.current_views
          if (r.eligible_views > 0) acc.verified_count++
          if (r.has_snapshots) acc.tracked_count++
          return acc
        },
        { total_eligible_views: 0, total_current_views: 0, verified_count: 0, tracked_count: 0 }
      )
    },
    enabled: () => submissionIds.value.length > 0,
    staleTime: 30_000,
  })
}

// ---------------------------------------------------------------------------
// Submissions
// ---------------------------------------------------------------------------

export type SubmissionStatus = 'pending' | 'approved' | 'rejected' | 'auto_approved' | 'disputed'

export interface Submission {
  id: string
  campaign_id: string
  clipper_id: string
  post_url: string
  platform: string
  platform_post_id: string | null
  status: SubmissionStatus
  rejection_reason: string | null
  approved_at: string | null
  auto_approved_at: string | null
  created_at: string
  updated_at: string
}

export interface SubmissionListResponse {
  submissions: Submission[]
}

/** Fetch submissions for a campaign (owner view). */
export function useCampaignSubmissions(campaignId: Ref<string>) {
  const { fetchApi } = useApi()

  return useQuery({
    queryKey: ['campaignSubmissions', campaignId],
    queryFn: () => fetchApi<SubmissionListResponse>(`/campaigns/${campaignId.value}/submissions`),
    enabled: () => !!campaignId.value,
  })
}

/** Fetch the authenticated user's own submissions. */
export function useMySubmissions() {
  const { fetchApi } = useApi()

  return useQuery({
    queryKey: ['mySubmissions'],
    queryFn: () => fetchApi<SubmissionListResponse>('/me/submissions'),
  })
}

/** Submit a clip to a campaign. */
export function useSubmitClip(campaignId: Ref<string>) {
  const queryClient = useQueryClient()
  const { fetchApi } = useApi()

  return useMutation({
    mutationFn: (body: { post_url: string; platform: string }) =>
      fetchApi<Submission>(`/campaigns/${campaignId.value}/submissions`, {
        method: 'POST',
        body: JSON.stringify(body),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['mySubmissions'] })
      queryClient.invalidateQueries({ queryKey: ['campaignSubmissions', campaignId] })
    },
  })
}

/** Approve a submission (campaign owner). */
export function useApproveSubmission(campaignId: Ref<string>) {
  const queryClient = useQueryClient()
  const { fetchApi } = useApi()

  return useMutation({
    mutationFn: (submissionId: string) =>
      fetchApi<Submission>(`/submissions/${submissionId}/approve`, {
        method: 'POST',
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['campaignSubmissions', campaignId] })
      queryClient.invalidateQueries({ queryKey: ['mySubmissions'] })
    },
  })
}

/** Reject a submission (campaign owner). */
export function useRejectSubmission(campaignId: Ref<string>) {
  const queryClient = useQueryClient()
  const { fetchApi } = useApi()

  return useMutation({
    mutationFn: ({ submissionId, reason }: { submissionId: string; reason?: string }) =>
      fetchApi<Submission>(`/submissions/${submissionId}/reject`, {
        method: 'POST',
        body: JSON.stringify({ reason }),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['campaignSubmissions', campaignId] })
      queryClient.invalidateQueries({ queryKey: ['mySubmissions'] })
    },
  })
}

// ---------------------------------------------------------------------------
// Batch Review
// ---------------------------------------------------------------------------

export interface BatchResult {
  approved: number
  failed: number
  errors: Array<{ id: string; reason: string }>
}

/** Batch approve submissions (campaign owner). */
export function useBatchApproveSubmissions(campaignId: Ref<string>) {
  const queryClient = useQueryClient()
  const { fetchApi } = useApi()

  return useMutation({
    mutationFn: (ids: string[]) =>
      fetchApi<BatchResult>(`/campaigns/${campaignId.value}/submissions/batch-approve`, {
        method: 'POST',
        body: JSON.stringify({ submission_ids: ids }),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['campaignSubmissions', campaignId] })
    },
  })
}

/** Batch reject submissions (campaign owner). */
export function useBatchRejectSubmissions(campaignId: Ref<string>) {
  const queryClient = useQueryClient()
  const { fetchApi } = useApi()

  return useMutation({
    mutationFn: ({ ids, reason }: { ids: string[]; reason?: string }) =>
      fetchApi<BatchResult>(`/campaigns/${campaignId.value}/submissions/batch-reject`, {
        method: 'POST',
        body: JSON.stringify({ submission_ids: ids, reason }),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['campaignSubmissions', campaignId] })
    },
  })
}

// ---------------------------------------------------------------------------
// Payouts
// ---------------------------------------------------------------------------

export type PayoutStatus = 'pending' | 'processing' | 'completed' | 'failed'

export interface PayoutRequest {
  id: string
  clipper_id: string
  amount: number
  upi_id: string
  status: PayoutStatus
  provider_ref: string | null
  failure_reason: string | null
  idempotency_key: string
  created_at: string
  processed_at: string | null
  updated_at: string
}

export interface PayoutListResponse {
  payout_requests: PayoutRequest[]
}

/** Update the authenticated user's UPI ID. */
export function useUpdateUPI() {
  const queryClient = useQueryClient()
  const { fetchApi } = useApi()

  return useMutation({
    mutationFn: (upi_id: string) =>
      fetchApi<{ upi_id: string }>('/me/upi', {
        method: 'PATCH',
        body: JSON.stringify({ upi_id }),
      }),
    onSuccess: (data) => {
      queryClient.setQueryData(['me'], (old: UserProfile | undefined) =>
        old ? { ...old, upi_id: data.upi_id } : old,
      )
    },
  })
}

/** Request a payout. */
export function useRequestPayout() {
  const queryClient = useQueryClient()
  const { fetchApi } = useApi()

  return useMutation({
    mutationFn: (amount: number) =>
      fetchApi<PayoutRequest>('/me/payouts', {
        method: 'POST',
        body: JSON.stringify({ amount }),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['payouts'] })
      queryClient.invalidateQueries({ queryKey: ['earnings', 'me'] })
    },
  })
}

/** Fetch the authenticated user's payout history. */
export function useMyPayouts() {
  const { fetchApi } = useApi()

  return useQuery({
    queryKey: ['payouts'],
    queryFn: () => fetchApi<PayoutListResponse>('/me/payouts'),
  })
}

// ---------------------------------------------------------------------------
// Ledger & Earnings
// ---------------------------------------------------------------------------

export type LedgerEntryType = 'platform_fee' | 'earning' | 'refund' | 'escrow_lock' | 'escrow_release'

export interface LedgerEntry {
  id: string
  idempotency_key: string
  entry_type: LedgerEntryType
  campaign_id: string
  submission_id: string | null
  clipper_id: string | null
  amount: number
  description: string | null
  metadata: Record<string, unknown> | null
  created_at: string
}

export interface EarningsSummary {
  total_earnings: number
  campaign_earnings: Array<{ campaign_id: string; campaign_title: string; amount: number }>
  recent_entries: LedgerEntry[]
}

export interface CampaignLedger {
  entries: LedgerEntry[]
  total_fees: number
  total_earnings: number
  remaining_budget: number
}

/** Fetch authenticated user's earnings summary. */
export function useMyEarnings() {
  const { fetchApi } = useApi()

  return useQuery({
    queryKey: ['earnings', 'me'],
    queryFn: () => fetchApi<EarningsSummary>('/me/earnings'),
  })
}

/** Fetch ledger for a specific campaign (owner view). */
export function useCampaignLedger(campaignId: Ref<string>) {
  const { fetchApi } = useApi()

  return useQuery({
    queryKey: ['ledger', campaignId],
    queryFn: () => fetchApi<CampaignLedger>(`/me/campaigns/${campaignId.value}/ledger`),
    enabled: () => !!campaignId.value,
  })
}

// ---------------------------------------------------------------------------
// Social Accounts
// ---------------------------------------------------------------------------

export interface SocialAccount {
  id: string
  user_id: string
  platform: 'youtube' | 'instagram' | 'tiktok'
  platform_user_id: string
  platform_username: string | null
  created_at: string
  updated_at: string
}

/** Fetch authenticated user's social accounts. */
export function useSocialAccounts() {
  const { fetchApi } = useApi()

  return useQuery({
    queryKey: ['me', 'social-accounts'],
    queryFn: () => fetchApi<SocialAccount[]>('/me/social-accounts'),
  })
}

/** Connect a social account. */
export function useConnectSocialAccount() {
  const queryClient = useQueryClient()
  const { fetchApi } = useApi()

  return useMutation({
    mutationFn: (body: { platform: string; platform_user_id: string; platform_username: string }) =>
      fetchApi<SocialAccount>('/me/social-accounts', {
        method: 'POST',
        body: JSON.stringify(body),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['me', 'social-accounts'] })
    },
  })
}

/** Disconnect a social account. */
export function useDisconnectSocialAccount() {
  const queryClient = useQueryClient()
  const { fetchApi } = useApi()

  return useMutation({
    mutationFn: (id: string) =>
      fetchApi(`/me/social-accounts/${id}`, { method: 'DELETE' }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['me', 'social-accounts'] })
    },
  })
}

// ---------------------------------------------------------------------------
// Campaign Analytics
// ---------------------------------------------------------------------------

export interface CampaignAnalytics {
  submissions: {
    total: number
    pending: number
    approved: number
    rejected: number
    unique_clippers: number
  }
  views: {
    total_views: number
    total_likes: number
    total_comments: number
    total_shares: number
  }
  financial: {
    total_earnings: number
    total_fees: number
    total_refunds: number
  }
  progress: {
    budget_consumed_pct: number
    time_remaining: string | null
  }
}

/** Fetch analytics for a campaign (owner view). */
export function useCampaignAnalytics(campaignId: Ref<string>) {
  const { fetchApi } = useApi()

  return useQuery({
    queryKey: ['campaignAnalytics', campaignId],
    queryFn: () => fetchApi<CampaignAnalytics>(`/me/campaigns/${campaignId.value}/analytics`),
    enabled: () => !!campaignId.value,
  })
}

// ---------------------------------------------------------------------------
// Templates
// ---------------------------------------------------------------------------

export interface CampaignTemplate {
  id: string
  name: string
  platform: 'youtube' | 'instagram' | 'tiktok' | 'multi'
  cpm_rate: number
  total_budget: number
  max_clips_per_clipper: number
  min_views_per_clip: number
  auto_approve_hours: number
  description_template: string | null
  created_at: string
}

/** Fetch all campaign templates. */
export function useTemplates() {
  const { fetchApi } = useApi()

  return useQuery({
    queryKey: ['templates'],
    queryFn: () => fetchApi<CampaignTemplate[]>('/templates'),
  })
}

// ---------------------------------------------------------------------------
// Notifications
// ---------------------------------------------------------------------------

export interface Notification {
  id: string
  user_id: string
  type: 'submission_approved' | 'submission_rejected' | 'campaign_update' | 'payout_completed' | 'system'
  title: string
  body: string | null
  link: string | null
  is_read: boolean
  created_at: string
}

/** Fetch paginated notifications. */
export function useNotifications(limit = 20, offset = 0) {
  const { fetchApi } = useApi()

  return useQuery({
    queryKey: ['me', 'notifications', { limit, offset }],
    queryFn: () => fetchApi<Notification[]>(`/me/notifications?limit=${limit}&offset=${offset}`),
  })
}

/** Polling unread notification count. */
export function useUnreadNotificationCount() {
  const { fetchApi } = useApi()

  return useQuery({
    queryKey: ['me', 'notifications', 'unread-count'],
    queryFn: () => fetchApi<{ count: number }>('/me/notifications/unread-count'),
    refetchInterval: 30_000,
  })
}

/** Mark a notification as read. */
export function useMarkNotificationRead() {
  const queryClient = useQueryClient()
  const { fetchApi } = useApi()

  return useMutation({
    mutationFn: (id: string) =>
      fetchApi(`/me/notifications/${id}/read`, { method: 'POST' }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['me', 'notifications'] })
    },
  })
}

/** Mark all notifications as read. */
export function useMarkAllNotificationsRead() {
  const queryClient = useQueryClient()
  const { fetchApi } = useApi()

  return useMutation({
    mutationFn: () =>
      fetchApi('/me/notifications/read-all', { method: 'POST' }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['me', 'notifications'] })
    },
  })
}

/** Delete a notification. */
export function useDeleteNotification() {
  const queryClient = useQueryClient()
  const { fetchApi } = useApi()

  return useMutation({
    mutationFn: (id: string) =>
      fetchApi(`/me/notifications/${id}`, { method: 'DELETE' }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['me', 'notifications'] })
    },
  })
}

// ---------------------------------------------------------------------------
// Leaderboard
// ---------------------------------------------------------------------------

export interface LeaderboardEntry {
  rank: number
  user_id: string
  display_name: string | null
  avatar_url: string | null
  total_earnings: number
  total_submissions: number
  campaigns_participated: number
}

export function useLeaderboard(sort: Ref<string> = ref('earnings'), limit: number = 20) {
  const { fetchApi } = useApi()

  return useQuery({
    queryKey: ['leaderboard', sort, limit],
    queryFn: () => fetchApi<LeaderboardEntry[]>(`/leaderboard?sort=${sort.value}&limit=${limit}`),
  })
}

// ---------------------------------------------------------------------------
// Clipper Profile
// ---------------------------------------------------------------------------

export interface ClipperProfile {
  id: string
  display_name: string | null
  avatar_url: string | null
  bio: string | null
  created_at: string
  stats: {
    total_submissions: number
    approved_submissions: number
    pending_submissions: number
    rejected_submissions: number
    total_views: number
    total_earnings: number
    campaigns_participated: number
  }
  social_accounts: Array<{ platform: string; platform_username: string | null }>
}

/** Fetch a public clipper profile by ID. */
export function useClipperProfile(id: Ref<string>) {
  const { fetchApi } = useApi()

  return useQuery({
    queryKey: ['clipper', id],
    queryFn: () => fetchApi<ClipperProfile>(`/clippers/${id.value}`),
    enabled: () => !!id.value,
  })
}

// ---------------------------------------------------------------------------
// Admin
// ---------------------------------------------------------------------------

export interface FraudFlag {
  id: string
  submission_id: string | null
  user_id: string | null
  flag_type: string
  severity: 'low' | 'medium' | 'high' | 'critical'
  description: string | null
  status: 'open' | 'investigating' | 'resolved' | 'dismissed'
  created_at: string
}

export interface AuditLog {
  id: string
  actor_id: string
  action: string
  resource_type: string
  resource_id: string
  details: Record<string, unknown> | null
  ip_address: string | null
  created_at: string
}

export interface AdminUser {
  id: string
  email: string
  display_name: string
  role: UserRole
  created_at: string
  flag_count: number
}

export interface AdminUsersResponse {
  users: AdminUser[]
  total: number
}

export interface AdminFraudFlagsResponse {
  flags: FraudFlag[]
}

export interface AdminAuditLogsResponse {
  logs: AuditLog[]
  total: number
}

export interface AdminStatsResponse {
  total_users: number
  open_fraud_flags: number
  total_campaigns: number
}

/** Admin: fetch paginated users. */
export function useAdminUsers(page: Ref<number> = ref(1), enabled?: Ref<boolean>) {
  const { fetchApi } = useApi()

  return useQuery({
    queryKey: ['admin', 'users', page],
    queryFn: () =>
      fetchApi<AdminUsersResponse>(`/admin/users?page=${page.value}&page_size=20`),
    enabled: enabled ? () => enabled.value : true,
  })
}

/** Admin: fetch user detail. */
export function useAdminUser(userId: Ref<string>) {
  const { fetchApi } = useApi()

  return useQuery({
    queryKey: ['admin', 'user', userId],
    queryFn: () => fetchApi<{ user: AdminUser; submissions: Submission[]; payouts: PayoutRequest[]; fraud_flags: FraudFlag[] }>(`/admin/users/${userId.value}`),
    enabled: () => !!userId.value,
  })
}

/** Admin: fetch all fraud flags. */
export function useAdminFraudFlags(enabled?: Ref<boolean>) {
  const { fetchApi } = useApi()

  return useQuery({
    queryKey: ['admin', 'fraud-flags'],
    queryFn: () => fetchApi<AdminFraudFlagsResponse>('/admin/fraud-flags'),
    enabled: enabled ? () => enabled.value : true,
  })
}

/** Admin: resolve a fraud flag. */
export function useResolveFraudFlag() {
  const queryClient = useQueryClient()
  const { fetchApi } = useApi()

  return useMutation({
    mutationFn: ({ id, resolution }: { id: string; resolution: string }) =>
      fetchApi(`/admin/fraud-flags/${id}/resolve`, {
        method: 'POST',
        body: JSON.stringify({ resolution }),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'fraud-flags'] })
    },
  })
}

/** Admin: dismiss a fraud flag. */
export function useDismissFraudFlag() {
  const queryClient = useQueryClient()
  const { fetchApi } = useApi()

  return useMutation({
    mutationFn: (id: string) =>
      fetchApi(`/admin/fraud-flags/${id}/dismiss`, { method: 'POST' }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'fraud-flags'] })
    },
  })
}

/** Admin: fetch paginated audit logs. */
export function useAdminAuditLogs(page: Ref<number> = ref(1), enabled?: Ref<boolean>) {
  const { fetchApi } = useApi()

  return useQuery({
    queryKey: ['admin', 'audit-logs', page],
    queryFn: () =>
      fetchApi<AdminAuditLogsResponse>(`/admin/audit-logs?page=${page.value}&page_size=20`),
    enabled: enabled ? () => enabled.value : true,
  })
}

/** Admin: fetch dashboard stats. */
export function useAdminStats(enabled?: Ref<boolean>) {
  const { fetchApi } = useApi()

  return useQuery({
    queryKey: ['admin', 'stats'],
    queryFn: () => fetchApi<AdminStatsResponse>('/admin/stats'),
    enabled: enabled ? () => enabled.value : true,
  })
}