<script setup lang="ts">
// 管理后台：单页管理
definePageMeta({ layout: 'admin', middleware: 'admin' })

const api = useApi()
const { data, refresh } = await useAsyncData('admin-pages', () => api.get<any>('/api/v1/admin/pages'))

const pages = computed<any[]>(() => {
  const d = data.value
  return Array.isArray(d) ? d : ((d as any)?.data ?? [])
})

const showCreate = ref(false)
const form = reactive({
  title: '',
  slug: '',
  content: '',
})
const loading = ref(false)
const msg = ref('')

async function create() {
  if (!form.title.trim() || !form.slug.trim()) {
    msg.value = '请输入标题和别名'
    return
  }
  loading.value = true
  msg.value = ''
  try {
    await api.post('/api/v1/admin/pages', form)
    msg.value = '创建成功'
    form.title = ''
    form.slug = ''
    form.content = ''
    showCreate.value = false
    refresh()
  } catch (err: any) {
    msg.value = err?.statusMessage || '创建失败'
  } finally {
    loading.value = false
  }
}

async function remove(id: number) {
  if (!confirm('确定删除该单页？')) return
  await api.del(`/api/v1/admin/pages/${id}`)
  refresh()
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-4">
      <h1 class="text-2xl font-bold text-gray-900">单页管理</h1>
      <AppButton @click="showCreate = !showCreate">{{ showCreate ? '取消' : '新建单页' }}</AppButton>
    </div>

    <div v-if="showCreate" class="mb-6 p-4 bg-white border border-gray-200 rounded-lg space-y-3">
      <div>
        <label class="block text-sm text-gray-700 mb-1">标题</label>
        <input v-model="form.title" class="w-full px-3 py-2 border border-gray-300 rounded-md" />
      </div>
      <div>
        <label class="block text-sm text-gray-700 mb-1">别名（URL 中使用）</label>
        <input v-model="form.slug" class="w-full px-3 py-2 border border-gray-300 rounded-md" placeholder="如 about" />
      </div>
      <div>
        <label class="block text-sm text-gray-700 mb-1">内容</label>
        <textarea v-model="form.content" rows="8" class="w-full px-3 py-2 border border-gray-300 rounded-md"></textarea>
      </div>
      <AppButton :loading="loading" @click="create">创建</AppButton>
      <p v-if="msg" class="text-sm" :class="msg.includes('失败') ? 'text-red-500' : 'text-primary-600'">{{ msg }}</p>
    </div>

    <div v-if="pages.length === 0" class="text-center py-12 text-gray-500">
      暂无单页
    </div>

    <div v-else class="bg-white border border-gray-200 rounded-lg divide-y">
      <div v-for="p in pages" :key="p.id" class="p-4 flex items-center justify-between">
        <div class="flex-1">
          <div class="text-sm font-medium text-gray-900">{{ p.title }}</div>
          <div class="mt-1 text-xs text-gray-500">/page/{{ p.slug }}</div>
        </div>
        <div class="flex gap-2">
          <a :href="`/page/${p.slug}`" target="_blank" class="text-sm text-primary-600 hover:underline">查看</a>
          <button class="text-sm text-red-500" @click="remove(p.id)">删除</button>
        </div>
      </div>
    </div>
  </div>
</template>
