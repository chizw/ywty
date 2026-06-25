<script setup lang="ts">
// 支付选择组件：套餐 + 渠道选择
import { Check } from '@lucide/vue'

interface PaymentChannel {
  key: string
  name: string
  icon: string
  description?: string
  disabled?: boolean
}

const props = withDefaults(defineProps<{
  channels?: PaymentChannel[]
  modelValue?: string
}>(), {
  channels: () => [
    { key: 'alipay', name: '支付宝', icon: '💰', description: '推荐' },
    { key: 'wxpay', name: '微信支付', icon: '💬' },
    { key: 'qqpay', name: 'QQ支付', icon: '🐧' },
    { key: 'paypal', name: 'PayPal', icon: '🌐' },
    { key: 'stripe', name: '信用卡', icon: '💳' },
  ],
})

const emit = defineEmits<{
  'update:modelValue': [string]
  select: [channel: string]
}>()

const selected = computed({
  get: () => props.modelValue || '',
  set: (v) => {
    emit('update:modelValue', v)
    emit('select', v)
  },
})

function selectChannel(key: string) {
  const ch = props.channels.find((c) => c.key === key)
  if (ch?.disabled) return
  selected.value = key
}
</script>

<template>
  <div class="space-y-2">
    <Label class="mb-2 block">选择支付方式</Label>
    <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
      <button
        v-for="ch in channels"
        :key="ch.key"
        type="button"
        class="flex items-center gap-3 p-3 border rounded-lg text-left transition-colors"
        :class="{
          'border-primary bg-primary/5 ring-1 ring-primary': selected === ch.key,
          'border-border hover:border-primary/50': selected !== ch.key && !ch.disabled,
          'border-border bg-muted opacity-50 cursor-not-allowed': ch.disabled,
        }"
        :disabled="ch.disabled"
        @click="selectChannel(ch.key)"
      >
        <span class="text-2xl">{{ ch.icon }}</span>
        <div class="flex-1">
          <div class="text-sm font-medium text-foreground">{{ ch.name }}</div>
          <div v-if="ch.description" class="text-xs text-muted-foreground">{{ ch.description }}</div>
        </div>
        <div v-if="selected === ch.key" class="text-primary">
          <Check class="h-5 w-5" />
        </div>
      </button>
    </div>
  </div>
</template>
