<script setup lang="ts">
// 404 兜底页
const { t } = useI18n()

useSeoMeta({
  title: '404 - 页面未找到',
  robots: 'noindex',
})
</script>

<template>
  <div class="max-w-3xl mx-auto px-4 py-24 text-center">
    <div class="text-7xl font-bold text-muted-foreground/40">404</div>
    <h1 class="mt-4 text-2xl font-semibold text-foreground">页面未找到</h1>
    <p class="mt-2 text-sm text-muted-foreground">您访问的页面不存在或已被移除。</p>
    <div class="mt-8 flex items-center justify-center gap-3">
      <Button as="NuxtLink" to="/" size="lg">
        {{ t('nav.home') }}
      </Button>
      <Button as="NuxtLink" to="/explore" variant="outline" size="lg">
        {{ t('nav.explore') }}
      </Button>
    </div>
  </div>
</template>
