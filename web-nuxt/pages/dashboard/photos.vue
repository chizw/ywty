<script setup lang="ts">
// 我的图片列表：多选 + 拖动框选 + 批量操作 + 筛选 + 排序 + 分页
definePageMeta({ layout: 'dashboard', middleware: 'auth' })

import { MousePointerClick, X } from 'lucide-vue-next'

const api = useApi()
const message = useMessage()
const statsStore = useStatsStore()

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

// 多选 + 拖动框选
// 新交互：无需先点"多选"按钮，在网格空白处按下并拖动鼠标即自动进入多选模式
const selectedIds = ref<number[]>([])
const dragSelecting = ref(false)
const dragStart = ref<{ x: number; y: number } | null>(null)
const dragEnd = ref<{ x: number; y: number } | null>(null)
const dragMoved = ref(false) // 是否产生有效拖动（>5px），用于区分单击和框选
const containerRef = ref<HTMLElement | null>(null)
const itemRefs = ref<HTMLElement[]>([])

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

// 手动管理数据，不使用 useAsyncData（避免 setup 阶段 auth 未就绪导致请求失败）
const rawData = ref<any>(null)
const loading = ref(false)

async function fetchPhotos() {
  loading.value = true
  try {
    rawData.value = await api.get<any>('/api/v1/photos', { query: query.value, raw: true })
  } catch {
    rawData.value = null
  } finally {
    loading.value = false
  }
}

const photos = computed<any[]>(() => {
  const d = rawData.value
  if (Array.isArray(d)) return d
  if (d && Array.isArray(d.data)) return d.data
  return []
})
const total = computed(() => Number(rawData.value?.meta?.total ?? photos.value.length))
const lastPage = computed(() => {
  const lp = rawData.value?.meta?.last_page
  if (lp) return Number(lp)
  return Math.max(1, Math.ceil(total.value / perPage))
})

const albumOptions = ref<{ label: string; value: number }[]>([])
const tagOptions = ref<{ label: string; value: string }[]>([])

onMounted(async () => {
  // 并行加载：图片列表 + 相册筛选项 + 标签筛选项
  const [, albumsRes, tagsRes] = await Promise.all([
    fetchPhotos(),
    api.get<any>('/api/v1/albums').catch(() => []),
    api.get<any>('/api/v1/tags').catch(() => []),
  ])
  const aList = Array.isArray(albumsRes) ? albumsRes : ((albumsRes as any)?.data ?? [])
  albumOptions.value = aList.map((a: any) => ({ label: a.name, value: a.id }))
  const tList = Array.isArray(tagsRes) ? tagsRes : ((tagsRes as any)?.data ?? [])
  tagOptions.value = tList.map((t: any) => ({ label: t.name, value: t.name }))
})

// refresh 供上传/删除后调用
function refresh() {
  fetchPhotos()
}

const allIds = computed(() => photos.value.map((p) => p.id))

// --- 自定义确认弹窗（替代 confirm()）---
const confirmState = reactive({
  show: false,
  title: '确认',
  message: '',
  okText: '确定',
  cancelText: '取消',
  danger: false,
  resolver: null as null | ((ok: boolean) => void),
})

function openConfirm(opts: { title?: string; message: string; okText?: string; danger?: boolean }): Promise<boolean> {
  confirmState.show = true
  confirmState.title = opts.title || '确认'
  confirmState.message = opts.message
  confirmState.okText = opts.okText || '确定'
  confirmState.danger = !!opts.danger
  return new Promise<boolean>((resolve) => { confirmState.resolver = resolve })
}
function onConfirmOk() { confirmState.resolver?.(true); confirmState.resolver = null }
function onConfirmCancel() { confirmState.resolver?.(false); confirmState.resolver = null }

// --- 拖动框选逻辑 ---
const dragBox = computed(() => {
  if (!dragSelecting.value || !dragStart.value || !dragEnd.value) return null
  const x1 = Math.min(dragStart.value.x, dragEnd.value.x)
  const y1 = Math.min(dragStart.value.y, dragEnd.value.y)
  const x2 = Math.max(dragStart.value.x, dragEnd.value.x)
  const y2 = Math.max(dragStart.value.y, dragEnd.value.y)
  return { left: x1, top: y1, width: x2 - x1, height: y2 - y1 }
})

function inBox(el: HTMLElement, box: { left: number; top: number; width: number; height: number }) {
  const r = el.getBoundingClientRect()
  return !(r.right < box.left || r.bottom < box.top || r.left > box.left + box.width || r.top > box.top + box.height)
}

// 计算 selectMode：只要有选中项或正在拖动选择，就视为多选模式
const selectMode = computed(() => selectedIds.value.length > 0 || dragSelecting.value)

function onPointerDown(e: PointerEvent) {
  // 只响应左键、且点击空白区域（不是图片）
  if (e.button !== 0) return
  const target = e.target as HTMLElement
  if (target.closest('[data-photo-item]')) return
  e.preventDefault()
  dragStart.value = { x: e.clientX, y: e.clientY }
  dragEnd.value = { x: e.clientX, y: e.clientY }
  dragMoved.value = false
  document.addEventListener('pointermove', onPointerMove)
  document.addEventListener('pointerup', onPointerUp, { once: true })
}

