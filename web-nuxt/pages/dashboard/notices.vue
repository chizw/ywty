<script setup lang="ts">
// 用户中心：我的通知
definePageMeta({ layout: 'dashboard', middleware: 'auth' })

const api = useApi()
const { data, refresh } = await useAsyncData('my-notices', () => api.get<any>('/api/v1/notices'))

const notices = computed<any[]>(() => {
  const d = data.value
  return Array.isArray(d) ? d : ((d as any)?.data ?? [])
})

async function markAsRead(id: number) {
  await api.put(`/api/v1/notices/${id}/read`)
  refresh()
}

async function markAllAsRead() {
  await api.put('/api/v1/notices/read-all')
  refresh()
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-4">
      <h1 class="text-2xl font-bold text-gray-900">我的通知</h1>
      <AppButton @click="markAllAsRead">全部已读</AppButton>
    </div>

    <div v-if="notices.length === 0" class="text-center py-12 text-gray-500">
      暂无通知
    </div>

    <div v-else class="bg-white border border-gray-200 rounded-lg divide-y">
      <div v-for="n in notices" :key="n.id" class="p-4 flex items-start gap-3" :class="{ 'bg-blue-50': !n.read_at }">
        <div class="flex-1">
          <div class="text-sm font-medium text-gray-900">{{ n.title }}</div>
          <div class="mt-1 text-sm text-gray-600 whitespace-pre-line">{{ n.content }}</div>
          <div class="mt-2 text-xs text-gray-500">{{ new Date(n.created_at).toLocaleString() }}</div>
        </div>
        <button v-if="!n.read_at" class="text-sm text-primary-600 hover:underline" @click="markAsRead(n.id)">
          标记已读
        </button>
      </div>
    </div>
  </div>
</template>
