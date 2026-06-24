<script setup lang="ts">
// 相册详情：批量移除照片 / 设置封面 / 添加图片
definePageMeta({ layout: 'dashboard', middleware: 'auth' })

import { ArrowLeft, Plus, Check, Image as ImageIcon } from '@lucide/vue'

const route = useRoute()
const albumId = Number(route.params.id)
const api = useApi()
const message = useMessage()

const album = ref<any>(null)
const photosRaw = ref<any>(null)
const loading = ref(false)

async function fetchAlbum() {
  try { album.value = await api.get<any>(`/api/v1/albums/${albumId}`).catch(() => null) } catch { album.value = null }
}

async function fetchPhotos() {
  loading.value = true
  try { photosRaw.value = await api.get<any>(`/api/v1/albums/${albumId}/photos`).catch(() => ({ data: [] })) } catch { photosRaw.value = { data: [] } } finally { loading.value = false }
}

const photoList = computed<any[]>(() => {
  const d = photosRaw.value
  return Array.isArray(d) ? d : ((d as any)?.data ?? [])
})

onMounted(() => { fetchAlbum(); fetchPhotos() })

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
  fetchPhotos()
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
    fetchPhotos()
  } catch (err: any) {
    message.error(err?.statusMessage || '添加失败')
  }
}

async function removeFromAlbum(photoId: number) {
  try {
    await api.del(`/api/v1/albums/${albumId}/photos/${photoId}`)
    message.success('已移除')
    fetchPhotos()
  } catch (err: any) {
    message.error(err?.statusMessage || '移除失败')
  }
}

async function setCover(photoId: number) {
  try {
    await api.request(`/api/v1/albums/${albumId}`, { method: 'PATCH', body: { cover_photo_id: photoId } })
    message.success('已设为封面')
    fetchAlbum()
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
        <NuxtLink to="/dashboard/albums" class="text-xs text-muted-foreground hover:text-primary flex items-center gap-1">
          <ArrowLeft class="h-3 w-3" />
          返回相册列表
        </NuxtLink>
        <h1 class="text-2xl font-bold text-foreground mt-1">{{ album.name }}</h1>
        <p v-if="album.intro" class="mt-1 text-sm text-muted-foreground">{{ album.intro }}</p>
      </div>
      <Button
        :variant="selectMode ? 'default' : 'outline'"
        @click="toggleSelectMode"
      >{{ selectMode ? '退出多选' : '多选' }}</Button>
    </div>

    <Card class="mb-6">
      <CardHeader>
        <CardTitle class="text-sm font-medium">添加已有图片到该相册</CardTitle>
      </CardHeader>
      <CardContent>
        <form class="flex gap-2" @submit.prevent="moveIn">
          <Input v-model="newPhotoId" type="number" min="1" placeholder="图片 ID" class="flex-1" />
          <Button type="submit">
            <Plus class="mr-1 h-4 w-4" />
            添加
          </Button>
        </form>
      </CardContent>
    </Card>

    <!-- 批量操作栏 -->
    <div v-if="selectMode" class="mb-4 flex items-center gap-3">
      <span class="text-sm text-muted-foreground">已选 <b class="text-primary">{{ selectedIds.length }}</b> 项</span>
      <Button variant="outline" size="sm" @click="toggleSelectAll">
        {{ allSelected ? '取消全选' : '全选' }}
      </Button>
      <Button
        variant="destructive"
        size="sm"
        :disabled="selectedIds.length === 0"
        @click="batchRemove"
      >批量移除</Button>
    </div>

    <AppEmpty v-if="photoList.length === 0" title="相册内还没有图片" />
    <div v-else class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-3">
      <div
        v-for="p in photoList"
        :key="p.id"
        class="group relative aspect-square bg-muted rounded overflow-hidden cursor-pointer"
        :class="{ 'ring-2 ring-primary': selectMode && selectedIds.includes(p.id) }"
        @click="selectMode ? toggleSelect(p.id) : null"
      >
        <img :src="`/uploads/${p.pathname}`" :alt="p.name" class="w-full h-full object-cover" />

        <!-- 封面标记 -->
        <Badge
          v-if="isCover(p)"
          variant="default"
          class="absolute top-1 left-1 text-[10px]"
        >封面</Badge>

        <!-- 多选勾选框 -->
        <div v-if="selectMode" class="absolute top-1 right-1 z-10">
          <span
            class="inline-flex items-center justify-center w-5 h-5 rounded border-2 text-white text-xs"
            :class="selectedIds.includes(p.id) ? 'bg-primary border-primary' : 'bg-black/30 border-white'"
          >{{ selectedIds.includes(p.id) ? '✓' : '' }}</span>
        </div>

        <!-- 单张操作（非多选模式） -->
        <div v-else class="absolute inset-0 bg-black bg-opacity-0 group-hover:bg-opacity-50 transition flex flex-col justify-between p-2 opacity-0 group-hover:opacity-100">
          <div class="flex justify-end gap-1">
            <Button
              variant="secondary"
              size="sm"
              class="h-6 px-1.5 text-[10px]"
              :disabled="isCover(p)"
              :class="{ 'opacity-40 cursor-not-allowed': isCover(p) }"
              @click.stop="setCover(p.id)"
            >
              <ImageIcon class="mr-1 h-3 w-3" />
              设为封面
            </Button>
          </div>
          <div class="flex justify-end">
            <Button
              variant="destructive"
              size="sm"
              class="h-6 px-1.5 text-[10px]"
              @click.stop="removeFromAlbum(p.id)"
            >解绑</Button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