function onPointerMove(e: PointerEvent) {
  if (!dragStart.value) return
  dragEnd.value = { x: e.clientX, y: e.clientY }
  // 当鼠标移动超过 5px 才视为有效框选
  if (!dragMoved.value) {
    const dx = e.clientX - dragStart.value.x
    const dy = e.clientY - dragStart.value.y
    if (Math.abs(dx) > 5 || Math.abs(dy) > 5) {
      dragMoved.value = true
      dragSelecting.value = true
    }
  }
}

function onPointerUp() {
  document.removeEventListener('pointermove', onPointerMove)
  if (dragMoved.value) {
    const box = dragBox.value
    if (box) {
      // 选择所有在拖动框内的图片
      const ids: number[] = []
      for (const el of itemRefs.value) {
        if (el && inBox(el, box)) {
          const id = Number(el.dataset.photoId)
          if (id) ids.push(id)
        }
      }
      // 追加到现有选择
      const set = new Set(selectedIds.value)
      for (const id of ids) set.add(id)
      selectedIds.value = Array.from(set)
    }
  }
  dragSelecting.value = false
  dragStart.value = null
  dragEnd.value = null
  dragMoved.value = false
}

function onItemPointerDown(e: PointerEvent, id: number) {
  // 当已经处于多选模式（用户已选过图片），点击图片切换选中
  if (selectedIds.value.length > 0) {
    e.stopPropagation()
    e.preventDefault()
    toggleSelect(id)
    return
  }
  // 否则正常进入详情页（不阻止默认）
}

function onUploaded() {
  page.value = 1
  refresh()
  statsStore.refresh()
}

watch([filterAlbumId, filterTag, startDate, endDate, sortBy, sortOrder], () => {
  page.value = 1
  refresh()
})
watch(page, () => refresh())

function exitSelectMode() {
  selectedIds.value = []
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
  statsStore.refresh()
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
  const ok = await openConfirm({
    title: '删除图片',
    message: '确定删除这张图片？此操作不可撤销。',
    okText: '删除',
    danger: true,
  })
  if (!ok) return
  try {
    await api.del(`/api/v1/photos/${id}`)
    message.success('已删除')
    refresh()
    statsStore.refresh()
  } catch (err: any) {
    message.error(err?.statusMessage || '删除失败')
  }
}

