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
    <div class="text-7xl font-bold text-gray-300">404</div>
    <h1 class="mt-4 text-2xl font-semibold text-gray-800">页面未找到</h1>
    <p class="mt-2 text-sm text-gray-500">您访问的页面不存在或已被移除。</p>
    <div class="mt-8 flex items-center justify-center gap-3">
      <NuxtLink
        to="/"
        class="px-6 py-2.5 bg-primary-600 text-white rounded-md hover:bg-primary-700"
      >
        {{ t('nav.home') }}
      </NuxtLink>
      <NuxtLink
        to="/explore"
        class="px-6 py-2.5 border border-gray-300 text-gray-700 rounded-md hover:bg-gray-50"
      >
        {{ t('nav.explore') }}
      </NuxtLink>
    </div>
  </div>
</template>
