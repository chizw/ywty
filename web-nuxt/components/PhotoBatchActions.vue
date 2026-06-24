<script setup lang="ts">
// 照片批量操作组件：批量删除 / 移入相册 / 移出相册 / 公开 / 私有 / 分享 / 全选
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
const showPublicModal = ref(false)
const albumId = ref<number | null>(null)
const albumOptions = ref<{ label: string; value: number }[]>([])
const sharePassword = ref('')
const shareExpire = ref(0)
const publicTarget = ref<boolean>(true)

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
  const ok = await openConfirm({
    title: '批量删除',
    message: `确定删除选中的 ${props.selectedIds.length} 张图片？此操作不可撤销。`,
    okText: '全部删除',
    danger: true,
  })
  if (!ok) return
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

async function batchRemoveFromAlbum() {
  if (!hasSelection.value) return
  const ok = await openConfirm({
    title: '移出相册',
    message: `将选中的 ${props.selectedIds.length} 张图片移出所有相册？`,
    okText: '移出',
  })
  if (!ok) return
  loading.value = true
  let ok2 = 0
  let fail = 0
  try {
    for (const id of props.selectedIds) {
      try {
        await api.post(`/api/v1/photos/${id}/move-to-album`, { album_id: 0 })
        ok2++
      } catch {
        fail++
      }
    }
    if (fail === 0) message.success(`已移出 ${ok2} 张图片`)
    else message.warning(`成功 ${ok2} 张，失败 ${fail} 张`)
    clearSelection()
    emit('done')
  } finally {
    loading.value = false
  }
}

function openPublicModal(target: boolean) {
  if (!hasSelection.value) return
  publicTarget.value = target
  showPublicModal.value = true
}

async function confirmPublic() {
  loading.value = true
  try {
    const res = await api.patch<any>('/api/v1/photos/batch-update', {
      ids: props.selectedIds,
      is_public: publicTarget.value,
    })
    const n = (res as any)?.updated ?? props.selectedIds.length
    message.success(`已${publicTarget.value ? '公开' : '转私有'} ${n} 张图片`)
    showPublicModal.value = false
    clearSelection()
    emit('done')
  } catch (err: any) {
    message.error(err?.statusMessage || '操作失败')
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
    <div class="flex flex-wrap items-center gap-2 bg-primary-50 border border-primary-200 rounded-lg p-2">
      <span class="text-sm text-gray-700 px-2">
        已选 <b class="text-primary-600">{{ selectedIds.length }}</b> 项
      </span>

      <div class="flex-1" />

      <NButtonGroup size="small">
        <NButton @click="toggleSelectAll" :disabled="allIds.length === 0">
          {{ allSelected ? '取消全选' : '全选' }}
        </NButton>
        <NButton :disabled="!hasSelection" @click="clearSelection">清空</NButton>
      </NButtonGroup>

      <NButtonGroup size="small">
        <NButton :disabled="!hasSelection" @click="openPublicModal(true)">批量公开</NButton>
        <NButton :disabled="!hasSelection" @click="openPublicModal(false)">批量私有</NButton>
        <NButton :disabled="!hasSelection" @click="openMoveModal">移入相册</NButton>
        <NButton :disabled="!hasSelection" :loading="loading" @click="batchRemoveFromAlbum">移出相册</NButton>
      </NButtonGroup>

      <NButtonGroup size="small">
        <NButton type="primary" :disabled="!hasSelection" @click="openShareModal">批量分享</NButton>
        <NButton type="error" :disabled="!hasSelection" :loading="loading" @click="batchDelete">
          批量删除
        </NButton>
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

    <!-- 公开/私有弹窗 -->
    <NModal
      v-model:show="showPublicModal"
      preset="card"
      :title="publicTarget ? '批量公开' : '批量转私有'"
      style="max-width: 380px"
      :mask-closable="false"
    >
      <p class="text-sm text-gray-500">
        确定将选中的 {{ selectedIds.length }} 张图片{{ publicTarget ? '设为公开' : '设为私有' }}？
      </p>
      <template #footer>
        <div class="flex justify-end gap-2">
          <NButton @click="showPublicModal = false">取消</NButton>
          <NButton type="primary" :loading="loading" @click="confirmPublic">确定</NButton>
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

    <!-- 自定义确认弹窗 -->
    <AppConfirm
      :show="confirmState.show"
      :title="confirmState.title"
      :message="confirmState.message"
      :ok-text="confirmState.okText"
      :danger="confirmState.danger"
      @update:show="(v) => confirmState.show = v"
      @confirm="onConfirmOk"
      @cancel="onConfirmCancel"
    />

    <template #fallback>
      <div class="text-sm text-gray-400">已选 {{ selectedIds.length }} 项</div>
    </template>
  </ClientOnly>
</template>
