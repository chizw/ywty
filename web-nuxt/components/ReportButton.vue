<script setup lang="ts">
// 举报弹窗
import { Button } from '~/components/ui/button'
import { Textarea } from '~/components/ui/textarea'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '~/components/ui/dialog'
import { Flag } from '@lucide/vue'

const props = defineProps<{
  targetType: string
  targetId: number | string
  size?: 'sm' | 'md'
}>()

const api = useApi()
const auth = useAuthStore()
const open = ref(false)
const reason = ref('')
const submitting = ref(false)
const msg = ref('')

async function submit() {
  if (!auth.isLoggedIn) {
    navigateTo('/auth/login')
    return
  }
  submitting.value = true
  msg.value = ''
  try {
    await api.post('/api/v1/reports', {
      target_type: props.targetType,
      target_id: Number(props.targetId),
      content: reason.value || '违规内容',
    })
    msg.value = '已提交，感谢反馈'
    reason.value = ''
    setTimeout(() => {
      open.value = false
      msg.value = ''
    }, 800)
  } catch (err: any) {
    msg.value = err?.statusMessage || '提交失败'
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <span class="inline-block">
    <Button
      type="button"
      variant="outline"
      :size="size === 'sm' ? 'sm' : 'default'"
      class="gap-1"
      @click.stop.prevent="open = true"
    >
      <Flag class="h-3 w-3" />
      举报
    </Button>

    <Dialog :open="open" @update:open="open = $event">
      <DialogContent class="max-w-md">
        <DialogHeader>
          <DialogTitle>举报内容</DialogTitle>
          <DialogDescription>请简要说明原因（可选）</DialogDescription>
        </DialogHeader>
        <Textarea
          v-model="reason"
          :rows="3"
          :maxlength="200"
          placeholder="如：违规、色情、侵权�?
        />
        <p v-if="msg" class="text-sm" :class="msg.includes('失败') ? 'text-destructive' : 'text-primary'">{{ msg }}</p>
        <DialogFooter>
          <Button variant="outline" @click="open = false">取消</Button>
          <Button :loading="submitting" @click="submit">提交</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </span>
</template>
