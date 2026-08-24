// 分享管理：列出分享 + 复制链接 + 编辑（密码/过期时间）+ 删除
import { useEffect, useState } from 'react'
import { Share2, Copy, Trash2, ExternalLink, Lock, Pencil } from 'lucide-react'
import { AppShell } from './AppShell'
import { PageHeader } from './PageHeader'
import { useConfirm } from './ConfirmDialog'
import { ShareEditDialog } from './ShareEditDialog'
import { useApi } from '@/lib/api'
import { toast } from '@/lib/react-store'
import { formatDate, timeAgo } from '@/lib/utils'
import { Button } from '../ui/button'
import { Badge } from '../ui/badge'

interface Share {
  id: number
  shareable_type: string
  shareable_id: number
  slug: string
  has_password: boolean
  views: number
  expires_at: string | null
  created_at: string
}

export function SharesPage() {
  const api = useApi()
  const { confirm, node } = useConfirm()

  const [shares, setShares] = useState<Share[]>([])
  const [loading, setLoading] = useState(true)
  const [editing, setEditing] = useState<Share | null>(null)

  const load = () => {
    setLoading(true)
    api.get<any>('/api/v1/shares', { raw: true })
      .then((r) => {
        const data = r?.data?.data ?? r?.data ?? []
        setShares(Array.isArray(data) ? data : [])
      })
      .catch(() => setShares([]))
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const shareUrl = (slug: string) => `${window.location.origin}/share/${slug}`

  const copyLink = async (slug: string) => {
    try {
      await navigator.clipboard.writeText(shareUrl(slug))
      toast.success('链接已复制')
    } catch {
      toast.error('复制失败，请手动复制')
    }
  }

  const remove = async (s: Share) => {
    const ok = await confirm({ title: '删除分享', message: '确定删除这个分享？删除后链接将失效。', okText: '删除', danger: true })
    if (!ok) return
    try {
      await api.del(`/api/v1/shares/${s.id}`)
      toast.success('已删除')
      load()
    } catch (e: any) {
      toast.error(e?.message || '删除失败')
    }
  }

  const expired = (s: Share) => !!s.expires_at && new Date(s.expires_at).getTime() < Date.now()

  return (
    <AppShell>
      <PageHeader title="分享管理" description={`共 ${shares.length} 条分享`} />

      {loading ? (
        <div className="space-y-3">
          {Array.from({ length: 4 }).map((_, i) => (
            <div key={i} className="skeleton h-16 rounded-md" />
          ))}
        </div>
      ) : shares.length === 0 ? (
        <div className="rounded-md border border-dashed border-border py-20 text-center">
          <p className="text-sm text-muted-foreground">还没有分享。在图片页选择图片后点「批量分享」，或创建分享链接。</p>
        </div>
      ) : (
        <div className="space-y-3">
          {shares.map((s) => (
            <div key={s.id} className="flex flex-wrap items-center gap-3 rounded-md border border-border bg-card p-4">
              <div className="grid h-9 w-9 shrink-0 place-items-center rounded-md bg-accent">
                <Share2 className="h-4 w-4 text-accent-foreground" />
              </div>
              <div className="min-w-0 flex-1">
                <div className="flex items-center gap-2">
                  <span className="font-mono text-sm">{s.slug}</span>
                  {s.shareable_type === 'photo' ? <Badge variant="brand" className="text-[10px]">图片</Badge> : <Badge variant="secondary" className="text-[10px]">相册</Badge>}
                  {s.has_password && <Lock className="h-3 w-3 text-muted-foreground" />}
                  {expired(s) && <Badge variant="destructive" className="text-[10px]">已过期</Badge>}
                </div>
                <p className="mt-1 text-xs text-muted-foreground">
                  {s.views} 次浏览 · 创建于 {timeAgo(s.created_at)}
                  {s.expires_at && ` · 过期于 ${formatDate(s.expires_at)}`}
                </p>
              </div>
              <div className="flex shrink-0 items-center gap-1.5">
                <Button variant="outline" size="sm" onClick={() => copyLink(s.slug)}>
                  <Copy className="h-3.5 w-3.5" /> 复制链接
                </Button>
                <Button variant="ghost" size="icon" aria-label="编辑" onClick={() => setEditing(s)}>
                  <Pencil className="h-4 w-4" />
                </Button>
                <a href={`/share/${s.slug}`} target="_blank" rel="noreferrer">
                  <Button variant="ghost" size="icon" aria-label="打开">
                    <ExternalLink className="h-4 w-4" />
                  </Button>
                </a>
                <Button variant="ghost" size="icon" className="text-destructive" onClick={() => remove(s)} aria-label="删除">
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>
            </div>
          ))}
        </div>
      )}

      <ShareEditDialog share={editing} open={!!editing} onOpenChange={(o) => !o && setEditing(null)} onSaved={load} />

      {node}
    </AppShell>
  )
}
