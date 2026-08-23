<script setup lang="ts">
// Hero 右侧「相片标本卡」：雾霭相片窗 + 直链 Markdown 标本（带复制）
import { ref } from 'vue'

const demoUrl = 'https://yunwu.site/i/ab12cd'
const demoMd = `![晨雾](https://yunwu.site/i/ab12cd)`
const copied = ref(false)
const copyError = ref(false)

async function copyMarkdown() {
  try {
    await navigator.clipboard.writeText(demoMd)
    copied.value = true
    copyError.value = false
    setTimeout(() => (copied.value = false), 1600)
  } catch {
    copyError.value = true
    setTimeout(() => (copyError.value = false), 1600)
  }
}
</script>

<template>
  <figure class="relative w-full max-w-md">
    <!-- 卡片后光晕 -->
    <div class="animate-mist absolute -inset-5 rounded-full bg-brand/[0.06] blur-2xl" aria-hidden="true" />

    <div class="relative rotate-[0.4deg] rounded-[0.45rem] border hairline bg-card p-3 shadow-[0_24px_50px_-28px_hsl(var(--brand)/0.35)]">
      <!-- 相片窗：雾霭 + 山形剪影 -->
      <div class="relative aspect-[16/10] overflow-hidden rounded-[0.25rem] border hairline bg-muted/60">
        <div class="animate-mist absolute inset-0" aria-hidden="true">
          <div class="absolute -left-1/4 top-1/3 h-36 w-3/4 rounded-full bg-brand/10 blur-2xl" />
          <div class="absolute -right-1/4 top-1/2 h-44 w-3/4 rounded-full bg-secondary/50 blur-2xl" />
          <div class="absolute -bottom-1/3 left-0 h-24 w-full rounded-full bg-brand/5 blur-2xl" />
        </div>
        <svg
          class="absolute bottom-0 left-0 h-2/5 w-full text-foreground/[0.08]"
          viewBox="0 0 400 100"
          preserveAspectRatio="none"
          aria-hidden="true"
        >
          <path d="M0 74 L64 34 L124 64 L196 20 L276 66 L336 40 L400 58 L400 100 L0 100 Z" fill="currentColor" />
          <path d="M0 90 L88 54 L158 82 L238 46 L328 80 L400 62 L400 100 L0 100 Z" fill="currentColor" opacity="0.55" />
        </svg>
        <span class="absolute left-2.5 top-2 text-[0.6rem] font-mono tracking-[0.2em] text-muted-foreground/70">
          YUNWU · SPECIMEN
        </span>
        <span class="absolute right-2.5 top-2.5 h-1.5 w-1.5 rounded-full bg-brand/60" aria-hidden="true" />
      </div>

      <!-- 标本标签 -->
      <div class="pt-3">
        <div class="flex items-baseline justify-between gap-2">
          <span class="font-display text-sm font-semibold">晨雾.jpg</span>
          <span class="shrink-0 text-[0.65rem] tabular-nums text-muted-foreground">2.4 MB · 4000×3000</span>
        </div>

        <div class="mt-2 overflow-x-auto rounded-[0.25rem] bg-foreground/[0.03] px-3 py-2.5 font-mono text-[0.72rem] leading-relaxed">
          <span class="text-muted-foreground">![晨雾](</span>
          <span class="text-brand">{{ demoUrl }}</span>
          <span class="text-muted-foreground">)</span>
        </div>

        <div class="mt-2 flex items-center justify-between">
          <span class="text-[0.65rem] font-mono tabular-nums text-muted-foreground">YUNWU-0001</span>
          <button
            type="button"
            class="inline-flex items-center gap-1 text-[0.72rem] transition-colors"
            :class="copyError ? 'text-destructive' : copied ? 'text-brand' : 'text-foreground hover:text-brand'"
            @click="copyMarkdown"
          >
            {{ copyError ? '复制失败，请手动复制' : copied ? '已复制 ✓' : '复制 Markdown' }}
          </button>
        </div>
      </div>
    </div>
  </figure>
</template>
