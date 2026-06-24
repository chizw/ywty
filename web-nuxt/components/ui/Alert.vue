<script setup lang="ts">
import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from '~/lib/utils'

const alertVariants = cva(
  'relative w-full rounded-lg border p-4 [&>svg~*]:pl-7 [&>svg+div]:translate-y-[-3px] [&>svg]:absolute [&>svg]:left-4 [&>svg]:top-4 [&>svg]:text-foreground',
  {
    variants: {
      variant: {
        default: 'bg-background text-foreground',
        destructive: 'border-destructive/50 text-destructive dark:border-destructive [&>svg]:text-destructive',
        success: 'border-green-500/50 text-green-700 bg-green-50 [&>svg]:text-green-600',
        info: 'border-blue-500/50 text-blue-700 bg-blue-50 [&>svg]:text-blue-600',
      },
    },
    defaultVariants: { variant: 'default' },
  },
)

interface Props {
  variant?: VariantProps<typeof alertVariants>['variant']
  class?: string
}

const props = defineProps<Props>()
</script>

<template>
  <div role="alert" :class="cn(alertVariants({ variant }), props.class)">
    <slot />
  </div>
</template>