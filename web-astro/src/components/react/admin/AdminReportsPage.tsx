// 举报管理：列表 + 状态处理
import { useEffect, useState } from 'react'
import { AdminShell } from './AdminShell'
import { AdminPageHeader } from './AdminPageHeader'
import { useApi } from '@/lib/api'
import { toast } from '@/lib/react-store'
import { formatDate } from '@/lib/utils'
import { Badge } from '../ui/badge'
import { Button } from '../ui/button'

interface Report {
  id: number
  reportable_type: string
  reportable_id: number
  reason: string
  status: number
  created_at: string
}

const TYPE_LABEL: Record<string, string> = { photo: '图片', album: '相册', user: '用户', comment: '评论' }

export function AdminReportsPage() {
  const api = useApi()
  const [reports, setReports] = useState<Report[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(true)

  const load = () => {
    setLoading(true)
    api.get<any>('/api/v1/admin/reports', { query: { page, per_page: 20 }, raw: true })
      .then((r) => {
        setReports(Array.isArray(r?.data) ? r.data : [])
        setTotal(Number(r?.meta?.total ?? 0))
      })
      .catch(() => setReports([]))
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [page])

  const handle = async (r: Report, status: number) => {
    try {
      await api.patch(`/api/v1/admin/reports/${r.id}`, { status })
      toast.success(status === 1 ? '已处理' : '已驳回')
      load()
    } catch (e: any) {
      toast.error(e?.message || '操作失败')
    }
  }

  const lastPage = Math.max(1, Math.ceil(total / 20))

  return (
    <AdminShell>
      <AdminPageHeader title="举报管理" description={`共 ${total} 条举报`} />

      <div className="space-y-3">
        {loading ? (
          <div className="skeleton h-20 rounded-md" />
        ) : reports.length === 0 ? (
          <div className="rounded-md border border-dashed border-border py-20 text-center">
            <p className="text-sm text-muted-foreground">没有举报。</p>
          </div>
        ) : reports.map((r) => (
          <div key={r.id} className="flex flex-wrap items-center gap-3 rounded-md border border-border bg-card p-4">
            <div className="min-w-0 flex-1">
              <div className="flex items-center gap-2">
                <Badge variant="secondary">{TYPE_LABEL[r.reportable_type] || r.reportable_type}</Badge>
                <span className="font-mono text-xs text-muted-foreground">#{r.reportable_id}</span>
                <Badge variant={r.status === 0 ? 'warning' : 'success'}>{r.status === 0 ? '待处理' : '已处理'}</Badge>
              </div>
              <p className="mt-1.5 text-sm">{r.reason}</p>
              <p className="mt-0.5 text-xs text-muted-foreground">{formatDate(r.created_at)}</p>
            </div>
            {r.status === 0 && (
              <div className="flex shrink-0 gap-1.5">
                <Button size="sm" onClick={() => handle(r, 1)}>标记已处理</Button>
                <Button size="sm" variant="outline" onClick={() => handle(r, 2)}>驳回</Button>
              </div>
            )}
          </div>
        ))}
      </div>

      {total > 20 && (
        <div className="mt-4 flex items-center justify-center gap-3 text-sm">
          <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>上一页</Button>
          <span className="text-muted-foreground">第 {page} / {lastPage} 页</span>
          <Button variant="outline" size="sm" disabled={page >= lastPage} onClick={() => setPage((p) => p + 1)}>下一页</Button>
        </div>
      )}
    </AdminShell>
  )
}
