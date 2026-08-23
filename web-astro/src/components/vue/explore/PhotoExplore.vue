<script setup lang="ts">
// 探索页：筛选栏 + 瀑布流 + 骨架屏 + 无限滚动
import { computed, ref, onMounted, onBeforeUnmount } from 'vue'
import { Flame, Clock, Loader2 } from '@lucide/vue'
import PhotoGallery from '../photo/PhotoGallery.vue'
import PhotoGridSkeleton from '../photo/PhotoGridSkeleton.vue'
import { useApi } from '../../../lib/api'
import type { PublicPhoto } from '../../../lib/types'

const props = defineProps<{ initial: PublicPhoto[] }>()

const items = ref<PublicPhoto[]>([...props.initial])
const page = ref(1)
const total = ref(items.value.length)
const loading = ref(false)
const done = ref(false)

const sortMode = ref<'latest' | 'hot'>('latest')

const displayed = computed(() => {
  const list = [...items.value]
  if (sortMode.value === 'hot') {
    list.sort((a, b) => (b.views ?? 0) - (a.views ?? 0) || (b.likes ?? 0) - (a.likes ?? 0))
  }
  return list
})

async function loadMore() {
  if (loading.value || done.value) return
  loading.value = true
  try {
    const res = await useApi().get<PublicPhoto[]>('/api/v1/public/photos', {
      query: { page: page.value + 1, per_page: 24 },
      raw: true,
    })
    const env = res as any
    const arr = Array.isArray(env?.data) ? (env.data as PublicPhoto[]) : []
    if (arr.length > 0) {
      items.value = [...items.value, ...arr]
      page.value += 1
      total.value = Number(env?.meta?.total ?? total.value)
    }
    if (arr.length === 0 || env?.meta?.last_page === page.value) {
      done.value = true
    }
  } catch {
    done.value = true
  } finally {
    loading.value = false
  }
}

let observer: IntersectionObserver | null = null
const sentinel = ref<HTMLElement | null>(null)

onMounted(() => {
  observer = new IntersectionObserver(
    (entries) => {
      if (entries.some((e) => e.isIntersecting)) loadMore()
    },
    { rootMargin: '600px 0px' }
  )
  if (sentinel.value) observer.observe(sentinel.value)
})

onBeforeUnmount(() => observer?.disconnect())

const tabs = [
  { key: 'latest' as const, label: '最新', icon: Clock },
  { key: 'hot' as const, label: '最热', icon: Flame },
]
</script>

<template>
  <div>
    <!-- 筛选栏：文本式分段 -->
    <div class="mb-8 flex flex-wrap items-center justify-between gap-3 border-b hairline pb-0">
      <div class="flex items-center gap-6">
        <button
          v-for="t in tabs"
          :key="t.key"
          type="button"
          class="relative flex items-center gap-1.5 pb-3 text-sm transition-colors"
          :class="sortMode === t.key ? 'text-foreground' : 'text-muted-foreground hover:text-foreground'"
          @click="sortMode = t.key"
        >
          <component :is="t.icon" class="h-3.5 w-3.5" />
          {{ t.label }}
          <span
            v-if="sortMode === t.key"
            class="absolute inset-x-0 -bottom-px h-[1.5px] bg-brand"
          />
        </button>
      </div>
      <span class="text-xs tabular-nums text-muted-foreground">
        共 {{ total }} 张 · {{ sortMode === 'hot' ? '按浏览热度' : '按发布时间' }}
      </span>
    </div>

    <!-- 瀑布流 -->
    <PhotoGallery :photos="displayed" interactive />

    <!-- 加载中骨架屏 -->
    <PhotoGridSkeleton v-if="loading" :count="8" class="mt-3" />

    <!-- 加载更多哨兵 -->
    <div ref="sentinel" class="h-10" />

    <!-- 到底提示 -->
    <p v-if="done && items.length > 0" class="py-6 text-center text-sm text-muted-foreground">
      已经到底啦
    </p>
    <div v-else-if="loading" class="flex justify-center py-6 text-muted-foreground">
      <Loader2 class="h-5 w-5 animate-spin" />
    </div>
  </div>
</template>
