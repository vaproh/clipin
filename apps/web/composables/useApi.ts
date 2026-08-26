import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'

export type UserRole = 'clipper' | 'owner'

export interface UserProfile {
  id: string
  email: string
  display_name: string
  role: UserRole
  created_at: string
  updated_at: string
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