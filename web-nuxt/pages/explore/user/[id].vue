<script setup lang="ts">
// 用户公开主页
const route = useRoute()
const userId = route.params.id as string
const api = useApi()
const { t } = useI18n()

const { data: user } = await useAsyncData(`explore-user-${userId}`, async () => {
  try {
    return await api.get<any>(`/api/v1/users/${userId}`)
  } catch {
    return null
  }
})

const { data: photosRaw } = await useAsyncData(`explore-user-photos-${userId}`, async () => {
  try {
    return await api.get<any>('/api/v1/photos', { query: { user_id: userId, is_public: 1 } })
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
  title: () => user.value?.name || user.value?.username || '用户主页',
  description: () => `查看 ${user.value?.name || user.value?.username || ''} 的公开图片`,
  ogTitle: () => user.value?.name || '用户主页',
  ogType: 'profile',
})
</script>

<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <NuxtLink to="/explore" class="text-xs text-gray-500 hover:text-primary-600">← {{ t('nav.explore') }}</NuxtLink>

    <template v-if="user">
      <div class="mt-4 flex items-center gap-4">
        <img
          v-if="user.avatar"
          :src="user.avatar"
          :alt="user.name || user.username"
          class="w-16 h-16 rounded-full object-cover bg-gray-100"
        />
        <div
          v-else
          class="w-16 h-16 rounded-full bg-primary-100 text-primary-600 flex items-center justify-center text-xl font-semibold"
        >
          {{ (user.name || user.username || '?').charAt(0).toUpperCase() }}
        </div>
        <div>
          <h1 class="text-2xl font-bold text-gray-900">{{ user.name || user.username }}</h1>
          <p v-if="user.intro" class="mt-1 text-sm text-gray-500">{{ user.intro }}</p>
          <p class="mt-1 text-xs text-gray-400">{{ photos.length }} 张公开图片</p>
        </div>
      </div>

      <AppEmpty v-if="photos.length === 0" title="该用户暂无公开图片" />
      <PhotoMasonry
        v-else
        :photos="photos"
        class="mt-8"
        @click="({ index }) => openLightbox(index)"
      />

      <PhotoLightbox
        v-model:visible="lightboxVisible"
        v-model:index="lightboxIndex"
        :photos="photos"
      />
    </template>
    <AppEmpty v-else title="用户不存在" />
  </div>
</template>
