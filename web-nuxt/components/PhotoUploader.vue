<script setup lang="ts">
// 图片上传组件：拖拽 / 粘贴 / 批量选择，独立进度条，缩略图预览，失败重试
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { NUpload, NProgress, NButton } from 'naive-ui'
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
    if (e.lengthComputable) {
      item.progress = Math.ceil((e.loaded / e.total) * 100)
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
          emit('uploaded', data)
          callbacks?.onFinish?.()
        } else {
          item.status = 'error'
          item.errorMsg = '响应数据异常'
          emit('error', `${item.name} 上传失败：响应数据异常`)
          callbacks?.onError?.()
        }
      } catch {
        item.status = 'error'
        item.errorMsg = '解析响应失败'
        emit('error', `${item.name} 上传失败：解析响应失败`)
        callbacks?.onError?.()
      }
    } else {
      item.status = 'error'
      item.errorMsg = `HTTP ${xhr.status}`
      emit('error', `${item.name} 上传失败：HTTP ${xhr.status}`)
      callbacks?.onError?.()
    }
  }

  xhr.onerror = () => {
    item.status = 'error'
    item.errorMsg = '网络错误'
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

      <!-- 文件列表 + 独立进度条 -->
      <div v-if="tasks.length > 0" class="mt-4 space-y-3">
        <div
          v-for="task in tasks"
          :key="task.id"
          class="flex items-center gap-3 bg-gray-50 px-3 py-2 rounded"
        >
          <!-- 缩略图预览 -->
          <div class="w-12 h-12 flex-shrink-0 bg-gray-200 rounded overflow-hidden">
            <img :src="task.preview" :alt="task.name" class="w-full h-full object-cover" />
          </div>

          <!-- 文件信息 + 进度条 -->
          <div class="flex-1 min-w-0">
            <div class="flex items-center justify-between text-sm">
              <span class="truncate text-gray-800">{{ task.name }}</span>
              <span class="text-xs text-gray-500 ml-2 shrink-0">{{ formatSize(task.size) }}</span>
            </div>
            <div class="mt-1">
              <NProgress
                v-if="task.status === 'uploading' || task.status === 'success'"
                :percentage="task.progress"
                :status="task.status === 'success' ? 'success' : 'default'"
                :show-indicator="false"
                :height="6"
              />
              <div v-else-if="task.status === 'pending'" class="text-xs text-gray-400">
                等待上传...
              </div>
              <div v-else class="text-xs text-red-500">
                {{ task.errorMsg || '上传失败' }}
              </div>
            </div>
          </div>

          <!-- 操作按钮 -->
          <div class="flex-shrink-0 flex gap-1 items-center">
            <NButton
              v-if="task.status === 'error'"
              size="tiny"
              type="primary"
              ghost
              @click="retryTask(task)"
            >
              重试
            </NButton>
            <button
              class="text-red-500 text-xs px-1 hover:text-red-700"
              title="移除"
              @click="removeTask(task)"
            >
              ✕
            </button>
          </div>
        </div>
      </div>

      <!-- 汇总 -->
      <div v-if="successCount > 0" class="mt-3 text-xs text-gray-500">
        已成功上传 {{ successCount }} 张
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
