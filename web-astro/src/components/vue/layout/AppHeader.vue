<script setup lang="ts">
// 顶部导航（Vue island）：宋体字标 + 朱印 + 发丝线
// 登录态由 .astro 布局以 props 注入（SSR/首帧渲染一致，避免水合闪烁）
import { onMounted, ref, computed } from 'vue'
import { Moon, Sun, Menu, X, LayoutDashboard } from '@lucide/vue'
import { authState, isLoggedIn, logout } from '../../../lib/vue-store'
import { getTheme, toggleTheme } from '../../../lib/theme'
import { weavatarUrl } from '../../../lib/weavatar'

interface MeInfo {
  name?: string | null
  username?: string | null
  email?: string | null
  avatar?: string | null
  avatar_url?: string | null
}

const props = withDefaults(
  defineProps<{
    /** 站点名称（由 .astro 布局传入） */
    name?: string
    /** 是否开放注册（false 时隐藏注册入口） */
    allowRegister?: boolean
    /** 是否已登录（SSR 注入） */
    authed?: boolean
    /** 登录用户信息（SSR 注入） */
    me?: MeInfo | null
  }>(),
  { name: '', allowRegister: true, authed: false, me: null },
)

const sealChar = computed(() => (props.name || '驿').slice(-1))

const navLinks = [
  { to: '/explore', label: '探索' },
  { to: '/plans', label: '套餐' },
]

const theme = ref<'light' | 'dark'>(getTheme())
const mobileOpen = ref(false)
const currentPath = ref('')

onMounted(() => {
  // 客户端挂载后才同步路径与主题（不影响首帧水合一致性）
  theme.value = getTheme()
  currentPath.value = window.location.pathname
})

function onThemeToggle() {
  theme.value = toggleTheme()
}

function onLogout() {
  logout()
  window.location.assign('/')
}

const isActive = (to: string) => currentPath.value === to || currentPath.value.startsWith(to + '/')
const isLogged = computed(() => props.authed || isLoggedIn())
const displayName = computed(
  () => props.me?.name || props.me?.username || authState.user?.name || authState.user?.username,
)
const avatarUrl = computed(() => {
  const m = props.me
  if (m) return m.avatar_url || m.avatar || weavatarUrl(m.email || m.username || 'user')
  const u = authState.user as (typeof authState.user & { email?: string }) | null
  if (!u) return ''
  return u.avatar_url || u.avatar || weavatarUrl(u.email || u.username || 'user')
})
</script>

<template>
  <header class="sticky top-0 z-30 border-b hairline bg-background/85 backdrop-blur-sm">
    <div class="container-site flex h-16 items-center justify-between">
      <!-- 字标 + 印章 -->
      <a href="/" class="group flex items-center gap-2.5">
        <span class="font-display text-[1.35rem] font-bold leading-none tracking-tight text-foreground">
          {{ name || '驿' }}
        </span>
        <span class="seal h-[1.35rem] w-[1.35rem] text-[0.65rem] transition-transform group-hover:-rotate-3" :aria-label="sealChar">
          {{ sealChar }}
        </span>
      </a>

      <!-- 导航 -->
      <nav class="hidden items-center gap-8 md:flex">
        <a
          v-for="link in navLinks"
          :key="link.to"
          :href="link.to"
          class="relative text-sm transition-colors"
          :class="isActive(link.to) ? 'text-foreground' : 'text-muted-foreground hover:text-foreground'"
        >
          {{ link.label }}
          <span
            v-if="isActive(link.to)"
            class="absolute -bottom-[1.7rem] left-0 h-px w-full bg-brand"
          />
        </a>
      </nav>

      <!-- 右侧 -->
      <div class="flex items-center gap-1.5">
        <button
          class="hidden h-9 w-9 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-muted hover:text-foreground sm:inline-flex"
          :aria-label="theme === 'light' ? '切换暗色模式' : '切换亮色模式'"
          @click="onThemeToggle"
        >
          <Sun v-if="theme === 'light'" class="h-[1.05rem] w-[1.05rem]" />
          <Moon v-else class="h-[1.05rem] w-[1.05rem]" />
        </button>

        <template v-if="isLogged">
          <a href="/dashboard" class="ml-1 hidden items-center rounded-md px-2 py-1.5 transition-colors hover:bg-muted sm:flex">
            <img
              v-if="avatarUrl"
              :src="avatarUrl"
              :alt="displayName || 'avatar'"
              class="h-8 w-8 rounded-full object-cover ring-1 ring-border"
            />
          </a>
          <button
            class="ml-1 hidden h-9 items-center rounded-md px-2.5 text-sm text-muted-foreground transition-colors hover:text-foreground sm:inline-flex"
            @click="onLogout"
          >
            退出
          </button>
          <a href="/dashboard" class="inline-flex sm:hidden">
            <span class="grid h-9 w-9 place-items-center rounded-md text-muted-foreground transition-colors hover:bg-muted">
              <LayoutDashboard class="h-5 w-5" />
            </span>
          </a>
        </template>
        <template v-else>
          <a href="/auth/login" class="hidden rounded-md px-3 py-2 text-sm text-muted-foreground transition-colors hover:text-foreground sm:inline-block">
            登录
          </a>
          <a
            v-if="allowRegister"
            href="/auth/register"
            class="ml-1 hidden items-center rounded-md border border-border bg-transparent px-3.5 py-2 text-sm text-foreground transition-colors hover:border-brand/50 hover:text-brand sm:inline-flex"
          >
            注册
          </a>
          <a href="/auth/login" class="inline-flex px-2 py-2 text-sm text-muted-foreground sm:hidden">
            登录
          </a>
        </template>

        <!-- 移动端菜单 -->
        <button class="inline-flex h-9 w-9 items-center justify-center rounded-md text-muted-foreground hover:bg-muted md:hidden" aria-label="菜单" @click="mobileOpen = !mobileOpen">
          <Menu v-if="!mobileOpen" class="h-5 w-5" />
          <X v-else class="h-5 w-5" />
        </button>
      </div>
    </div>

    <!-- 移动端抽屉 -->
    <div v-if="mobileOpen" class="border-t hairline bg-background px-5 py-3 md:hidden">
      <nav class="flex flex-col">
        <a
          v-for="link in navLinks"
          :key="link.to"
          :href="link.to"
          class="border-b hairline py-3 text-sm"
          :class="isActive(link.to) ? 'font-medium text-foreground' : 'text-muted-foreground'"
        >
          {{ link.label }}
        </a>
        <button class="flex items-center gap-2 py-3 text-left text-sm text-muted-foreground" @click="onThemeToggle">
          <Sun v-if="theme === 'light'" class="h-4 w-4" />
          <Moon v-else class="h-4 w-4" />
          {{ theme === 'light' ? '切换到暗色模式' : '切换到亮色模式' }}
        </button>
      </nav>
    </div>
  </header>
</template>
