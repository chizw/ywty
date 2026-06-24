<script setup lang="ts">
// 设置中心
definePageMeta({ layout: 'dashboard', middleware: 'auth' })

import { Settings } from '@lucide/vue'

const { user, fetchMe } = useAuth()
const api = useApi()

const form = reactive({
  name: '',
  avatar: '',
  location: '',
  url: '',
  company: '',
  company_title: '',
  tagline: '',
  bio: '',
})

onMounted(async () => {
  try {
    const p: any = await api.get('/api/v1/user/profile')
    Object.assign(form, {
      name: p.name || '',
      avatar: p.avatar || '',
      location: p.location || '',
      url: p.url || '',
      company: p.company || '',
      company_title: p.company_title || '',
      tagline: p.tagline || '',
      bio: p.bio || '',
    })
  } catch { /* noop */ }
})

const saving = ref(false)
const msg = ref('')

async function save() {
  saving.value = true
  msg.value = ''
  try {
    await api.request('/api/v1/user/profile', { method: 'PATCH', body: form })
    await fetchMe()
    msg.value = '已保�?
  } catch (err: any) {
    msg.value = err?.statusMessage || '保存失败'
  } finally {
    saving.value = false
  }
}

const navItems = [
  { to: '/dashboard/settings', label: '个人资料' },
  { to: '/dashboard/change-password', label: '修改密码' },
  { to: '/dashboard/change-email', label: '更换邮箱' },
  { to: '/dashboard/change-phone', label: '更换手机' },
]
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-foreground mb-4">我的云雾图驿</h1>

    <div class="mb-6 flex flex-wrap gap-2">
      <NuxtLink
        v-for="item in navItems"
        :key="item.to"
        :to="item.to"
        class="px-3 py-1.5 text-sm rounded-md"
        :class="item.to === '/dashboard/settings' ? 'bg-primary text-primary-foreground' : 'bg-muted text-muted-foreground hover:bg-muted/80'"
      >{{ item.label }}</NuxtLink>
    </div>

    <Card class="max-w-2xl">
      <form @submit.prevent="save">
        <CardContent class="pt-6 space-y-4">
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <Label>用户�?/Label>
              <Input :model-value="user?.username" disabled class="mt-1 bg-muted" />
            </div>
            <div>
              <Label>邮箱</Label>
              <Input :model-value="user?.email" disabled class="mt-1 bg-muted" />
            </div>
            <div>
              <Label>姓名</Label>
              <Input v-model="form.name" class="mt-1" />
            </div>
            <div>
              <Label>所在地</Label>
              <Input v-model="form.location" class="mt-1" />
            </div>
            <div>
              <Label>个人网站</Label>
              <Input v-model="form.url" class="mt-1" />
            </div>
            <div>
              <Label>公司</Label>
              <Input v-model="form.company" class="mt-1" />
            </div>
            <div>
              <Label>职位</Label>
              <Input v-model="form.company_title" class="mt-1" />
            </div>
            <div>
              <Label>签名</Label>
              <Input v-model="form.tagline" class="mt-1" />
            </div>
          </div>
          <div>
            <Label>个人简�?/Label>
            <Textarea v-model="form.bio" rows="3" class="mt-1" />
          </div>
          <p v-if="msg" class="text-sm" :class="msg.includes('失败') ? 'text-destructive' : 'text-green-600'">{{ msg }}</p>
        </CardContent>
        <CardFooter>
          <Button type="submit" :loading="saving">
            <Settings class="mr-2 h-4 w-4" />
            保存
          </Button>
        </CardFooter>
      </form>
    </Card>
  </div>
</template>
