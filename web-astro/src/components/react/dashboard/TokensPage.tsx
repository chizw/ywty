// API Token 管理
import { useEffect, useState } from 'react'
import { Key, Plus, Trash2, Copy } from 'lucide-react'
import { AppShell } from './AppShell'
import { PageHeader } from './PageHeader'
import { useConfirm } from './ConfirmDialog'
import { useApi } from '@/lib/api'
import { toast } from '@/lib/react-store'
import { formatDate } from '@/lib/utils'
import { Button } from '../ui/button'
import { Input } from '../ui/input'
import { Label } from '../ui/label'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '../ui/dialog'

interface Token {
  id: number
  name: string
  token: string
  scopes: string | null
  last_used_at: string | null
  expires_at: string | null
  created_at: string
}

export function TokensPage() {
  const api = useApi()
  const { confirm, node } = useConfirm()

  const [tokens, setTokens] = useState<Token[]>([])
  const [loading, setLoading] = useState(true)
  const [showNew, setShowNew] = useState(false)
  const [name, setName] = useState('')
  const [creating, setCreating] = useState(false)
  const [revealed, setRevealed] = useState<number | null>(null)

  const load = () => {
    setLoading(true)
    api.get<any>('/api/v1/tokens', { raw: true })
      .then((r) => {
        const data = r?.data?.data ?? r?.data ?? []
        setTokens(Array.isArray(data) ? data : [])
      })
      .catch(() => setTokens([]))
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const create = async () => {
    if (!name.trim()) return
    setCreating(true)
    try {
      await api.post('/api/v1/tokens', { name: name.trim() })
      toast.success('Token 已创建')
      setShowNew(false)
      setName('')
      load()
    } catch (e: any) {
      toast.error(e?.message || '创建失败')
    } finally {
      setCreating(false)
    }
  }

  const revoke = async (t: Token) => {
    const ok = await confirm({ title: '撤销 Token', message: `确定撤销 Token「${t.name}」？使用它的应用将立即失效。`, okText: '撤销', danger: true })
    if (!ok) return
    try {
      await api.del(`/api/v1/tokens/${t.id}`)
      toast.success('已撤销')
      load()
    } catch (e: any) {
      toast.error(e?.message || '撤销失败')
    }
  }

  const copy = async (s: string) => {
    try {
      await navigator.clipboard.writeText(s)
      toast.success('已复制')
    } catch {
      toast.error('复制失败')
    }
  }

  const masked = (t: string) => (revealed === null ? `${t.slice(0, 8)}…${t.slice(-4)}` : t)

  return (
    <AppShell>
      <PageHeader title="API Token" description="用于第三方应用访问你的图库">
        <Button size="sm" onClick={() => setShowNew(true)}>
          <Plus className="h-4 w-4" /> 新建 Token
        </Button>
      </PageHeader>

      {loading ? (
        <div className="space-y-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className="skeleton h-16 rounded-md" />
          ))}
        </div>
      ) : tokens.length === 0 ? (
        <div className="rounded-md border border-dashed border-border py-20 text-center">
          <Key className="mx-auto mb-3 h-8 w-8 text-muted-foreground" />
          <p className="text-sm text-muted-foreground">还没有 Token。</p>
        </div>
      ) : (
        <div className="space-y-3">
          {tokens.map((t) => (
            <div key={t.id} className="flex flex-wrap items-center gap-3 rounded-md border border-border bg-card p-4">
              <div className="grid h-9 w-9 shrink-0 place-items-center rounded-md bg-accent">
                <Key className="h-4 w-4 text-accent-foreground" />
              </div>
              <div className="min-w-0 flex-1">
                <p className="text-sm font-medium">{t.name}</p>
                <button className="mt-0.5 font-mono text-xs text-muted-foreground hover:text-foreground" onClick={() => setRevealed((r) => (r === t.id ? null : t.id))}>
                  {masked(t.token)}
                </button>
                <p className="mt-0.5 text-xs text-muted-foreground">
                  创建于 {formatDate(t.created_at)}
                  {t.last_used_at && ` · 最近使用 ${formatDate(t.last_used_at)}`}
                </p>
              </div>
              <div className="flex shrink-0 items-center gap-1.5">
                <Button variant="outline" size="sm" onClick={() => copy(t.token)}>
                  <Copy className="h-3.5 w-3.5" /> 复制
                </Button>
                <Button variant="ghost" size="icon" className="text-destructive" onClick={() => revoke(t)} aria-label="撤销">
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>
            </div>
          ))}
        </div>
      )}

      <Dialog open={showNew} onOpenChange={setShowNew}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>新建 Token</DialogTitle>
          </DialogHeader>
          <div className="space-y-2">
            <Label>名称</Label>
            <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="例如：我的博客" autoFocus />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowNew(false)}>取消</Button>
            <Button onClick={create} loading={creating} disabled={!name.trim()}>创建</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {node}
    </AppShell>
  )
}
