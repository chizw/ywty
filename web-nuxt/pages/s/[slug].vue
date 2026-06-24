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

const password = ref('')
const unlocking = ref(false)
const errorMsg = ref('')

async function unlock() {
  if (!password.value) return
  unlocking.value = true
  errorMsg.value = ''
  try {
    const res = await $fetch<any>(`/s/${slug}`, {
      query: { password: password.value },
      baseURL: useRuntimeConfig().apiBase as string,
    })
    data.value = res
  } catch (e: any) {
    errorMsg.value = e?.statusMessage || e?.message || '密码错误'
  } finally {
    unlocking.value = false
  }
}

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
  <div class="min-h-screen bg-background">
    <header class="bg-card border-b border-border">
      <div class="max-w-5xl mx-auto px-4 h-14 flex items-center justify-between">
        <NuxtLink to="/" class="text-lg font-semibold text-primary">云雾图驿</NuxtLink>
        <span class="text-sm text-muted-foreground">分享</span>
      </div>
    </header>
    <main class="max-w-5xl mx-auto px-4 py-8">
      <div v-if="!data" class="text-center text-muted-foreground">分享不存在或已过期</div>
      <Card v-else-if="needPwd" class="max-w-md mx-auto">
        <CardHeader>
          <CardTitle>{{ share?.type === 'photo' ? '图片分享' : '相册分享' }}</CardTitle>
          <CardDescription>该分享需要密码访问</CardDescription>
        </CardHeader>
        <CardContent>
          <form class="space-y-4" @submit.prevent="unlock">
            <div class="space-y-2">
              <Label for="share-pwd">密码</Label>
              <Input id="share-pwd" v-model="password" type="password" placeholder="密码" autofocus />
            </div>
            <Alert v-if="errorMsg" variant="destructive">
              <AlertDescription>{{ errorMsg }}</AlertDescription>
            </Alert>
            <Button type="submit" :loading="unlocking" class="w-full">解锁</Button>
          </form>
        </CardContent>
      </Card>
      <div v-else>
        <h1 class="text-2xl font-bold text-foreground">
          {{ share?.type === 'photo' ? '图片分享' : '相册分享' }}
        </h1>
        <p class="mt-1 text-sm text-muted-foreground">共 {{ items.length }} 项 · 浏览 {{ share?.view_count ?? 0 }} 次</p>
        <div v-if="share?.type === 'photo'" class="mt-6 grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-4">
          <a
            v-for="it in items"
            :key="it.id"
            :href="`/uploads/${it.pathname}`"
            target="_blank"
            class="block aspect-square bg-muted rounded overflow-hidden"
          >
            <img :src="`/uploads/${it.pathname}`" :alt="it.name" class="w-full h-full object-cover" />
          </a>
        </div>
        <div v-else class="mt-6 grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4">
          <Card v-for="it in items" :key="it.id">
            <CardContent class="p-4">
              <h3 class="font-medium text-foreground">{{ it.name }}</h3>
              <p v-if="it.intro" class="mt-1 text-sm text-muted-foreground line-clamp-2">{{ it.intro }}</p>
            </CardContent>
          </Card>
        </div>
      </div>
    </main>
  </div>
</template>
