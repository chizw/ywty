<script setup lang="ts">
// 个人资料 - 更换邮箱
definePageMeta({ layout: 'dashboard', middleware: 'auth' })

const api = useApi()
const message = useMessage()
const { user, fetchMe } = useAuth()

const form = reactive({
  email: '',
  code: '',
})

const sending = ref(false)
const changing = ref(false)
const countdown = ref(0)

async function sendCode() {
  if (!form.email) {
    message.error('请先输入新邮箱')
    return
  }
  if (countdown.value > 0) return

  sending.value = true
  try {
    await api.post('/api/v1/verify-codes', {
      channel: 'email',
      account: form.email,
      event: 'change_email',
    })
    message.success('验证码已发送')
    countdown.value = 60
    const t = setInterval(() => {
      countdown.value--
      if (countdown.value <= 0) clearInterval(t)
    }, 1000)
  } catch (err: any) {
    message.error(err?.statusMessage || '发送失败')
  } finally {
    sending.value = false
  }
}

async function changeEmail() {
  if (!form.email || !form.code) {
    message.error('请填写完整')
    return
  }

  changing.value = true
  try {
    await api.post('/api/v1/user/change-email', {
      email: form.email,
      code: form.code,
    })
    await fetchMe()
    message.success('邮箱已更换')
    form.email = ''
    form.code = ''
  } catch (err: any) {
    message.error(err?.statusMessage || '更换失败')
  } finally {
    changing.value = false
  }
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-900 mb-4">更换邮箱</h1>

    <div class="mb-6 flex flex-wrap gap-2">
      <NuxtLink to="/dashboard/profile" class="px-3 py-1.5 text-sm rounded-md bg-gray-100 text-gray-700 hover:bg-gray-200">基本信息</NuxtLink>
      <NuxtLink to="/dashboard/profile/email" class="px-3 py-1.5 text-sm rounded-md bg-primary-50 text-primary-700">更换邮箱</NuxtLink>
      <NuxtLink to="/dashboard/profile/phone" class="px-3 py-1.5 text-sm rounded-md bg-gray-100 text-gray-700 hover:bg-gray-200">更换手机</NuxtLink>
      <NuxtLink to="/dashboard/profile/password" class="px-3 py-1.5 text-sm rounded-md bg-gray-100 text-gray-700 hover:bg-gray-200">修改密码</NuxtLink>
      <NuxtLink to="/dashboard/profile/social" class="px-3 py-1.5 text-sm rounded-md bg-gray-100 text-gray-700 hover:bg-gray-200">社交账号</NuxtLink>
    </div>

    <div class="bg-white border border-gray-200 rounded-lg p-6 space-y-4 max-w-2xl">
      <div>
        <label class="block text-sm text-gray-700 mb-1">当前邮箱</label>
        <input :value="user?.email" disabled class="w-full px-3 py-2 border border-gray-300 rounded-md bg-gray-50" />
      </div>

      <div>
        <label class="block text-sm text-gray-700 mb-1">新邮箱</label>
        <input v-model="form.email" type="email" class="w-full px-3 py-2 border border-gray-300 rounded-md" placeholder="请输入新邮箱" />
      </div>

      <div>
        <label class="block text-sm text-gray-700 mb-1">验证码</label>
        <div class="flex gap-2">
          <input v-model="form.code" type="text" maxlength="6" class="flex-1 px-3 py-2 border border-gray-300 rounded-md" placeholder="请输入验证码" />
          <button
            type="button"
            class="px-4 py-2 border border-gray-300 text-sm rounded-md disabled:opacity-50 whitespace-nowrap"
            :disabled="countdown > 0 || sending"
            @click="sendCode"
          >
            {{ countdown > 0 ? `${countdown}s` : '发送验证码' }}
          </button>
        </div>
      </div>

      <AppButton type="button" :loading="changing" @click="changeEmail">更换邮箱</AppButton>
    </div>
  </div>
</template>
