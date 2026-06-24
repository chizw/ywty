<script setup lang="ts">
// 个人资料 - 基本信息
definePageMeta({ layout: 'dashboard', middleware: 'auth' })

const api = useApi()
const message = useMessage()

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

const loading = ref(true)
const saving = ref(false)

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
  } catch (err: any) {
    message.error(err?.statusMessage || '加载失败')
  } finally {
    loading.value = false
  }
})

async function save() {
  saving.value = true
  try {
    await api.request('/api/v1/user/profile', { method: 'PATCH', body: form })
    message.success('已保�?)
  } catch (err: any) {
    message.error(err?.statusMessage || '保存失败')
  } finally {
    saving.value = false
  }
}

const navItems = [
  { to: '/dashboard/profile', label: '基本信息', active: true },
  { to: '/dashboard/profile/email', label: '更换邮箱', active: false },
  { to: '/dashboard/profile/phone', label: '更换手机', active: false },
  { to: '/dashboard/profile/password', label: '修改密码', active: false },
  { to: '/dashboard/profile/social', label: '社交账号', active: false },
]
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-foreground mb-4">个人资料</h1>

    <div class="mb-6 flex flex-wrap gap-2">
      <NuxtLink
        v-for="item in navItems"
        :key="item.to"
        :to="item.to"
        class="px-3 py-1.5 text-sm rounded-md"
        :class="item.to === '/dashboard/profile' ? 'bg-primary text-primary-foreground' : 'bg-muted text-muted-foreground hover:bg-muted/80'"
      >{{ item.label }}</NuxtLink>
    </div>

    <Skeleton v-if="loading" class="h-64 w-full max-w-2xl" />

    <Card v-else class="max-w-2xl">
      <form @submit.prevent="save">
        <CardContent class="pt-6 space-y-4">
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
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
        </CardContent>
        <CardFooter>
          <Button type="submit" :loading="saving">保存</Button>
        </CardFooter>
      </form>
    </Card>
  </div>
</template>
