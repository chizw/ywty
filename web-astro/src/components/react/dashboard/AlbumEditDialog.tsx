// 编辑相册对话框：名称 / 描述 / 是否公开
// PATCH /api/v1/albums/:id  body: { name?, description?, is_public? }
import { useEffect, useState } from 'react'
import { useApi } from '@/lib/api'
import { toast } from '@/lib/react-store'
import { Button } from '../ui/button'
import { Checkbox } from '../ui/checkbox'
import { Input } from '../ui/input'
import { Label } from '../ui/label'
import { Textarea } from '../ui/textarea'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '../ui/dialog'

interface AlbumRow {
  id: number
  name: string
  description: string | null
  is_public: boolean
}

interface AlbumEditDialogProps {
  album: AlbumRow | null
  open: boolean
  onOpenChange: (open: boolean) => void
  onSaved?: () => void
}

export function AlbumEditDialog({ album, open, onOpenChange, onSaved }: AlbumEditDialogProps) {
  const api = useApi()

  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [isPublic, setIsPublic] = useState(false)
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    if (open && album) {
      setName(album.name)
      setDescription(album.description ?? '')
      setIsPublic(!!album.is_public)
      setSaving(false)
    }
  }, [open, album])

  const submit = async () => {
    if (!album || !name.trim()) return
    setSaving(true)
    try {
      await api.patch(`/api/v1/albums/${album.id}`, {
        name: name.trim(),
        description: description.trim() || null,
        is_public: isPublic,
      })
      toast.success('相册已更新')
      onOpenChange(false)
      onSaved?.()
    } catch (e: any) {
      toast.error(e?.message || '更新失败')
    } finally {
      setSaving(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>编辑相册</DialogTitle>
        </DialogHeader>
        <div className="space-y-4">
          <div className="space-y-2">
            <Label>相册名称</Label>
            <Input value={name} onChange={(e) => setName(e.target.value)} maxLength={100} autoFocus />
          </div>
          <div className="space-y-2">
            <Label>描述（可选）</Label>
            <Textarea value={description} onChange={(e) => setDescription(e.target.value)} placeholder="一句话描述这个相册" />
          </div>
          <label className="flex cursor-pointer items-center gap-2 text-sm">
            <Checkbox checked={isPublic} onCheckedChange={(v) => setIsPublic(v === true)} />
            公开相册（任何人可通过链接访问）
          </label>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={saving}>取消</Button>
          <Button onClick={submit} loading={saving} disabled={!name.trim()}>保存</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
