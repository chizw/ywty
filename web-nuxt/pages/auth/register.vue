<script setup lang="ts">
definePageMeta({ layout: 'auth' })

const { register } = useAuth()
const router = useRouter()

const form = reactive({ username: '', email: '', password: '', phone: '' })
const loading = ref(false)
const errorMsg = ref('')

async function onSubmit() {
  loading.value = true
  errorMsg.value = ''
  try {
    await register({
      username: form.username,
      email: form.email,
      password: form.password,
      phone: form.phone || undefined,
    })
    await router.push('/dashboard')
  } catch (err: any) {
    errorMsg.value = err?.statusMessage || err?.message || '注册失败'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-900 mb-1">注册</h1>
    <p class="text-sm text-gray-500 mb-6">创建你的 ywty 账号</p>
    <form class="space-y-4" @submit.prevent="onSubmit">
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">用户名</label>
        <input v-model="form.username" required minlength="3" maxlength="32" class="w-full px-3 py-2 border border-gray-300 rounded-md" />
      </div>
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">邮箱</label>
        <input v-model="form.email" type="email" required class="w-full px-3 py-2 border border-gray-300 rounded-md" />
      </div>
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">手机号（可选）</label>
        <input v-model="form.phone" class="w-full px-3 py-2 border border-gray-300 rounded-md" />
      </div>
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">密码</label>
        <input v-model="form.password" type="password" required minlength="6" class="w-full px-3 py-2 border border-gray-300 rounded-md" />
      </div>
      <p v-if="errorMsg" class="text-sm text-red-500">{{ errorMsg }}</p>
      <AppButton type="submit" :loading="loading" block>注册</AppButton>
    </form>
    <p class="mt-6 text-center text-sm text-gray-500">
      已有账号？
      <NuxtLink to="/auth/login" class="text-primary-600 hover:underline">去登录</NuxtLink>
    </p>
  </div>
</template>
