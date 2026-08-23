// 图片上传（React）：拖拽/粘贴/批量，XHR 真实进度，失败重试
import { useEffect, useRef, useState } from 'react'
import { Upload, X, RefreshCw, ChevronDown, ChevronUp } from 'lucide-react'
import { useAuthStore } from '@/lib/react-store'
import { toast } from '@/lib/react-store'
import { ACCEPT_IMAGE_TYPES, ACCEPT_IMAGE_ATTR, MAX_UPLOAD_MB } from '@/lib/constants'
import { formatBytes } from '@/lib/utils'
import { cn } from '@/lib/utils'
import { Button } from '../ui/button'
import type { UploadResult } from '@/lib/types'

interface Task {
  id: string
  file: File
  name: string
  size: number
  preview: string
  progress: number
  status: 'pending' | 'uploading' | 'success' | 'error'
  errorMsg?: string
  result?: UploadResult
  xhr?: XMLHttpRequest
}

let taskSeq = 0
const nextId = () => `task-${Date.now()}-${++taskSeq}`

export function PhotoUploader({
  onUploaded,
  multiple = true,
  maxSizeMB = MAX_UPLOAD_MB,
}: {
  onUploaded: (r: UploadResult) => void
  multiple?: boolean
  maxSizeMB?: number
}) {
  const auth = useAuthStore()
  const [tasks, setTasks] = useState<Task[]>([])
  const [dragOver, setDragOver] = useState(false)
  const [showDetail, setShowDetail] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)

  const patch = (id: string, p: Partial<Task>) =>
    setTasks((ts) => ts.map((t) => (t.id === id ? { ...t, ...p } : t)))

  function validate(file: File): string | null {
    if (!ACCEPT_IMAGE_TYPES.includes(file.type)) {
      return `${file.name} 类型不被允许，仅支持 jpeg/png/gif/webp/bmp`
    }
    if (file.size > maxSizeMB * 1024 * 1024) return `${file.name} 超过 ${maxSizeMB}MB 限制`
    return null
  }

  function addFiles(list: FileList | File[]) {
    const arr = Array.from(list)
    const newTasks: Task[] = []
    for (const f of arr) {
      const dup = tasks.some((t) => t.file.name === f.name && t.file.size === f.size && t.file.lastModified === f.lastModified)
      if (dup) continue
      const err = validate(f)
      if (err) {
        toast.error(err)
        continue
      }
      newTasks.push({
        id: nextId(),
        file: f,
        name: f.name,
        size: f.size,
        preview: URL.createObjectURL(f),
        progress: 0,
        status: 'pending',
      })
    }
    if (newTasks.length) setTasks((ts) => [...ts, ...newTasks])
    newTasks.forEach(uploadTask)
  }

  function uploadTask(item: Task) {
    if (item.status === 'uploading') return
    patch(item.id, { status: 'uploading', progress: 0, errorMsg: undefined })

    const form = new FormData()
    form.append('file', item.file)

    const xhr = new XMLHttpRequest()
    xhr.open('POST', `/api/v1/photos`)
    if (auth.accessToken) xhr.setRequestHeader('Authorization', `Bearer ${auth.accessToken}`)

    xhr.upload.onprogress = (e) => {
      if (e.lengthComputable && e.total > 0) {
        patch(item.id, { progress: Math.min(99, Math.ceil((e.loaded / e.total) * 100)) })
      }
    }
    xhr.onload = () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        try {
          const res = JSON.parse(xhr.responseText)
          const data = res?.data as UploadResult | undefined
          if (data) {
            patch(item.id, { status: 'success', progress: 100, result: data })
            onUploaded(data)
          } else {
            patch(item.id, { status: 'error', errorMsg: '响应数据异常' })
            toast.error(`${item.name} 上传失败`)
          }
        } catch {
          patch(item.id, { status: 'error', errorMsg: '解析响应失败' })
          toast.error(`${item.name} 上传失败`)
        }
      } else {
        patch(item.id, { status: 'error', errorMsg: `HTTP ${xhr.status}` })
        toast.error(`${item.name} 上传失败（HTTP ${xhr.status}）`)
      }
    }
    xhr.onerror = () => {
      patch(item.id, { status: 'error', errorMsg: '网络错误' })
      toast.error(`${item.name} 上传失败：网络错误`)
    }
    item.xhr = xhr
    xhr.send(form)
  }

  function retry(id: string) {
    const t = tasks.find((x) => x.id === id)
    if (t) uploadTask(t)
  }

  function remove(id: string) {
    const t = tasks.find((x) => x.id === id)
    if (t?.xhr && t.status === 'uploading') t.xhr.abort()
    if (t?.preview) URL.revokeObjectURL(t.preview)
    setTasks((ts) => ts.filter((x) => x.id !== id))
  }

  // 全局粘贴
  useEffect(() => {
    const onPaste = (e: ClipboardEvent) => {
      if (!e.clipboardData) return
      const files = Array.from(e.clipboardData.items)
        .filter((it) => it.kind === 'file')
        .map((it) => it.getAsFile())
        .filter((f): f is File => !!f)
      if (files.length) addFiles(files)
    }
    window.addEventListener('paste', onPaste)
    return () => {
      window.removeEventListener('paste', onPaste)
      tasks.forEach((t) => t.preview && URL.revokeObjectURL(t.preview))
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [tasks.length])

  const uploading = tasks.filter((t) => t.status === 'uploading' || t.status === 'pending')
  const successCount = tasks.filter((t) => t.status === 'success').length
  const errorCount = tasks.filter((t) => t.status === 'error').length
  const isAllDone = tasks.length > 0 && uploading.length === 0

  const overall = tasks.length
    ? Math.ceil(
        tasks.reduce((s, t) => s + t.progress * (t.size || 1), 0) /
          tasks.reduce((s, t) => s + (t.size || 1), 0)
      )
    : 0

  return (
    <div>
      <div
        className={cn(
          'cursor-pointer rounded-md border-2 border-dashed p-8 text-center transition-colors',
          dragOver ? 'border-brand bg-brand/5' : 'border-border hover:border-brand/50'
        )}
        onClick={() => inputRef.current?.click()}
        onDragOver={(e) => {
          e.preventDefault()
          setDragOver(true)
        }}
        onDragLeave={() => setDragOver(false)}
        onDrop={(e) => {
          e.preventDefault()
          setDragOver(false)
          if (e.dataTransfer) addFiles(e.dataTransfer.files)
        }}
      >
        <input
          ref={inputRef}
          type="file"
          className="hidden"
          accept={ACCEPT_IMAGE_ATTR}
          multiple={multiple}
          onChange={(e) => {
            if (e.target.files) addFiles(e.target.files)
            e.target.value = ''
          }}
        />
        <Upload className="mx-auto mb-2 h-8 w-8 text-muted-foreground" />
        <p className="text-sm text-foreground">点击选择，或拖拽文件到此处</p>
        <p className="mt-1 text-xs text-muted-foreground">
          支持 jpeg/png/gif/webp/bmp，单文件不超过 {maxSizeMB}MB（也可 Ctrl+V 粘贴）
        </p>
      </div>

      {tasks.length > 0 && (
        <div className="mt-3 rounded-md border border-border bg-card p-3">
          <div className="flex items-center gap-3 text-sm">
            <div className="flex-1 min-w-0">
              {!isAllDone ? (
                <span>
                  正在上传 <b className="text-brand">{tasks.length}</b> 个文件
                  {successCount > 0 && <span className="text-emerald-600"> · 完成 {successCount}</span>}
                  {errorCount > 0 && <span className="text-destructive"> · 失败 {errorCount}</span>}
                </span>
              ) : errorCount > 0 ? (
                <span className="text-destructive">上传完成 · 成功 {successCount} · 失败 {errorCount}</span>
              ) : (
                <span className="text-emerald-600">上传完成 · 成功 {successCount} 个文件</span>
              )}
            </div>
            <div className="flex shrink-0 items-center gap-2">
              {isAllDone && errorCount > 0 && (
                <Button variant="outline" size="sm" className="h-7 text-xs text-destructive" onClick={() => tasks.filter((t) => t.status === 'error').forEach((t) => retry(t.id))}>
                  <RefreshCw className="h-3 w-3" /> 重试失败
                </Button>
              )}
              <Button variant="ghost" size="sm" className="h-7 text-xs text-brand" onClick={() => setShowDetail((v) => !v)}>
                {showDetail ? '收起' : '详情'}
                {showDetail ? <ChevronUp className="h-3 w-3" /> : <ChevronDown className="h-3 w-3" />}
              </Button>
            </div>
          </div>
          <div className="mt-2 h-1 w-full overflow-hidden rounded-full bg-muted">
            <div
              className={cn(
                'h-full rounded-full transition-all duration-200',
                isAllDone && errorCount > 0 ? 'bg-destructive' : isAllDone ? 'bg-emerald-500' : 'bg-brand'
              )}
              style={{ width: `${overall}%` }}
            />
          </div>

          {showDetail && (
            <div className="mt-2 max-h-72 space-y-2 overflow-y-auto pr-1">
              {tasks.map((t) => (
                <div key={t.id} className="flex items-center gap-2 rounded-md bg-muted px-2 py-1.5 text-sm">
                  <div className="h-8 w-8 flex-shrink-0 overflow-hidden rounded bg-muted">
                    <img src={t.preview} alt={t.name} className="h-full w-full object-cover" />
                  </div>
                  <div className="min-w-0 flex-1">
                    <div className="flex items-center justify-between text-xs">
                      <span className="truncate">{t.name}</span>
                      <span className="ml-2 shrink-0 text-muted-foreground">
                        {t.status === 'success' ? '✓' : t.status === 'error' ? '✗' : `${t.progress}%`} · {formatBytes(t.size)}
                      </span>
                    </div>
                    <div className="mt-1 h-[3px] w-full overflow-hidden rounded-full bg-muted">
                      <div
                        className={cn('h-full rounded-full', t.status === 'success' ? 'bg-emerald-500' : t.status === 'error' ? 'bg-destructive' : 'bg-brand')}
                        style={{ width: `${t.progress}%` }}
                      />
                    </div>
                  </div>
                  <div className="flex shrink-0 items-center gap-1">
                    {t.status === 'error' && (
                      <Button variant="ghost" size="sm" className="h-6 px-1 text-xs text-brand" onClick={() => retry(t.id)}>
                        重试
                      </Button>
                    )}
                    <Button variant="ghost" size="sm" className="h-6 w-6 p-0 text-muted-foreground hover:text-destructive" onClick={() => remove(t.id)}>
                      <X className="h-3 w-3" />
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  )
}
