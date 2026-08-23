// 意见反馈
import { useEffect, useState } from 'react'
import { Trash2 } from 'lucide-react'
import { AdminShell } from './AdminShell'
import { AdminPageHeader } from './AdminPageHeader'
import { useConfirm } from '../dashboard/ConfirmDialog'
import { useApi } from '@/lib/api'
import { toast } from '@/lib/react-store'
import { formatDate } from '@/lib/utils'
import { Badge } from '../ui/badge'
import { Button } from '../ui/button'

interface Feedback {
  id: number
  type: string
  title: string
  name: string
  email: string
  content: string
  created_at: string
}

const TYPE_LABEL: Record<string, string> = { feedback: '反馈', bug: '缺陷', suggestion: '建议', other: '其他' }

export function AdminFeedbacksPage() {
  const api = useApi()
  const { confirm, node } = useConfirm()
  const [items, setItems] = useState<Feedback[]>([])
  const [loading, setLoading] = useState(true)

  const load = () => {
    setLoading(true)
    api.get<any>('/api/v1/admin/feedbacks', { raw: true })
      .then((r) => setItems(Array.isArray(r?.data) ? r.data : []))
      .catch(() => setItems([]))
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const remove = async (f: Feedback) => {
    const ok = await confirm({ title: '删除反馈', message: `确定删除「${f.title}」？`, okText: '删除', danger: true })
    if (!ok) return
    try {
      await api.del(`/api/v1/admin/feedbacks/${f.id}`)
      toast.success('已删除')
      load()
    } catch (e: any) {
      toast.error(e?.message || '删除失败')
    }
  }

  return (
    <AdminShell>
      <AdminPageHeader title="意见反馈" description={`共 ${items.length} 条`} />

      <div className="space-y-3">
        {loading ? (
          <div className="skeleton h-24 rounded-md" />
        ) : items.length === 0 ? (
          <div className="rounded-md border border-dashed border-border py-20 text-center">
            <p className="text-sm text-muted-foreground">暂无反馈。</p>
          </div>
        ) : items.map((f) => (
          <div key={f.id} className="flex flex-wrap items-start gap-3 rounded-md border border-border bg-card p-4">
            <div className="min-w-0 flex-1">
              <div className="flex items-center gap-2">
                <Badge variant="secondary">{TYPE_LABEL[f.type] || f.type}</Badge>
                <span className="font-medium">{f.title}</span>
              </div>
              <p className="mt-1.5 whitespace-pre-line text-sm text-muted-foreground">{f.content}</p>
              <p className="mt-1 text-xs text-muted-foreground">
                {f.name} · {f.email} · {formatDate(f.created_at)}
              </p>
            </div>
            <Button variant="ghost" size="icon" className="text-destructive" onClick={() => remove(f)} aria-label="删除">
              <Trash2 className="h-4 w-4" />
            </Button>
          </div>
        ))}
      </div>

      {node}
    </AdminShell>
  )
}
