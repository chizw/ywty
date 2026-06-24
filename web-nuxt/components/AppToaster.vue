<script setup lang="ts">
// AppToaster 全局消息提示渲染器（配合 useMessage 使用）
import { cn } from '~/lib/utils'

const { toasts } = useMessage()

const kindClasses: Record<string, string> = {
  success: 'border-green-500 bg-green-500 text-white',
  error: 'border-destructive bg-destructive text-destructive-foreground',
  warning: 'border-yellow-500 bg-yellow-500 text-white',
  info: 'border-foreground bg-foreground text-background',
}
</script>

<template>
  <div class="fixed top-4 right-4 z-[9999] flex flex-col gap-2 max-w-sm pointer-events-none">
    <transition-group name="toast" tag="div" class="flex flex-col gap-2">
      <div
        v-for="t in toasts"
        :key="t.id"
        class="pointer-events-auto px-4 py-3 rounded-md shadow-lg text-sm border"
        :class="cn(kindClasses[t.kind] || kindClasses.info)"
      >
        {{ t.text }}
      </div>
    </transition-group>
  </div>
</template>

<style scoped>
.toast-enter-active, .toast-leave-active {
  transition: opacity 0.2s, transform 0.2s;
}
.toast-enter-from, .toast-leave-to {
  opacity: 0;
  transform: translateX(20px);
}
</style>
