<script setup lang="ts">
// 个人资料 - 修改密码
definePageMeta({ layout: 'dashboard', middleware: 'auth' })

import { Lock } from '@lucide/vue'

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
    <h1 class="text-2xl font-bold text-foreground mb-4">修改密码</h1>

    <div class="mb-6 flex flex-wrap gap-2">
      <NuxtLink
        v-for="item in navItems"
        :key="item.to"
        :to="item.to"
        class="px-3 py-1.5 text-sm rounded-md"
        :class="item.to === '/dashboard/profile/password' ? 'bg-primary text-primary-foreground' : 'bg-muted text-muted-foreground hover:bg-muted/80'"
      >{{ item.label }}</NuxtLink>
    </div>

    <Card class="max-w-2xl">
      <CardContent class="pt-6 space-y-4">
        <div>
          <Label>当前密码</Label>
          <Input v-model="form.old_password" type="password" placeholder="请输入当前密码" class="mt-1" />
        </div>

        <div>
          <Label>新密码</Label>
          <Input v-model="form.password" type="password" placeholder="请输入新密码（至少 6 位）" class="mt-1" />
        </div>

        <div>
          <Label>确认新密码</Label>
          <Input v-model="form.password_confirmation" type="password" placeholder="请再次输入新密码" class="mt-1" />
        </div>
      </CardContent>
      <CardFooter>
        <Button :loading="changing" @click="changePassword">
          <Lock class="mr-2 h-4 w-4" />
          修改密码
        </Button>
      </CardFooter>
    </Card>
  </div>
</template>
