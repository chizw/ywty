// 批量打标签对话框：列出我的标签，勾选 = 应用到全部选中图片，取消勾选 = 从选中图片移除
// attach/detach: POST /api/v1/tags/attach | /detach  body: { target_type, target_id, tag_id }
import { useEffect, useState } from 'react'
import { Plus } from 'lucide-react'
import { useApi } from '@/lib/api'
import { toast } from '@/lib/react-store'
import { Button } from '../ui/button'
import { Checkbox } from '../ui/checkbox'
import { Input } from '../ui/input'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '../ui/dialog'

interface TagItem {
  id: number
  name: string
}

interface TagAttachDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  photoIds: number[]
  onSaved?: () => void
}

function extractList(r: any): any[] {
  if (Array.isArray(r)) return r
  return Array.isArray(r?.data) ? r.data : []
}

export function TagAttachDialog({ open, onOpenChange, photoIds, onSaved }: TagAttachDialogProps) {
  const api = useApi()

  const [tags, setTags] = useState<TagItem[]>([])
  const [checked, setChecked] = useState<Set<number>>(new Set())
  // 每张图片当前已绑定的标签（用于计算增量请求）
  const [initialMap, setInitialMap] = useState<Map<number, Set<number>>>(new Map())
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [newName, setNewName] = useState('')
  const [creating, setCreating] = useState(false)

  useEffect(() => {
    if (!open || photoIds.length === 0) return
    let cancelled = false
    setLoading(true)
    Promise.all([
      api.get<any>('/api/v1/tags').catch(() => null),
      ...photoIds.map((id) =>
        api
          .get<any>('/api/v1/tags', { query: { target_type: 'photo', target_id: id } })
          .catch(() => null)
      ),
    ])
      .then((results) => {
        if (cancelled) return
        const allTags: TagItem[] = extractList(results[0]).map((t: any) => ({ id: Number(t.id), name: String(t.name) }))
        const map = new Map<number, Set<number>>()
        photoIds.forEach((pid, i) => {
          const list = extractList(results[i + 1])
          map.set(pid, new Set(list.map((t: any) => Number(t.id))))
        })
        setTags(allTags)
        setInitialMap(map)
        // 初始勾选 = 被全部选中图片共有的标签
        const common = allTags
          .filter((t) => photoIds.every((pid) => map.get(pid)?.has(t.id)))
          .map((t) => t.id)
        setChecked(new Set(common))
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open])

  useEffect(() => {
    if (!open) {
      setTags([])
      setChecked(new Set())
      setInitialMap(new Map())
      setNewName('')
    }
  }, [open])

  const toggle = (id: number) =>
    setChecked((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })

  const createTag = async () => {
    const name = newName.trim()
    if (!name || creating) return
    setCreating(true)
    try {
      const res = await api.post<any>('/api/v1/tags', { name })
      const tag = res?.data ?? res
      const item: TagItem = { id: Number(tag?.id ?? Date.now()), name: String(tag?.name ?? name) }
      setTags((prev) => (prev.some((t) => t.name === item.name) ? prev : [...prev, item]))
      setChecked((prev) => new Set(prev).add(item.id))
      setNewName('')
    } catch (e: any) {
      toast.error(e?.message || '创建标签失败')
    } finally {
      setCreating(false)
    }
  }

  const submit = async () => {
    setSaving(true)
    try {
      const ops: Promise<unknown>[] = []
      for (const t of tags) {
        for (const pid of photoIds) {
          const had = initialMap.get(pid)?.has(t.id) ?? false
          const want = checked.has(t.id)
          if (want && !had) {
            ops.push(api.post('/api/v1/tags/attach', { target_type: 'photo', target_id: pid, tag_id: t.id }))
          } else if (!want && had) {
            ops.push(api.post('/api/v1/tags/detach', { target_type: 'photo', target_id: pid, tag_id: t.id }))
          }
        }
      }
      const results = await Promise.allSettled(ops)
      const failed = results.filter((r) => r.status === 'rejected').length
      if (failed > 0) {
        toast.error(`${failed} 项标签操作失败，请重试`)
      } else {
        toast.success(`已更新 ${photoIds.length} 张图片的标签`)
      }
      onOpenChange(false)
      onSaved?.()
    } finally {
      setSaving(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>批量打标签</DialogTitle>
          <DialogDescription>对选中的 {photoIds.length} 张图片：勾选的标签将应用到所有图片，取消勾选则从所有图片移除。</DialogDescription>
        </DialogHeader>

        {loading ? (
          <div className="space-y-2">
            {Array.from({ length: 4 }).map((_, i) => (
              <div key={i} className="skeleton h-8 rounded-md" />
            ))}
          </div>
        ) : (
          <>
            <div className="flex max-h-60 space-y-1 overflow-y-auto rounded-md border border-border p-2">
              {tags.length === 0 ? (
                <p className="py-6 text-center text-sm text-muted-foreground">还没有标签，先在下方新建一个。</p>
              ) : (
                tags.map((t) => (
                  <label key={t.id} className="flex cursor-pointer items-center gap-2.5 rounded px-2 py-1.5 text-sm hover:bg-accent">
                    <Checkbox checked={checked.has(t.id)} onCheckedChange={() => toggle(t.id)} />
                    {t.name}
                  </label>
                ))
              )}
            </div>
            <div className="flex gap-2">
              <Input
                value={newName}
                onChange={(e) => setNewName(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && createTag()}
                placeholder="新标签名称"
                maxLength={50}
              />
              <Button variant="outline" onClick={createTag} loading={creating} disabled={!newName.trim()}>
                <Plus className="h-4 w-4" /> 新建
              </Button>
            </div>
          </>
        )}

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={saving}>取消</Button>
          <Button onClick={submit} loading={saving} disabled={loading}>保存</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
