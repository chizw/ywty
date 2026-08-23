<script setup lang="ts">
// 瀑布流骨架屏（配合 PhotoMasonry 的占位）
interface Props {
  count?: number
  columns?: number
}
const props = withDefaults(defineProps<Props>(), {
  count: 8,
  columns: 0,
})
</script>

<template>
  <div :class="columns > 0 ? `masonry-cols masonry-cols-${columns}` : 'masonry-cols'">
    <div v-for="i in props.count" :key="i" class="mb-3 break-inside-avoid">
      <div class="skeleton aspect-[3/4] rounded-lg" :style="{ height: 180 + ((i * 37) % 120) + 'px' }" />
    </div>
  </div>
</template>

<style scoped>
.masonry-cols { column-gap: 0.75rem; column-count: 1; }
@media (min-width: 640px) { .masonry-cols { column-count: 2; } }
@media (min-width: 768px) { .masonry-cols { column-count: 3; } }
@media (min-width: 1024px) { .masonry-cols { column-count: 4; } }
.masonry-cols-1 { column-count: 1; }
.masonry-cols-2 { column-count: 2; }
.masonry-cols-3 { column-count: 3; }
.masonry-cols-4 { column-count: 4; }
</style>
