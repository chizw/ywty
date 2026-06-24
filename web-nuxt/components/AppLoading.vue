<script setup lang="ts">
// 通用加载占位 �?使用 shadcn Skeleton
import { Skeleton } from '~/components/ui/skeleton'

withDefaults(defineProps<{
  rows?: number
  block?: boolean
}>(), {
  rows: 3,
  block: false,
})
</script>

<template>
  <div :class="block ? 'w-full space-y-2' : 'space-y-2'">
    <Skeleton
      v-for="i in rows"
      :key="i"
      class="h-4"
      :class="block ? 'w-full' : ''"
      :style="!block ? `width: ${100 - i * 8}%` : ''"
    />
  </div>
</template>
