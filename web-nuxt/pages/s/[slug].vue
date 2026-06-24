<script setup lang="ts">
// 公开分享页
const route = useRoute()
const slug = String(route.params.slug || '')

const { data } = await useAsyncData(`share-${slug}`, async () => {
  return await $fetch<any>(`/s/${slug}`, { baseURL: useRuntimeConfig().apiBase as string })
})

const share = computed<any>(() => (data.value as any)?.data?.share)
const items = computed<any[]>(() => (data.value as any)?.data?.items ?? [])
const needPwd = computed(() => (data.value as any)?.data?.need_pwd === true)

const ogImage = computed(() => {
  const first = items.value[0]
  return first?.pathname ? `/uploads/${first.pathname}` : undefined
})

useSeoMeta({
  title: () => share.value?.type === 'photo' ? '图片分享' : (share.value?.type === 'album' ? '相册分享' : '分享'),
  description: () => `共 ${items.value.length} 项 · 浏览 ${share.value?.view_count ?? 0} 次`,
  ogTitle: () => share.value?.type === 'photo' ? '图片分享' : '相册分享',
  ogDescription: () => `共 ${items.value.length} 项`,
  ogImage: () => ogImage.value,
  ogType: 'website',
})
</script>

<template>
  <div class="min-h-screen bg-gray-50">
    <header class="bg-white border-b border-gray-200">
      <div class="max-w-5xl mx-auto px-4 h-14 flex items-center justify-between">
        <NuxtLink to="/" class="text-lg font-semibold text-primary-600">ywty</NuxtLink>
        <span class="text-sm text-gray-500">分享</span>
      </div>
    </header>
    <main class="max-w-5xl mx-auto px-4 py-8">
      <div v-if="!data" class="text-center text-gray-500">分享不存在或已过期</div>
      <div v-else-if="needPwd" class="max-w-md mx-auto bg-white border border-gray-200 rounded-lg p-6">
        <h1 class="text-lg font-medium mb-2">{{ share?.type === 'photo' ? '图片分享' : '相册分享' }}</h1>
        <p class="text-sm text-gray-500">该分享需要密码访问</p>
        <form class="mt-4" @submit.prevent="() => $fetch(`/s/${slug}?password=${(($event.target as HTMLFormElement).querySelector('input') as HTMLInputElement).value}`).then((r) => data = r)">
          <input type="password" placeholder="密码" class="w-full px-3 py-2 border border-gray-300 rounded-md" />
          <AppButton type="submit" block class="mt-3">解锁</AppButton>
        </form>
      </div>
      <div v-else>
        <h1 class="text-2xl font-bold text-gray-900">
          {{ share?.type === 'photo' ? '图片分享' : '相册分享' }}
        </h1>
        <p class="mt-1 text-sm text-gray-500">共 {{ items.length }} 项 · 浏览 {{ share?.view_count ?? 0 }} 次</p>
        <div v-if="share?.type === 'photo'" class="mt-6 grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-4">
          <a
            v-for="it in items"
            :key="it.id"
            :href="`/uploads/${it.pathname}`"
            target="_blank"
            class="block aspect-square bg-gray-100 rounded overflow-hidden"
          >
            <img :src="`/uploads/${it.pathname}`" :alt="it.name" class="w-full h-full object-cover" />
          </a>
        </div>
        <div v-else class="mt-6 grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4">
          <div v-for="it in items" :key="it.id" class="bg-white border border-gray-200 rounded-lg p-4">
            <h3 class="font-medium">{{ it.name }}</h3>
            <p v-if="it.intro" class="mt-1 text-sm text-gray-500 line-clamp-2">{{ it.intro }}</p>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>
