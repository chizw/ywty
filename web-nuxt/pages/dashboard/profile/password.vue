<script setup lang="ts">
// 个人资料 - 修改密码
definePageMeta({ layout: 'dashboard', middleware: 'auth' })

const api = useApi()
const message = useMessage()

const form = reactive({
  old_password: '',
  password: '',
  password_confirmation: '',
})

const changing = ref(false)

async function changePassword() {
  if (!form.old_password || !form.password || !form.password_confirmation) {
    message.error('请填写完整')
    return
  }
  if (form.password !== form.password_confirmation) {
    message.error('两次密码不一致')
    return
  }
  if (form.password.length < 6) {
    message.error('密码至少 6 位')
    return
  }

  changing.value = true
  try {
    await api.post('/api/v1/user/change-password', {
      old_password: form.old_password,
      password: form.password,
      password_confirmation: form.password_confirmation,
    })
    message.success('密码已修改')
    form.old_password = ''
    form.password = ''
    form.password_confirmation = ''
  } catch (err: any) {
    message.error(err?.statusMessage || '修改失败')
  } finally {
    changing.value = false
  }
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-900 mb-4">修改密码</h1>

    <div class="mb-6 flex flex-wrap gap-2">
      <NuxtLink to="/dashboard/profile" class="px-3 py-1.5 text-sm rounded-md bg-gray-100 text-gray-700 hover:bg-gray-200">基本信息</NuxtLink>
      <NuxtLink to="/dashboard/profile/email" class="px-3 py-1.5 text-sm rounded-md bg-gray-100 text-gray-700 hover:bg-gray-200">更换邮箱</NuxtLink>
      <NuxtLink to="/dashboard/profile/phone" class="px-3 py-1.5 text-sm rounded-md bg-gray-100 text-gray-700 hover:bg-gray-200">更换手机</NuxtLink>
      <NuxtLink to="/dashboard/profile/password" class="px-3 py-1.5 text-sm rounded-md bg-primary-50 text-primary-700">修改密码</NuxtLink>
      <NuxtLink to="/dashboard/profile/social" class="px-3 py-1.5 text-sm rounded-md bg-gray-100 text-gray-700 hover:bg-gray-200">社交账号</NuxtLink>
    </div>

    <div class="bg-white border border-gray-200 rounded-lg p-6 space-y-4 max-w-2xl">
      <div>
        <label class="block text-sm text-gray-700 mb-1">当前密码</label>
        <input v-model="form.old_password" type="password" class="w-full px-3 py-2 border border-gray-300 rounded-md" placeholder="请输入当前密码" />
      </div>

      <div>
        <label class="block text-sm text-gray-700 mb-1">新密码</label>
        <input v-model="form.password" type="password" class="w-full px-3 py-2 border border-gray-300 rounded-md" placeholder="请输入新密码（至少 6 位）" />
      </div>

      <div>
        <label class="block text-sm text-gray-700 mb-1">确认新密码</label>
        <input v-model="form.password_confirmation" type="password" class="w-full px-3 py-2 border border-gray-300 rounded-md" placeholder="请再次输入新密码" />
      </div>

      <AppButton type="button" :loading="changing" @click="changePassword">修改密码</AppButton>
    </div>
  </div>
</template>
