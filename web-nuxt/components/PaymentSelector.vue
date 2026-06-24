<script setup lang="ts">
// 支付选择组件：图标 + 渠道选择
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
    <label class="block text-sm text-gray-700 mb-2">选择支付方式</label>
    <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
      <button
        v-for="ch in channels"
        :key="ch.key"
        type="button"
        class="flex items-center gap-3 p-3 border rounded-lg text-left transition"
        :class="{
          'border-primary-500 bg-primary-50 ring-1 ring-primary-500': selected === ch.key,
          'border-gray-300 hover:border-gray-400': selected !== ch.key && !ch.disabled,
          'border-gray-200 bg-gray-50 opacity-50 cursor-not-allowed': ch.disabled,
        }"
        :disabled="ch.disabled"
        @click="selectChannel(ch.key)"
      >
        <span class="text-2xl">{{ ch.icon }}</span>
        <div class="flex-1">
          <div class="text-sm font-medium text-gray-900">{{ ch.name }}</div>
          <div v-if="ch.description" class="text-xs text-gray-500">{{ ch.description }}</div>
        </div>
        <div v-if="selected === ch.key" class="text-primary-600">
          <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
            <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
          </svg>
        </div>
      </button>
    </div>
  </div>
</template>
