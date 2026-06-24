<script setup lang="ts">
// 点赞按钮（通用：支持任意 target_type + target_id）
const props = defineProps<{
  targetType: string // 'photo' | 'album' | ...
  targetId: number | string
  initialLiked?: boolean
  initialCount?: number
  size?: 'sm' | 'md'
}>()

const api = useApi()
const auth = useAuthStore()
const liked = ref(!!props.initialLiked)
const count = ref(props.initialCount ?? 0)
const loading = ref(false)

async function fetchStatus() {
  try {
    const res: any = await api.get('/api/v1/likes', {
      query: { target_type: props.targetType, target_id: props.targetId },
    })
    if (res) {
      liked.value = !!res.liked
      count.value = res.count ?? 0
    }
  } catch {
    /* 公开资源时未登录也允许 */
  }
}

onMounted(fetchStatus)
watch(
  () => [props.targetType, props.targetId],
  () => fetchStatus()
)

async function toggle() {
  if (!auth.isLoggedIn) {
    navigateTo('/auth/login')
    return
  }
  if (loading.value) return
  loading.value = true
  try {
    const res: any = await api.post('/api/v1/likes', {
      target_type: props.targetType,
      target_id: Number(props.targetId),
    })
    liked.value = !!res.liked
    count.value = res.count ?? 0
  } catch (err: any) {
    alert(err?.statusMessage || '操作失败')
  } finally {
    loading.value = false
  }
}

const sizeClass = computed(() =>
  props.size === 'sm' ? 'text-xs px-2 py-0.5' : 'text-sm px-3 py-1'
)
</script>

<template>
  <button
    type="button"
    :disabled="loading"
    :class="[
      'inline-flex items-center gap-1 rounded-md border transition',
      liked
        ? 'bg-red-50 text-red-600 border-red-200 hover:bg-red-100'
        : 'bg-white text-gray-600 border-gray-200 hover:bg-gray-50',
      sizeClass,
    ]"
    @click.stop.prevent="toggle"
  >
    <span aria-hidden="true">{{ liked ? '♥' : '♡' }}</span>
    <span>{{ count }}</span>
  </button>
</template>
