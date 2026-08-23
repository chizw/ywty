<script setup lang="ts">
// 注册表单（Vue island）
import { reactive, ref } from 'vue'
import Button from '../ui/Button.vue'
import Input from '../ui/Input.vue'
import Label from '../ui/Label.vue'
import { register } from '../../../lib/vue-store'

withDefaults(
  defineProps<{
    /** 站点名称（由注册页传入，用于标语） */
    siteName?: string
  }>(),
  { siteName: '' },
)

const form = reactive({ username: '', email: '', password: '', phone: '' })
const loading = ref(false)
const errorMsg = ref('')

async function onSubmit() {
  if (form.username.length < 3) {
    errorMsg.value = '用户名至少 3 个字符'
    return
  }
  if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.email)) {
    errorMsg.value = '邮箱格式不正确'
    return
  }
  if (form.password.length < 6) {
    errorMsg.value = '密码至少 6 位'
    return
  }
  loading.value = true
  errorMsg.value = ''
  try {
    await register({
      username: form.username,
      email: form.email,
      password: form.password,
      phone: form.phone || undefined,
    })
    window.location.assign('/dashboard')
  } catch (err: any) {
    errorMsg.value = err?.message || '注册失败'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div>
    <h1 class="font-display text-2xl font-bold tracking-tight">注册</h1>
    <p class="mt-1.5 text-sm text-muted-foreground">
      {{ siteName ? `加入 ${siteName}` : '创建你的新账号' }}
    </p>

    <form class="mt-7 space-y-5" @submit.prevent="onSubmit">
      <div class="space-y-2">
        <Label for="reg-username">用户名</Label>
        <Input id="reg-username" v-model="form.username" required minlength="3" maxlength="32" placeholder="3-32 个字符" variant="underline" />
      </div>
      <div class="space-y-2">
        <Label for="reg-email">邮箱</Label>
        <Input id="reg-email" v-model="form.email" type="email" required placeholder="you@example.com" variant="underline" />
      </div>
      <div class="space-y-2">
        <Label for="reg-phone">手机号（选填）</Label>
        <Input id="reg-phone" v-model="form.phone" placeholder="可用于登录" variant="underline" />
      </div>
      <div class="space-y-2">
        <Label for="reg-password">密码</Label>
        <Input id="reg-password" v-model="form.password" type="password" required minlength="6" placeholder="至少 6 位" variant="underline" />
      </div>

      <p v-if="errorMsg" class="text-[13px] text-destructive">{{ errorMsg }}</p>

      <Button type="submit" :loading="loading" class="w-full">注册</Button>
    </form>

    <p class="mt-6 text-center text-sm text-muted-foreground">
      已有账号？
      <a href="/auth/login" class="font-medium text-brand hover:underline">去登录</a>
    </p>
  </div>
</template>
