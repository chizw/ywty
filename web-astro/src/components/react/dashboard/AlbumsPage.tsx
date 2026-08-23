// 相册列表：卡片 + 新建相册
import { useEffect, useState } from 'react'
import { BookImage, Plus } from 'lucide-react'
import { AppShell } from './AppShell'
import { PageHeader } from './PageHeader'
import { useConfirm } from './ConfirmDialog'
import { useApi } from '@/lib/api'
import { useStatsStore, toast } from '@/lib/react-store'
import { timeAgo } from '@/lib/utils'
import { Button } from '../ui/button'
import { Input } from '../ui/input'
import { Label } from '../ui/label'
import { Textarea } from '../ui/textarea'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '../ui/dialog'

interface Album {
  id: number
  name: string
  description: string | null
  is_public: boolean
  photo_count: number
  views: number
  created_at: string
}

export function AlbumsPage() {
  const api = useApi()
  const stats = useStatsStore()
  const { confirm, node } = useConfirm()

  const [albums, setAlbums] = useState<Album[]>([])
  const [loading, setLoading] = useState(true)
  const [showCreate, setShowCreate] = useState(false)
  const [creating, setCreating] = useState(false)
  const [form, setForm] = useState({ name: '', description: '' })

  const load = () => {
    setLoading(true)
    api.get<any>('/api/v1/albums', { raw: true })
      .then((r) => setAlbums(Array.isArray(r?.data) ? r.data : []))
      .catch(() => setAlbums([]))
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const create = async () => {
    if (!form.name.trim()) return
    setCreating(true)
    try {
      await api.post('/api/v1/albums', { name: form.name.trim(), description: form.description.trim() || null })
      toast.success('已创建相册')
      setShowCreate(false)
      setForm({ name: '', description: '' })
      load()
      stats.refresh()
    } catch (e: any) {
      toast.error(e?.message || '创建失败')
    } finally {
      setCreating(false)
    }
  }

  const remove = async (a: Album) => {
    const ok = await confirm({ title: '删除相册', message: `确定删除相册「${a.name}」？相册内的图片不会被删除。`, okText: '删除', danger: true })
    if (!ok) return
    try {
      await api.del(`/api/v1/albums/${a.id}`)
      toast.success('已删除')
      load()
      stats.refresh()
    } catch (e: any) {
      toast.error(e?.message || '删除失败')
    }
  }

  return (
    <AppShell>
      <PageHeader title="我的相册" description={`共 ${albums.length} 个`}>
        <Button size="sm" onClick={() => setShowCreate(true)}>
          <Plus className="h-4 w-4" /> 新建相册
        </Button>
      </PageHeader>

      {loading ? (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 6 }).map((_, i) => (
            <div key={i} className="skeleton h-32 rounded-md" />
          ))}
        </div>
      ) : albums.length === 0 ? (
        <div className="rounded-md border border-dashed border-border py-20 text-center">
          <p className="text-sm text-muted-foreground">还没有相册，创建一个来整理图片。</p>
        </div>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {albums.map((a) => (
            <a key={a.id} href={`/dashboard/albums/${a.id}`} className="card-hover group rounded-md border border-border bg-card p-5">
              <div className="flex items-start justify-between">
                <div className="grid h-10 w-10 place-items-center rounded-md bg-accent">
                  <BookImage className="h-5 w-5 text-accent-foreground" />
                </div>
                {a.is_public && <span className="text-xs text-muted-foreground">公开</span>}
              </div>
              <h3 className="mt-3 font-display text-lg font-semibold leading-snug">{a.name}</h3>
              {a.description && <p className="mt-1 line-clamp-2 text-sm text-muted-foreground">{a.description}</p>}
              <div className="mt-4 flex items-center justify-between text-xs text-muted-foreground">
                <span className="tabular-nums">{a.photo_count} 张图片 · {a.views} 浏览</span>
                <span>{timeAgo(a.created_at)}</span>
              </div>
              <div className="mt-3 opacity-0 transition-opacity group-hover:opacity-100">
                <Button variant="ghost" size="sm" className="h-6 text-xs text-destructive" onClick={(e) => { e.preventDefault(); remove(a) }}>
                  删除
                </Button>
              </div>
            </a>
          ))}
        </div>
      )}

      {/* 新建相册弹窗 */}
      <Dialog open={showCreate} onOpenChange={setShowCreate}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>新建相册</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label>相册名称</Label>
              <Input value={form.name} onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))} placeholder="例如：旅行记录" autoFocus />
            </div>
            <div className="space-y-2">
              <Label>描述（可选）</Label>
              <Textarea value={form.description} onChange={(e) => setForm((f) => ({ ...f, description: e.target.value }))} placeholder="一句话描述这个相册" />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowCreate(false)}>取消</Button>
            <Button onClick={create} loading={creating} disabled={!form.name.trim()}>创建</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {node}
    </AppShell>
  )
}
