<script setup lang="ts">
// 我的相册
definePageMeta({ layout: 'dashboard', middleware: 'auth' })

const api = useApi()

const rawData = ref<any>(null)
const loading = ref(false)

async function fetchAlbums() {
  loading.value = true
  try {
    rawData.value = await api.get<any[]>('/api/v1/albums').catch(() => [] as any[])
  } catch {
    rawData.value = []
  } finally {
    loading.value = false
  }
}

const albums = computed<any[]>(() => {
  const d = rawData.value
  return Array.isArray(d) ? d : ((d as any)?.data ?? [])
})

onMounted(() => fetchAlbums())

const showCreate = ref(false)
const newAlbum = reactive({ name: '', intro: '', is_public: false })

async function create() {
  if (!newAlbum.name) return
  await api.post('/api/v1/albums', newAlbum)
  newAlbum.name = ''
  newAlbum.intro = ''
  newAlbum.is_public = false
  showCreate.value = false
  fetchAlbums()
}

async function remove(id: number) {
  if (!confirm('确定删除该相册？')) return
  await api.del(`/api/v1/albums/${id}`)
  fetchAlbums()
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-4">
      <h1 class="text-2xl font-bold text-gray-900">我的相册</h1>
      <AppButton @click="showCreate = !showCreate">{{ showCreate ? '取消' : '新建相册' }}</AppButton>
    </div>

    <div v-if="showCreate" class="mb-6 p-4 bg-white border border-gray-200 rounded-lg space-y-3">
      <input v-model="newAlbum.name" placeholder="相册名" class="w-full px-3 py-2 border border-gray-300 rounded-md" />
      <textarea v-model="newAlbum.intro" placeholder="介绍（可选）" class="w-full px-3 py-2 border border-gray-300 rounded-md" rows="2" />
      <label class="flex items-center gap-2 text-sm text-gray-700">
        <input v-model="newAlbum.is_public" type="checkbox" />
        公开相册
      </label>
      <AppButton @click="create">创建</AppButton>
    </div>

    <AppEmpty v-if="albums.length === 0" title="还没有相册" description="点击右上角新建一个相册，把图片归类管理" />
    <div v-else class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4">
      <div
        v-for="a in albums"
        :key="a.id"
        class="bg-white border border-gray-200 rounded-lg p-4"
      >
        <div class="flex items-center justify-between">
          <h3 class="font-medium text-gray-900">{{ a.name }}</h3>
          <span v-if="a.is_public" class="text-xs text-primary-600">公开</span>
        </div>
        <p v-if="a.intro" class="mt-1 text-sm text-gray-500 line-clamp-2">{{ a.intro }}</p>
        <p class="mt-2 text-xs text-gray-400">{{ a.photo_count }} 张图片</p>
        <div class="mt-3 flex gap-2">
          <NuxtLink
            :to="`/dashboard/albums/${a.id}`"
            class="px-2 py-1 text-xs border border-gray-300 rounded hover:bg-gray-50"
          >查看</NuxtLink>
          <button class="px-2 py-1 text-xs text-red-500" @click="remove(a.id)">删除</button>
        </div>
      </div>
    </div>
  </div>
</template>
