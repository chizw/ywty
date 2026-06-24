<script setup lang="ts">
// 图片上传组件：拖拽 / 粘贴 / 批量选择，独立进度条，缩略图预览，失败重试
import { ref, computed, onMounted, onBeforeUnmount, triggerRef, watch } from 'vue'
import { NUpload, NButton } from 'naive-ui'
import type { UploadCustomRequestOptions } from 'naive-ui'
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

// 与 useApi 一致：从 useAuthStore 取 token，从 runtimeConfig 取 baseURL
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

// NUpload custom-request：将 NUpload 选择的文件桥接到任务系统
function customRequest(options: UploadCustomRequestOptions) {
  const rawFile = options.file.file as File | null
  if (!rawFile) {
    options.onError()
    return
  }
  const item = createTask(rawFile)
  if (!item) {
    options.onError()
    return
  }
  if (props.autoUpload) {
    uploadTask(item, {
      onFinish: () => options.onFinish(),
      onError: () => options.onError(),
    })
  } else {
    options.onFinish()
  }
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
    // size 缺失时退化为按文件数等权
    const sum = list.reduce((s, t) => s + t.progress, 0)
    return Math.ceil(sum / list.length)
  }
  const sum = list.reduce((s, t) => s + (t.progress * (t.size || 0)), 0)
  return Math.ceil(sum / totalSize)
})

// 详情展开：上传中默认折叠（节省空间），全部完成后默认折叠，可点击展开
const showDetail = ref(false)
watch(isAllDone, (v) => {
  // 上传完毕后 2s 自动折叠
  if (v) {
    setTimeout(() => {
      if (isAllDone.value) showDetail.value = false
    }, 2000)
  }
})

function clearFinished() {
  // 仅清除 success 的任务，保留 error / uploading / pending
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
      <!-- 拖拽区 + NUpload 触发器 -->
      <div
        class="border-2 border-dashed rounded-lg p-8 text-center cursor-pointer transition"
        :class="dragOver ? 'border-primary-500 bg-primary-50' : 'border-gray-300 hover:border-primary-400'"
        @dragover="onDragOver"
        @dragleave="onDragLeave"
        @drop="onDrop"
      >
        <NUpload
          :custom-request="customRequest"
          :accept="ACCEPT_ATTR"
          :multiple="multiple"
          :show-file-list="false"
          :default-upload="true"
        >
          <div class="text-center">
            <div class="text-4xl mb-2">📤</div>
            <p class="text-sm text-gray-700">点击选择，或拖拽文件到此处</p>
            <p class="mt-1 text-xs text-gray-500">
              支持 jpeg/png/gif/webp/bmp，单文件不超过 {{ maxSizeMB }}MB（也可 Ctrl+V 粘贴）
            </p>
          </div>
        </NUpload>
      </div>

      <!-- 手动上传按钮（autoUpload=false 时） -->
      <div v-if="!autoUpload && hasPending" class="mt-3 text-center">
        <NButton type="primary" @click="uploadPending">上传全部</NButton>
      </div>

      <!-- 聚合进度条（始终单行，不占大量空间） -->
      <div
        v-if="totalCount > 0"
        class="mt-3 bg-white border border-gray-200 rounded-lg p-3"
      >
        <!-- 状态行 -->
        <div class="flex items-center gap-3 text-sm">
          <div class="flex-1 min-w-0">
            <span v-if="uploadingCount > 0 || pendingCount > 0" class="text-gray-800">
              正在上传 <b class="text-primary-600">{{ totalCount }}</b> 个文件
              <span v-if="successCount > 0" class="text-green-600">· 已完成 {{ successCount }}</span>
              <span v-if="errorCount > 0" class="text-red-500">· 失败 {{ errorCount }}</span>
            </span>
            <span v-else-if="isAllDone && errorCount > 0" class="text-red-600">
              上传完成 · 成功 {{ successCount }} · 失败 {{ errorCount }}
            </span>
            <span v-else class="text-green-600">
              上传完成 · 成功 {{ successCount }} 个文件
            </span>
          </div>
          <div class="flex items-center gap-2 shrink-0">
            <button
              v-if="hasFailures && isAllDone"
              class="px-2 py-1 text-xs border border-red-300 text-red-600 rounded hover:bg-red-50"
              @click="retryAllFailed"
            >重试失败</button>
            <button
              v-if="isAllDone && successCount > 0"
              class="px-2 py-1 text-xs text-gray-500 hover:text-gray-700"
              @click="clearFinished"
            >清除已完成</button>
            <button
              class="px-2 py-1 text-xs text-primary-600 hover:underline"
              @click="showDetail = !showDetail"
            >{{ showDetail ? '收起' : '详情' }} {{ showDetail ? '▴' : '▾' }}</button>
          </div>
        </div>
        <!-- 整体进度条 -->
        <div class="mt-2 w-full bg-gray-200 rounded-full overflow-hidden" style="height: 4px;">
          <div
            class="h-full rounded-full transition-all duration-200 ease-out"
            :class="hasFailures && isAllDone ? 'bg-red-500' : (isAllDone ? 'bg-green-500' : 'bg-primary-600')"
            :style="{ width: overallProgress + '%' }"
          ></div>
        </div>
      </div>

      <!-- 文件详情列表（按需展开，默认折叠） -->
      <div v-if="showDetail && totalCount > 0" class="mt-2 max-h-72 overflow-y-auto space-y-2 pr-1">
        <div
          v-for="task in tasks"
          :key="task.id"
          class="flex items-center gap-2 bg-gray-50 px-2 py-1.5 rounded text-sm"
        >
          <!-- 缩略图预览 -->
          <div class="w-8 h-8 flex-shrink-0 bg-gray-200 rounded overflow-hidden">
            <img :src="task.preview" :alt="task.name" class="w-full h-full object-cover" />
          </div>

          <!-- 文件名 + 进度条 -->
          <div class="flex-1 min-w-0">
            <div class="flex items-center justify-between text-xs">
              <span class="truncate text-gray-800">{{ task.name }}</span>
              <span class="text-gray-500 ml-2 shrink-0">
                <template v-if="task.status === 'success'">✓</template>
                <template v-else-if="task.status === 'error'">✕</template>
                <template v-else>{{ task.progress }}%</template>
                · {{ formatSize(task.size) }}
              </span>
            </div>
            <div class="mt-1 w-full bg-gray-200 rounded-full overflow-hidden" style="height: 3px;">
              <div
                class="h-full rounded-full transition-all duration-200 ease-out"
                :class="task.status === 'success' ? 'bg-green-500' : (task.status === 'error' ? 'bg-red-500' : 'bg-primary-600')"
                :style="{ width: task.progress + '%' }"
              ></div>
            </div>
          </div>

          <!-- 操作按钮 -->
          <div class="flex-shrink-0 flex gap-1 items-center">
            <button
              v-if="task.status === 'error'"
              class="text-primary-600 text-xs px-1 hover:underline"
              @click="retryTask(task)"
            >重试</button>
            <button
              class="text-gray-400 hover:text-red-500 text-sm px-1"
              title="移除"
              @click="removeTask(task)"
            >✕</button>
          </div>
        </div>
      </div>
    </div>

    <template #fallback>
      <div class="border-2 border-dashed rounded-lg p-8 text-center border-gray-300">
        <div class="text-4xl mb-2">📤</div>
        <p class="text-sm text-gray-700">加载上传组件...</p>
      </div>
    </template>
  </ClientOnly>
</template>
