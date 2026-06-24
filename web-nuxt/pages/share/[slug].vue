<script setup lang="ts">
// 分享公开页：通过 slug 查看分享的照片/相册
definePageMeta({ layout: 'default' })

const route = useRoute()
const slug = route.params.slug as string
const api = useApi()

const form = reactive({ password: '' })
const showPasswordForm = ref(false)
const loading = ref(true)
const errorMsg = ref('')

const { data: share, refresh } = await useAsyncData(`share-${slug}`, async () => {
  try {
    return await api.get<any>(`/api/v1/shares/${slug}`)
  } catch (err: any) {
    if (err?.status === 403 || err?.statusMessage?.includes('密码')) {
      showPasswordForm.value = true
    } else {
      errorMsg.value = err?.statusMessage || '分享不存在或已过期'
    }
    return null
  }
})

async function submitPassword() {
  if (!form.password) return
  loading.value = true
  errorMsg.value = ''
  try {
    await api.post(`/api/v1/shares/${slug}/verify`, { password: form.password })
    showPasswordForm.value = false
    await refresh()
  } catch (err: any) {
    errorMsg.value = err?.statusMessage || '密码错误'
  } finally {
    loading.value = false
  }
}

// Lightbox
const lightboxVisible = ref(false)
const lightboxIndex = ref(0)
function openLightbox(idx: number) {
  lightboxIndex.value = idx
  lightboxVisible.value = true
}

const photos = computed<any[]>(() => {
  if (!share.value) return []
  if (share.value.type === 'photo') {
    return share.value.photos || [share.value.photo]
  }
  if (share.value.type === 'album') {
    return share.value.album?.photos || []
  }
  return []
})

useSeoMeta({
  title: () => share.value?.title || '分享',
  description: () => share.value?.description || '查看分享的内容',
  ogTitle: () => share.value?.title || '分享',
  ogDescription: () => share.value?.description || '查看分享的内容',
  ogType: 'article',
})

onMounted(() => {
  loading.value = false
})
</script>

<template>
  <div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <!-- 加载中 -->
    <div v-if="loading" class="text-center py-24 text-gray-500">
      加载中...
    </div>

    <!-- 错误提示 -->
    <AppEmpty v-else-if="errorMsg && !share" :title="errorMsg" />

    <!-- 需要密码 -->
    <div v-else-if="showPasswordForm" class="max-w-md mx-auto py-16">
      <h1 class="text-2xl font-bold text-gray-900 mb-4 text-center">此分享需要密码</h1>
      <form class="space-y-4" @submit.prevent="submitPassword">
        <input
          v-model="form.password"
          type="password"
          placeholder="请输入访问密码"
          class="w-full px-3 py-2 border border-gray-300 rounded-md"
          autofocus
        />
        <p v-if="errorMsg" class="text-sm text-red-500">{{ errorMsg }}</p>
        <AppButton type="submit" :loading="loading" block>确认</AppButton>
      </form>
    </div>

    <!-- 分享内容 -->
    <template v-else-if="share">
      <!-- 分享信息 -->
      <div class="bg-white border border-gray-200 rounded-lg p-6 mb-6">
        <h1 class="text-2xl font-bold text-gray-900">{{ share.title || '分享' }}</h1>
        <p v-if="share.description" class="mt-2 text-sm text-gray-600">{{ share.description }}</p>
        <div class="mt-4 flex flex-wrap gap-4 text-xs text-gray-500">
          <span>由 {{ share.user?.name || share.user?.username || '匿名' }} 分享</span>
          <span v-if="share.expire_at">有效期至 {{ new Date(share.expire_at).toLocaleString() }}</span>
          <span>共 {{ photos.length }} 项</span>
        </div>
      </div>

      <!-- 照片列表 -->
      <AppEmpty v-if="photos.length === 0" title="分享内容为空" />
      <PhotoMasonry
        v-else
        :photos="photos"
        @click="({ index }) => openLightbox(index)"
      />

      <PhotoLightbox
        v-model:visible="lightboxVisible"
        v-model:index="lightboxIndex"
        :photos="photos"
      />
    </template>
  </div>
</template>
