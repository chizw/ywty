<script setup lang="ts">
// 首页「接触印相」：一排带发丝线的图片底片，点击开灯箱
// 静态部署：客户端拉取公开图片
import { onMounted, ref } from 'vue'
import LazyImage from '../photo/LazyImage.vue'
import PhotoLightbox from '../photo/PhotoLightbox.vue'
import { useApi } from '../../../lib/api'
import type { PublicPhoto } from '../../../lib/types'

const photos = ref<PublicPhoto[]>([])
const loaded = ref(false)

onMounted(async () => {
  try {
    const res = await useApi().get<PublicPhoto[]>('/api/v1/public/photos', {
      query: { page: 1, per_page: 12 },
    })
    photos.value = Array.isArray(res) ? res : []
  } catch {
    photos.value = []
  } finally {
    loaded.value = true
  }
})

const lightboxVisible = ref(false)
const lightboxIndex = ref(0)

function open(i: number) {
  lightboxIndex.value = i
  lightboxVisible.value = true
}
</script>

<template>
  <div>
    <!-- 加载中 -->
    <div v-if="!loaded" class="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-6 lg:gap-4">
      <div v-for="i in 6" :key="i" class="skeleton aspect-[4/3] rounded-[0.35rem]" />
    </div>

    <!-- 空态 -->
    <div v-else-if="photos.length === 0" class="rounded-[0.35rem] border hairline py-16 text-center">
      <p class="text-sm text-muted-foreground">这里还空着。</p>
      <a href="/auth/register" class="mt-3 inline-block text-sm text-brand hover:underline">
        创建账号，放上第一张图片 →
      </a>
    </div>

    <!-- 底片条 -->
    <div v-else class="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-6 lg:gap-4">
      <button
        v-for="(p, i) in photos"
        :key="p.id"
        class="group block cursor-zoom-in text-left"
        @click="open(i)"
      >
        <div class="overflow-hidden rounded-[0.35rem] border hairline bg-card">
          <LazyImage
            :src="p.url"
            :alt="p.name || p.username || ''"
            :thumb="p.thumbnail_url"
            class="aspect-[4/3] w-full"
          />
        </div>
        <div class="mt-1.5 flex items-center justify-between gap-2 text-[0.7rem] leading-none">
          <img
            v-if="p.avatar_url"
            :src="p.avatar_url"
            :alt="p.username || ''"
            loading="lazy"
            class="h-5 w-5 rounded-full object-cover"
            :title="p.username || ''"
          />
          <span class="flex-1" />
          <span v-if="p.views" class="shrink-0 tabular-nums text-muted-foreground/70">{{ p.views }}</span>
        </div>
      </button>
    </div>

    <PhotoLightbox v-model:visible="lightboxVisible" v-model:index="lightboxIndex" :photos="photos" />
  </div>
</template>
