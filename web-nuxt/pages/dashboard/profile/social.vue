<script setup lang="ts">
// 个人资料 - 社交账号
definePageMeta({ layout: 'dashboard', middleware: 'auth' })

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
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-900 mb-4">社交账号</h1>

    <div class="mb-6 flex flex-wrap gap-2">
      <NuxtLink to="/dashboard/profile" class="px-3 py-1.5 text-sm rounded-md bg-gray-100 text-gray-700 hover:bg-gray-200">基本信息</NuxtLink>
      <NuxtLink to="/dashboard/profile/email" class="px-3 py-1.5 text-sm rounded-md bg-gray-100 text-gray-700 hover:bg-gray-200">更换邮箱</NuxtLink>
      <NuxtLink to="/dashboard/profile/phone" class="px-3 py-1.5 text-sm rounded-md bg-gray-100 text-gray-700 hover:bg-gray-200">更换手机</NuxtLink>
      <NuxtLink to="/dashboard/profile/password" class="px-3 py-1.5 text-sm rounded-md bg-gray-100 text-gray-700 hover:bg-gray-200">修改密码</NuxtLink>
      <NuxtLink to="/dashboard/profile/social" class="px-3 py-1.5 text-sm rounded-md bg-primary-50 text-primary-700">社交账号</NuxtLink>
    </div>

    <div v-if="loading" class="text-center py-12 text-gray-500">加载中...</div>

    <div v-else class="bg-white border border-gray-200 rounded-lg divide-y max-w-2xl">
      <div v-for="p in providers" :key="p.key" class="p-4 flex items-center justify-between">
        <div class="flex items-center gap-3">
          <span class="text-2xl">{{ p.icon }}</span>
          <div>
            <div class="text-sm font-medium text-gray-900">{{ p.name }}</div>
            <div v-if="isBound(p.key)" class="text-xs text-gray-500">
              已绑定：{{ getAccount(p.key)?.union_id }}
            </div>
            <div v-else class="text-xs text-gray-400">未绑定</div>
          </div>
        </div>

        <div>
          <a
            v-if="!isBound(p.key)"
            :href="bindUrl(p.key)"
            class="px-3 py-1.5 text-sm bg-primary-600 text-white rounded-md hover:bg-primary-700"
          >
            绑定
          </a>
          <button
            v-else
            class="px-3 py-1.5 text-sm border border-gray-300 text-gray-700 rounded-md hover:bg-gray-50"
            @click="unbind(p.key)"
          >
            解绑
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
