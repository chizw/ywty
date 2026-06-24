<script setup lang="ts">
// 照片批量操作组件：批量删除 / 移入相册 / 分享 / 全选
import { NButtonGroup, NButton, NModal, NSelect, NInput, NInputNumber } from 'naive-ui'

const props = defineProps<{
  selectedIds: number[]
  allIds: number[]
}>()

const emit = defineEmits<{
  done: []
  'update:selectedIds': [number[]]
}>()

const api = useApi()
const message = useMessage()

const loading = ref(false)
const showMoveModal = ref(false)
const showShareModal = ref(false)
const albumId = ref<number | null>(null)
const albumOptions = ref<{ label: string; value: number }[]>([])
const sharePassword = ref('')
const shareExpire = ref(0)

const hasSelection = computed(() => props.selectedIds.length > 0)
const allSelected = computed(() => props.allIds.length > 0 && props.selectedIds.length === props.allIds.length)

async function loadAlbums() {
  try {
    const res = await api.get<any>('/api/v1/albums')
    const list = Array.isArray(res) ? res : ((res as any)?.data ?? [])
    albumOptions.value = list.map((a: any) => ({ label: a.name, value: a.id }))
  } catch {
    albumOptions.value = []
  }
}

function toggleSelectAll() {
  if (allSelected.value) emit('update:selectedIds', [])
  else emit('update:selectedIds', [...props.allIds])
}

function clearSelection() {
  emit('update:selectedIds', [])
}

async function batchDelete() {
  if (!hasSelection.value) return
  if (!confirm(`确定删除选中的 ${props.selectedIds.length} 张图片？此操作不可撤销。`)) return
  loading.value = true
  try {
    await api.post('/api/v1/photos/batch-delete', { ids: props.selectedIds })
    message.success(`已删除 ${props.selectedIds.length} 张图片`)
    clearSelection()
    emit('done')
  } catch (err: any) {
    message.error(err?.statusMessage || '批量删除失败')
  } finally {
    loading.value = false
  }
}

function openMoveModal() {
  if (!hasSelection.value) return
  loadAlbums()
  albumId.value = null
  showMoveModal.value = true
}

async function confirmMove() {
  if (!albumId.value) {
    message.warning('请选择目标相册')
    return
  }
  loading.value = true
  let ok = 0
  let fail = 0
  try {
    for (const id of props.selectedIds) {
      try {
        await api.post(`/api/v1/photos/${id}/move-to-album`, { album_id: albumId.value })
        ok++
      } catch {
        fail++
      }
    }
    if (fail === 0) message.success(`已移动 ${ok} 张图片到相册`)
    else message.warning(`成功 ${ok} 张，失败 ${fail} 张`)
    showMoveModal.value = false
    clearSelection()
    emit('done')
  } finally {
    loading.value = false
  }
}

function openShareModal() {
  if (!hasSelection.value) return
  sharePassword.value = ''
  shareExpire.value = 0
  showShareModal.value = true
}

async function confirmShare() {
  loading.value = true
  try {
    const body: any = { type: 'photo', ids: props.selectedIds }
    if (sharePassword.value) body.password = sharePassword.value
    if (shareExpire.value > 0) body.expire_minutes = shareExpire.value
    await api.post('/api/v1/shares', body)
    message.success('已创建分享链接')
    showShareModal.value = false
    clearSelection()
    emit('done')
  } catch (err: any) {
    message.error(err?.statusMessage || '创建分享失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <ClientOnly>
    <div class="flex flex-wrap items-center gap-3">
      <span class="text-sm text-gray-600">已选 <b class="text-primary-600">{{ selectedIds.length }}</b> 项</span>

      <NButtonGroup size="small">
        <NButton @click="toggleSelectAll" :disabled="allIds.length === 0">
          {{ allSelected ? '取消全选' : '全选' }}
        </NButton>
        <NButton :disabled="!hasSelection" @click="clearSelection">清空</NButton>
      </NButtonGroup>

      <NButtonGroup size="small">
        <NButton type="error" :disabled="!hasSelection" :loading="loading" @click="batchDelete">
          批量删除
        </NButton>
        <NButton :disabled="!hasSelection" @click="openMoveModal">移入相册</NButton>
        <NButton type="primary" :disabled="!hasSelection" @click="openShareModal">批量分享</NButton>
      </NButtonGroup>
    </div>

    <!-- 移入相册弹窗 -->
    <NModal
      v-model:show="showMoveModal"
      preset="card"
      title="移入相册"
      style="max-width: 420px"
      :mask-closable="false"
    >
      <div class="space-y-3">
        <p class="text-sm text-gray-500">将选中的 {{ selectedIds.length }} 张图片移入以下相册：</p>
        <NSelect
          v-model:value="albumId"
          :options="albumOptions"
          placeholder="选择目标相册"
          clearable
          filterable
        />
      </div>
      <template #footer>
        <div class="flex justify-end gap-2">
          <NButton @click="showMoveModal = false">取消</NButton>
          <NButton type="primary" :loading="loading" @click="confirmMove">确定</NButton>
        </div>
      </template>
    </NModal>

    <!-- 批量分享弹窗 -->
    <NModal
      v-model:show="showShareModal"
      preset="card"
      title="批量分享"
      style="max-width: 420px"
      :mask-closable="false"
    >
      <div class="space-y-3">
        <p class="text-sm text-gray-500">为选中的 {{ selectedIds.length }} 张图片创建分享链接：</p>
        <div>
          <label class="block text-sm text-gray-700 mb-1">访问密码（可选）</label>
          <NInput v-model:value="sharePassword" placeholder="留空则公开访问" />
        </div>
        <div>
          <label class="block text-sm text-gray-700 mb-1">过期分钟数（0 = 永不过期）</label>
          <NInputNumber v-model:value="shareExpire" :min="0" class="w-full" />
        </div>
      </div>
      <template #footer>
        <div class="flex justify-end gap-2">
          <NButton @click="showShareModal = false">取消</NButton>
          <NButton type="primary" :loading="loading" @click="confirmShare">创建分享</NButton>
        </div>
      </template>
    </NModal>

    <template #fallback>
      <div class="text-sm text-gray-400">已选 {{ selectedIds.length }} 项</div>
    </template>
  </ClientOnly>
</template>
