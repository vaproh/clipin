import { readonly } from 'vue'
import { useState } from '#app'

export interface AuthUser {
  id: string
  email: string
  name: string
  role: 'clipper' | 'brand' | 'admin'
}

export interface AuthState {
  user: AuthUser | null
  isAuthenticated: boolean
  isLoading: boolean
}

/**
 * Placeholder composable for external auth integration (e.g. Clerk / Supabase / Custom OIDC).
 * Auth logic is explicitly deferred to later implementation phases.
 */
export const useAuth = () => {
  const authState = useState<AuthState>('auth-state', () => ({
    user: null,
    isAuthenticated: false,
    isLoading: false,
  }))

  const login = async () => {
    console.warn('Auth integration is not implemented in scaffolding phase.')
  }

  const logout = async () => {
    authState.value.user = null
    authState.value.isAuthenticated = false
  }

  return {
    state: readonly(authState),
    login,
    logout,
  }
}
