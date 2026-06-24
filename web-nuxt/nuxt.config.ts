// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2025-01-01',
  devtools: { enabled: true },

  modules: [
    '@nuxtjs/tailwindcss',
    '@nuxt/image',
    '@pinia/nuxt',
    '@nuxtjs/i18n',
  ],

  // i18n（P8 阶段启用）
  i18n: {
    strategy: 'no_prefix',
    defaultLocale: 'zh-CN',
    locales: [
      { code: 'zh-CN', name: '简体中文', file: 'zh-CN.json' },
      { code: 'en-US', name: 'English', file: 'en-US.json' },
    ],
    langDir: 'locales',
    detectBrowserLanguage: { useCookie: true, cookieKey: 'ywty_lang' },
  },

  // useHead 由 Nuxt 内置的 unhead 提供，无需额外模块

  // SSR 模式（公开端 + 用户中心都需 SSR；后台可切换为 SPA）
  ssr: true,

  // 运行时配置
  // 合并部署时 apiBase 留空 → 走相对路径，由 Nitro routeRules 反代到 Go
  runtimeConfig: {
    apiBase: process.env.NUXT_API_BASE || '',
    public: {
      appName: 'ywty',
      appVersion: '0.1.0',
    },
  },

  // Nitro 反代：将 API / 静态资源 / 健康检查转发到内部 Go 服务
  nitro: {
    routeRules: {
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
      title: 'ywty',
      titleTemplate: '%s · ywty',
      htmlAttrs: { lang: 'zh-CN' },
      meta: [
        { charset: 'utf-8' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
        { name: 'description', content: '自托管图床 / 云相册' },
        { property: 'og:site_name', content: 'ywty' },
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
