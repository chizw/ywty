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
    <NuxtLink to="/explore" class="text-xs text-muted-foreground hover:text-primary">← {{ t('nav.explore') }}</NuxtLink>

    <template v-if="user">
      <div class="mt-4 flex items-center gap-4">
        <Avatar class="h-16 w-16">
          <AvatarImage
            v-if="user.avatar"
            :src="user.avatar"
            :alt="user.name || user.username"
          />
          <AvatarFallback class="text-xl font-semibold bg-primary/10 text-primary">
            {{ (user.name || user.username || '?').charAt(0).toUpperCase() }}
          </AvatarFallback>
        </Avatar>
        <div>
          <h1 class="text-2xl font-bold text-foreground">{{ user.name || user.username }}</h1>
          <p v-if="user.intro" class="mt-1 text-sm text-muted-foreground">{{ user.intro }}</p>
          <Badge variant="secondary" class="mt-1">{{ photos.length }} 张公开图片</Badge>
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
