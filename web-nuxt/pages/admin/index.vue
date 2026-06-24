<script setup lang="ts">
// 管理后台：仪表盘
definePageMeta({ layout: 'admin', middleware: 'admin' })

const api = useApi()
const { data: stats, refresh } = await useAsyncData('admin-stats', () => api.get<any>('/api/v1/admin/stats'))
const s = computed(() => (stats.value as any) ?? {})

const refreshing = ref(false)
async function doRefresh() {
  refreshing.value = true
  try { await refresh() } finally { refreshing.value = false }
}
onMounted(() => {
  document.addEventListener('visibilitychange', onVis)
})
onBeforeUnmount(() => {
  document.removeEventListener('visibilitychange', onVis)
})
function onVis() {
  if (document.visibilityState === 'visible') doRefresh()
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-900">仪表盘</h1>
    <p class="mt-1 text-sm text-gray-500">管理后台概览</p>

    <div class="mt-6 grid grid-cols-2 lg:grid-cols-3 gap-4">
      <div class="bg-white border border-gray-200 rounded-lg p-4">
        <div class="text-xs text-gray-500">用户</div>
        <div class="mt-1 text-2xl font-semibold">{{ s.users ?? '-' }}</div>
      </div>
      <div class="bg-white border border-gray-200 rounded-lg p-4">
        <div class="text-xs text-gray-500">图片</div>
        <div class="mt-1 text-2xl font-semibold">{{ s.photos ?? '-' }}</div>
      </div>
      <div class="bg-white border border-gray-200 rounded-lg p-4">
        <div class="text-xs text-gray-500">相册</div>
        <div class="mt-1 text-2xl font-semibold">{{ s.albums ?? '-' }}</div>
      </div>
      <div class="bg-white border border-gray-200 rounded-lg p-4">
        <div class="text-xs text-gray-500">分享</div>
        <div class="mt-1 text-2xl font-semibold">{{ s.shares ?? '-' }}</div>
      </div>
      <div class="bg-white border border-gray-200 rounded-lg p-4">
        <div class="text-xs text-gray-500">待处理举报</div>
        <div class="mt-1 text-2xl font-semibold text-yellow-600">{{ s.pending_reports ?? 0 }}</div>
      </div>
      <div class="bg-white border border-gray-200 rounded-lg p-4">
        <div class="text-xs text-gray-500">总举报</div>
        <div class="mt-1 text-2xl font-semibold">{{ s.reports ?? 0 }}</div>
      </div>
    </div>

    <div class="mt-6 flex gap-2">
      <AppButton :loading="refreshing" @click="doRefresh">刷新数据</AppButton>
    </div>
  </div>
</template>
