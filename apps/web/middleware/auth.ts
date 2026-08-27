import { defineNuxtRouteMiddleware, navigateTo } from '#app'

export default defineNuxtRouteMiddleware((to) => {
  if (process.env.E2E_TEST) {
    return navigateTo('/sign-in')
  }

  const { userId } = useAuth()

  if (!userId.value) {
    const returnPath = to.fullPath.startsWith('/') && !to.fullPath.startsWith('//')
      ? to.fullPath
      : '/app'
    return navigateTo(`/sign-in?redirect_url=${encodeURIComponent(returnPath)}`)
  }
})
