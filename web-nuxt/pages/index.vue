<script setup lang="ts">
// 首页：展示最�?热门图片
const { t } = useI18n()
const api = useApi()

// 获取最新公开图片（无需鉴权�?const { data: latestData } = await useAsyncData('home-latest', async () => {
  try {
    return await api.get<any>('/api/v1/public/photos', { query: { page: 1, per_page: 12 } })
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
  description: '自托管图�?/ 云相�?· 重构�?,
  ogTitle: '云雾图驿 · 自托管图�?/ 云相�?,
  ogDescription: '自托管图�?/ 云相�?· 重构�?,
  ogType: 'website',
})
</script>

<template>
  <div
    :class="[
      'max-w-7xl mx-auto px-4 sm:px-6 lg:px-8',
      latestPhotos.length === 0
        ? 'min-h-[calc(100vh-12rem)] flex items-center justify-center'
        : 'py-12',
    ]"
  >
    <!-- Hero 区域 -->
    <div class="text-center w-full">
      <h1 class="text-5xl font-bold text-foreground">云雾图驿</h1>
      <p class="mt-3 text-lg text-muted-foreground">自托管图�?/ 云相�?· 重构�?/p>
      <div class="mt-8 flex items-center justify-center gap-4">
        <Button as="NuxtLink" to="/explore" size="lg">
          {{ t('nav.explore') }}
        </Button>
        <Button as="NuxtLink" to="/auth/login" variant="outline" size="lg">
          {{ t('nav.login') }}
        </Button>
      </div>
    </div>

    <!-- 最新图�?-->
    <section v-if="latestPhotos.length > 0" class="mt-16">
      <div class="flex items-center justify-between mb-6">
        <h2 class="text-2xl font-bold text-foreground">最新图�?/h2>
        <NuxtLink to="/explore" class="text-sm text-primary hover:underline">
          查看全部 �?        </NuxtLink>
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
