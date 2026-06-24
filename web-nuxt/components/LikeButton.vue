<script setup lang="ts">
// 点赞按钮（通用：支持任意 target_type + target_id）
import { Button } from '~/components/ui/button'
import { Heart } from '@lucide/vue'

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
</script>

<template>
  <Button
    type="button"
    :variant="liked ? 'outline' : 'outline'"
    :size="size === 'sm' ? 'sm' : 'default'"
    :disabled="loading"
    :class="[
      'gap-1',
      liked
        ? 'text-red-600 border-red-200 bg-red-50 hover:bg-red-100 hover:text-red-700'
        : 'text-muted-foreground',
    ]"
    @click.stop.prevent="toggle"
  >
    <Heart class="h-4 w-4" :class="liked ? 'fill-red-600' : ''" />
    <span>{{ count }}</span>
  </Button>
</template>
