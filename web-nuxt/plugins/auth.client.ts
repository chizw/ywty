// 在 Nuxt 启动时执行：客户端水合 auth state
export default defineNuxtPlugin(() => {
  if (import.meta.client) {
    const auth = useAuthStore()
    auth.hydrate()
  }
})
