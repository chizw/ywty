<script setup lang="ts">
// 用户中心布局：左侧导航 + 右侧内容
import { Home, Image, BookImage, Crown, Share2, ShoppingCart, Tag, Ticket, Bell, Key, Link, Settings } from '@lucide/vue'

const route = useRoute()

const navItems = [
  { to: '/dashboard', label: '概览', icon: Home },
  { to: '/dashboard/photos', label: '我的图片', icon: Image },
  { to: '/dashboard/albums', label: '我的相册', icon: BookImage },
  { to: '/dashboard/plans', label: '套餐', icon: Crown },
  { to: '/dashboard/shares', label: '分享管理', icon: Share2 },
  { to: '/dashboard/orders', label: '订单', icon: ShoppingCart },
  { to: '/dashboard/tags', label: '标签', icon: Tag },
  { to: '/dashboard/tickets', label: '工单', icon: Ticket },
  { to: '/dashboard/notices', label: '通知', icon: Bell },
  { to: '/dashboard/tokens', label: 'API Token', icon: Key },
  { to: '/dashboard/oauth', label: '三方账号', icon: Link },
  { to: '/dashboard/settings', label: '设置', icon: Settings },
]
</script>

<template>
  <div class="min-h-screen flex flex-col bg-background">
    <AppHeader />
    <div class="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-6 flex gap-6">
      <aside class="w-56 flex-shrink-0">
        <Card class="p-2">
          <nav class="space-y-1">
            <NuxtLink
              v-for="item in navItems"
              :key="item.to"
              :to="item.to"
              class="flex items-center gap-2 px-3 py-2 rounded-md text-sm transition-colors"
              :class="route.path === item.to
                ? 'bg-accent text-accent-foreground font-medium'
                : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'"
            >
              <component :is="item.icon" class="h-4 w-4 shrink-0" />
              <span>{{ item.label }}</span>
            </NuxtLink>
          </nav>
        </Card>
      </aside>
      <section class="flex-1 min-w-0">
        <slot />
      </section>
    </div>
    <AppFooter />
    <AppToaster />
  </div>
</template>
