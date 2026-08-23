<script setup lang="ts">
import { cn } from '../../../lib/utils'
import { cva, type VariantProps } from 'class-variance-authority'

const alertVariants = cva(
  'relative w-full rounded-md border-l-2 px-4 py-3 text-sm',
  {
    variants: {
      variant: {
        default: 'border-l-border bg-muted/40 text-foreground',
        destructive: 'border-l-destructive bg-destructive/[0.05] text-destructive dark:bg-destructive/[0.08]',
      },
    },
    defaultVariants: {
      variant: 'default',
    },
  }
)

const props = defineProps<{
  variant?: VariantProps<typeof alertVariants>['variant']
  class?: string
}>()
</script>

<template>
  <div role="alert" :class="cn(alertVariants({ variant }), props.class)">
    <slot />
  </div>
</template>
