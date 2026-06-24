<script setup lang="ts">
// 探索页 - 相册列表
const api = useApi()
const { t } = useI18n()

const { data } = await useAsyncData('explore-albums', async () => {
  try {
    return await api.get<any[]>('/api/v1/albums', { query: { page: 1, per_page: 20, is_public: 1 } })
  } catch {
    return []
  }
})

const albums = computed<any[]>(() => {
  const d = data.value
  if (Array.isArray(d)) return d
  if (d && Array.isArray((d as any).data)) return (d as any).data
  return []
})

useSeoMeta({
  title: t('nav.explore') + ' - ' + t('photo.albums'),
  description: '浏览公开相册',
})
</script>

<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <NuxtLink to="/explore" class="text-xs text-gray-500 hover:text-primary-600">← {{ t('nav.explore') }}</NuxtLink>

    <h1 class="mt-2 text-2xl font-bold text-gray-900">{{ t('photo.albums') }}</h1>
    <p class="mt-1 text-sm text-gray-500">浏览公开相册</p>

    <AppEmpty v-if="albums.length === 0" title="暂无公开相册" />

    <div v-else class="mt-6 grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
      <NuxtLink
        v-for="album in albums"
        :key="album.id"
        :to="`/explore/album/${album.id}`"
        class="group block bg-white border border-gray-200 rounded-lg overflow-hidden hover:shadow-md transition"
      >
        <div class="aspect-square bg-gray-100 relative overflow-hidden">
          <img
            v-if="album.cover"
            :src="`/uploads/${album.cover}`"
            :alt="album.name"
            class="w-full h-full object-cover group-hover:scale-105 transition"
            loading="lazy"
          />
          <div v-else class="w-full h-full flex items-center justify-center text-gray-400">
            <svg class="w-12 h-12" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
          </div>
        </div>
        <div class="p-3">
          <h3 class="text-sm font-medium text-gray-900 truncate">{{ album.name }}</h3>
          <p v-if="album.intro" class="mt-1 text-xs text-gray-500 truncate">{{ album.intro }}</p>
          <p class="mt-2 text-xs text-gray-400">{{ album.photos_count || 0 }} {{ t('album.photos') }}</p>
        </div>
      </NuxtLink>
    </div>
  </div>
</template>
