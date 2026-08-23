<script setup lang="ts">
// 全局 toast 渲染器（配合 vue-store 的 message）
import { CheckCircle2, AlertCircle, AlertTriangle, Info } from '@lucide/vue'
import { messageState } from '../../../lib/vue-store'
import type { ToastKind } from '../../../lib/vue-store'
import { cn } from '../../../lib/utils'

const icons: Record<ToastKind, unknown> = {
  success: CheckCircle2,
  error: AlertCircle,
  warning: AlertTriangle,
  info: Info,
}

const classes: Record<ToastKind, string> = {
  success: 'border-green-500/40 bg-green-500/95 text-white',
  error: 'border-destructive/40 bg-destructive/95 text-destructive-foreground',
  warning: 'border-yellow-500/40 bg-yellow-500/95 text-white',
  info: 'border-border bg-card text-card-foreground shadow-lg',
}
</script>

<template>
  <Teleport to="body">
    <div class="pointer-events-none fixed right-4 top-4 z-[9999] flex max-w-sm flex-col gap-2">
      <TransitionGroup name="toast" tag="div" class="flex flex-col gap-2">
        <div
          v-for="t in messageState.toasts"
          :key="t.id"
          class="pointer-events-auto flex items-center gap-2 rounded-md border px-4 py-3 text-sm shadow-lg"
          :class="cn(classes[t.kind])"
        >
          <component :is="icons[t.kind]" class="h-4 w-4 shrink-0" />
          <span>{{ t.text }}</span>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>

<style scoped>
.toast-enter-active,
.toast-leave-active {
  transition:
    opacity 0.2s,
    transform 0.2s;
}
.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translateX(20px);
}
</style>
