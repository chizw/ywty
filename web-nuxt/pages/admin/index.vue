<script setup lang="ts">
// 管理后台：仪表盘
definePageMeta({ layout: 'admin', middleware: 'admin' })

import { Users, Image, Folder, Share, Flag, AlertTriangle, RefreshCw } from '@lucide/vue'

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
    <h1 class="text-2xl font-bold text-foreground">仪表盘</h1>
    <p class="mt-1 text-sm text-muted-foreground">管理后台概览</p>

    <div class="mt-6 grid grid-cols-2 lg:grid-cols-3 gap-4">
      <Card>
        <CardContent class="p-4">
          <div class="flex items-center gap-2 text-xs text-muted-foreground">
            <Users class="h-4 w-4" />
            用户
          </div>
          <div class="mt-1 text-2xl font-semibold">{{ s.users ?? '-' }}</div>
        </CardContent>
      </Card>
      <Card>
        <CardContent class="p-4">
          <div class="flex items-center gap-2 text-xs text-muted-foreground">
            <Image class="h-4 w-4" />
            图片
          </div>
          <div class="mt-1 text-2xl font-semibold">{{ s.photos ?? '-' }}</div>
        </CardContent>
      </Card>
      <Card>
        <CardContent class="p-4">
          <div class="flex items-center gap-2 text-xs text-muted-foreground">
            <Folder class="h-4 w-4" />
            相册
          </div>
          <div class="mt-1 text-2xl font-semibold">{{ s.albums ?? '-' }}</div>
        </CardContent>
      </Card>
      <Card>
        <CardContent class="p-4">
          <div class="flex items-center gap-2 text-xs text-muted-foreground">
            <Share class="h-4 w-4" />
            分享
          </div>
          <div class="mt-1 text-2xl font-semibold">{{ s.shares ?? '-' }}</div>
        </CardContent>
      </Card>
      <Card>
        <CardContent class="p-4">
          <div class="flex items-center gap-2 text-xs text-muted-foreground">
            <AlertTriangle class="h-4 w-4" />
            待处理举报          </div>
          <div class="mt-1 text-2xl font-semibold text-yellow-600">{{ s.pending_reports ?? 0 }}</div>
        </CardContent>
      </Card>
      <Card>
        <CardContent class="p-4">
          <div class="flex items-center gap-2 text-xs text-muted-foreground">
            <Flag class="h-4 w-4" />
            总举报          </div>
          <div class="mt-1 text-2xl font-semibold">{{ s.reports ?? 0 }}</div>
        </CardContent>
      </Card>
    </div>

    <div class="mt-6 flex gap-2">
      <Button :loading="refreshing" @click="doRefresh">
        <RefreshCw class="h-4 w-4 mr-2" />
        刷新数据
      </Button>
    </div>
  </div>
</template>
