<script setup lang="ts">
// 修改密码
definePageMeta({ layout: 'dashboard', middleware: 'auth' })

import { Lock } from '@lucide/vue'

const api = useApi()
const form = reactive({ old_password: '', new_password: '' })
const loading = ref(false)
const msg = ref('')

async function submit() {
  if (form.new_password.length < 6) {
    msg.value = '新密码至少6位'
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
    <h1 class="text-2xl font-bold text-foreground mb-4">修改密码</h1>
    <Card class="max-w-md">
      <form @submit.prevent="submit">
        <CardContent class="pt-6 space-y-4">
          <div>
            <Label>原密码</Label>
            <Input v-model="form.old_password" type="password" required class="mt-1" />
          </div>
          <div>
            <Label>新密码</Label>
            <Input v-model="form.new_password" type="password" required minlength="6" class="mt-1" />
          </div>
          <p v-if="msg" class="text-sm" :class="msg.includes('成功') ? 'text-green-600' : 'text-destructive'">{{ msg }}</p>
        </CardContent>
        <CardFooter>
          <Button type="submit" :loading="loading" class="w-full">
            <Lock class="mr-2 h-4 w-4" />
            修改密码
          </Button>
        </CardFooter>
      </form>
    </Card>
  </div>
</template>
