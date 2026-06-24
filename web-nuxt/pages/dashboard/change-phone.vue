<script setup lang="ts">
// 更换手机�?definePageMeta({ layout: 'dashboard', middleware: 'auth' })

import { Phone } from '@lucide/vue'

const api = useApi()
const { user, fetchMe } = useAuth()

const form = reactive({ new_phone: '', code: '' })
const sending = ref(false)
const countdown = ref(0)
const loading = ref(false)
const msg = ref('')

async function sendCode() {
  if (!form.new_phone) {
    msg.value = '请填写新手机�?
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
    msg.value = err?.statusMessage || '发送失�?
  } finally {
    sending.value = false
  }
}

async function submit() {
  if (!form.new_phone || !form.code) {
    msg.value = '请填写完�?
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
    <h1 class="text-2xl font-bold text-foreground mb-4">更换手机�?/h1>
    <p class="text-sm text-muted-foreground mb-4">当前手机：{{ (user as any)?.phone || '未绑�? }}</p>
    <Card class="max-w-md">
      <form @submit.prevent="submit">
        <CardContent class="pt-6 space-y-4">
          <div>
            <Label>新手机号</Label>
            <Input v-model="form.new_phone" required class="mt-1" />
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
                <Phone class="mr-1 h-4 w-4" />
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
