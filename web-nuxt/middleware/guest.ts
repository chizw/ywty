// 游客中间件：已登录用户访问登�?注册页时跳到首页
export default defineNuxtRouteMiddleware(() => {
  const auth = useAuthStore()
  auth.hydrate()
  if (auth.isLoggedIn) {
    return navigateTo('/')
  }
})
