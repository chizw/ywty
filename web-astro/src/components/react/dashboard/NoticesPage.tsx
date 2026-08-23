// 通知列表
import { useEffect, useState } from 'react'
import { Bell } from 'lucide-react'
import { AppShell } from './AppShell'
import { PageHeader } from './PageHeader'
import { useApi } from '@/lib/api'
import { formatDate } from '@/lib/utils'

interface Notice {
  id: number
  title: string
  content: string | null
  view_count: number
  created_at: string
}

export function NoticesPage() {
  const api = useApi()
  const [notices, setNotices] = useState<Notice[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    api.get<any>('/api/v1/notices', { raw: true })
      .then((r) => {
        const data = r?.data?.data ?? r?.data ?? []
        setNotices(Array.isArray(data) ? data : [])
      })
      .catch(() => setNotices([]))
      .finally(() => setLoading(false))
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return (
    <AppShell>
      <PageHeader title="通知" description="站点公告与更新" />

      {loading ? (
        <div className="space-y-3">
          {Array.from({ length: 4 }).map((_, i) => (
            <div key={i} className="skeleton h-16 rounded-md" />
          ))}
        </div>
      ) : notices.length === 0 ? (
        <div className="rounded-md border border-dashed border-border py-20 text-center">
          <Bell className="mx-auto mb-3 h-8 w-8 text-muted-foreground" />
          <p className="text-sm text-muted-foreground">暂无通知。</p>
        </div>
      ) : (
        <div className="space-y-3">
          {notices.map((n) => (
            <div key={n.id} className="rounded-md border border-border bg-card p-4">
              <div className="flex items-center justify-between gap-3">
                <h3 className="font-display text-base font-semibold">{n.title}</h3>
                <span className="shrink-0 text-xs text-muted-foreground">{formatDate(n.created_at)}</span>
              </div>
              {n.content && <p className="mt-2 line-clamp-3 whitespace-pre-line text-sm text-muted-foreground">{n.content}</p>}
              <div className="mt-2 text-xs text-muted-foreground">{n.view_count} 次阅读</div>
            </div>
          ))}
        </div>
      )}
    </AppShell>
  )
}
