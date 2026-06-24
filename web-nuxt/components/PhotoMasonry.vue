<script setup lang="ts">
// 瀑布流图片展示：CSS columns 实现，响应式 1-4 列
interface Photo {
  id: number | string
  pathname: string
  name?: string
  intro?: string
  width?: number
  height?: number
  [key: string]: any
}

const props = withDefaults(defineProps<{
  photos: Photo[]
  /** 指定列数；不传则响应式自适应 */
  columns?: number
}>(), {
  columns: 0,
})

const emit = defineEmits<{
  (e: 'click', payload: { photo: Photo; index: number }): void
}>()

const containerClass = computed(() => {
  if (props.columns > 0) return `masonry-cols masonry-cols-${props.columns}`
  return 'masonry-cols'
})

function onClick(photo: Photo, index: number) {
  emit('click', { photo, index })
}
</script>

<template>
  <div :class="containerClass">
    <div
      v-for="(p, i) in photos"
      :key="p.id"
      class="mb-3 break-inside-avoid group relative cursor-zoom-in"
      @click="onClick(p, i)"
    >
      <div class="relative bg-gray-100 rounded overflow-hidden">
        <img
          :src="`/uploads/${p.pathname}`"
          :alt="p.name || ''"
          loading="lazy"
          class="w-full h-auto block"
        />
        <div
          class="absolute bottom-1 right-1 flex gap-1 opacity-0 group-hover:opacity-100 transition"
          @click.stop
        >
          <LikeButton size="sm" target-type="photo" :target-id="p.id" />
          <ReportButton size="sm" target-type="photo" :target-id="p.id" />
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
