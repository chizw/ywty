<script setup lang="ts">
// 更换邮箱
definePageMeta({ layout: 'dashboard', middleware: 'auth' })

import { Mail } from '@lucide/vue'

const api = useApi()
const { user, fetchMe } = useAuth()

const form = reactive({ new_email: '', code: '' })
const sending = ref(false)
const countdown = ref(0)
const loading = ref(false)
const msg = ref('')

async function sendCode() {
  if (!form.new_email) {
    msg.value = '请填写新邮箱'
    return
  }
  sending.value = true
  msg.value = ''
  try {
    await api.post('/api/v1/verify-codes', {
      channel: 'email',
      account: form.new_email,
      event: 'change_email',
    })
    countdown.value = 60
    const t = setInterval(() => {
      countdown.value--
      if (countdown.value <= 0) clearInterval(t)
    }, 1000)
  } catch (err: any) {
    msg.value = err?.statusMessage || '发送失�?
  } finally {
    sending.value = false
  }
}

async function submit() {
  if (!form.new_email || !form.code) {
    msg.value = '请填写完�?
    return
  }
  loading.value = true
  msg.value = ''
  try {
    await api.post('/api/v1/user/change-email', form)
    msg.value = '更换成功'
    await fetchMe()
    form.new_email = ''
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
    <h1 class="text-2xl font-bold text-foreground mb-4">更换邮箱</h1>
    <p class="text-sm text-muted-foreground mb-4">当前邮箱：{{ user?.email }}</p>
    <Card class="max-w-md">
      <form @submit.prevent="submit">
        <CardContent class="pt-6 space-y-4">
          <div>
            <Label>新邮�?/Label>
            <Input v-model="form.new_email" type="email" required class="mt-1" />
          </div>
          <div>
            <Label>验证�?/Label>
            <div class="flex gap-2 mt-1">
              <Input v-model="form.code" required maxlength="6" class="flex-1" />
              <Button
                variant="outline"
                :disabled="countdown > 0 || sending"
                @click="sendCode"
              >
                <Mail class="mr-1 h-4 w-4" />
                {{ countdown > 0 ? `${countdown}s` : '发送验证码' }}
              </Button>
            </div>
          </div>
          <p v-if="msg" class="text-sm" :class="msg.includes('失败') ? 'text-destructive' : 'text-green-600'">{{ msg }}</p>
        </CardContent>
        <CardFooter>
          <Button type="submit" :loading="loading" class="w-full">确认更换</Button>
        </CardFooter>
      </form>
    </Card>
  </div>
</template>
