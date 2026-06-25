<script setup lang="ts">
// 图片上传组件：拖拽 / 粘贴 / 批量选择，独立进度条，缩略图预览，失败重试
import { ref, computed, onMounted, onBeforeUnmount, triggerRef, watch } from 'vue'
import { Upload, X, RefreshCw, ChevronDown, ChevronUp } from '@lucide/vue'
import { useAuthStore } from '~/stores/auth'
import type { UploadResult } from '~/types/api'

const props = withDefaults(defineProps<{
  multiple?: boolean
  maxSizeMB?: number
  autoUpload?: boolean
}>(), {
  multiple: true,
  maxSizeMB: 20,
  autoUpload: true,
})

const emit = defineEmits<{
  uploaded: [UploadResult]
  error: [string]
}>()

// 与 useApi 一致：用 useAuthStore 取 token，从 runtimeConfig 取 baseURL
const config = useRuntimeConfig()
const auth = useAuthStore()

const ACCEPT_TYPES = ['image/jpeg', 'image/png', 'image/gif', 'image/webp', 'image/bmp']
const ACCEPT_ATTR = '.jpg,.jpeg,.png,.gif,.webp,.bmp,image/jpeg,image/png,image/gif,image/webp,image/bmp'

interface TaskItem {
  id: string
  file: File
  name: string
  size: number
  preview: string
  progress: number
  status: 'pending' | 'uploading' | 'success' | 'error'
  result?: UploadResult
  errorMsg?: string
  xhr?: XMLHttpRequest
}

const tasks = ref<TaskItem[]>([])
const dragOver = ref(false)
const inputRef = ref<HTMLInputElement | null>(null)

let taskSeq = 0
function nextId(): string {
  taskSeq += 1
  return `task-${Date.now()}-${taskSeq}`
}

function validateFile(file: File): string | null {
  if (!ACCEPT_TYPES.includes(file.type)) {
    return `${file.name} 类型不被允许，仅支持 jpeg/png/gif/webp/bmp`
  }
  if (file.size > props.maxSizeMB * 1024 * 1024) {
    return `${file.name} 超过 ${props.maxSizeMB}MB 限制`
  }
  return null
}

function isDuplicate(file: File): boolean {
  return tasks.value.some(
    (t) => t.file.name === file.name && t.file.size === file.size && t.file.lastModified === file.lastModified,
  )
}

function createTask(file: File): TaskItem | null {
  if (isDuplicate(file)) return null
  const err = validateFile(file)
  if (err) {
    emit('error', err)
    return null
  }
  const item: TaskItem = {
    id: nextId(),
    file,
    name: file.name,
    size: file.size,
    preview: URL.createObjectURL(file),
    progress: 0,
    status: 'pending',
  }
  tasks.value.push(item)
  return item
}

interface UploadCallbacks {
  onFinish?: () => void
  onError?: () => void
}

function uploadTask(item: TaskItem, callbacks?: UploadCallbacks) {
  if (item.status === 'uploading') return
  item.status = 'uploading'
  item.progress = 0
  item.errorMsg = undefined
  triggerRef(tasks)

  const formData = new FormData()
  formData.append('file', item.file)

  const xhr = new XMLHttpRequest()
  const baseURL = (config.apiBase as string) || ''
  xhr.open('POST', `${baseURL}/api/v1/photos`)

  // 注入认证 token（与 useApi 内部逻辑一致）
  if (auth.accessToken) {
    xhr.setRequestHeader('Authorization', `Bearer ${auth.accessToken}`)
  }

  xhr.upload.onprogress = (e: ProgressEvent) => {
    if (e.lengthComputable && e.total > 0) {
      item.progress = Math.min(99, Math.ceil((e.loaded / e.total) * 100))
      triggerRef(tasks)
    }
  }

  xhr.upload.onload = () => {
    if (item.progress < 100) {
      item.progress = 99
      triggerRef(tasks)
    }
  }

  xhr.onload = () => {
    if (xhr.status >= 200 && xhr.status < 300) {
      try {
        const res = JSON.parse(xhr.responseText)
        const data = res?.data as UploadResult | undefined
        if (data) {
          item.status = 'success'
          item.progress = 100
          item.result = data
          triggerRef(tasks)
          emit('uploaded', data)
          callbacks?.onFinish?.()
        } else {
          item.status = 'error'
          item.errorMsg = '响应数据异常'
          triggerRef(tasks)
          emit('error', `${item.name} 上传失败：响应数据异常`)
          callbacks?.onError?.()
        }
      } catch {
        item.status = 'error'
        item.errorMsg = '解析响应失败'
        triggerRef(tasks)
        emit('error', `${item.name} 上传失败：解析响应失败`)
        callbacks?.onError?.()
      }
    } else {
      item.status = 'error'
      item.errorMsg = `HTTP ${xhr.status}`
      triggerRef(tasks)
      emit('error', `${item.name} 上传失败：HTTP ${xhr.status}`)
      callbacks?.onError?.()
    }
  }

  xhr.onerror = () => {
    item.status = 'error'
    item.errorMsg = '网络错误'
    triggerRef(tasks)
    emit('error', `${item.name} 上传失败：网络错误`)
    callbacks?.onError?.()
  }

  item.xhr = xhr
  xhr.send(formData)
}

