// 移入相册对话框：把选中图片批量加入目标相册
// POST /api/v1/albums/:id/photos  body: { photo_ids: number[] }
import { useEffect, useState } from 'react'
import { useApi } from '@/lib/api'
import { toast } from '@/lib/react-store'
import { Button } from '../ui/button'
import { Label } from '../ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '../ui/select'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '../ui/dialog'

interface AlbumItem {
  id: number
  name: string
}

interface MoveToAlbumDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  photoIds: number[]
  onDone?: () => void
}

export function MoveToAlbumDialog({ open, onOpenChange, photoIds, onDone }: MoveToAlbumDialogProps) {
  const api = useApi()

  const [albums, setAlbums] = useState<AlbumItem[]>([])
  const [albumId, setAlbumId] = useState('')
  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    if (!open) return
    setLoading(true)
    setAlbumId('')
    api
      .get<any>('/api/v1/albums')
      .then((r) => {
        const list = Array.isArray(r) ? r : (r?.data ?? [])
        setAlbums(list.map((a: any) => ({ id: Number(a.id), name: String(a.name) })))
      })
      .catch(() => setAlbums([]))
      .finally(() => setLoading(false))
  }, [open])

  const submit = async () => {
    if (!albumId || submitting) return
    setSubmitting(true)
    try {
      const res = await api.post<any>(`/api/v1/albums/${albumId}/photos`, { photo_ids: photoIds })
      const added = Number(res?.added ?? photoIds.length)
      const name = albums.find((a) => String(a.id) === albumId)?.name ?? ''
      toast.success(`已将 ${added} 张图片移入「${name}」`)
      onOpenChange(false)
      onDone?.()
    } catch (e: any) {
      toast.error(e?.message || '移入失败')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle>移入相册</DialogTitle>
          <DialogDescription>将选中的 {photoIds.length} 张图片添加到目标相册。</DialogDescription>
        </DialogHeader>

        <div className="space-y-2">
          <Label>目标相册</Label>
          <Select value={albumId} onValueChange={setAlbumId} disabled={loading}>
            <SelectTrigger>
              <SelectValue placeholder={loading ? '加载中…' : '选择相册'} />
            </SelectTrigger>
            <SelectContent>
              {albums.map((a) => (
                <SelectItem key={a.id} value={String(a.id)}>{a.name}</SelectItem>
              ))}
            </SelectContent>
          </Select>
          {!loading && albums.length === 0 && (
            <p className="text-xs text-muted-foreground">还没有相册，请先在「我的相册」中创建。</p>
          )}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={submitting}>取消</Button>
          <Button onClick={submit} loading={submitting} disabled={!albumId || loading}>移入</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
