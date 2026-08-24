<script setup lang="ts">
// OAuth 回调处理（Vue island）：
// - 登录流程：后端 302 到 /auth/callback#access_token=..&refresh_token=..&user=<json>
// - 绑定流程：302 到 /dashboard/oauth?bound=<provider>（不会进入本页）
// hash 中数据不经过服务器，解析后写入本地会话再跳转 dashboard。
import { onMounted, ref } from 'vue'
import { writeAuthPair, type TokenPair } from '../../../lib/auth'

const status = ref<'working' | 'error'>('working')
const message = ref('正在完成登录…')

function fail(msg: string) {
  status.value = 'error'
  message.value = msg
}

onMounted(() => {
  try {
    const hash = window.location.hash.replace(/^#/, '')
    const params = new URLSearchParams(hash)
    const access_token = params.get('access_token') || ''
    const refresh_token = params.get('refresh_token') || ''
    const userRaw = params.get('user') || ''

    if (!access_token) {
      fail('登录数据缺失，请重新发起授权')
      return
    }

    let user: Record<string, unknown> = {}
    try {
      const parsed = JSON.parse(decodeURIComponent(userRaw))
      if (parsed && typeof parsed === 'object') user = parsed as Record<string, unknown>
    } catch {
      /* 保持空对象 */
    }

    const pair: TokenPair = {
      access_token,
      refresh_token,
      token_type: 'Bearer',
      expires_at: new Date(Date.now() + 7 * 864e5).toISOString(),
      user: {
        id: Number(user.id ?? 0),
        uuid: String(user.uuid ?? ''),
        username: String(user.username ?? ''),
        email: String(user.email ?? ''),
        avatar: (user.avatar as string | null) ?? null,
        role: String(user.role ?? 'user'),
        name: String(user.username ?? ''),
      },
    }
    writeAuthPair(pair)
    window.location.replace('/dashboard')
  } catch {
    fail('登录数据处理失败')
  }
})
</script>

<template>
  <div class="mx-auto max-w-md px-6 py-24 text-center">
    <div v-if="status !== 'error'" class="space-y-3">
      <div class="mx-auto h-8 w-8 animate-spin rounded-full border-2 border-border border-t-brand" />
      <p class="text-sm text-muted-foreground">{{ message }}</p>
    </div>
    <div v-else class="space-y-3">
      <p class="text-sm text-destructive">{{ message }}</p>
      <a href="/auth/login" class="text-sm font-medium text-brand hover:underline">返回登录</a>
    </div>
  </div>
</template>
