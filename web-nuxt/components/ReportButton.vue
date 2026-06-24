<script setup lang="ts">
// 举报弹窗
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

const sizeClass = computed(() =>
  props.size === 'sm' ? 'text-xs px-2 py-0.5' : 'text-sm px-3 py-1'
)
</script>

<template>
  <span class="inline-block">
    <button
      type="button"
      :class="[
        'inline-flex items-center gap-1 rounded-md border border-gray-200 bg-white text-gray-500 hover:bg-gray-50',
        sizeClass,
      ]"
      @click.stop.prevent="open = true"
    >举报</button>

    <div
      v-if="open"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4"
      @click.self="open = false"
    >
      <div class="w-full max-w-md bg-white rounded-lg shadow-lg p-5">
        <h3 class="text-lg font-semibold text-gray-900 mb-2">举报内容</h3>
        <p class="text-xs text-gray-500 mb-3">请简要说明原因（可选）</p>
        <textarea
          v-model="reason"
          rows="3"
          maxlength="200"
          class="w-full px-3 py-2 border border-gray-300 rounded-md text-sm"
          placeholder="如：违规、色情、侵权…"
        />
        <p v-if="msg" class="mt-2 text-sm" :class="msg.includes('失败') ? 'text-red-500' : 'text-primary-600'">{{ msg }}</p>
        <div class="mt-4 flex justify-end gap-2">
          <button class="px-3 py-1.5 text-sm border border-gray-300 rounded-md" @click="open = false">取消</button>
          <AppButton size="sm" :loading="submitting" @click="submit">提交</AppButton>
        </div>
      </div>
    </div>
  </span>
</template>
