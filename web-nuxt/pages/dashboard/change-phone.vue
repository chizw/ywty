<script setup lang="ts">
// 更换手机号
definePageMeta({ layout: 'dashboard', middleware: 'auth' })

const api = useApi()
const { user, fetchMe } = useAuth()

const form = reactive({ new_phone: '', code: '' })
const sending = ref(false)
const countdown = ref(0)
const loading = ref(false)
const msg = ref('')

async function sendCode() {
  if (!form.new_phone) {
    msg.value = '请填写新手机号'
    return
  }
  sending.value = true
  msg.value = ''
  try {
    await api.post('/api/v1/verify-codes', {
      channel: 'sms',
      account: form.new_phone,
      event: 'change_phone',
    })
    countdown.value = 60
    const t = setInterval(() => {
      countdown.value--
      if (countdown.value <= 0) clearInterval(t)
    }, 1000)
  } catch (err: any) {
    msg.value = err?.statusMessage || '发送失败'
  } finally {
    sending.value = false
  }
}

async function submit() {
  if (!form.new_phone || !form.code) {
    msg.value = '请填写完整'
    return
  }
  loading.value = true
  msg.value = ''
  try {
    await api.post('/api/v1/user/change-phone', form)
    msg.value = '更换成功'
    await fetchMe()
    form.new_phone = ''
    form.code = ''
  } catch (err: any) {
    msg.value = err?.statusMessage || '更换失败'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-900 mb-4">更换手机号</h1>
    <p class="text-sm text-gray-500 mb-4">当前手机：{{ (user as any)?.phone || '未绑定' }}</p>
    <form class="max-w-md bg-white border border-gray-200 rounded-lg p-6 space-y-4" @submit.prevent="submit">
      <div>
        <label class="block text-sm text-gray-700 mb-1">新手机号</label>
        <input v-model="form.new_phone" required class="w-full px-3 py-2 border border-gray-300 rounded-md" />
      </div>
      <div>
        <label class="block text-sm text-gray-700 mb-1">验证码</label>
        <div class="flex gap-2">
          <input v-model="form.code" required maxlength="6" class="flex-1 px-3 py-2 border border-gray-300 rounded-md" />
          <button
            type="button"
            class="px-4 py-2 border border-gray-300 text-sm rounded-md disabled:opacity-50"
            :disabled="countdown > 0 || sending"
            @click="sendCode"
          >{{ countdown > 0 ? `${countdown}s` : '发送验证码' }}</button>
        </div>
      </div>
      <p v-if="msg" class="text-sm" :class="msg.includes('失败') ? 'text-red-500' : 'text-primary-600'">{{ msg }}</p>
      <AppButton type="submit" :loading="loading" block>确认更换</AppButton>
    </form>
  </div>
</template>
