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
    message.success('已保存')
  } catch (err: any) {
    message.error(err?.statusMessage || '保存失败')
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-900 mb-4">个人资料</h1>

    <div class="mb-6 flex flex-wrap gap-2">
      <NuxtLink to="/dashboard/profile" class="px-3 py-1.5 text-sm rounded-md bg-primary-50 text-primary-700">基本信息</NuxtLink>
      <NuxtLink to="/dashboard/profile/email" class="px-3 py-1.5 text-sm rounded-md bg-gray-100 text-gray-700 hover:bg-gray-200">更换邮箱</NuxtLink>
      <NuxtLink to="/dashboard/profile/phone" class="px-3 py-1.5 text-sm rounded-md bg-gray-100 text-gray-700 hover:bg-gray-200">更换手机</NuxtLink>
      <NuxtLink to="/dashboard/profile/password" class="px-3 py-1.5 text-sm rounded-md bg-gray-100 text-gray-700 hover:bg-gray-200">修改密码</NuxtLink>
      <NuxtLink to="/dashboard/profile/social" class="px-3 py-1.5 text-sm rounded-md bg-gray-100 text-gray-700 hover:bg-gray-200">社交账号</NuxtLink>
    </div>

    <div v-if="loading" class="text-center py-12 text-gray-500">加载中...</div>

    <form v-else class="bg-white border border-gray-200 rounded-lg p-6 space-y-4 max-w-2xl" @submit.prevent="save">
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
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
      <AppButton type="submit" :loading="saving">保存</AppButton>
    </form>
  </div>
</template>