async function copy(id: number) {
  try {
    await api.post(`/api/v1/photos/${id}/copy`, {})
    message.success('已复制')
    refresh()
    statsStore.refresh()
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

function goPrev() { if (page.value > 1) page.value-- }
function goNext() { if (page.value < lastPage.value) page.value++ }
</script>

<template>
  <div>
    <!-- 顶部操作栏 -->
    <div class="flex items-center justify-between mb-4">
      <div class="flex items-center gap-3">
        <h1 class="text-2xl font-bold text-foreground">我的图片</h1>
        <span class="text-sm text-muted-foreground">共 {{ total }} 张</span>
        <Badge v-if="selectMode" variant="default">
          已选 <strong class="ml-1">{{ selectedIds.length }}</strong> 项
        </Badge>
      </div>
      <div class="flex items-center gap-2">
        <Button
          v-if="selectMode"
          variant="outline"
          size="sm"
          @click="exitSelectMode"
        >
          <X class="mr-1 h-3 w-3" />
          取消选择
        </Button>
      </div>
    </div>

    <!-- 多选使用提示 -->
    <div v-if="!selectMode" class="mb-3 text-xs text-muted-foreground flex items-center gap-1.5">
      <MousePointerClick class="w-3.5 h-3.5" />
      提示：在图片网格的空白处<strong class="font-semibold mx-0.5">按住鼠标左键拖动</strong>可框选多张图片
    </div>

    <PhotoUploader class="mb-6" @uploaded="onUploaded" @error="(m) => message.error(m)" />

    <!-- 筛选 / 排序 -->
    <Card class="mb-4">
      <CardContent class="p-3">
        <div class="flex flex-wrap items-end gap-3">
          <div>
            <Label class="text-xs text-muted-foreground mb-1">相册</Label>
            <Select v-model="filterAlbumId">
              <SelectTrigger class="w-[140px]">
                <SelectValue placeholder="全部" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem :value="null as any">全部</SelectItem>
                <SelectItem v-for="a in albumOptions" :key="a.value" :value="a.value">{{ a.label }}</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div>
            <Label class="text-xs text-muted-foreground mb-1">标签</Label>
            <Select v-model="filterTag">
              <SelectTrigger class="w-[140px]">
                <SelectValue placeholder="全部" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="">全部</SelectItem>
                <SelectItem v-for="t in tagOptions" :key="t.value" :value="t.value">{{ t.label }}</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div>
            <Label class="text-xs text-muted-foreground mb-1">开始日期</Label>
            <Input v-model="startDate" type="date" class="w-[150px]" />
          </div>
          <div>
            <Label class="text-xs text-muted-foreground mb-1">结束日期</Label>
            <Input v-model="endDate" type="date" class="w-[150px]" />
          </div>
          <div>
            <Label class="text-xs text-muted-foreground mb-1">排序</Label>
            <Select v-model="sortBy">
              <SelectTrigger class="w-[120px]">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="created_at">按时间</SelectItem>
                <SelectItem value="size">按大小</SelectItem>
                <SelectItem value="name">按名称</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div>
            <Label class="text-xs text-muted-foreground mb-1">方向</Label>
            <Select v-model="sortOrder">
              <SelectTrigger class="w-[100px]">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="desc">降序</SelectItem>
                <SelectItem value="asc">升序</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <Button variant="outline" size="sm" @click="resetFilters">重置</Button>
        </div>
      </CardContent>
    </Card>

    <!-- 批量操作栏 -->
    <div v-if="selectMode" class="mb-4">
      <PhotoBatchActions
        v-model:selected-ids="selectedIds"
        :all-ids="allIds"
        @done="onBatchDone"
      />
    </div>

    <AppEmpty v-if="photos.length === 0" title="还没有图片" description="拖拽或点击上方上传你的第一张图片" />
    <div
      v-else
      ref="containerRef"
      class="relative grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-3 select-none"
      :class="{
        'cursor-crosshair': !selectMode,
        'cursor-pointer': selectMode,
      }"
      @pointerdown="onPointerDown"
    >
      <div
        v-for="(p, idx) in photos"
        :key="p.id"
        :ref="(el) => { if (el) itemRefs[idx] = el as HTMLElement }"
        :data-photo-item="p.id"
        :data-photo-id="p.id"
        class="group relative bg-muted rounded overflow-hidden aspect-square"
        :class="{
          'cursor-pointer': !selectMode,
          'ring-2 ring-primary': selectMode && isSelected(p.id),
        }"
        @pointerdown="(e) => onItemPointerDown(e, p.id)"
      >
        <img :src="`/uploads/${p.pathname}`" :alt="p.name" class="w-full h-full object-cover pointer-events-none" loading="lazy" />

        <!-- 多选勾选框 -->
        <div v-if="selectMode" class="absolute top-1 left-1 z-10 pointer-events-none">
          <span
            class="inline-flex items-center justify-center w-5 h-5 rounded border-2 text-white text-xs"
            :class="isSelected(p.id) ? 'bg-primary border-primary' : 'bg-black/30 border-white'"
          >{{ isSelected(p.id) ? '✓' : '' }}</span>
        </div>

        <!-- 单张操作 -->
        <div
          v-if="!selectMode"
          class="absolute inset-0 bg-black bg-opacity-0 group-hover:bg-opacity-50 transition flex flex-col justify-between p-2 opacity-0 group-hover:opacity-100"
        >
          <div class="flex justify-end">
            <Badge v-if="p.is_public" variant="default" class="text-[10px]">公开</Badge>
          </div>
          <div class="flex gap-1 w-full justify-end items-center flex-wrap">
            <LikeButton size="sm" target-type="photo" :target-id="p.id" />
            <ReportButton size="sm" target-type="photo" :target-id="p.id" />
            <Button variant="secondary" size="sm" class="h-6 px-2 text-[10px]" @click.stop="togglePublic(p)">{{ p.is_public ? '转私有' : '转公开' }}</Button>
            <Button variant="secondary" size="sm" class="h-6 px-2 text-[10px]" @click.stop="copy(p.id)">复制</Button>
            <Button variant="destructive" size="sm" class="h-6 px-2 text-[10px]" @click.stop="remove(p.id)">删除</Button>
          </div>
        </div>
      </div>

      <!-- 拖动选区框 -->
      <div
        v-if="dragBox"
        class="fixed pointer-events-none z-50 border-2 border-primary bg-primary/10"
        :style="{
          left: dragBox.left + 'px',
          top: dragBox.top + 'px',
          width: dragBox.width + 'px',
          height: dragBox.height + 'px',
        }"
      />
    </div>

    <!-- 分页 -->
    <div v-if="total > perPage" class="mt-6 flex items-center justify-center gap-3 text-sm">
      <Button variant="outline" size="sm" :disabled="page <= 1" @click="goPrev">上一页</Button>
      <span class="text-muted-foreground">第 {{ page }} / {{ lastPage }} 页</span>
      <Button variant="outline" size="sm" :disabled="page >= lastPage" @click="goNext">下一页</Button>
    </div>

    <!-- 自定义确认弹窗 -->
    <Dialog v-model:open="confirmState.show">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{{ confirmState.title }}</DialogTitle>
        </DialogHeader>
        <p class="text-sm text-muted-foreground">{{ confirmState.message }}</p>
        <DialogFooter>
          <Button variant="outline" @click="onConfirmCancel">取消</Button>
          <Button :variant="confirmState.danger ? 'destructive' : 'default'" @click="onConfirmOk">{{ confirmState.okText }}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
