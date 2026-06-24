<script setup lang="ts">
// AppToaster 全局消息提示渲染器（配合 useMessage 使用）
const { toasts } = useMessage()
</script>

<template>
  <div class="fixed top-4 right-4 z-[9999] flex flex-col gap-2 max-w-sm pointer-events-none">
    <transition-group name="toast" tag="div" class="flex flex-col gap-2">
      <div
        v-for="t in toasts"
        :key="t.id"
        class="pointer-events-auto px-4 py-2 rounded-md shadow-lg text-sm text-white"
        :class="{
          'bg-emerald-500': t.kind === 'success',
          'bg-red-500': t.kind === 'error',
          'bg-amber-500': t.kind === 'warning',
          'bg-gray-700': t.kind === 'info',
        }"
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
