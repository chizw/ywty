// 工单列表 + 新建工单
import { useEffect, useState } from 'react'
import { Plus, Ticket } from 'lucide-react'
import { AppShell } from './AppShell'
import { PageHeader } from './PageHeader'
import { useApi } from '@/lib/api'
import { toast } from '@/lib/react-store'
import { formatDate } from '@/lib/utils'
import { Button } from '../ui/button'
import { Input } from '../ui/input'
import { Label } from '../ui/label'
import { Textarea } from '../ui/textarea'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '../ui/dialog'
import { Badge } from '../ui/badge'
import { TicketDetailPage } from './TicketDetailPage'

interface TicketItem {
  id: number
  issue_no: string
  title: string
  ticket_type: string
  status: string
  created_at: string
}

const STATUS_LABEL: Record<string, string> = { open: '处理中', closed: '已关闭', resolved: '已解决' }

export function TicketsPage() {
  // 静态部署：/dashboard/tickets/{id} 由服务器回退到本页，按路径切换详情视图
  const detailId = typeof window !== 'undefined' ? (window.location.pathname.match(/\/dashboard\/tickets\/(\d+)/) || [])[1] : undefined
  const api = useApi()
  const [tickets, setTickets] = useState<TicketItem[]>([])
  const [loading, setLoading] = useState(true)
  const [showNew, setShowNew] = useState(false)
  const [form, setForm] = useState({ title: '', type: 'question', content: '' })
  const [submitting, setSubmitting] = useState(false)

  const load = () => {
    setLoading(true)
    api.get<any>('/api/v1/tickets', { raw: true })
      .then((r) => setTickets(Array.isArray(r?.data) ? r.data : []))
      .catch(() => setTickets([]))
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const submit = async () => {
    if (!form.title.trim() || !form.content.trim()) return
    setSubmitting(true)
    try {
      await api.post('/api/v1/tickets', { title: form.title.trim(), ticket_type: form.type, content: form.content.trim() })
      toast.success('工单已提交')
      setShowNew(false)
      setForm({ title: '', type: 'question', content: '' })
      load()
    } catch (e: any) {
      toast.error(e?.message || '提交失败')
    } finally {
      setSubmitting(false)
    }
  }

  if (detailId) {
    return <TicketDetailPage id={Number(detailId)} />
  }

  return (
    <AppShell>
      <PageHeader title="工单" description="反馈问题或寻求帮助">
        <Button size="sm" onClick={() => setShowNew(true)}>
          <Plus className="h-4 w-4" /> 新建工单
        </Button>
      </PageHeader>

      {loading ? (
        <div className="space-y-3">
          {Array.from({ length: 4 }).map((_, i) => (
            <div key={i} className="skeleton h-14 rounded-md" />
          ))}
        </div>
      ) : tickets.length === 0 ? (
        <div className="rounded-md border border-dashed border-border py-20 text-center">
          <Ticket className="mx-auto mb-3 h-8 w-8 text-muted-foreground" />
          <p className="text-sm text-muted-foreground">还没有工单。</p>
        </div>
      ) : (
        <div className="space-y-3">
          {tickets.map((t) => (
            <a key={t.id} href={`/dashboard/tickets/${t.id}`} className="flex items-center gap-3 rounded-md border border-border bg-card p-4 transition-colors hover:bg-muted/40">
              <div className="grid h-9 w-9 shrink-0 place-items-center rounded-md bg-accent">
                <Ticket className="h-4 w-4 text-accent-foreground" />
              </div>
              <div className="min-w-0 flex-1">
                <p className="truncate text-sm font-medium">{t.title}</p>
                <p className="mt-0.5 font-mono text-xs text-muted-foreground">{t.issue_no}</p>
              </div>
              <Badge variant={t.status === 'open' ? 'warning' : 'secondary'}>{STATUS_LABEL[t.status] || t.status}</Badge>
              <span className="hidden text-xs text-muted-foreground sm:block">{formatDate(t.created_at)}</span>
            </a>
          ))}
        </div>
      )}

      <Dialog open={showNew} onOpenChange={setShowNew}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>新建工单</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label>标题</Label>
              <Input value={form.title} onChange={(e) => setForm((f) => ({ ...f, title: e.target.value }))} placeholder="一句话说明问题" autoFocus />
            </div>
            <div className="space-y-2">
              <Label>内容</Label>
              <Textarea value={form.content} onChange={(e) => setForm((f) => ({ ...f, content: e.target.value }))} placeholder="详细描述你遇到的问题" />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowNew(false)}>取消</Button>
            <Button onClick={submit} loading={submitting} disabled={!form.title.trim() || !form.content.trim()}>提交</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </AppShell>
  )
}
