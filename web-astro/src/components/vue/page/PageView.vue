<script setup lang="ts">
// 自定义单页（静态模式客户端渲染）：/page/{slug} 未知 slug 时由服务器回退到本页
import { onMounted, ref } from 'vue'
import { useApi } from '../../../lib/api'

interface PageData {
  title?: string
  name?: string
  content?: string | null
  description?: string | null
  keywords?: string | null
}

const page = ref<PageData | null>(null)
const loading = ref(true)
const errorMsg = ref('')

function resolveSlug(): string {
  if (typeof window === 'undefined') return ''
  const q = new URLSearchParams(window.location.search).get('slug')
  if (q) return q
  const m = window.location.pathname.match(/\/page\/([^/?#]+)/)
  return m ? decodeURIComponent(m[1]) : ''
}

onMounted(async () => {
  const slug = resolveSlug()
  if (!slug) {
    loading.value = false
    errorMsg.value = '页面不存在'
    return
  }
  try {
    const res = await useApi().get<any>(`/api/v1/pages/${encodeURIComponent(slug)}`, { raw: true })
    const d = res?.data ?? res ?? {}
    page.value = Array.isArray(d) ? (d[0] ?? null) : d
    if (page.value?.title) document.title = `${page.value.title}`
  } catch {
    errorMsg.value = '页面不存在或已被删除'
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="mx-auto max-w-3xl px-6 py-12">
    <div v-if="loading" class="py-16 text-center text-sm text-muted-foreground">加载中…</div>

    <div v-else-if="errorMsg" class="py-16 text-center">
      <p class="text-lg text-muted-foreground">{{ errorMsg }}</p>
      <a href="/" class="mt-4 inline-block text-sm text-brand hover:underline">返回首页</a>
    </div>

    <article v-else-if="page">
      <h1 class="font-display text-3xl font-bold tracking-tight">{{ page.title || page.name }}</h1>
      <p v-if="page.description" class="mt-2 text-sm text-muted-foreground">{{ page.description }}</p>
      <div class="mt-8 leading-relaxed text-foreground/90" v-html="page.content || ''" />
    </article>

    <p v-else class="py-16 text-center text-sm text-muted-foreground">暂无内容</p>
  </div>
</template>
