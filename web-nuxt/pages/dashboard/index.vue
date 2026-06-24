<script setup lang="ts">
definePageMeta({ layout: 'dashboard', middleware: 'auth' })

import { RefreshCw, Image, FolderOpen, HardDrive } from '@lucide/vue'

const { user, fetchMe } = useAuth()
const statsStore = useStatsStore()

const photos = computed(() => `${statsStore.photos} 张`)
const albums = computed(() => `${statsStore.albums} 个`)
const usedLabel = computed(() => {
  const b = statsStore.usedBytes || 0
  if (b < 1024) return `${b} B`
  if (b < 1024 * 1024) return `${(b / 1024).toFixed(2)} KB`
  if (b < 1024 * 1024 * 1024) return `${(b / 1024 / 1024).toFixed(2)} MB`
  return `${(b / 1024 / 1024 / 1024).toFixed(2)} GB`
})
const updatedLabel = computed(() => {
  if (!statsStore.lastUpdatedAt) return ''
  return new Date(statsStore.lastUpdatedAt).toLocaleTimeString('zh-CN', { hour12: false })
})

onMounted(() => {
  // fetchMe �?statsStore.refresh 并行执行，互不阻�?  fetchMe().catch(() => {})
  statsStore.refresh()
  document.addEventListener('visibilitychange', onVisibility)
})
onBeforeUnmount(() => {
  document.removeEventListener('visibilitychange', onVisibility)
})
function onVisibility() {
  if (document.visibilityState === 'visible') statsStore.refresh()
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-foreground">
          欢迎，{{ user?.name || user?.username }}
        </h1>
        <p class="mt-1 text-sm text-muted-foreground">
          这里是你的控制台<span v-if="updatedLabel"> · 上次更新 {{ updatedLabel }}</span>
        </p>
      </div>
      <Button size="sm" :loading="statsStore.loading" @click="statsStore.refresh()">
        <RefreshCw class="mr-2 h-4 w-4" :class="{ 'animate-spin': statsStore.loading }" />
        刷新
      </Button>
    </div>

    <div class="mt-6 grid grid-cols-1 sm:grid-cols-3 gap-4">
      <Card>
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium text-muted-foreground">图片总数</CardTitle>
          <Image class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-semibold">{{ photos }}</div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium text-muted-foreground">我的相册</CardTitle>
          <FolderOpen class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-semibold">{{ albums }}</div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium text-muted-foreground">已用容量</CardTitle>
          <HardDrive class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-semibold">{{ usedLabel }}</div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>
