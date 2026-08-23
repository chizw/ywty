<script setup lang="ts">
// 找回密码表单（Vue island）：邮箱验证 → 重置
import { reactive, ref } from 'vue'
import Button from '../ui/Button.vue'
import Input from '../ui/Input.vue'
import Label from '../ui/Label.vue'
import VerifyCodeInput from './VerifyCodeInput.vue'
import { useApi } from '../../../lib/api'

const form = reactive({ account: '', code: '', password: '' })
const loading = ref(false)
const msg = ref('')

async function submit() {
  loading.value = true
  msg.value = ''
  try {
    await useApi().post('/api/v1/auth/reset-password', {
      email: form.account,
      code: form.code,
      new_password: form.password,
    })
    msg.value = '重置成功，即将跳到登录页'
    setTimeout(() => window.location.assign('/auth/login'), 1000)
  } catch (err: any) {
    msg.value = err?.message || '重置失败'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div>
    <h1 class="font-display text-2xl font-bold tracking-tight">找回密码</h1>
    <p class="mt-1.5 text-sm text-muted-foreground">通过邮箱验证后重置</p>

    <form class="mt-7 space-y-5" @submit.prevent="submit">
      <div class="space-y-2">
        <Label for="reset-account">邮箱</Label>
        <Input
          id="reset-account"
          v-model="form.account"
          type="email"
          placeholder="you@example.com"
          variant="underline"
        />
      </div>
      <VerifyCodeInput :account="form.account" event="reset_password" />
      <div class="space-y-2">
        <Label for="reset-code">验证码</Label>
        <Input id="reset-code" v-model="form.code" placeholder="验证码（自动填充或手动输入）" variant="underline" />
      </div>
      <div class="space-y-2">
        <Label for="reset-password">新密码</Label>
        <Input id="reset-password" v-model="form.password" type="password" minlength="6" placeholder="新密码（至少 6 位）" variant="underline" />
      </div>

      <p v-if="msg" class="text-[13px]" :class="msg.includes('成功') ? 'text-emerald-600 dark:text-emerald-400' : 'text-destructive'">
        {{ msg }}
      </p>

      <Button type="submit" :loading="loading" class="w-full">重置密码</Button>
    </form>

    <p class="mt-6 text-center text-sm text-muted-foreground">
      记起密码了？
      <a href="/auth/login" class="font-medium text-brand hover:underline">去登录</a>
    </p>
  </div>
</template>
