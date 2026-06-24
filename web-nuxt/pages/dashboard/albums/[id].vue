<script setup lang="ts">
// 相册详情：批量移除照片 / 设置封面 / 添加图片
definePageMeta({ layout: 'dashboard', middleware: 'auth' })

const route = useRoute()
const albumId = Number(route.params.id)
const api = useApi()
const message = useMessage()

const { data: album, refresh: refreshAlbum } = await useAsyncData(`album-${albumId}`, () =>
  api.get<any>(`/api/v1/albums/${albumId}`).catch(() => null)
)
const { data: photos, refresh: refreshPhotos } = await useAsyncData(`album-photos-${albumId}`, () =>
  api.get<any>(`/api/v1/albums/${albumId}/photos`).catch(() => ({ data: [] }))
)

const photoList = computed<any[]>(() => {
  const d = photos.value
  return Array.isArray(d) ? d : ((d as any)?.data ?? [])
})

const newPhotoId = ref('')

// 多选 + 批量移除
const selectMode = ref(false)
const selectedIds = ref<number[]>([])

const allIds = computed(() => photoList.value.map((p) => p.id))
const allSelected = computed(() => allIds.value.length > 0 && selectedIds.value.length === allIds.value.length)

function toggleSelectMode() {
  selectMode.value = !selectMode.value
  if (!selectMode.value) selectedIds.value = []
}

function toggleSelect(id: number) {
  const idx = selectedIds.value.indexOf(id)
  if (idx >= 0) selectedIds.value.splice(idx, 1)
  else selectedIds.value.push(id)
}

function toggleSelectAll() {
  if (allSelected.value) selectedIds.value = []
  else selectedIds.value = [...allIds.value]
}

async function batchRemove() {
  if (selectedIds.value.length === 0) return
  if (!confirm(`确定从相册移除选中的 ${selectedIds.value.length} 张图片？（不会删除原图）`)) return
  let ok = 0
  let fail = 0
  for (const id of selectedIds.value) {
    try {
      await api.del(`/api/v1/albums/${albumId}/photos/${id}`)
      ok++
    } catch {
      fail++
    }
  }
  if (fail === 0) message.success(`已移除 ${ok} 张`)
  else message.warning(`成功 ${ok} 张，失败 ${fail} 张`)
  selectedIds.value = []
  refreshPhotos()
}

async function moveIn() {
  const id = Number(newPhotoId.value)
  if (!id) {
    message.warning('请填写图片 ID')
    return
  }
  try {
    await api.post(`/api/v1/photos/${id}/move-to-album`, { album_id: albumId })
    message.success('已添加到相册')
    newPhotoId.value = ''
    refreshPhotos()
  } catch (err: any) {
    message.error(err?.statusMessage || '添加失败')
  }
}

async function removeFromAlbum(photoId: number) {
  try {
    await api.del(`/api/v1/albums/${albumId}/photos/${photoId}`)
    message.success('已移除')
    refreshPhotos()
  } catch (err: any) {
    message.error(err?.statusMessage || '移除失败')
  }
}

async function setCover(photoId: number) {
  try {
    await api.request(`/api/v1/albums/${albumId}`, { method: 'PATCH', body: { cover_photo_id: photoId } })
    message.success('已设为封面')
    refreshAlbum()
  } catch (err: any) {
    message.error(err?.statusMessage || '设置封面失败')
  }
}

function isCover(p: any) {
  const a = album.value as any
  return a && (a.cover_photo_id === p.id || a.cover === p.pathname)
}
</script>

<template>
  <div v-if="album">
    <div class="flex items-center justify-between mb-4">
      <div>
        <NuxtLink to="/dashboard/albums" class="text-xs text-gray-500 hover:text-primary-600">← 返回相册列表</NuxtLink>
        <h1 class="text-2xl font-bold text-gray-900 mt-1">{{ album.name }}</h1>
        <p v-if="album.intro" class="mt-1 text-sm text-gray-500">{{ album.intro }}</p>
      </div>
      <button
        class="px-3 py-1.5 text-sm rounded-md border"
        :class="selectMode ? 'bg-primary-600 text-white border-primary-600' : 'border-gray-300 text-gray-700 hover:bg-gray-50'"
        @click="toggleSelectMode"
      >{{ selectMode ? '退出多选' : '多选' }}</button>
    </div>

    <div class="bg-white border border-gray-200 rounded-lg p-4 mb-6">
      <h3 class="text-sm font-medium text-gray-700 mb-2">添加已有图片到该相册</h3>
      <form class="flex gap-2" @submit.prevent="moveIn">
        <input v-model="newPhotoId" type="number" min="1" placeholder="图片 ID" class="flex-1 px-3 py-2 border border-gray-300 rounded-md" />
        <AppButton type="submit">添加</AppButton>
      </form>
    </div>

    <!-- 批量操作栏 -->
    <div v-if="selectMode" class="mb-4 flex items-center gap-3">
      <span class="text-sm text-gray-600">已选 <b class="text-primary-600">{{ selectedIds.length }}</b> 项</span>
      <button class="px-3 py-1.5 text-sm border border-gray-300 rounded-md text-gray-700 hover:bg-gray-50" @click="toggleSelectAll">
        {{ allSelected ? '取消全选' : '全选' }}
      </button>
      <button
        class="px-3 py-1.5 text-sm bg-red-500 text-white rounded-md disabled:opacity-50"
        :disabled="selectedIds.length === 0"
        @click="batchRemove"
      >批量移除</button>
    </div>

    <AppEmpty v-if="photoList.length === 0" title="相册内还没有图片" />
    <div v-else class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-3">
      <div
        v-for="p in photoList"
        :key="p.id"
        class="group relative aspect-square bg-gray-100 rounded overflow-hidden cursor-pointer"
        :class="{ 'ring-2 ring-primary-500': selectMode && selectedIds.includes(p.id) }"
        @click="selectMode ? toggleSelect(p.id) : null"
      >
        <img :src="`/uploads/${p.pathname}`" :alt="p.name" class="w-full h-full object-cover" />

        <!-- 封面标记 -->
        <span
          v-if="isCover(p)"
          class="absolute top-1 left-1 px-1.5 py-0.5 bg-primary-500 text-white text-[10px] rounded"
        >封面</span>

        <!-- 多选勾选框 -->
        <div v-if="selectMode" class="absolute top-1 right-1 z-10">
          <span
            class="inline-flex items-center justify-center w-5 h-5 rounded border-2 text-white text-xs"
            :class="selectedIds.includes(p.id) ? 'bg-primary-600 border-primary-600' : 'bg-black/30 border-white'"
          >{{ selectedIds.includes(p.id) ? '✓' : '' }}</span>
        </div>

        <!-- 单张操作（非多选模式） -->
        <div v-else class="absolute inset-0 bg-black bg-opacity-0 group-hover:bg-opacity-50 transition flex flex-col justify-between p-2 opacity-0 group-hover:opacity-100">
          <div class="flex justify-end gap-1">
            <button
              class="px-1.5 py-0.5 bg-white text-[10px] rounded hover:bg-gray-100"
              :disabled="isCover(p)"
              :class="{ 'opacity-40 cursor-not-allowed': isCover(p) }"
              @click.stop="setCover(p.id)"
            >设为封面</button>
          </div>
          <div class="flex justify-end">
            <button
              class="px-1.5 py-0.5 bg-red-500 text-white text-[10px] rounded"
              @click.stop="removeFromAlbum(p.id)"
            >解绑</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
