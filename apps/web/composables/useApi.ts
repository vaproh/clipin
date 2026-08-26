import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import type { Ref } from 'vue'

export type UserRole = 'clipper' | 'owner'

export interface UserProfile {
  id: string
  email: string
  display_name: string
  role: UserRole
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
      throw new Error(`API ${res.status} on ${path}`)
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
    queryFn: () => fetchApi<CampaignListResponse>('/campaigns/mine'),
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