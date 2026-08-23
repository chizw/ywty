<script setup lang="ts">
import { computed } from 'vue'
import { cn } from '../../../lib/utils'

interface Props {
  modelValue?: string | number
  type?: string
  placeholder?: string
  disabled?: boolean
  /** boxed=常规方框（后台/数据密集区）；underline=墨线表单（公开/认证页） */
  variant?: 'boxed' | 'underline'
  class?: string
}
const props = withDefaults(defineProps<Props>(), {
  variant: 'boxed',
})

const emit = defineEmits<{
  'update:modelValue': [value: string | number | null]
}>()

const inputClass = computed(() =>
  cn(
    props.variant === 'underline'
      ? 'flex h-10 w-full rounded-none border-0 border-b border-border bg-transparent px-0 py-2 text-[0.95rem] shadow-none transition-colors placeholder:text-muted-foreground/60 focus-visible:border-brand focus-visible:outline-none focus-visible:ring-0 disabled:cursor-not-allowed disabled:opacity-50'
      : 'flex h-9 w-full rounded-md border border-input bg-background px-3 py-1 text-sm shadow-sm transition-colors placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50',
    props.class
  )
)
</script>

<template>
  <input
    :type="props.type || 'text'"
    :value="props.modelValue"
    :placeholder="props.placeholder"
    :disabled="props.disabled"
    :class="inputClass"
    @input="emit('update:modelValue', ($event.target as HTMLInputElement).value)"
  />
</template>
