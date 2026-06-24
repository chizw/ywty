<script setup lang="ts">
// 验证码输入组件（封装发�?验证�?// - 支持邮箱 / 短信验证码发�?// - 支持图形验证码（可选，captcha=true 时显示）
// - 60 秒倒计�?// - 验证码输入框�?-6 位）
import { Input } from '~/components/ui/input'
import { Button } from '~/components/ui/button'
import { Label } from '~/components/ui/label'

const props = withDefaults(defineProps<{
  // 兼容 'phone'：内部归一化为 'sms'
  channel: 'email' | 'sms' | 'phone'
  account: string
  event: string
  /** 是否需要图形验证码 */
  captcha?: boolean
  /** 验证码长度（默认 6�?*/
  length?: number
}>(), {
  captcha: false,
  length: 6,
})

const emit = defineEmits<{
  verified: [code: string]
}>()

const api = useApi()
const code = ref('')
const countdown = ref(0)
const sending = ref(false)
const error = ref('')

// 图形验证�?const captchaImage = ref('')
const captchaKey = ref('')
const captchaCode = ref('')
const captchaLoading = ref(false)

function normalizeChannel(c: string): 'email' | 'sms' {
  return c === 'email' ? 'email' : 'sms'
}

async function loadCaptcha() {
  if (!props.captcha) return
  captchaLoading.value = true
  try {
    const res = await api.get<any>('/api/v1/captcha')
    const d = res as any
    const img = d?.image || d?.data?.image || d?.img || ''
    const key = d?.key || d?.data?.key || d?.captcha_id || ''
    captchaImage.value = img.startsWith('data:') ? img : (img ? `data:image/png;base64,${img}` : '')
    captchaKey.value = key
  } catch {
    captchaImage.value = ''
    captchaKey.value = ''
  } finally {
    captchaLoading.value = false
  }
}

onMounted(() => {
  if (props.captcha) loadCaptcha()
})

watch(() => props.captcha, (v) => {
  if (v) loadCaptcha()
})

async function send() {
  if (countdown.value > 0 || sending.value) return
  if (!props.account) {
    error.value = '请先填写账号'
    return
  }
  if (props.captcha && !captchaCode.value) {
    error.value = '请先输入图形验证�?
    return
  }
  sending.value = true
  error.value = ''
  try {
    const body: any = {
      channel: normalizeChannel(props.channel),
      account: props.account,
      event: props.event,
    }
    if (props.captcha) {
      body.captcha_key = captchaKey.value
      body.captcha_code = captchaCode.value
    }
    await api.post('/api/v1/verify-codes', body)
    countdown.value = 60
    const t = setInterval(() => {
      countdown.value--
      if (countdown.value <= 0) clearInterval(t)
    }, 1000)
    captchaCode.value = ''
  } catch (err: any) {
    error.value = err?.statusMessage || '发送失�?
    // 发送失败时刷新图形验证�?    if (props.captcha) loadCaptcha()
  } finally {
    sending.value = false
  }
}

function commit() {
  if (code.value.length >= Math.min(4, props.length)) emit('verified', code.value)
}

defineExpose({ code, captchaCode })
</script>

<template>
  <div class="space-y-2">
    <!-- 图形验证�?-->
    <div v-if="captcha" class="flex items-center gap-2">
      <Input
        v-model="captchaCode"
        type="text"
        :maxlength="6"
        placeholder="图形验证�?
        class="flex-1"
      />
      <button
        type="button"
        class="h-10 w-[100px] flex-shrink-0 border border-input rounded-md overflow-hidden bg-muted flex items-center justify-center text-xs text-muted-foreground"
        :disabled="captchaLoading"
        @click="loadCaptcha"
      >
        <img v-if="captchaImage" :src="captchaImage" alt="captcha" class="w-full h-full object-cover" />
        <span v-else>{{ captchaLoading ? '加载�?..' : '点击获取' }}</span>
      </button>
    </div>

    <!-- 验证�?+ 发送按�?-->
    <div class="flex gap-2">
      <Input
        v-model="code"
        type="text"
        :maxlength="length"
        placeholder="验证�?
        class="flex-1"
        @input="commit"
      />
      <Button
        type="button"
        variant="outline"
        :disabled="countdown > 0 || sending"
        class="whitespace-nowrap"
        @click="send"
      >
        {{ countdown > 0 ? `${countdown}s` : '发送验证码' }}
      </Button>
    </div>
    <p v-if="error" class="text-xs text-destructive">{{ error }}</p>
  </div>
</template>
