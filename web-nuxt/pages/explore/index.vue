<script setup lang="ts">
// 公开探索页：浏览公开图片，支持点赞/举报
const api = useApi()
const { t } = useI18n()

const { data } = await useAsyncData('explore', async () => {
  try {
    return await api.get<any[]>('/api/v1/photos', { query: { page: 1, per_page: 24 } })
  } catch {
    return []
  }
})

const items = computed<any[]>(() => {
  const d = data.value
  if (Array.isArray(d)) return d
  if (d && Array.isArray((d as any).data)) return (d as any).data
  return []
})

// Lightbox
const lightboxVisible = ref(false)
const lightboxIndex = ref(0)
function openLightbox(idx: number) {
  lightboxIndex.value = idx
  lightboxVisible.value = true
}

useSeoMeta({
  title: t('nav.explore'),
  description: '浏览公开图片与相册',
})
</script>

<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <h1 class="text-2xl font-bold text-gray-900">{{ t('nav.explore') }}</h1>
    <p class="mt-1 text-sm text-gray-500">浏览公开图片</p>

    <AppEmpty v-if="items.length === 0" title="暂无公开图片" />
    <PhotoMasonry
      v-else
      :photos="items"
      class="mt-6"
      @click="({ index }) => openLightbox(index)"
    />

    <PhotoLightbox
      v-model:visible="lightboxVisible"
      v-model:index="lightboxIndex"
      :photos="items"
    />
  </div>
</template>
