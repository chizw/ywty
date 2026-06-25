<script setup lang="ts">
// 通用按钮组件：包装 shadcn Button，兼容旧接口
import { Button } from '~/components/ui/button'

const variantMap = {
  primary: 'default',
  secondary: 'secondary',
  ghost: 'ghost',
  danger: 'destructive',
} as const

const sizeMap = {
  sm: 'sm',
  md: 'default',
  lg: 'lg',
} as const

const props = withDefaults(defineProps<{
  variant?: 'primary' | 'secondary' | 'ghost' | 'danger'
  size?: 'sm' | 'md' | 'lg'
  type?: 'button' | 'submit' | 'reset'
  loading?: boolean
  disabled?: boolean
  block?: boolean
}>(), {
  variant: 'primary',
  size: 'md',
  type: 'button',
  loading: false,
  disabled: false,
  block: false,
})

const shadVariant = computed(() => variantMap[props.variant])
const shadSize = computed(() => sizeMap[props.size])
</script>

<template>
  <Button
    :variant="shadVariant"
    :size="shadSize"
    :disabled="disabled"
    :loading="loading"
    :type="type"
    :class="block ? 'w-full' : ''"
  >
    <slot />
  </Button>
</template>
