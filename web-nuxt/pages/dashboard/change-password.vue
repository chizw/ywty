<script setup lang="ts">
// 修改密码
definePageMeta({ layout: 'dashboard', middleware: 'auth' })

const api = useApi()
const form = reactive({ old_password: '', new_password: '' })
const loading = ref(false)
const msg = ref('')

async function submit() {
  if (form.new_password.length < 6) {
    msg.value = '新密码至少 6 位'
    return
  }
  loading.value = true
  msg.value = ''
  try {
    await api.post('/api/v1/user/change-password', form)
    msg.value = '修改成功，请重新登录'
    setTimeout(async () => {
      const { logout } = useAuth()
      await logout()
    }, 1000)
  } catch (err: any) {
    msg.value = err?.statusMessage || '修改失败'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-900 mb-4">修改密码</h1>
    <form class="max-w-md bg-white border border-gray-200 rounded-lg p-6 space-y-4" @submit.prevent="submit">
      <div>
        <label class="block text-sm text-gray-700 mb-1">原密码</label>
        <input v-model="form.old_password" type="password" required class="w-full px-3 py-2 border border-gray-300 rounded-md" />
      </div>
      <div>
        <label class="block text-sm text-gray-700 mb-1">新密码</label>
        <input v-model="form.new_password" type="password" required minlength="6" class="w-full px-3 py-2 border border-gray-300 rounded-md" />
      </div>
      <p v-if="msg" class="text-sm" :class="msg.includes('成功') ? 'text-primary-600' : 'text-red-500'">{{ msg }}</p>
      <AppButton type="submit" :loading="loading" block>修改密码</AppButton>
    </form>
  </div>
</template>
