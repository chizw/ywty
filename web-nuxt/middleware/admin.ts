// 管理员中间件：未登录或非管理员跳走
export default defineNuxtRouteMiddleware((to) => {
  if (import.meta.server) return
  const auth = useAuthStore()
  auth.hydrate()
  if (!auth.isLoggedIn) {
    return navigateTo({ path: '/auth/login', query: { redirect: to.fullPath } })
  }
  if (!auth.isAdmin) {
    return navigateTo('/')
  }
})
