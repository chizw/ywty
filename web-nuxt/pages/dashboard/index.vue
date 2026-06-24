<script setup lang="ts">
definePageMeta({ layout: 'dashboard', middleware: 'auth' })

const { user, fetchMe } = useAuth()
const api = useApi()

const stats = reactive({ photos: 0, albums: 0, usedKb: 0 })

onMounted(async () => {
  try {
    await fetchMe()
    const list = await api.get<any[]>('/api/v1/photos', { query: { page: 1, per_page: 1 } })
    stats.photos = (list as any)?.length || 0
  } catch {
    // ignore
  }
})
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-900">
      欢迎，{{ user?.name || user?.username }}
    </h1>
    <p class="mt-1 text-sm text-gray-500">这里是你的控制台</p>

    <div class="mt-6 grid grid-cols-1 sm:grid-cols-3 gap-4">
      <div class="bg-white border border-gray-200 rounded-lg p-4">
        <div class="text-xs text-gray-500">图片总数</div>
        <div class="mt-1 text-2xl font-semibold">{{ stats.photos }}</div>
      </div>
      <div class="bg-white border border-gray-200 rounded-lg p-4">
        <div class="text-xs text-gray-500">相册</div>
        <div class="mt-1 text-2xl font-semibold">{{ stats.albums }}</div>
      </div>
      <div class="bg-white border border-gray-200 rounded-lg p-4">
        <div class="text-xs text-gray-500">已用容量</div>
        <div class="mt-1 text-2xl font-semibold">{{ stats.usedKb }} KB</div>
      </div>
    </div>
  </div>
</template>
