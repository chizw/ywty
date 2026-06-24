<script setup lang="ts">
import { Separator, type SeparatorProps } from 'radix-vue'
import { cn } from '~/lib/utils'

const props = withDefaults(defineProps<SeparatorProps & { class?: string }>(), {
  orientation: 'horizontal',
})
</script>

<template>
  <Separator
    v-bind="props"
    :class="cn(
      'shrink-0 bg-border',
      orientation === 'horizontal' ? 'h-[1px] w-full' : 'h-full w-[1px]',
      props.class,
    )"
  />
</template>