<script setup lang="ts">
// 管理后台布局：左侧深色侧边栏 + 顶部 admin 栏 + 右侧内容
const route = useRoute()

interface NavItem {
  to: string
  label: string
  icon?: string
  group?: string
}

const navItems: NavItem[] = [
  { group: '概览', to: '/admin', label: '仪表盘' },
  { group: '内容', to: '/admin/photos', label: '图片管理' },
  { group: '内容', to: '/admin/reports', label: '举报管理' },
  { group: '运营', to: '/admin/tickets', label: '工单管理' },
  { group: '运营', to: '/admin/notices', label: '通知管理' },
  { group: '运营', to: '/admin/pages', label: '单页管理' },
  { group: '运营', to: '/admin/violations', label: '违规记录' },
  { group: '运营', to: '/admin/feedbacks', label: '意见反馈' },
  { group: '用户', to: '/admin/users', label: '用户' },
  { group: '用户', to: '/admin/groups', label: '角色组' },
  { group: '系统', to: '/admin/storage', label: '存储策略' },
  { group: '系统', to: '/admin/drivers', label: '驱动管理' },
  { group: '系统', to: '/admin/license', label: 'License' },
]
</script>

<template>
  <div class="min-h-screen flex bg-gray-100">
    <aside class="w-60 bg-gray-900 text-gray-200 flex flex-col">
      <NuxtLink to="/admin" class="h-14 flex items-center px-5 text-lg font-semibold text-white border-b border-gray-800">
        云雾图驿 · 后台
      </NuxtLink>
      <nav class="flex-1 overflow-y-auto py-3 space-y-3">
        <template v-for="(group, gi) in [...new Set(navItems.map(i => i.group))]" :key="gi">
          <div class="px-5 text-xs uppercase tracking-wider text-gray-500 mt-2">{{ group }}</div>
          <NuxtLink
            v-for="item in navItems.filter(i => i.group === group)"
            :key="item.to"
            :to="item.to"
            class="block px-5 py-2 text-sm hover:bg-gray-800"
            :class="{ 'bg-gray-800 text-white border-l-2 border-primary-500': route.path === item.to }"
          >
            {{ item.label }}
          </NuxtLink>
        </template>
      </nav>
      <div class="p-4 border-t border-gray-800 text-xs text-gray-500">
        v{{ useRuntimeConfig().public.appVersion }}
      </div>
    </aside>

    <div class="flex-1 flex flex-col min-w-0">
      <header class="h-14 bg-white border-b border-gray-200 flex items-center justify-between px-6">
        <div class="text-sm text-gray-500">
          <slot name="breadcrumb">控制台</slot>
        </div>
        <div class="flex items-center gap-3">
          <NuxtLink to="/" class="text-sm text-gray-600 hover:text-primary-600">前台</NuxtLink>
          <div class="w-8 h-8 rounded-full bg-primary-500 text-white flex items-center justify-center text-sm">
            A
          </div>
        </div>
      </header>
      <main class="flex-1 overflow-y-auto p-6">
        <slot />
      </main>
    </div>
    <AppToaster />
  </div>
</template>
