<script setup lang="ts">
// 登录表单（Vue island）
import { onMounted, reactive, ref } from 'vue'
import Button from '../ui/Button.vue'
import Input from '../ui/Input.vue'
import Label from '../ui/Label.vue'
import { login } from '../../../lib/vue-store'

const props = withDefaults(
  defineProps<{
    /** 站点名称（由登录页传入，用于标语） */
    siteName?: string
    /** 是否开放注册（false 时隐藏注册入口） */
    allowRegister?: boolean
  }>(),
  { siteName: '', allowRegister: true },
)

const redirectTo = ref<string>('/dashboard')

if (!import.meta.env.SSR) {
  const params = new URLSearchParams(window.location.search)
  redirectTo.value = params.get('redirect') || '/dashboard'
}

const form = reactive({ account: '', password: '' })
const loading = ref(false)
const errorMsg = ref('')

// 已配置的 OAuth 提供商
const oauthProviders = ref<{ provider: string; name: string }[]>([])

onMounted(async () => {
  try {
    const api = (await import('../../../lib/api')).useApi()
    const r = await api.get<any>('/api/v1/oauth/providers')
    oauthProviders.value = Array.isArray(r?.providers) ? r.providers : []
  } catch {
    oauthProviders.value = []
  }
})

function startOAuth(provider: string) {
  // 登录态下跳到 authorize 会进入绑定流程，这里始终以登录意图发起
  const url = `/api/v1/oauth/${provider}/authorize`
  fetch(url, { headers: { Accept: 'application/json' } })
    .then((r) => r.json())
    .then((body) => {
      const target = body?.data?.url
      if (target) window.location.href = target
      else errorMsg.value = '获取授权地址失败'
    })
    .catch(() => {
      errorMsg.value = '获取授权地址失败'
    })
}

async function onSubmit() {
  if (!form.account || !form.password) {
    errorMsg.value = '请输入账号和密码'
    return
  }
  loading.value = true
  errorMsg.value = ''
  try {
    await login(form.account, form.password)
    window.location.assign(redirectTo.value)
  } catch (err: any) {
    errorMsg.value = err?.message || '登录失败'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div>
    <h1 class="font-display text-2xl font-bold tracking-tight">登录</h1>
    <p class="mt-1.5 text-sm text-muted-foreground">
      {{ siteName ? `欢迎回到 ${siteName}` : '回到你的图库' }}
    </p>

    <form class="mt-7 space-y-5" @submit.prevent="onSubmit">
      <div class="space-y-2">
        <Label for="login-account">账号</Label>
        <Input id="login-account" v-model="form.account" type="text" required placeholder="用户名 / 邮箱" variant="underline" />
      </div>
      <div class="space-y-2">
        <div class="flex items-center justify-between">
          <Label for="login-password">密码</Label>
          <a href="/auth/reset-password" class="text-xs text-muted-foreground transition-colors hover:text-brand">
            忘记密码？
          </a>
        </div>
        <Input id="login-password" v-model="form.password" type="password" required minlength="6" placeholder="至少 6 位" variant="underline" />
      </div>

      <p v-if="errorMsg" class="text-[13px] text-destructive">{{ errorMsg }}</p>

      <Button type="submit" :loading="loading" class="w-full">登录</Button>
    </form>

    <div v-if="oauthProviders.length" class="mt-6">
      <div class="flex items-center gap-3 text-xs text-muted-foreground">
        <span class="h-px flex-1 bg-border" /> 或使用第三方登录 <span class="h-px flex-1 bg-border" />
      </div>
      <div class="mt-3 grid gap-2" :class="oauthProviders.length > 1 ? 'grid-cols-2' : 'grid-cols-1'">
        <Button
          v-for="p in oauthProviders"
          :key="p.provider"
          type="button"
          variant="outline"
          @click="startOAuth(p.provider)"
        >
          {{ p.name }}
        </Button>
      </div>
    </div>

    <p v-if="allowRegister" class="mt-6 text-center text-sm text-muted-foreground">
      还没有账号？
      <a href="/auth/register" class="font-medium text-brand hover:underline">立即注册</a>
    </p>
  </div>
</template>
