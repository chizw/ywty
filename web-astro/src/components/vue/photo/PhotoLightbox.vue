<script setup lang="ts">
// 全屏图片查看器：上一张/下一张、键盘控制、缩略图导航、下载、点击关闭
import { computed, onMounted, onBeforeUnmount, watch } from 'vue'
import { X, ChevronLeft, ChevronRight, Download } from '@lucide/vue'
import Button from '../ui/Button.vue'
import { cn } from '../../../lib/utils'
import type { PublicPhoto as Photo } from '../../../lib/types'

const props = defineProps<{
  photos: Photo[]
  visible: boolean
  index: number
}>()

const emit = defineEmits<{
  'update:visible': [v: boolean]
  'update:index': [v: number]
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

watch(
  () => props.visible,
  (v) => {
    if (!import.meta.env.SSR) {
      document.body.style.overflow = v ? 'hidden' : ''
    }
  }
)

function fileName(photo: Photo, index: number): string {
  return photo.name || `photo-${index + 1}.jpg`
}
</script>

<template>
  <Teleport to="body">
    <div
      v-if="visible && current"
      class="fixed inset-0 z-[100] flex flex-col bg-black/95"
      @click.self="close"
    >
      <!-- 顶部栏 -->
      <div class="flex items-center justify-between px-4 py-3 text-sm text-white/80">
        <span>{{ index + 1 }} / {{ photos.length }}</span>
        <div class="flex items-center gap-2">
          <a
            :href="current.url"
            :download="fileName(current, index)"
            target="_blank"
            class="inline-flex h-8 w-8 items-center justify-center rounded-md text-white/80 transition-colors hover:bg-white/10 hover:text-white"
            aria-label="下载"
            @click.stop
          >
            <Download class="h-4 w-4" />
          </a>
          <button
            class="inline-flex h-8 w-8 items-center justify-center rounded-md text-white/80 transition-colors hover:bg-white/10 hover:text-white"
            aria-label="关闭"
            @click="close"
          >
            <X class="h-4 w-4" />
          </button>
        </div>
      </div>

      <!-- 主图区 -->
      <div class="relative flex min-h-0 flex-1 items-center justify-center px-12">
        <button
          v-if="photos.length > 1"
          class="absolute left-2 top-1/2 inline-flex h-10 w-10 -translate-y-1/2 items-center justify-center rounded-full bg-white/10 text-white transition-colors hover:bg-white/20"
          aria-label="上一张"
          @click.stop="prev"
        >
          <ChevronLeft class="h-6 w-6" />
        </button>

        <img
          :src="current.url"
          :alt="current.name || ''"
          class="max-h-full max-w-full object-contain"
          @click.stop
        />

        <button
          v-if="photos.length > 1"
          class="absolute right-2 top-1/2 inline-flex h-10 w-10 -translate-y-1/2 items-center justify-center rounded-full bg-white/10 text-white transition-colors hover:bg-white/20"
          aria-label="下一张"
          @click.stop="next"
        >
          <ChevronRight class="h-6 w-6" />
        </button>
      </div>

      <!-- 标题 -->
      <div v-if="current.name" class="px-4 py-2 text-center text-sm text-white/70">
        {{ current.name }}
      </div>

      <!-- 缩略图导航 -->
      <div
        v-if="photos.length > 1"
        class="flex justify-center gap-2 overflow-x-auto px-4 py-3"
        @click.stop
      >
        <button
          v-for="(p, i) in photos"
          :key="p.id"
          class="h-14 w-14 flex-shrink-0 overflow-hidden rounded border-2 transition"
          :class="i === index ? 'border-primary' : 'border-transparent opacity-60 hover:opacity-100'"
          @click="select(i)"
        >
          <img
            :src="p.thumbnail_url || p.url"
            :alt="p.name || ''"
            class="h-full w-full object-cover"
          />
        </button>
      </div>
    </div>
  </Teleport>
</template>
