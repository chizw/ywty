<script setup lang="ts">
// 首页：展示最新/热门图片
const { t } = useI18n()
const api = useApi()

// 获取最新图片
const { data: latestData } = await useAsyncData('home-latest', async () => {
  try {
    return await api.get<any[]>('/api/v1/photos', { query: { page: 1, per_page: 12, sort: 'created_at', order: 'desc' } })
  } catch {
    return []
  }
})

const latestPhotos = computed<any[]>(() => {
  const d = latestData.value
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
  title: t('nav.home'),
  description: '自托管图床 / 云相册 · 重构版',
  ogTitle: '云雾图驿 · 自托管图床 / 云相册',
  ogDescription: '自托管图床 / 云相册 · 重构版',
  ogType: 'website',
})
</script>

<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
    <!-- Hero 区域 -->
    <div class="text-center mb-12">
      <h1 class="text-5xl font-bold text-gray-900">云雾图驿</h1>
      <p class="mt-3 text-lg text-gray-600">自托管图床 / 云相册 · 重构版</p>
      <div class="mt-8 flex items-center justify-center gap-4">
        <NuxtLink
          to="/explore"
          class="px-6 py-2.5 bg-primary-600 text-white rounded-md hover:bg-primary-700"
        >
          {{ t('nav.explore') }}
        </NuxtLink>
        <NuxtLink
          to="/auth/login"
          class="px-6 py-2.5 border border-gray-300 text-gray-700 rounded-md hover:bg-gray-50"
        >
          {{ t('nav.login') }}
        </NuxtLink>
      </div>
    </div>

    <!-- 最新图片 -->
    <section v-if="latestPhotos.length > 0" class="mt-16">
      <div class="flex items-center justify-between mb-6">
        <h2 class="text-2xl font-bold text-gray-900">最新图片</h2>
        <NuxtLink to="/explore" class="text-sm text-primary-600 hover:underline">
          查看全部 →
        </NuxtLink>
      </div>
      <PhotoMasonry
        :photos="latestPhotos"
        @click="({ index }) => openLightbox(index)"
      />
      <PhotoLightbox
        v-model:visible="lightboxVisible"
        v-model:index="lightboxIndex"
        :photos="latestPhotos"
      />
    </section>
  </div>
</template>
