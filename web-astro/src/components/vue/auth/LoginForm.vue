<script setup lang="ts">
// 登录表单（Vue island）
import { reactive, ref } from 'vue'
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
        <Input id="login-account" v-model="form.account" type="text" required placeholder="用户名 / 邮箱 / 手机号" variant="underline" />
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

    <p v-if="allowRegister" class="mt-6 text-center text-sm text-muted-foreground">
      还没有账号？
      <a href="/auth/register" class="font-medium text-brand hover:underline">立即注册</a>
    </p>
  </div>
</template>
