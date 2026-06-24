<script setup lang="ts">
definePageMeta({ layout: 'dashboard', middleware: 'auth' })

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

onMounted(async () => {
  await fetchMe()
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
        <h1 class="text-2xl font-bold text-gray-900">
          欢迎，{{ user?.name || user?.username }}
        </h1>
        <p class="mt-1 text-sm text-gray-500">
          这里是你的控制台<span v-if="updatedLabel"> · 上次更新 {{ updatedLabel }}</span>
        </p>
      </div>
      <AppButton size="sm" :loading="statsStore.loading" @click="statsStore.refresh()">刷新</AppButton>
    </div>

    <div class="mt-6 grid grid-cols-1 sm:grid-cols-3 gap-4">
      <div class="bg-white border border-gray-200 rounded-lg p-4">
        <div class="text-xs text-gray-500">图片总数</div>
        <div class="mt-1 text-2xl font-semibold">{{ photos }}</div>
      </div>
      <div class="bg-white border border-gray-200 rounded-lg p-4">
        <div class="text-xs text-gray-500">我的相册</div>
        <div class="mt-1 text-2xl font-semibold">{{ albums }}</div>
      </div>
      <div class="bg-white border border-gray-200 rounded-lg p-4">
        <div class="text-xs text-gray-500">已用容量</div>
        <div class="mt-1 text-2xl font-semibold">{{ usedLabel }}</div>
      </div>
    </div>
  </div>
</template>
