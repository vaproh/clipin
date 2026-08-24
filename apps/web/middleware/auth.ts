import { defineNuxtRouteMiddleware, navigateTo } from '#app'

export default defineNuxtRouteMiddleware((to) => {
  const { userId } = useAuth()

  if (!userId.value) {
    return navigateTo(`/sign-in?redirect_url=${encodeURIComponent(to.fullPath)}`)
  }
})
