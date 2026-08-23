<script setup lang="ts">
// 首页「接触印相」：一排带发丝线的图片底片，点击开灯箱
import { ref } from 'vue'
import LazyImage from '../photo/LazyImage.vue'
import PhotoLightbox from '../photo/PhotoLightbox.vue'
import type { PublicPhoto } from '../../../lib/types'

const props = defineProps<{ photos: PublicPhoto[] }>()

const lightboxVisible = ref(false)
const lightboxIndex = ref(0)

function open(i: number) {
  lightboxIndex.value = i
  lightboxVisible.value = true
}
</script>

<template>
  <div>
    <div class="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-6 lg:gap-4">
      <button
        v-for="(p, i) in props.photos"
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
        <div class="mt-1.5 flex items-baseline justify-between gap-2 text-[0.7rem] leading-none">
          <span class="truncate text-muted-foreground">{{ p.username || p.name }}</span>
          <span v-if="p.views" class="shrink-0 tabular-nums text-muted-foreground/70">{{ p.views }}</span>
        </div>
      </button>
    </div>

    <PhotoLightbox v-model:visible="lightboxVisible" v-model:index="lightboxIndex" :photos="props.photos" />
  </div>
</template>
