<script setup lang="ts">
// 管理后台布局：左侧深色侧边栏 + 顶部 admin 栏 + 右侧内容
import { LayoutDashboard, Image, AlertTriangle, Ticket, Bell, FileText, ShieldAlert, MessageSquare, Users, Shield, HardDrive, Cpu, KeyRound } from '@lucide/vue'

const route = useRoute()

interface NavItem {
  to: string
  label: string
  icon?: Component
  group?: string
}

const navItems: NavItem[] = [
  { group: '概览', to: '/admin', label: '仪表盘', icon: LayoutDashboard },
  { group: '内容', to: '/admin/photos', label: '图片管理', icon: Image },
  { group: '内容', to: '/admin/reports', label: '举报管理', icon: AlertTriangle },
  { group: '运营', to: '/admin/tickets', label: '工单管理', icon: Ticket },
  { group: '运营', to: '/admin/notices', label: '通知管理', icon: Bell },
  { group: '运营', to: '/admin/pages', label: '单页管理', icon: FileText },
  { group: '运营', to: '/admin/violations', label: '违规记录', icon: ShieldAlert },
  { group: '运营', to: '/admin/feedbacks', label: '意见反馈', icon: MessageSquare },
  { group: '用户', to: '/admin/users', label: '用户', icon: Users },
  { group: '用户', to: '/admin/groups', label: '角色组', icon: Shield },
  { group: '系统', to: '/admin/storage', label: '存储策略', icon: HardDrive },
  { group: '系统', to: '/admin/drivers', label: '驱动管理', icon: Cpu },
  { group: '系统', to: '/admin/license', label: 'License', icon: KeyRound },
]

const groups = computed(() => [...new Set(navItems.map(i => i.group))])
</script>

<template>
  <div class="min-h-screen flex bg-background">
    <!-- 深色侧边栏 -->
    <aside class="w-60 bg-card border-r border-border flex flex-col">
      <NuxtLink to="/admin" class="h-14 flex items-center px-5 text-lg font-semibold text-foreground border-b border-border">
        云雾图驿 · 后台
      </NuxtLink>
      <nav class="flex-1 overflow-y-auto py-3 space-y-3">
        <template v-for="group in groups" :key="group">
          <div class="px-5 text-xs uppercase tracking-wider text-muted-foreground mt-2">{{ group }}</div>
          <NuxtLink
            v-for="item in navItems.filter(i => i.group === group)"
            :key="item.to"
            :to="item.to"
            class="flex items-center gap-2 px-5 py-2 text-sm hover:bg-accent hover:text-accent-foreground transition-colors"
            :class="{ 'bg-accent text-accent-foreground font-medium border-l-2 border-primary': route.path === item.to }"
          >
            <component v-if="item.icon" :is="item.icon" class="h-4 w-4 shrink-0" />
            <span>{{ item.label }}</span>
          </NuxtLink>
        </template>
      </nav>
      <div class="p-4 border-t border-border text-xs text-muted-foreground">
        v{{ useRuntimeConfig().public.appVersion }}
      </div>
    </aside>

    <!-- 主内容区 -->
    <div class="flex-1 flex flex-col min-w-0">
      <header class="h-14 bg-card border-b border-border flex items-center justify-between px-6">
        <div class="text-sm text-muted-foreground">
          <slot name="breadcrumb">控制台</slot>
        </div>
        <div class="flex items-center gap-3">
          <Button as-child variant="ghost" size="sm">
            <NuxtLink to="/">前台</NuxtLink>
          </Button>
          <Avatar class="h-8 w-8">
            <AvatarFallback class="bg-primary text-primary-foreground text-sm">A</AvatarFallback>
          </Avatar>
        </div>
      </header>
      <main class="flex-1 overflow-y-auto p-6">
        <slot />
      </main>
    </div>
    <AppToaster />
  </div>
</template>
