<script setup lang="ts">
// 公开端底部：站名 + 版权 + 可选页脚文案 / ICP 备案（均由 .astro 布局传入）
withDefaults(
  defineProps<{
    /** 站点名称 */
    name?: string
    /** 自定义页脚文案（非空时展示） */
    footer?: string
    /** ICP 备案号（非空时展示） */
    icp?: string
  }>(),
  { name: '', footer: '', icp: '' },
)

const year = new Date().getFullYear()
</script>

<template>
  <footer class="border-t hairline py-8">
    <div class="container-site flex flex-col items-center justify-between gap-3 text-sm text-muted-foreground sm:flex-row">
      <div class="flex items-center gap-2">
        <span class="font-display text-[0.95rem] font-semibold text-foreground">{{ name || '驿' }}</span>
        <span class="text-xs">&copy; {{ year }}</span>
      </div>
      <div class="flex items-center gap-5">
        <a href="/page/about" class="transition-colors hover:text-foreground">关于</a>
        <a href="/page/terms" class="transition-colors hover:text-foreground">条款</a>
        <a href="/page/privacy" class="transition-colors hover:text-foreground">隐私</a>
      </div>
    </div>
    <p v-if="footer" class="container-site mt-4 whitespace-pre-line text-center text-xs leading-relaxed text-muted-foreground">
      {{ footer }}
    </p>
    <p v-if="icp" class="container-site mt-1.5 text-center text-xs text-muted-foreground/70">
      <a
        href="https://beian.miit.gov.cn/"
        target="_blank"
        rel="noopener noreferrer"
        class="transition-colors hover:text-muted-foreground"
      >
        {{ icp }}
      </a>
    </p>
  </footer>
</template>
