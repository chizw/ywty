// 鉴权 hydrate 插件
// 修复以下问题：
//  1. 访问首页/公开页时没有 auth 中间件触发 hydrate()，store 始终是空
//  2. 刷新后 SSR 拿到空状态，hydration 后仍是空
//  3. 登录跳转后才正确显示用户名
//
// 必须在 Pinia 实例就绪后才能调用 useAuthStore()，所以用 onNuxtReady
// 延迟到 Nuxt 初始化完毕（此时 @pinia/nuxt 已注入 pinia 实例）。
export default defineNuxtPlugin({
  name: 'auth-hydrate',
  setup(nuxtApp) {
    nuxtApp.hook('app:created', () => {
      const auth = useAuthStore()
      // 每次入口都强制重读 cookie，避免上一次的 hydrated 标记锁住空状态
      auth.$patch({ hydrated: false })
      auth.hydrate()
    })
  },
})
