<script setup lang="ts">
// 相册探索详情页
const route = useRoute()
const albumId = route.params.id as string
const api = useApi()
const { t } = useI18n()

const { data: album } = await useAsyncData(`explore-album-${albumId}`, async () => {
  try {
    return await api.get<any>(`/api/v1/albums/${albumId}`)
  } catch {
    return null
  }
})

const { data: photosRaw } = await useAsyncData(`explore-album-photos-${albumId}`, async () => {
  try {
    return await api.get<any>(`/api/v1/albums/${albumId}/photos`)
  } catch {
    return []
  }
})

const photos = computed<any[]>(() => {
  const d = photosRaw.value
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
  title: () => album.value?.name || '相册',
  description: () => album.value?.intro || '浏览公开相册',
  ogTitle: () => album.value?.name || '相册',
  ogDescription: () => album.value?.intro || '浏览公开相册',
  ogType: 'website',
})
</script>

<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <NuxtLink to="/explore" class="text-xs text-muted-foreground hover:text-primary">← {{ t('nav.explore') }}</NuxtLink>

    <template v-if="album">
      <h1 class="mt-2 text-2xl font-bold text-foreground">{{ album.name }}</h1>
      <p v-if="album.intro" class="mt-1 text-sm text-muted-foreground">{{ album.intro }}</p>
      <p class="mt-1 text-xs text-muted-foreground/70">共 {{ photos.length }} 张</p>

      <AppEmpty v-if="photos.length === 0" title="相册内还没有图片" />
      <PhotoMasonry
        v-else
        :photos="photos"
        class="mt-6"
        @click="({ index }) => openLightbox(index)"
      />

      <PhotoLightbox
        v-model:visible="lightboxVisible"
        v-model:index="lightboxIndex"
        :photos="photos"
      />
    </template>
    <AppEmpty v-else title="相册不存在或不可见" />
  </div>
</template>
