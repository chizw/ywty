// 鉴权中间件：未登录跳�?/auth/login
export default defineNuxtRouteMiddleware((to) => {
  const auth = useAuthStore()
  auth.hydrate()
  if (!auth.isLoggedIn) {
    return navigateTo({ path: '/auth/login', query: { redirect: to.fullPath } })
  }
})
