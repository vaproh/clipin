import { defineNuxtRouteMiddleware, navigateTo } from '#app'

export default defineNuxtRouteMiddleware((to) => {
  if (process.env.E2E_TEST) {
    return navigateTo('/sign-in')
  }

  const { userId } = useAuth()

  if (!userId.value) {
    return navigateTo(`/sign-in?redirect_url=${encodeURIComponent(to.fullPath)}`)
  }
})