function retryTask(item: TaskItem) {
  uploadTask(item)
}

function removeTask(item: TaskItem) {
  if (item.xhr && item.status === 'uploading') {
    item.xhr.abort()
  }
  if (item.preview) {
    URL.revokeObjectURL(item.preview)
  }
  const idx = tasks.value.findIndex((t) => t.id === item.id)
  if (idx >= 0) tasks.value.splice(idx, 1)
}

function uploadPending() {
  for (const t of tasks.value) {
    if (t.status === 'pending') uploadTask(t)
  }
}

// 文件选择
function pickFiles() {
  inputRef.value?.click()
}

function onFileInputChange(e: Event) {
  const input = e.target as HTMLInputElement
  if (!input.files) return
  for (const f of Array.from(input.files)) {
    const item = createTask(f)
    if (item && props.autoUpload) uploadTask(item)
  }
  input.value = ''
}

// 拖拽上传
function onDrop(e: DragEvent) {
  e.preventDefault()
  dragOver.value = false
  if (!e.dataTransfer) return
  for (const f of Array.from(e.dataTransfer.files)) {
    const item = createTask(f)
    if (item && props.autoUpload) uploadTask(item)
  }
}

function onDragOver(e: DragEvent) {
  e.preventDefault()
  dragOver.value = true
}

function onDragLeave() {
  dragOver.value = false
}

// 粘贴上传
function onPaste(e: ClipboardEvent) {
  if (!e.clipboardData) return
  const items = Array.from(e.clipboardData.items)
  for (const it of items) {
    if (it.kind === 'file') {
      const f = it.getAsFile()
      if (f) {
        const item = createTask(f)
        if (item && props.autoUpload) uploadTask(item)
      }
    }
  }
}

onMounted(() => {
  window.addEventListener('paste', onPaste)
})
onBeforeUnmount(() => {
  window.removeEventListener('paste', onPaste)
  for (const t of tasks.value) {
    if (t.preview) URL.revokeObjectURL(t.preview)
  }
})

