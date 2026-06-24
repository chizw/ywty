<script setup lang="ts">
// 重置密码（公开页）
definePageMeta({ layout: 'auth', middleware: 'guest' })

const route = useRoute()
const router = useRouter()
const api = useApi()
const mode = ref<'email' | 'phone'>('email')
const form = reactive({ account: '', code: '', password: '' })
const loading = ref(false)
const msg = ref('')

async function submit() {
  loading.value = true
  msg.value = ''
  try {
    const path = mode.value === 'email'
      ? '/api/v1/auth/reset-password'
      : '/api/v1/auth/reset-password/phone'
    await api.post(path, {
      [mode.value === 'email' ? 'email' : 'phone']: form.account,
      code: form.code,
      new_password: form.password,
    })
    msg.value = '重置成功，即将跳到登录页'
    setTimeout(() => router.push('/auth/login'), 1000)
  } catch (err: any) {
    msg.value = err?.statusMessage || '重置失败'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-900 mb-1">找回密码</h1>
    <p class="text-sm text-gray-500 mb-6">通过 {{ mode === 'email' ? '邮箱' : '手机' }} 验证后重置</p>

    <div class="flex gap-2 mb-4">
      <button
        type="button"
        class="flex-1 py-2 text-sm rounded-md"
        :class="mode === 'email' ? 'bg-primary-50 text-primary-700' : 'bg-gray-50 text-gray-600'"
        @click="mode = 'email'"
      >邮箱验证</button>
      <button
        type="button"
        class="flex-1 py-2 text-sm rounded-md"
        :class="mode === 'phone' ? 'bg-primary-50 text-primary-700' : 'bg-gray-50 text-gray-600'"
        @click="mode = 'phone'"
      >手机验证</button>
    </div>

    <form class="space-y-4" @submit.prevent="submit">
      <input
        v-model="form.account"
        :placeholder="mode === 'email' ? '邮箱' : '手机号'"
        class="w-full px-3 py-2 border border-gray-300 rounded-md"
      />
      <VerifyCodeInput
        :channel="mode"
        :account="form.account"
        :event="'reset_password'"
      />
      <input
        v-model="form.code"
        placeholder="验证码（自动填充，或手动输入）"
        class="w-full px-3 py-2 border border-gray-300 rounded-md"
      />
      <input
        v-model="form.password"
        type="password"
        minlength="6"
        placeholder="新密码（至少 6 位）"
        class="w-full px-3 py-2 border border-gray-300 rounded-md"
      />
      <p v-if="msg" class="text-sm text-center" :class="msg.includes('成功') ? 'text-primary-600' : 'text-red-500'">{{ msg }}</p>
      <AppButton type="submit" :loading="loading" block>重置密码</AppButton>
    </form>

    <p class="mt-6 text-center text-sm text-gray-500">
      记起密码了？<NuxtLink to="/auth/login" class="text-primary-600 hover:underline">去登录</NuxtLink>
    </p>
  </div>
</template>
