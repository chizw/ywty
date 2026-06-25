<script setup lang="ts">
// 全屏图片查看器：上一张/下一张、键盘控制、缩略图导航、点击关闭
import { X, ChevronLeft, ChevronRight } from '@lucide/vue'

interface Photo {
  id: number | string
  pathname: string
  name?: string
  [key: string]: any
}

const props = defineProps<{
  photos: Photo[]
  visible: boolean
  index: number
}>()

const emit = defineEmits<{
  (e: 'update:visible', v: boolean): void
  (e: 'update:index', v: number): void
}>()

const current = computed(() => props.photos[props.index])

function close() {
  emit('update:visible', false)
}
function prev() {
  if (props.photos.length === 0) return
  const next = (props.index - 1 + props.photos.length) % props.photos.length
  emit('update:index', next)
}
function next() {
  if (props.photos.length === 0) return
  const n = (props.index + 1) % props.photos.length
  emit('update:index', n)
}
function select(i: number) {
  emit('update:index', i)
}

function onKey(e: KeyboardEvent) {
  if (!props.visible) return
  if (e.key === 'Escape') close()
  else if (e.key === 'ArrowLeft') prev()
  else if (e.key === 'ArrowRight') next()
}

onMounted(() => window.addEventListener('keydown', onKey))
onBeforeUnmount(() => window.removeEventListener('keydown', onKey))

// 锁定背景滚动
watch(
  () => props.visible,
  (v) => {
    if (import.meta.client) {
      document.body.style.overflow = v ? 'hidden' : ''
    }
  }
)
</script>

<template>
  <Teleport to="body">
    <div
      v-if="visible && current"
      class="fixed inset-0 z-[100] bg-black/90 flex flex-col"
      @click.self="close"
    >
      <!-- 顶部栏 -->
      <div class="flex items-center justify-between px-4 py-3 text-white/80 text-sm">
        <span>{{ index + 1 }} / {{ photos.length }}</span>
        <Button
          variant="ghost"
          size="icon"
          class="text-white hover:bg-white/10 hover:text-white h-8 w-8"
          aria-label="关闭"
          @click="close"
        >
          <X class="h-4 w-4" />
        </Button>
      </div>

      <!-- 主图区 -->
      <div class="flex-1 relative flex items-center justify-center px-12 min-h-0">
        <Button
          v-if="photos.length > 1"
          variant="ghost"
          size="icon"
          class="absolute left-2 top-1/2 -translate-y-1/2 h-10 w-10 rounded-full bg-white/10 hover:bg-white/20 text-white"
          aria-label="上一张"
          @click.stop="prev"
        >
          <ChevronLeft class="h-6 w-6" />
        </Button>

        <img
          :src="`/uploads/${current.pathname}`"
          :alt="current.name || ''"
          class="max-w-full max-h-full object-contain"
          @click.stop
        />

        <Button
          v-if="photos.length > 1"
          variant="ghost"
          size="icon"
          class="absolute right-2 top-1/2 -translate-y-1/2 h-10 w-10 rounded-full bg-white/10 hover:bg-white/20 text-white"
          aria-label="下一张"
          @click.stop="next"
        >
          <ChevronRight class="h-6 w-6" />
        </Button>
      </div>

      <!-- 标题 -->
      <div v-if="current.name" class="text-center text-white/70 text-sm py-2 px-4">
        {{ current.name }}
      </div>

      <!-- 缩略图导航 -->
      <div
        v-if="photos.length > 1"
        class="flex justify-center gap-2 px-4 py-3 overflow-x-auto"
        @click.stop
      >
        <button
          v-for="(p, i) in photos"
          :key="p.id"
          class="flex-shrink-0 w-14 h-14 rounded overflow-hidden border-2 transition"
          :class="i === index ? 'border-primary' : 'border-transparent opacity-60 hover:opacity-100'"
          @click="select(i)"
        >
          <img :src="`/uploads/${p.pathname}`" :alt="p.name || ''" class="w-full h-full object-cover" />
        </button>
      </div>
    </div>
  </Teleport>
</template>