function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / 1024 / 1024).toFixed(2)} MB`
}

const hasPending = computed(() => tasks.value.some((t) => t.status === 'pending'))
const successCount = computed(() => tasks.value.filter((t) => t.status === 'success').length)
const errorCount = computed(() => tasks.value.filter((t) => t.status === 'error').length)
const uploadingCount = computed(() => tasks.value.filter((t) => t.status === 'uploading').length)
const pendingCount = computed(() => tasks.value.filter((t) => t.status === 'pending').length)
const totalCount = computed(() => tasks.value.length)
const isAllDone = computed(() => totalCount.value > 0 && uploadingCount.value === 0 && pendingCount.value === 0)
const hasFailures = computed(() => errorCount.value > 0)

// 整体进度：按文件 size 加权平均
const overallProgress = computed(() => {
  const list = tasks.value
  if (list.length === 0) return 0
  const totalSize = list.reduce((s, t) => s + (t.size || 0), 0)
  if (totalSize === 0) {
    const sum = list.reduce((s, t) => s + t.progress, 0)
    return Math.ceil(sum / list.length)
  }
  const sum = list.reduce((s, t) => s + (t.progress * (t.size || 0)), 0)
  return Math.ceil(sum / totalSize)
})

const showDetail = ref(false)
watch(isAllDone, (v) => {
  if (v) {
    setTimeout(() => {
      if (isAllDone.value) showDetail.value = false
    }, 2000)
  }
})

function clearFinished() {
  for (let i = tasks.value.length - 1; i >= 0; i--) {
    const t = tasks.value[i]
    if (t.status === 'success') {
      if (t.preview) URL.revokeObjectURL(t.preview)
      tasks.value.splice(i, 1)
    }
  }
}

function retryAllFailed() {
  for (const t of tasks.value) {
    if (t.status === 'error') retryTask(t)
  }
}
</script>

<template>
  <ClientOnly>
    <div class="photo-uploader">
      <!-- 拖拽区 -->
      <div
        class="border-2 border-dashed rounded-lg p-8 text-center cursor-pointer transition-colors"
        :class="dragOver ? 'border-primary bg-primary/5' : 'border-border hover:border-primary/50'"
        @dragover="onDragOver"
        @dragleave="onDragLeave"
        @drop="onDrop"
        @click="pickFiles"
      >
        <input
          ref="inputRef"
          type="file"
          class="hidden"
          :accept="ACCEPT_ATTR"
          :multiple="multiple"
          @change="onFileInputChange"
        />
        <Upload class="mx-auto h-10 w-10 text-muted-foreground mb-2" />
        <p class="text-sm text-foreground">点击选择，或拖拽文件到此处</p>
        <p class="mt-1 text-xs text-muted-foreground">
          支持 jpeg/png/gif/webp/bmp，单文件不超过 {{ maxSizeMB }}MB（也可 Ctrl+V 粘贴）
        </p>
      </div>

      <!-- 手动上传按钮（autoUpload=false 时） -->
      <div v-if="!autoUpload && hasPending" class="mt-3 text-center">
        <Button @click="uploadPending">上传全部</Button>
      </div>

      <!-- 聚合进度条 -->
      <div
        v-if="totalCount > 0"
        class="mt-3 bg-card border border-border rounded-lg p-3"
      >
        <!-- 状态行 -->
        <div class="flex items-center gap-3 text-sm">
          <div class="flex-1 min-w-0">
            <span v-if="uploadingCount > 0 || pendingCount > 0" class="text-foreground">
              正在上传 <b class="text-primary">{{ totalCount }}</b> 个文件
              <span v-if="successCount > 0" class="text-green-600">· 已完成 {{ successCount }}</span>
              <span v-if="errorCount > 0" class="text-destructive">· 失败 {{ errorCount }}</span>
            </span>
            <span v-else-if="isAllDone && errorCount > 0" class="text-destructive">
              上传完成 · 成功 {{ successCount }} · 失败 {{ errorCount }}
            </span>
            <span v-else class="text-green-600">
              上传完成 · 成功 {{ successCount }} 个文件
            </span>
          </div>
          <div class="flex items-center gap-2 shrink-0">
            <Button
              v-if="hasFailures && isAllDone"
              variant="outline"
              size="sm"
              class="h-7 text-xs text-destructive border-destructive/30"
              @click="retryAllFailed"
            >
              <RefreshCw class="h-3 w-3 mr-1" />
              重试失败
            </Button>
            <Button
              v-if="isAllDone && successCount > 0"
              variant="ghost"
              size="sm"
              class="h-7 text-xs text-muted-foreground"
              @click="clearFinished"
            >
              清除已完成
            </Button>
            <Button
              variant="ghost"
              size="sm"
              class="h-7 text-xs text-primary"
              @click="showDetail = !showDetail"
            >
              {{ showDetail ? '收起' : '详情' }}
              <ChevronUp v-if="showDetail" class="h-3 w-3 ml-1" />
              <ChevronDown v-else class="h-3 w-3 ml-1" />
            </Button>
          </div>
        </div>
        <!-- 整体进度条 -->
        <div class="mt-2 w-full bg-muted rounded-full overflow-hidden" style="height: 4px;">
          <div
            class="h-full rounded-full transition-all duration-200 ease-out"
            :class="hasFailures && isAllDone ? 'bg-destructive' : (isAllDone ? 'bg-green-500' : 'bg-primary')"
            :style="{ width: overallProgress + '%' }"
          ></div>
        </div>
      </div>

      <!-- 文件详情列表 -->
      <div v-if="showDetail && totalCount > 0" class="mt-2 max-h-72 overflow-y-auto space-y-2 pr-1">
        <div
          v-for="task in tasks"
          :key="task.id"
          class="flex items-center gap-2 bg-muted px-2 py-1.5 rounded-md text-sm"
        >
          <!-- 缩略图预览 -->
          <div class="w-8 h-8 flex-shrink-0 bg-muted rounded overflow-hidden">
            <img :src="task.preview" :alt="task.name" class="w-full h-full object-cover" />
          </div>

          <!-- 文件名 + 进度条 -->
          <div class="flex-1 min-w-0">
            <div class="flex items-center justify-between text-xs">
              <span class="truncate text-foreground">{{ task.name }}</span>
              <span class="text-muted-foreground ml-2 shrink-0">
                <template v-if="task.status === 'success'">✓</template>
                <template v-else-if="task.status === 'error'">✗</template>
                <template v-else>{{ task.progress }}%</template>
                · {{ formatSize(task.size) }}
              </span>
            </div>
            <div class="mt-1 w-full bg-muted rounded-full overflow-hidden" style="height: 3px;">
              <div
                class="h-full rounded-full transition-all duration-200 ease-out"
                :class="task.status === 'success' ? 'bg-green-500' : (task.status === 'error' ? 'bg-destructive' : 'bg-primary')"
                :style="{ width: task.progress + '%' }"
              ></div>
            </div>
          </div>

          <!-- 操作按钮 -->
          <div class="flex-shrink-0 flex gap-1 items-center">
            <Button
              v-if="task.status === 'error'"
              variant="ghost"
              size="sm"
              class="h-6 text-xs text-primary px-1"
              @click="retryTask(task)"
            >
              重试
            </Button>
            <Button
              variant="ghost"
              size="sm"
              class="h-6 w-6 p-0 text-muted-foreground hover:text-destructive"
              title="移除"
              @click="removeTask(task)"
            >
              <X class="h-3 w-3" />
            </Button>
          </div>
        </div>
      </div>
    </div>

    <template #fallback>
      <div class="border-2 border-dashed rounded-lg p-8 text-center border-border">
        <Upload class="mx-auto h-10 w-10 text-muted-foreground mb-2" />
        <p class="text-sm text-foreground">加载上传组件...</p>
      </div>
    </template>
  </ClientOnly>
</template>
