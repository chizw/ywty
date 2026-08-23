<script setup lang="ts">
// 图片画廊：瀑布流 + 灯箱（含打开状态）
import { ref } from 'vue'
import PhotoMasonry from './PhotoMasonry.vue'
import PhotoLightbox from './PhotoLightbox.vue'
import type { PublicPhoto as Photo } from '../../../lib/types'

withDefaults(
  defineProps<{
    photos: Photo[]
    columns?: number
    interactive?: boolean
    /** 网格下方额外插槽（如加载更多按钮） */
  }>(),
  {
    columns: 0,
    interactive: true,
  }
)

const lightboxVisible = ref(false)
const lightboxIndex = ref(0)

function openLightbox(index: number) {
  lightboxIndex.value = index
  lightboxVisible.value = true
}
</script>

<template>
  <div>
    <PhotoMasonry :photos="photos" :columns="columns" :interactive="interactive" @click="openLightbox($event.index)" />
    <slot />
    <PhotoLightbox v-model:visible="lightboxVisible" v-model:index="lightboxIndex" :photos="photos" />
  </div>
</template>
