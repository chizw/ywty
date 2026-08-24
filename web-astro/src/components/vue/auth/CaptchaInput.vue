<script setup lang="ts">
// 图形验证码：拉取 /captcha 展示 base64 图，点击刷新；code 双向绑定
import { onMounted, ref, watch } from 'vue'
import Input from '../ui/Input.vue'
import { useApi } from '../../../lib/api'

const props = defineProps<{ captchaId: string }>()
const emit = defineEmits<{ 'update:captchaId': [id: string]; 'update:code': [code: string] }>()

const imageBase64 = ref('')
const loading = ref(false)
const code = ref('')

async function refresh() {
  loading.value = true
  try {
    const api = useApi()
    const r = await api.get<any>('/api/v1/captcha', { raw: true })
    const d = r?.data ?? r ?? {}
    imageBase64.value = d.image_base64 || ''
    emit('update:captchaId', d.captcha_id || '')
  } catch {
    imageBase64.value = ''
  } finally {
    loading.value = false
  }
}

watch(code, (v) => emit('update:code', v))
onMounted(refresh)
defineExpose({ refresh })
</script>

<template>
  <div class="space-y-2">
    <Input v-model="code" type="text" maxlength="4" placeholder="图形验证码" variant="underline" />
    <button
      type="button"
      class="block h-10 w-full max-w-[140px] cursor-pointer overflow-hidden rounded border border-border bg-muted/40"
      title="点击刷新"
      @click.prevent="refresh"
    >
      <img v-if="imageBase64" :src="imageBase64" alt="图形验证码" class="h-full w-full object-contain" />
      <span v-else class="text-xs text-muted-foreground">{{ loading ? '加载中…' : '点击获取' }}</span>
    </button>
  </div>
</template>
