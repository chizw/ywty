<script setup lang="ts">
definePageMeta({ layout: 'auth' })

const { login } = useAuth()
const route = useRoute()
const router = useRouter()

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
    const redirect = (route.query.redirect as string) || '/dashboard'
    await router.push(redirect)
  } catch (err: any) {
    errorMsg.value = err?.statusMessage || err?.message || '登录失败'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-900 mb-1">登录</h1>
    <p class="text-sm text-gray-500 mb-6">使用账号密码登录到 ywty</p>
    <form class="space-y-4" @submit.prevent="onSubmit">
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">账号</label>
        <input
          v-model="form.account"
          type="text"
          required
          placeholder="用户名 / 邮箱 / 手机号"
          class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-primary-500 focus:border-primary-500"
        />
      </div>
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">密码</label>
        <input
          v-model="form.password"
          type="password"
          required
          minlength="6"
          placeholder="至少 6 位"
          class="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-primary-500 focus:border-primary-500"
        />
      </div>
      <p v-if="errorMsg" class="text-sm text-red-500">{{ errorMsg }}</p>
      <AppButton type="submit" :loading="loading" block>登录</AppButton>
    </form>
    <p class="mt-6 text-center text-sm text-gray-500">
      还没有账号？
      <NuxtLink to="/auth/register" class="text-primary-600 hover:underline">立即注册</NuxtLink>
    </p>
  </div>
</template>
