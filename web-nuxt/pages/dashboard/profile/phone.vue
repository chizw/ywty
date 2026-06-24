<script setup lang="ts">
// 个人资料 - 更换手机
definePageMeta({ layout: 'dashboard', middleware: 'auth' })

import { Phone } from '@lucide/vue'

const api = useApi()
const message = useMessage()
const { user, fetchMe } = useAuth()

const form = reactive({
  phone: '',
  code: '',
})

const sending = ref(false)
const changing = ref(false)
const countdown = ref(0)

async function sendCode() {
  if (!form.phone) {
    message.error('请先输入新手机号')
    return
  }
  if (countdown.value > 0) return

  sending.value = true
  try {
    await api.post('/api/v1/verify-codes', {
      channel: 'sms',
      account: form.phone,
      event: 'change_phone',
    })
    message.success('验证码已发送')
    countdown.value = 60
    const t = setInterval(() => {
      countdown.value--
      if (countdown.value <= 0) clearInterval(t)
    }, 1000)
  } catch (err: any) {
    message.error(err?.statusMessage || '发送失�?)
  } finally {
    sending.value = false
  }
}

async function changePhone() {
  if (!form.phone || !form.code) {
    message.error('请填写完�?)
    return
  }

  changing.value = true
  try {
    await api.post('/api/v1/user/change-phone', {
      phone: form.phone,
      code: form.code,
    })
    await fetchMe()
    message.success('手机号已更换')
    form.phone = ''
    form.code = ''
  } catch (err: any) {
    message.error(err?.statusMessage || '更换失败')
  } finally {
    changing.value = false
  }
}

const navItems = [
  { to: '/dashboard/profile', label: '基本信息' },
  { to: '/dashboard/profile/email', label: '更换邮箱' },
  { to: '/dashboard/profile/phone', label: '更换手机' },
  { to: '/dashboard/profile/password', label: '修改密码' },
  { to: '/dashboard/profile/social', label: '社交账号' },
]
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-foreground mb-4">更换手机</h1>

    <div class="mb-6 flex flex-wrap gap-2">
      <NuxtLink
        v-for="item in navItems"
        :key="item.to"
        :to="item.to"
        class="px-3 py-1.5 text-sm rounded-md"
        :class="item.to === '/dashboard/profile/phone' ? 'bg-primary text-primary-foreground' : 'bg-muted text-muted-foreground hover:bg-muted/80'"
      >{{ item.label }}</NuxtLink>
    </div>

    <Card class="max-w-2xl">
      <CardContent class="pt-6 space-y-4">
        <div>
          <Label>当前手机�?/Label>
          <Input :model-value="user?.phone" disabled class="mt-1 bg-muted" />
        </div>

        <div>
          <Label>新手机号</Label>
          <Input v-model="form.phone" type="tel" placeholder="请输入新手机�? class="mt-1" />
        </div>

        <div>
          <Label>验证�?/Label>
          <div class="flex gap-2 mt-1">
            <Input v-model="form.code" type="text" maxlength="6" placeholder="请输入验证码" class="flex-1" />
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
      </CardContent>
      <CardFooter>
        <Button :loading="changing" @click="changePhone">更换手机</Button>
      </CardFooter>
    </Card>
  </div>
</template>
