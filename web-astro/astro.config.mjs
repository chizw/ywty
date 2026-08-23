// https://astro.build/config
import { defineConfig } from 'astro/config'
import vue from '@astrojs/vue'
import react from '@astrojs/react'
import tailwind from '@astrojs/tailwind'
import node from '@astrojs/node'

// 后端地址：开发时 Rust 服务默认跑在 3000 端口
const BACKEND = process.env.YWTY_API_INTERNAL || 'http://127.0.0.1:3000'

// https://astro.build/config
export default defineConfig({
  output: 'server',
  // SSR 部署：standalone 模式，由外层反代（nginx/Docker）转发
  adapter: node({ mode: 'standalone' }),
  site: process.env.SITE_URL || 'http://localhost:4321',

  // Vue(公开页) + React(用户中心/后台) 双 Islands
  integrations: [
    tailwind({
      // global.css 自带 @tailwind 指令，关闭自动注入避免重复
      applyBaseStyles: false,
    }),
    vue(),
    react(),
  ],

  // SSR 输出目录（合并部署时由外层反代到 Astro）
  build: {
    inlineStylesheets: 'auto',
  },

  vite: {
    // 预打包全部依赖：避免 Vite 中途发现新依赖重新优化，
    // 导致已发出的旧 ?v= 哈希 504（Outdated Optimize Dep），进而 island 注水失败
    optimizeDeps: {
      include: [
        'vue',
        'react',
        'react-dom',
        'react-dom/client',
        'react/jsx-runtime',
        'zustand',
        '@radix-ui/react-slot',
        '@radix-ui/react-label',
        '@radix-ui/react-dialog',
        '@radix-ui/react-select',
        '@radix-ui/react-tabs',
        '@radix-ui/react-avatar',
        '@radix-ui/react-checkbox',
        '@radix-ui/react-separator',
        '@radix-ui/react-dropdown-menu',
        'class-variance-authority',
        'clsx',
        'tailwind-merge',
        '@lucide/vue',
        'lucide-react',
        'radix-vue',
      ],
    },
    server: {
      proxy: {
        // 开发期把 API 与资源代理到 Rust 后端
        // 注意：'/s/'（带尾斜杠）而非 '/s'，避免前缀匹配吞掉 /share/*、/search 等
        '/api': BACKEND,
        '/uploads': BACKEND,
        '/i/': BACKEND,
        '/s/': BACKEND,
        '/healthz': BACKEND,
      },
    },
  },
})
