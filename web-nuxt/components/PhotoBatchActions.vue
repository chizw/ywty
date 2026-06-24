<script setup lang="ts">
// 照片批量操作组件：批量删除 / 移入相册 / 移出相册 / 公开 / 私有 / 分享 / 全选
import { Button } from '~/components/ui/button'
import { Input } from '~/components/ui/input'
import { Label } from '~/components/ui/label'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '~/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '~/components/ui/select'

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
const albumId = ref<string>('')
const albumOptions = ref<{ label: string; value: string }[]>([])
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
    albumOptions.value = list.map((a: any) => ({ label: a.name, value: String(a.id) }))
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
  albumId.value = ''
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
        await api.post(`/api/v1/photos/${id}/move-to-album`, { album_id: Number(albumId.value) })
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
    <div class="flex flex-wrap items-center gap-2 bg-primary/5 border border-primary/20 rounded-lg p-2">
      <span class="text-sm text-foreground px-2">
        已选 <b class="text-primary">{{ selectedIds.length }}</b> 项
      </span>

      <div class="flex-1" />

      <div class="flex items-center gap-1">
        <Button variant="outline" size="sm" :disabled="allIds.length === 0" @click="toggleSelectAll">
          {{ allSelected ? '取消全选' : '全选' }}
        </Button>
        <Button variant="outline" size="sm" :disabled="!hasSelection" @click="clearSelection">清空</Button>
      </div>

      <div class="flex items-center gap-1">
        <Button variant="outline" size="sm" :disabled="!hasSelection" @click="openPublicModal(true)">批量公开</Button>
        <Button variant="outline" size="sm" :disabled="!hasSelection" @click="openPublicModal(false)">批量私有</Button>
        <Button variant="outline" size="sm" :disabled="!hasSelection" @click="openMoveModal">移入相册</Button>
        <Button variant="outline" size="sm" :disabled="!hasSelection" :loading="loading" @click="batchRemoveFromAlbum">移出相册</Button>
      </div>

      <div class="flex items-center gap-1">
        <Button size="sm" :disabled="!hasSelection" @click="openShareModal">批量分享</Button>
        <Button variant="destructive" size="sm" :disabled="!hasSelection" :loading="loading" @click="batchDelete">
          批量删除
        </Button>
      </div>
    </div>

    <!-- 移入相册弹窗 -->
    <Dialog :open="showMoveModal" @update:open="showMoveModal = $event">
      <DialogContent class="max-w-md">
        <DialogHeader>
          <DialogTitle>移入相册</DialogTitle>
        </DialogHeader>
        <div class="space-y-3">
          <p class="text-sm text-muted-foreground">将选中的 {{ selectedIds.length }} 张图片移入以下相册：</p>
          <Select v-model="albumId">
            <SelectTrigger>
              <SelectValue placeholder="选择目标相册" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem v-for="opt in albumOptions" :key="opt.value" :value="opt.value">
                {{ opt.label }}
              </SelectItem>
            </SelectContent>
          </Select>
        </div>
        <DialogFooter>
          <Button variant="outline" @click="showMoveModal = false">取消</Button>
          <Button :loading="loading" @click="confirmMove">确定</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- 公开/私有弹窗 -->
    <Dialog :open="showPublicModal" @update:open="showPublicModal = $event">
      <DialogContent class="max-w-sm">
        <DialogHeader>
          <DialogTitle>{{ publicTarget ? '批量公开' : '批量转私有' }}</DialogTitle>
        </DialogHeader>
        <p class="text-sm text-muted-foreground">
          确定将选中的 {{ selectedIds.length }} 张图片{{ publicTarget ? '设为公开' : '设为私有' }}？
        </p>
        <DialogFooter>
          <Button variant="outline" @click="showPublicModal = false">取消</Button>
          <Button :loading="loading" @click="confirmPublic">确定</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- 批量分享弹窗 -->
    <Dialog :open="showShareModal" @update:open="showShareModal = $event">
      <DialogContent class="max-w-md">
        <DialogHeader>
          <DialogTitle>批量分享</DialogTitle>
        </DialogHeader>
        <div class="space-y-3">
          <p class="text-sm text-muted-foreground">为选中的 {{ selectedIds.length }} 张图片创建分享链接：</p>
          <div class="space-y-1">
            <Label>访问密码（可选）</Label>
            <Input v-model="sharePassword" placeholder="留空则公开访问" />
          </div>
          <div class="space-y-1">
            <Label>过期分钟数（0 = 永不过期）</Label>
            <Input v-model="shareExpire" type="number" :min="0" />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" @click="showShareModal = false">取消</Button>
          <Button :loading="loading" @click="confirmShare">创建分享</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

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
      <div class="text-sm text-muted-foreground">已选 {{ selectedIds.length }} 项</div>
    </template>
  </ClientOnly>
</template>
