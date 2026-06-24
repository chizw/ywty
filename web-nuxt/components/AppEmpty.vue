<script setup lang="ts">
// 通用空状态
withDefaults(defineProps<{
  title?: string
  description?: string
  icon?: string
}>(), {
  title: '暂无数据',
  description: '',
  icon: '📭',
})
</script>

<template>
  <div class="flex flex-col items-center justify-center py-16 text-center">
    <div class="text-5xl mb-4">{{ icon }}</div>
    <h3 class="text-lg font-medium text-gray-800">{{ title }}</h3>
    <p v-if="description" class="mt-2 text-sm text-gray-500 max-w-md">{{ description }}</p>
    <div class="mt-6">
      <slot />
    </div>
  </div>
</template>
