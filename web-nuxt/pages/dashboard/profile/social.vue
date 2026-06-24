<script setup lang="ts">
// 个人资料 - 社交账号
definePageMeta({ layout: 'dashboard', middleware: 'auth' })

import { Link2, Unlink } from '@lucide/vue'

const api = useApi()
const message = useMessage()

interface OAuthAccount {
  provider: string
  provider_name: string
  union_id: string
  created_at: string
}

const accounts = ref<OAuthAccount[]>([])
const loading = ref(true)

const providers = [
  { key: 'github', name: 'GitHub', icon: '🐙' },
  { key: 'google', name: 'Google', icon: '🔍' },
  { key: 'weixin', name: '微信', icon: '💬' },
  { key: 'qq', name: 'QQ', icon: '🐧' },
  { key: 'weibo', name: '微博', icon: '📢' },
  { key: 'gitee', name: 'Gitee', icon: '🍊' },
]

onMounted(async () => {
  await loadAccounts()
})

async function loadAccounts() {
  loading.value = true
  try {
    const res = await api.get<any>('/api/v1/oauth/accounts')
    const list = Array.isArray(res) ? res : ((res as any)?.data ?? [])
    accounts.value = list
  } catch {
    accounts.value = []
  } finally {
    loading.value = false
  }
}

function isBound(provider: string): boolean {
  return accounts.value.some((a) => a.provider === provider)
}

function getAccount(provider: string): OAuthAccount | undefined {
  return accounts.value.find((a) => a.provider === provider)
}

function bindUrl(provider: string): string {
  return `/api/v1/oauth/${provider}/bind`
}

async function unbind(provider: string) {
  if (!confirm(`确定解绑 ${provider}？`)) return

  try {
    await api.post(`/api/v1/oauth/${provider}/unbind`, {})
    message.success('已解绑')
    await loadAccounts()
  } catch (err: any) {
    message.error(err?.statusMessage || '解绑失败')
  }
}

const navItems = [
  { to: '/dashboard/profile', label: '基本信息' },
  { to: '/dashboard/profile/email', label: '更换邮箱' },
  { to: '/dashboard/profile/phone', label: '更换手机' },
  { to: '/dashboard/profile/password', label: '修改密码' },
  { to: '/dashboard/profile/social', label: '社交账号' },
]
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-foreground mb-4">社交账号</h1>

    <div class="mb-6 flex flex-wrap gap-2">
      <NuxtLink
        v-for="item in navItems"
        :key="item.to"
        :to="item.to"
        class="px-3 py-1.5 text-sm rounded-md"
        :class="item.to === '/dashboard/profile/social' ? 'bg-primary text-primary-foreground' : 'bg-muted text-muted-foreground hover:bg-muted/80'"
      >{{ item.label }}</NuxtLink>
    </div>

    <Skeleton v-if="loading" class="h-64 w-full max-w-2xl" />

    <Card v-else class="max-w-2xl">
      <CardContent class="p-0 divide-y divide-border">
        <div v-for="p in providers" :key="p.key" class="p-4 flex items-center justify-between">
          <div class="flex items-center gap-3">
            <span class="text-2xl">{{ p.icon }}</span>
            <div>
              <div class="text-sm font-medium text-foreground">{{ p.name }}</div>
              <div v-if="isBound(p.key)" class="text-xs text-muted-foreground">
                已绑定：{{ getAccount(p.key)?.union_id }}
              </div>
              <div v-else class="text-xs text-muted-foreground">未绑定</div>
            </div>
          </div>

          <div>
            <a
              v-if="!isBound(p.key)"
              :href="bindUrl(p.key)"
            >
              <Button size="sm">
                <Link2 class="mr-1 h-3 w-3" />
                绑定
              </Button>
            </a>
            <Button
              v-else
              variant="outline"
              size="sm"
              @click="unbind(p.key)"
            >
              <Unlink class="mr-1 h-3 w-3" />
              解绑
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>
  </div>
</template>
