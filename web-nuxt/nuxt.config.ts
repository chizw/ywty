// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2025-01-01',
  devtools: { enabled: true },

  modules: [
    '@nuxtjs/tailwindcss',
    '@nuxt/image',
    '@pinia/nuxt',
    '@nuxtjs/i18n',
    'shadcn-nuxt',
  ],

  shadcn: {
    prefix: '',
    componentDir: '@/components/ui',
  },

  // i18n（P8 阶段启用）
  i18n: {
    strategy: 'no_prefix',
    defaultLocale: 'zh-CN',
    locales: [
      { code: 'zh-CN', name: '简体中文', file: 'zh-CN.json' },
      { code: 'en-US', name: 'English', file: 'en-US.json' },
    ],
    langDir: 'locales',
    detectBrowserLanguage: { useCookie: true, cookieKey: 'yunwu_lang' },
  },

  // useHead 由 Nuxt 内置的 unhead 提供，无需额外模块

  // SSR 模式（公开端 + 用户中心都需 SSR；后台可切换为 SPA）
  ssr: true,

  // 运行时配置
  // 合并部署时 apiBase 留空 → 走相对路径，由 Nitro routeRules 反代到 Go
  runtimeConfig: {
    apiBase: process.env.NUXT_API_BASE || '',
    public: {
      appName: '云雾图驿',
      appVersion: '0.1.0',
    },
  },

  // Nitro 反代：将 API / 静态资源 / 健康检查转发到内部 Go 服务
  nitro: {
    routeRules: {
      // 后台 / 用户中心 走 SPA 模式：避免 SSR 阶段空 token 渲染导致的
      // 中间件 302、Pinia state 覆盖、useAsyncData 401 等问题
      '/admin/**': { ssr: false },
      '/dashboard/**': { ssr: false },
      '/api/**': { proxy: `${process.env.NUXT_API_INTERNAL || 'http://127.0.0.1:8080'}/api/**` },
      '/uploads/**': { proxy: `${process.env.NUXT_API_INTERNAL || 'http://127.0.0.1:8080'}/uploads/**` },
      '/i/**': { proxy: `${process.env.NUXT_API_INTERNAL || 'http://127.0.0.1:8080'}/i/**` },
      '/s/**': { proxy: `${process.env.NUXT_API_INTERNAL || 'http://127.0.0.1:8080'}/s/**` },
      '/healthz': { proxy: `${process.env.NUXT_API_INTERNAL || 'http://127.0.0.1:8080'}/healthz` },
    },
  },

  // 主题
  app: {
    head: {
      title: '云雾图驿',
      titleTemplate: '%s · 云雾图驿',
      htmlAttrs: { lang: 'zh-CN' },
      meta: [
        { charset: 'utf-8' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
        { name: 'description', content: '云雾图驿 - 开箱即用的图床/相册系统' },
        { property: 'og:site_name', content: '云雾图驿' },
        { property: 'og:type', content: 'website' },
        { name: 'twitter:card', content: 'summary_large_image' },
      ],
    },
  },

  // 样式
  css: ['~/assets/css/main.css'],

  // TypeScript
  typescript: {
    strict: true,
    shim: false,
  },

  // Vite
  vite: {
    server: {
      hmr: { protocol: 'ws' },
    },
  },

  // 实验性：异步入口（更快的冷启动）
  experimental: {
    asyncContext: true,
  },
})
