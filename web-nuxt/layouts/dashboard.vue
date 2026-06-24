<script setup lang="ts">
// 用户中心布局：左侧导航 + 右侧内容
const route = useRoute()
const navItems = [
  { to: '/dashboard', label: '概览', icon: 'home' },
  { to: '/dashboard/photos', label: '我的图片', icon: 'image' },
  { to: '/dashboard/albums', label: '我的相册', icon: 'album' },
  { to: '/dashboard/plans', label: '套餐', icon: 'crown' },
  { to: '/dashboard/shares', label: '分享管理', icon: 'share' },
  { to: '/dashboard/orders', label: '订单', icon: 'order' },
  { to: '/dashboard/tags', label: '标签', icon: 'tag' },
  { to: '/dashboard/tickets', label: '工单', icon: 'ticket' },
  { to: '/dashboard/notices', label: '通知', icon: 'bell' },
  { to: '/dashboard/tokens', label: 'API Token', icon: 'key' },
  { to: '/dashboard/oauth', label: '三方账号', icon: 'link' },
  { to: '/dashboard/settings', label: '设置', icon: 'cog' },
]
</script>

<template>
  <div class="min-h-screen flex flex-col bg-gray-50">
    <AppHeader />
    <div class="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-6 flex gap-6">
      <aside class="w-56 flex-shrink-0">
        <nav class="bg-white rounded-lg border border-gray-200 p-2 space-y-1">
          <NuxtLink
            v-for="item in navItems"
            :key="item.to"
            :to="item.to"
            class="block px-3 py-2 rounded text-sm text-gray-700 hover:bg-gray-100"
            :class="{ 'bg-primary-50 text-primary-700 font-medium': route.path === item.to }"
          >
            {{ item.label }}
          </NuxtLink>
        </nav>
      </aside>
      <section class="flex-1 min-w-0">
        <slot />
      </section>
    </div>
    <AppFooter />
    <AppToaster />
  </div>
</template>
