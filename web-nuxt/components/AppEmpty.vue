<script setup lang="ts">
// 通用空状�?import { Inbox } from '@lucide/vue'

withDefaults(defineProps<{
  title?: string
  description?: string
  icon?: string
}>(), {
  title: '暂无数据',
  description: '',
  icon: '',
})
</script>

<template>
  <div class="flex flex-col items-center justify-center py-16 text-center">
    <div class="mb-4">
      <slot name="icon">
        <span v-if="icon" class="text-5xl">{{ icon }}</span>
        <Inbox v-else class="h-12 w-12 text-muted-foreground" />
      </slot>
    </div>
    <h3 class="text-lg font-medium text-foreground">{{ title }}</h3>
    <p v-if="description" class="mt-2 text-sm text-muted-foreground max-w-md">{{ description }}</p>
    <div class="mt-6">
      <slot />
    </div>
  </div>
</template>
