<script setup lang="ts">
import { CheckboxIndicator, CheckboxRoot, type CheckboxRootEmits, type CheckboxRootProps } from 'radix-vue'
import { cn } from '~/lib/utils'

const props = defineProps<CheckboxRootProps & { class?: string }>()
const emits = defineEmits<CheckboxRootEmits>()
</script>

<template>
  <CheckboxRoot
    v-bind="props"
    :class="cn(
      'peer h-4 w-4 shrink-0 rounded-sm border border-primary ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 data-[state=checked]:bg-primary data-[state=checked]:text-primary-foreground',
      props.class,
    )"
    @update:checked="emits('update:checked', $event)"
  >
    <CheckboxIndicator class="flex h-full w-full items-center justify-center text-current">
      <svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="20 6 9 17 4 12" /></svg>
    </CheckboxIndicator>
  </CheckboxRoot>
</template>