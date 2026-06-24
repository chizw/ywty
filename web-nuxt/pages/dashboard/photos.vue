<script setup lang="ts">
// 我的图片列表：多选 + 批量操作 + 筛选 + 排序 + 分页
definePageMeta({ layout: 'dashboard', middleware: 'auth' })

const api = useApi()
const message = useMessage()

const page = ref(1)
const perPage = 24

// 筛选条件
const filterAlbumId = ref<number | null>(null)
const filterTag = ref('')
const startDate = ref('')
const endDate = ref('')

// 排序
const sortBy = ref<'created_at' | 'size' | 'name'>('created_at')
const sortOrder = ref<'desc' | 'asc'>('desc')

// 多选
const selectMode = ref(false)
const selectedIds = ref<number[]>([])

const query = computed(() => {
  const q: Record<string, any> = {
    page: page.value,
    per_page: perPage,
    sort: sortBy.value,
    order: sortOrder.value,
  }
  if (filterAlbumId.value) q.album_id = filterAlbumId.value
  if (filterTag.value) q.tag = filterTag.value
  if (startDate.value) q.start_date = startDate.value
  if (endDate.value) q.end_date = endDate.value
  return q
})

const { data, refresh } = await useAsyncData('my-photos', () =>
  api.get<any>('/api/v1/photos', { query: query.value })
)

const photos = computed<any[]>(() => {
  const d = data.value as any
  if (Array.isArray(d)) return d
  if (d && Array.isArray(d.data)) return d.data
  return []
})
const total = computed(() => (data.value as any)?.meta?.total ?? photos.value.length)
const lastPage = computed(() => (data.value as any)?.meta?.last_page ?? Math.max(1, Math.ceil(total.value / perPage)))

// 相册 / 标签筛选项
const { data: albumsData } = await useAsyncData('photos-filter-albums', () =>
  api.get<any>('/api/v1/albums').catch(() => [])
)
const albumOptions = computed(() => {
  const d = albumsData.value as any
  const list = Array.isArray(d) ? d : ((d as any)?.data ?? [])
  return list.map((a: any) => ({ label: a.name, value: a.id }))
})

const { data: tagsData } = await useAsyncData('photos-filter-tags', () =>
  api.get<any>('/api/v1/tags').catch(() => [])
)
const tagOptions = computed(() => {
  const d = tagsData.value as any
  const list = Array.isArray(d) ? d : ((d as any)?.data ?? [])
  return list.map((t: any) => ({ label: t.name, value: t.name }))
})

const allIds = computed(() => photos.value.map((p) => p.id))

function onUploaded() {
  page.value = 1
  refresh()
}

// 筛选/排序变化时回到第一页并刷新
watch([filterAlbumId, filterTag, startDate, endDate, sortBy, sortOrder], () => {
  page.value = 1
  refresh()
})
watch(page, () => refresh())

function toggleSelectMode() {
  selectMode.value = !selectMode.value
  if (!selectMode.value) selectedIds.value = []
}

function toggleSelect(id: number) {
  const idx = selectedIds.value.indexOf(id)
  if (idx >= 0) selectedIds.value.splice(idx, 1)
  else selectedIds.value.push(id)
}

function isSelected(id: number) {
  return selectedIds.value.includes(id)
}

function onBatchDone() {
  refresh()
}

function resetFilters() {
  filterAlbumId.value = null
  filterTag.value = ''
  startDate.value = ''
  endDate.value = ''
  sortBy.value = 'created_at'
  sortOrder.value = 'desc'
}

async function remove(id: number) {
  if (!confirm('确定删除这张图片？')) return
  try {
    await api.del(`/api/v1/photos/${id}`)
    message.success('已删除')
    refresh()
  } catch (err: any) {
    message.error(err?.statusMessage || '删除失败')
  }
}

async function copy(id: number) {
  try {
    await api.post(`/api/v1/photos/${id}/copy`, {})
    message.success('已复制')
    refresh()
  } catch (err: any) {
    message.error(err?.statusMessage || '复制失败')
  }
}

async function togglePublic(p: any) {
  try {
    await api.request(`/api/v1/photos/${p.id}`, { method: 'PATCH', body: { is_public: !p.is_public } })
    refresh()
  } catch (err: any) {
    message.error(err?.statusMessage || '操作失败')
  }
}

