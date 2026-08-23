// 工单详情 + 回复
import { useEffect, useState } from 'react'
import { ArrowLeft, Send } from 'lucide-react'
import { AppShell } from './AppShell'
import { useApi } from '@/lib/api'
import { toast } from '@/lib/react-store'
import { formatDate } from '@/lib/utils'
import { Button } from '../ui/button'
import { Textarea } from '../ui/textarea'
import { Badge } from '../ui/badge'

interface Reply {
  id: number
  content: string
  is_admin: number | boolean
  created_at: string
}

interface TicketInfo {
  id: number
  issue_no: string
  title: string
  status: string
  created_at: string
}

interface TicketDetail {
  ticket: TicketInfo
  replies: Reply[]
}

const STATUS_LABEL: Record<string, string> = { in_progress: '处理中', resolved: '已解决', closed: '已关闭' }

export function TicketDetailPage({ id }: { id: number }) {
  const api = useApi()
  const [detail, setDetail] = useState<TicketDetail | null>(null)
  const [content, setContent] = useState('')
  const [sending, setSending] = useState(false)

  const load = () => {
    api.get<any>(`/api/v1/tickets/${id}`, { raw: true })
      .then((r) => setDetail(r?.data?.data ?? null))
      .catch(() => setDetail(null))
  }

  useEffect(() => {
    load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id])

  const reply = async () => {
    if (!content.trim()) return
    setSending(true)
    try {
      await api.post(`/api/v1/tickets/${id}/replies`, { content: content.trim() })
      toast.success('已回复')
      setContent('')
      load()
    } catch (e: any) {
      toast.error(e?.message || '回复失败')
    } finally {
      setSending(false)
    }
  }

  const close = async () => {
    try {
      await api.post(`/api/v1/tickets/${id}/close`, {})
      toast.success('工单已关闭')
      load()
    } catch (e: any) {
      toast.error(e?.message || '关闭失败')
    }
  }

  const ticket = detail?.ticket
  const closed = ticket?.status === 'closed'

  return (
    <AppShell>
      <a href="/dashboard/tickets" className="mb-4 inline-flex items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground">
        <ArrowLeft className="h-4 w-4" /> 返回工单
      </a>

      {!detail || !ticket ? (
        <div className="rounded-md border border-dashed border-border py-20 text-center">
          <p className="text-sm text-muted-foreground">工单不存在。</p>
        </div>
      ) : (
        <div className="max-w-3xl">
          <div className="border-b border-border pb-4">
            <div className="flex items-center gap-2">
              <h1 className="font-display text-2xl font-bold tracking-tight">{ticket.title}</h1>
              <Badge variant={ticket.status === 'in_progress' ? 'warning' : 'secondary'}>{STATUS_LABEL[ticket.status] || ticket.status}</Badge>
            </div>
            <p className="mt-1 font-mono text-xs text-muted-foreground">{ticket.issue_no} · 创建于 {formatDate(ticket.created_at)}</p>
          </div>

          {/* 回复列表（含首条问题描述，时间正序） */}
          <div className="mt-6 space-y-3">
            {(detail.replies ?? []).map((r) => (
              <div key={r.id} className={`rounded-md border p-4 ${r.is_admin ? 'border-brand/30 bg-brand/5' : 'border-border bg-card'}`}>
                <div className="mb-1.5 flex items-center gap-2 text-xs text-muted-foreground">
                  <span className={r.is_admin ? 'font-medium text-brand' : ''}>{r.is_admin ? '客服' : '我'}</span>
                  <span>{formatDate(r.created_at)}</span>
                </div>
                <p className="whitespace-pre-line text-sm leading-relaxed">{r.content}</p>
              </div>
            ))}
            {(detail.replies ?? []).length === 0 && (
              <p className="py-8 text-center text-sm text-muted-foreground">暂无回复。</p>
            )}
          </div>

          {/* 回复框 */}
          {!closed && (
            <div className="mt-6 space-y-3">
              <Textarea value={content} onChange={(e) => setContent(e.target.value)} placeholder="输入回复内容…" rows={4} />
              <div className="flex justify-between">
                <Button variant="outline" onClick={close}>关闭工单</Button>
                <Button onClick={reply} loading={sending} disabled={!content.trim()}>
                  <Send className="h-4 w-4" /> 回复
                </Button>
              </div>
            </div>
          )}
        </div>
      )}
    </AppShell>
  )
}
