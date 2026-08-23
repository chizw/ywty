<script setup lang="ts">
// 验证码输入组件：发送验证码 + 60s 倒计时
import { ref, onMounted, onBeforeUnmount } from 'vue'
import Input from '../ui/Input.vue'
import { useApi } from '../../../lib/api'

const props = withDefaults(
  defineProps<{
    account: string
    event: string
    length?: number
  }>(),
  {
    length: 6,
  }
)

const emit = defineEmits<{
  verified: [code: string]
}>()

const code = ref('')
const countdown = ref(0)
const sending = ref(false)
const error = ref('')
let timer: ReturnType<typeof setInterval> | null = null

async function send() {
  if (countdown.value > 0 || sending.value) return
  if (!props.account) {
    error.value = '请先填写账号'
    return
  }
  sending.value = true
  error.value = ''
  try {
    await useApi().post('/api/v1/verify-codes', {
      email: props.account,
      event: props.event,
    })
    countdown.value = 60
    timer = setInterval(() => {
      countdown.value--
      if (countdown.value <= 0 && timer) clearInterval(timer)
    }, 1000)
  } catch (err: any) {
    error.value = err?.message || '发送失败'
  } finally {
    sending.value = false
  }
}

function commit() {
  if (code.value.length >= Math.min(4, props.length)) emit('verified', code.value)
}

onMounted(() => {
  // 组件卸载时清理计时器
})
onBeforeUnmount(() => {
  if (timer) clearInterval(timer)
})

defineExpose({ code })
</script>

<template>
  <div class="space-y-2">
    <div class="flex gap-3">
      <Input
        v-model="code"
        type="text"
        :maxlength="length"
        placeholder="验证码"
        variant="underline"
        class="flex-1"
        @input="commit"
      />
      <button
        type="button"
        class="shrink-0 self-end pb-2 text-[13px] transition-colors"
        :class="countdown > 0 || sending ? 'text-muted-foreground' : 'text-brand hover:underline'"
        :disabled="countdown > 0 || sending"
        @click="send"
      >
        {{ countdown > 0 ? `${countdown}s 后重发` : '发送验证码' }}
      </button>
    </div>
    <p v-if="error" class="text-xs text-destructive">{{ error }}</p>
  </div>
</template>
