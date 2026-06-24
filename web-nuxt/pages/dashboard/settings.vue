<script setup lang="ts">
// 设置中心
definePageMeta({ layout: 'dashboard', middleware: 'auth' })

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
    msg.value = '已保存'
  } catch (err: any) {
    msg.value = err?.statusMessage || '保存失败'
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-900 mb-4">我的云雾图驿</h1>

    <div class="mb-6 flex flex-wrap gap-2">
      <NuxtLink to="/dashboard/settings" class="px-3 py-1.5 text-sm rounded-md bg-primary-50 text-primary-700">个人资料</NuxtLink>
      <NuxtLink to="/dashboard/change-password" class="px-3 py-1.5 text-sm rounded-md bg-gray-100 text-gray-700 hover:bg-gray-200">修改密码</NuxtLink>
      <NuxtLink to="/dashboard/change-email" class="px-3 py-1.5 text-sm rounded-md bg-gray-100 text-gray-700 hover:bg-gray-200">更换邮箱</NuxtLink>
      <NuxtLink to="/dashboard/change-phone" class="px-3 py-1.5 text-sm rounded-md bg-gray-100 text-gray-700 hover:bg-gray-200">更换手机</NuxtLink>
    </div>

    <form class="bg-white border border-gray-200 rounded-lg p-6 space-y-4 max-w-2xl" @submit.prevent="save">
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <div>
          <label class="block text-sm text-gray-700 mb-1">用户名</label>
          <input :value="user?.username" disabled class="w-full px-3 py-2 border border-gray-300 rounded-md bg-gray-50" />
        </div>
        <div>
          <label class="block text-sm text-gray-700 mb-1">邮箱</label>
          <input :value="user?.email" disabled class="w-full px-3 py-2 border border-gray-300 rounded-md bg-gray-50" />
        </div>
        <div>
          <label class="block text-sm text-gray-700 mb-1">姓名</label>
          <input v-model="form.name" class="w-full px-3 py-2 border border-gray-300 rounded-md" />
        </div>
        <div>
          <label class="block text-sm text-gray-700 mb-1">所在地</label>
          <input v-model="form.location" class="w-full px-3 py-2 border border-gray-300 rounded-md" />
        </div>
        <div>
          <label class="block text-sm text-gray-700 mb-1">个人网站</label>
          <input v-model="form.url" class="w-full px-3 py-2 border border-gray-300 rounded-md" />
        </div>
        <div>
          <label class="block text-sm text-gray-700 mb-1">公司</label>
          <input v-model="form.company" class="w-full px-3 py-2 border border-gray-300 rounded-md" />
        </div>
        <div>
          <label class="block text-sm text-gray-700 mb-1">职位</label>
          <input v-model="form.company_title" class="w-full px-3 py-2 border border-gray-300 rounded-md" />
        </div>
        <div>
          <label class="block text-sm text-gray-700 mb-1">签名</label>
          <input v-model="form.tagline" class="w-full px-3 py-2 border border-gray-300 rounded-md" />
        </div>
      </div>
      <div>
        <label class="block text-sm text-gray-700 mb-1">个人简介</label>
        <textarea v-model="form.bio" rows="3" class="w-full px-3 py-2 border border-gray-300 rounded-md" />
      </div>
      <p v-if="msg" class="text-sm" :class="msg.includes('失败') ? 'text-red-500' : 'text-primary-600'">{{ msg }}</p>
      <AppButton type="submit" :loading="saving">保存</AppButton>
    </form>
  </div>
</template>