function goPrev() {
  if (page.value > 1) page.value--
}
function goNext() {
  if (page.value < lastPage.value) page.value++
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-4">
      <h1 class="text-2xl font-bold text-gray-900">我的图片</h1>
      <div class="flex items-center gap-3">
        <span class="text-sm text-gray-500">共 {{ total }} 张</span>
        <button
          class="px-3 py-1.5 text-sm rounded-md border"
          :class="selectMode ? 'bg-primary-600 text-white border-primary-600' : 'border-gray-300 text-gray-700 hover:bg-gray-50'"
          @click="toggleSelectMode"
        >{{ selectMode ? '退出多选' : '多选' }}</button>
      </div>
    </div>

    <PhotoUploader class="mb-6" @uploaded="onUploaded" @error="(m) => message.error(m)" />

    <!-- 筛选 / 排序工具栏 -->
    <div class="bg-white border border-gray-200 rounded-lg p-3 mb-4 flex flex-wrap items-end gap-3">
      <div>
        <label class="block text-xs text-gray-500 mb-1">相册</label>
        <select v-model="filterAlbumId" class="px-2 py-1.5 border border-gray-300 rounded-md text-sm min-w-[120px]">
          <option :value="null">全部</option>
          <option v-for="a in albumOptions" :key="a.value" :value="a.value">{{ a.label }}</option>
        </select>
      </div>
      <div>
        <label class="block text-xs text-gray-500 mb-1">标签</label>
        <select v-model="filterTag" class="px-2 py-1.5 border border-gray-300 rounded-md text-sm min-w-[120px]">
          <option value="">全部</option>
          <option v-for="t in tagOptions" :key="t.value" :value="t.value">{{ t.label }}</option>
        </select>
      </div>
      <div>
        <label class="block text-xs text-gray-500 mb-1">开始日期</label>
        <input v-model="startDate" type="date" class="px-2 py-1.5 border border-gray-300 rounded-md text-sm" />
      </div>
      <div>
        <label class="block text-xs text-gray-500 mb-1">结束日期</label>
        <input v-model="endDate" type="date" class="px-2 py-1.5 border border-gray-300 rounded-md text-sm" />
      </div>
      <div>
        <label class="block text-xs text-gray-500 mb-1">排序</label>
        <select v-model="sortBy" class="px-2 py-1.5 border border-gray-300 rounded-md text-sm">
          <option value="created_at">按时间</option>
          <option value="size">按大小</option>
          <option value="name">按名称</option>
        </select>
      </div>
      <div>
        <label class="block text-xs text-gray-500 mb-1">方向</label>
        <select v-model="sortOrder" class="px-2 py-1.5 border border-gray-300 rounded-md text-sm">
          <option value="desc">降序</option>
          <option value="asc">升序</option>
        </select>
      </div>
      <button class="px-3 py-1.5 text-sm border border-gray-300 rounded-md text-gray-600 hover:bg-gray-50" @click="resetFilters">重置</button>
    </div>

    <!-- 批量操作栏 -->
    <div v-if="selectMode" class="mb-4">
      <PhotoBatchActions
        v-model:selected-ids="selectedIds"
        :all-ids="allIds"
        @done="onBatchDone"
      />
    </div>

    <AppEmpty v-if="photos.length === 0" title="还没有图片" description="拖拽或点击上方上传你的第一张图片" />
    <div v-else class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-3">
      <div
        v-for="p in photos"
        :key="p.id"
        class="group relative bg-gray-100 rounded overflow-hidden aspect-square cursor-pointer"
        :class="{ 'ring-2 ring-primary-500': selectMode && isSelected(p.id) }"
        @click="selectMode ? toggleSelect(p.id) : null"
      >
        <img :src="`/uploads/${p.pathname}`" :alt="p.name" class="w-full h-full object-cover" loading="lazy" />

        <!-- 多选勾选框 -->
        <div v-if="selectMode" class="absolute top-1 left-1 z-10">
          <span
            class="inline-flex items-center justify-center w-5 h-5 rounded border-2 text-white text-xs"
            :class="isSelected(p.id) ? 'bg-primary-600 border-primary-600' : 'bg-black/30 border-white'"
          >{{ isSelected(p.id) ? '✓' : '' }}</span>
        </div>

        <!-- 单张操作（非多选模式） -->
        <div
          v-if="!selectMode"
          class="absolute inset-0 bg-black bg-opacity-0 group-hover:bg-opacity-50 transition flex flex-col justify-between p-2 opacity-0 group-hover:opacity-100"
        >
          <div class="flex justify-end">
            <span v-if="p.is_public" class="px-1.5 py-0.5 bg-primary-500 text-white text-[10px] rounded">公开</span>
          </div>
          <div class="flex gap-1 w-full justify-end items-center flex-wrap">
            <LikeButton size="sm" target-type="photo" :target-id="p.id" />
            <ReportButton size="sm" target-type="photo" :target-id="p.id" />
            <button class="px-2 py-1 bg-white text-xs rounded hover:bg-gray-100" @click.stop="togglePublic(p)">{{ p.is_public ? '转私有' : '转公开' }}</button>
            <button class="px-2 py-1 bg-white text-xs rounded hover:bg-gray-100" @click.stop="copy(p.id)">复制</button>
            <button class="px-2 py-1 bg-red-500 text-white text-xs rounded" @click.stop="remove(p.id)">删除</button>
          </div>
        </div>
      </div>
    </div>

    <!-- 分页 -->
    <div v-if="total > perPage" class="mt-6 flex items-center justify-center gap-3 text-sm">
      <button
        class="px-3 py-1.5 border border-gray-300 rounded-md disabled:opacity-40"
        :disabled="page <= 1"
        @click="goPrev"
      >上一页</button>
      <span class="text-gray-600">第 {{ page }} / {{ lastPage }} 页</span>
      <button
        class="px-3 py-1.5 border border-gray-300 rounded-md disabled:opacity-40"
        :disabled="page >= lastPage"
        @click="goNext"
      >下一页</button>
    </div>
  </div>
</template>
