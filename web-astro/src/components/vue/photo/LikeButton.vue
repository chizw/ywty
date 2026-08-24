<script setup lang="ts">
// 点赞按钮：挂载时回显点赞状态与计数（GET /api/v1/likes）；未登录跳登录页；已登录 POST /api/v1/likes 切换
import { ref, onMounted } from 'vue'
import { Heart } from '@lucide/vue'
import { useApi } from '../../../lib/api'
import { readAuthPair } from '../../../lib/auth'
import { message } from '../../../lib/vue-store'
import { cn } from '../../../lib/utils'

const props = withDefaults(
  defineProps<{
    targetType: 'photo' | 'album'
    targetId: number | string
    count?: number
    size?: 'sm' | 'md'
  }>(),
  {
    count: 0,
    size: 'md',
  }
)

const liked = ref(false)
const count = ref(Number(props.count) || 0)
const busy = ref(false)

function requireLogin(): boolean {
  if (!readAuthPair()) {
    window.location.assign('/auth/login?redirect=' + encodeURIComponent(window.location.pathname))
    return false
  }
  return true
}

// 接口把 { liked, count } 又包了一层 data（{ code, message, data: { data } }），解包时做兼容
function unwrapLikePayload(raw: any): { liked?: boolean; count?: number } {
  return raw?.data ?? raw ?? {}
}

onMounted(async () => {
  // /likes 需要登录态，匿名访客保持服务端下发的初始计数即可
  if (!readAuthPair()) return
  try {
    const res = await useApi().get<any>('/api/v1/likes', {
      query: {
        target_type: props.targetType,
        target_id: Number(props.targetId),
      },
    })
    const payload = unwrapLikePayload(res)
    liked.value = !!payload.liked
    if (payload.count !== undefined && payload.count !== null) count.value = Math.max(0, Number(payload.count) || 0)
  } catch {
    // 回显失败不影响交互，保持初始值
  }
})

async function onClick() {
  if (busy.value) return
  if (!requireLogin()) return
  busy.value = true
  try {
    const res = await useApi().post<any>('/api/v1/likes', {
      target_type: props.targetType,
      target_id: Number(props.targetId),
    })
    const payload = unwrapLikePayload(res)
    const nowLiked = payload.liked ?? !liked.value
    liked.value = nowLiked
    if (payload.count !== undefined && payload.count !== null) {
      count.value = Math.max(0, Number(payload.count) || 0)
    } else {
      count.value = Math.max(0, count.value + (nowLiked ? 1 : -1))
    }
  } catch {
    message.error('操作失败')
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <button
    class="inline-flex items-center gap-1 rounded-full px-2 py-1 text-xs font-medium transition-colors"
    :class="[
      size === 'sm' ? 'h-6' : 'h-8 px-2.5',
      liked
        ? 'bg-primary/15 text-primary'
        : 'bg-black/30 text-white backdrop-blur hover:bg-black/40',
    ]"
    :disabled="busy"
    :aria-label="liked ? '取消点赞' : '点赞'"
    @click.stop="onClick"
  >
    <Heart class="h-3.5 w-3.5" :class="cn(liked && 'fill-current')" />
    <span>{{ count > 0 ? count : '' }}</span>
  </button>
</template>
