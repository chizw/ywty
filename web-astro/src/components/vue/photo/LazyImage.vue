<script setup lang="ts">
// LazyImage：IntersectionObserver 懒加载 + 缩略图模糊占位 + 渐显
import { ref, onMounted, onBeforeUnmount } from 'vue'

const props = withDefaults(
  defineProps<{
    src: string
    alt?: string
    /** 低清占位图（可选）：加载完成前显示模糊缩略图 */
    thumb?: string | null
    class?: string
    eager?: boolean
  }>(),
  {
    alt: '',
    thumb: null,
    eager: false,
  }
)

const root = ref<HTMLElement | null>(null)
const inView = ref(props.eager)
const loaded = ref(false)
const failed = ref(false)

let observer: IntersectionObserver | null = null

onMounted(() => {
  if (props.eager) {
    inView.value = true
    return
  }
  if (!root.value || typeof IntersectionObserver === 'undefined') {
    inView.value = true
    return
  }
  observer = new IntersectionObserver(
    (entries) => {
      for (const entry of entries) {
        if (entry.isIntersecting) {
          inView.value = true
          observer?.disconnect()
        }
      }
    },
    { rootMargin: '400px 0px' }
  )
  observer.observe(root.value)
})

onBeforeUnmount(() => observer?.disconnect())
</script>

<template>
  <div ref="root" :class="['relative overflow-hidden bg-muted', props.class]">
    <!-- 缩略图模糊占位 -->
    <img
      v-if="props.thumb && !loaded && !failed"
      :src="props.thumb"
      :alt="props.alt"
      loading="lazy"
      decoding="async"
      class="absolute inset-0 h-full w-full scale-105 object-cover blur-md transition-opacity duration-500"
    />
    <!-- 主图 -->
    <img
      v-if="inView"
      :src="props.src"
      :alt="props.alt"
      :loading="props.eager ? 'eager' : 'lazy'"
      decoding="async"
      class="h-full w-full object-cover transition-all duration-500"
      :class="loaded && !failed ? 'opacity-100' : 'opacity-0'"
      @load="loaded = true"
      @error="failed = true"
    />
    <!-- 加载中骨架（无缩略图时） -->
    <div v-if="!props.thumb && !loaded && !failed && inView" class="absolute inset-0 skeleton" />
  </div>
</template>
