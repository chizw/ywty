<script setup lang="ts">
// 单页 CMS 渲染：通过 slug 获取页面并渲染 HTML
const route = useRoute()
const slug = route.params.slug as string
const api = useApi()

const { data: page } = await useAsyncData(`page-${slug}`, async () => {
  try {
    return await api.get<any>(`/api/v1/pages/${slug}`)
  } catch {
    return null
  }
})

useSeoMeta({
  title: () => page.value?.title || '页面',
  description: () => page.value?.description || page.value?.intro || '',
  ogTitle: () => page.value?.title || '页面',
  ogDescription: () => page.value?.description || page.value?.intro || '',
  ogType: 'article',
})
</script>

<template>
  <div class="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 py-10">
    <template v-if="page">
      <h1 class="text-3xl font-bold text-foreground">{{ page.title }}</h1>
      <p v-if="page.updated_at" class="mt-2 text-xs text-muted-foreground">更新于 {{ page.updated_at }}</p>
      <div
        class="prose prose-sm sm:prose lg:prose-lg max-w-none mt-6 cms-content"
        v-html="page.content || ''"
      />
    </template>
    <AppEmpty v-else title="页面不存在" />
  </div>
</template>

<style scoped>
.cms-content :deep(img) {
  max-width: 100%;
  height: auto;
}
.cms-content :deep(a) {
  color: hsl(var(--primary));
  text-decoration: underline;
}
</style>
