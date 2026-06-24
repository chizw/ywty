<script setup lang="ts">
// 通用文件上传组件
import { Button } from '~/components/ui/button'
import { Card, CardContent } from '~/components/ui/card'
import { Upload, X } from '@lucide/vue'
import type { UploadResult } from '~/types/api'

const props = withDefaults(defineProps<{
  multiple?: boolean
  accept?: string
  maxSizeMB?: number
  autoUpload?: boolean
}>(), {
  multiple: true,
  accept: 'image/*',
  maxSizeMB: 20,
  autoUpload: true,
})

const emit = defineEmits<{
  uploaded: [UploadResult]
  error: [string]
}>()

const api = useApi()

const dragOver = ref(false)
const files = ref<File[]>([])
const uploading = ref(false)
const results = ref<UploadResult[]>([])
const errors = ref<string[]>([])
const inputRef = ref<HTMLInputElement | null>(null)

function pickFiles() {
  inputRef.value?.click()
}

function onFileChange(e: Event) {
  const input = e.target as HTMLInputElement
  if (!input.files) return
  addFiles(Array.from(input.files))
  input.value = ''
}

function onDrop(e: DragEvent) {
  e.preventDefault()
  dragOver.value = false
  if (!e.dataTransfer) return
  addFiles(Array.from(e.dataTransfer.files))
}

function onDragOver(e: DragEvent) {
  e.preventDefault()
  dragOver.value = true
}

function onDragLeave() {
  dragOver.value = false
}

function addFiles(list: File[]) {
  for (const f of list) {
    if (f.size > props.maxSizeMB * 1024 * 1024) {
      errors.value.push(`${f.name} 超过 ${props.maxSizeMB}MB`)
      continue
    }
    files.value.push(f)
  }
  if (props.autoUpload) uploadAll()
}

function removeFile(i: number) {
  files.value.splice(i, 1)
}

async function uploadAll() {
  if (files.value.length === 0 || uploading.value) return
  uploading.value = true
  errors.value = []
  results.value = []
  try {
    for (const f of files.value) {
      const form = new FormData()
      form.append('file', f)
      const res = await api.post<UploadResult>('/api/v1/photos', form, { raw: true })
      const data = (res as any).data as UploadResult
      if (data?.photo) {
        results.value.push(data)
        emit('uploaded', data)
      } else {
        errors.value.push(`${f.name} 上传失败`)
      }
    }
    files.value = []
  } catch (err: any) {
    const msg = err?.statusMessage || err?.message || 'upload failed'
    errors.value.push(msg)
    emit('error', msg)
  } finally {
    uploading.value = false
  }
}

function pasteFromClipboard(e: ClipboardEvent) {
  if (!e.clipboardData) return
  const items = Array.from(e.clipboardData.items)
  const list: File[] = []
  for (const it of items) {
    if (it.kind === 'file') {
      const f = it.getAsFile()
      if (f) list.push(f)
    }
  }
  if (list.length > 0) addFiles(list)
}

onMounted(() => {
  window.addEventListener('paste', pasteFromClipboard)
})
onBeforeUnmount(() => {
  window.removeEventListener('paste', pasteFromClipboard)
})
</script>

<template>
  <div>
    <div
      class="border-2 border-dashed rounded-lg p-8 text-center cursor-pointer transition-colors"
      :class="dragOver ? 'border-primary bg-primary/5' : 'border-border hover:border-primary/50'"
      @click="pickFiles"
      @drop="onDrop"
      @dragover="onDragOver"
      @dragleave="onDragLeave"
    >
      <input
        ref="inputRef"
        type="file"
        class="hidden"
        :accept="accept"
        :multiple="multiple"
        @change="onFileChange"
      />
      <Upload class="mx-auto h-10 w-10 text-muted-foreground mb-2" />
      <p class="text-sm text-foreground">拖拽文件到此处，或点击选择</p>
      <p class="mt-1 text-xs text-muted-foreground">
        支持 {{ accept }}，单文件不超�?{{ maxSizeMB }}MB（也�?Ctrl+V 粘贴图片�?      </p>
    </div>

    <div v-if="files.length > 0" class="mt-3 space-y-2">
      <div
        v-for="(f, i) in files"
        :key="i"
        class="flex items-center justify-between bg-muted px-3 py-2 rounded-md text-sm"
      >
        <span class="truncate text-foreground">{{ f.name }}</span>
        <span class="text-muted-foreground ml-3 shrink-0">{{ (f.size / 1024).toFixed(1) }} KB</span>
        <Button variant="ghost" size="sm" class="ml-3 h-6 w-6 p-0 text-destructive" @click.stop="removeFile(i)">
          <X class="h-3 w-3" />
        </Button>
      </div>
    </div>

    <div v-if="results.length > 0" class="mt-4">
      <h4 class="text-sm font-medium text-foreground mb-2">已上�?{{ results.length }} �?/h4>
      <div class="grid grid-cols-4 sm:grid-cols-6 gap-2">
        <a
          v-for="r in results"
          :key="r.photo.id"
          :href="r.url"
          target="_blank"
          class="block aspect-square bg-muted rounded-md overflow-hidden"
        >
          <img :src="r.url" :alt="r.photo.name" class="w-full h-full object-cover" />
        </a>
      </div>
    </div>

    <div v-if="errors.length > 0" class="mt-3 space-y-1">
      <p v-for="(e, i) in errors" :key="i" class="text-xs text-destructive">{{ e }}</p>
    </div>
  </div>
</template>
