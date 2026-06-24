<script setup lang="ts">
// 管理后台：意见反馈管理
definePageMeta({ layout: 'admin', middleware: 'admin' })

const api = useApi()
const { data, refresh } = await useAsyncData('admin-feedbacks', () => api.get<any>('/api/v1/admin/feedbacks'))

const feedbacks = computed<any[]>(() => {
  const d = data.value
  return Array.isArray(d) ? d : ((d as any)?.data ?? [])
})

async function remove(id: number) {
  if (!confirm('确定删除该反馈？')) return
  await api.del(`/api/v1/admin/feedbacks/${id}`)
  refresh()
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-900 mb-4">意见反馈</h1>

    <div v-if="feedbacks.length === 0" class="text-center py-12 text-gray-500">
      暂无反馈
    </div>

    <div v-else class="bg-white border border-gray-200 rounded-lg divide-y">
      <div v-for="f in feedbacks" :key="f.id" class="p-4">
        <div class="flex items-start justify-between">
          <div class="flex-1">
            <div class="text-sm font-medium text-gray-900">{{ f.type }}</div>
            <div class="mt-2 text-sm text-gray-600 whitespace-pre-line">{{ f.content }}</div>
            <div class="mt-2 text-xs text-gray-500">
              {{ f.email || f.phone || '匿名用户' }} · {{ new Date(f.created_at).toLocaleString() }}
            </div>
          </div>
          <button class="ml-4 text-red-500 text-sm" @click="remove(f.id)">删除</button>
        </div>
      </div>
    </div>
  </div>
</template>
