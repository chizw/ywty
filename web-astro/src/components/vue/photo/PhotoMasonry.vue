<script setup lang="ts">
// 瀑布流图片网格：CSS columns 响应式 1-4 列，懒加载 + hover 操作
import { computed } from 'vue'
import LazyImage from './LazyImage.vue'
import LikeButton from './LikeButton.vue'
import type { PublicPhoto as Photo } from '../../../lib/types'

const props = withDefaults(
  defineProps<{
    photos: Photo[]
    columns?: number
    /** 是否显示 hover 操作（点赞） */
    interactive?: boolean
  }>(),
  {
    columns: 0,
    interactive: true,
  }
)

const emit = defineEmits<{
  click: [payload: { photo: Photo; index: number }]
}>()

const containerClass = computed(() => {
  if (props.columns > 0) return `masonry-cols masonry-cols-${props.columns}`
  return 'masonry-cols'
})

function ratio(photo: Photo): string {
  const w = Number(photo.width)
  const h = Number(photo.height)
  if (w > 0 && h > 0) return `${w} / ${h}`
  return '4 / 3'
}

function onClick(photo: Photo, index: number) {
  emit('click', { photo, index })
}
</script>

<template>
  <div :class="containerClass">
    <div
      v-for="(p, i) in photos"
      :key="p.id"
      class="group relative mb-3 cursor-zoom-in break-inside-avoid"
      @click="onClick(p, i)"
    >
      <div class="relative overflow-hidden rounded-lg bg-muted" :style="{ aspectRatio: ratio(p) }">
        <LazyImage
          :src="p.url"
          :alt="p.name || p.username || ''"
          :thumb="p.thumbnail_url"
          class="h-full w-full"
        />

        <!-- hover 遮罩 -->
        <div
          class="absolute inset-0 flex flex-col justify-between bg-gradient-to-t from-black/60 via-transparent to-transparent p-2 opacity-0 transition-opacity duration-200 group-hover:opacity-100"
          @click.stop
        >
          <div class="flex justify-end">
            <LikeButton
              v-if="interactive"
              size="sm"
              target-type="photo"
              :target-id="p.id"
              :count="p.likes"
            />
          </div>
          <div class="flex items-center justify-between gap-2 text-xs text-white/90">
            <img
              v-if="p.avatar_url"
              :src="p.avatar_url"
              :alt="p.username || ''"
              loading="lazy"
              class="h-5 w-5 rounded-full object-cover"
              :title="p.username || ''"
            />
            <span v-else class="truncate">{{ p.username || p.name || '' }}</span>
            <span v-if="p.views" class="shrink-0 opacity-80">{{ p.views }} 浏览</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.masonry-cols {
  column-gap: 0.75rem;
  column-count: 1;
}
@media (min-width: 640px) {
  .masonry-cols {
    column-count: 2;
  }
}
@media (min-width: 768px) {
  .masonry-cols {
    column-count: 3;
  }
}
@media (min-width: 1024px) {
  .masonry-cols {
    column-count: 4;
  }
}
.masonry-cols-1 { column-count: 1; }
.masonry-cols-2 { column-count: 2; }
.masonry-cols-3 { column-count: 3; }
.masonry-cols-4 { column-count: 4; }
</style>
