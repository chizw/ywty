<script setup lang="ts">
// 通用加载占位
withDefaults(defineProps<{
  rows?: number
  block?: boolean
}>(), {
  rows: 3,
  block: false,
})
</script>

<template>
  <div :class="block ? 'w-full' : 'inline-block'">
    <div
      v-for="i in rows"
      :key="i"
      class="bg-gray-200 rounded animate-pulse mb-2 last:mb-0"
      :class="block ? 'h-4 w-full' : 'h-4'"
      :style="!block ? `width: ${100 - i * 8}%` : ''"
    />
  </div>
</template>
