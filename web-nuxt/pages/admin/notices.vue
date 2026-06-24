<script setup lang="ts">
// 管理后台：通知管理
definePageMeta({ layout: 'admin', middleware: 'admin' })

const api = useApi()
const { data, refresh } = await useAsyncData('admin-notices', () => api.get<any>('/api/v1/admin/notices'))

const notices = computed<any[]>(() => {
  const d = data.value
  return Array.isArray(d) ? d : ((d as any)?.data ?? [])
})

const showCreate = ref(false)
const form = reactive({
  title: '',
  content: '',
})
const loading = ref(false)
const msg = ref('')

async function create() {
  if (!form.title.trim()) {
    msg.value = '请输入标题'
    return
  }
  loading.value = true
  msg.value = ''
  try {
    await api.post('/api/v1/admin/notices', form)
    msg.value = '创建成功'
    form.title = ''
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
  if (!confirm('确定删除该通知？')) return
  await api.del(`/api/v1/admin/notices/${id}`)
  refresh()
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-4">
      <h1 class="text-2xl font-bold text-gray-900">通知管理</h1>
      <AppButton @click="showCreate = !showCreate">{{ showCreate ? '取消' : '新建通知' }}</AppButton>
    </div>

    <div v-if="showCreate" class="mb-6 p-4 bg-white border border-gray-200 rounded-lg space-y-3">
      <div>
        <label class="block text-sm text-gray-700 mb-1">标题</label>
        <input v-model="form.title" class="w-full px-3 py-2 border border-gray-300 rounded-md" />
      </div>
      <div>
        <label class="block text-sm text-gray-700 mb-1">内容</label>
        <textarea v-model="form.content" rows="5" class="w-full px-3 py-2 border border-gray-300 rounded-md"></textarea>
      </div>
      <AppButton :loading="loading" @click="create">创建</AppButton>
      <p v-if="msg" class="text-sm" :class="msg.includes('失败') ? 'text-red-500' : 'text-primary-600'">{{ msg }}</p>
    </div>

    <div v-if="notices.length === 0" class="text-center py-12 text-gray-500">
      暂无通知
    </div>

    <div v-else class="bg-white border border-gray-200 rounded-lg divide-y">
      <div v-for="n in notices" :key="n.id" class="p-4 flex items-start justify-between">
        <div class="flex-1">
          <div class="text-sm font-medium text-gray-900">{{ n.title }}</div>
          <div class="mt-1 text-sm text-gray-600 whitespace-pre-line">{{ n.content }}</div>
          <div class="mt-2 text-xs text-gray-500">{{ new Date(n.created_at).toLocaleString() }}</div>
        </div>
        <button class="ml-4 text-red-500 text-sm" @click="remove(n.id)">删除</button>
      </div>
    </div>
  </div>
</template>
