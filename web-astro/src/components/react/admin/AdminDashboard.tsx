// 仪表盘：统计卡
import { useEffect, useState } from 'react'
import { Users, Image, Folder, Share, AlertTriangle, ReceiptText } from 'lucide-react'
import { AdminShell } from './AdminShell'
import { AdminPageHeader } from './AdminPageHeader'
import { useApi } from '@/lib/api'

interface Stats {
  users?: number
  photos?: number
  albums?: number
  shares?: number
  reports?: number
  orders?: number
  paid_orders?: number
  total_income?: number
}

const CARDS = [
  { key: 'users', label: '用户', icon: Users },
  { key: 'photos', label: '图片', icon: Image },
  { key: 'albums', label: '相册', icon: Folder },
  { key: 'shares', label: '分享', icon: Share },
  { key: 'reports', label: '待处理举报', icon: AlertTriangle, accent: true },
  { key: 'orders', label: '订单', icon: ReceiptText },
]

export function AdminDashboard() {
  const api = useApi()
  const [stats, setStats] = useState<Stats>({})
  const [loading, setLoading] = useState(true)

  const load = () => {
    setLoading(true)
    api.get<Stats>('/api/v1/admin/stats')
      .then(setStats)
      .catch(() => setStats({}))
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return (
    <AdminShell>
      <AdminPageHeader title="仪表盘" description="站点整体情况" />

      <div className="grid grid-cols-2 gap-3 md:grid-cols-3 xl:grid-cols-6">
        {CARDS.map((c) => (
          <div key={c.key} className="rounded-md border border-border bg-card p-4">
            <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
              <c.icon className="h-3.5 w-3.5" />
              {c.label}
            </div>
            <div className={`mt-1.5 font-display text-2xl font-bold tracking-tight ${c.accent ? 'text-amber-600 dark:text-amber-400' : ''}`}>
              {loading ? '—' : (stats as any)[c.key] ?? 0}
            </div>
          </div>
        ))}
      </div>

      {/* 收入 */}
      <div className="mt-6 grid grid-cols-2 gap-3 md:grid-cols-3">
        <div className="rounded-md border border-border bg-card p-4">
          <div className="text-xs text-muted-foreground">已支付订单</div>
          <div className="mt-1.5 font-display text-2xl font-bold tracking-tight">{stats.paid_orders ?? 0}</div>
        </div>
        <div className="rounded-md border border-border bg-card p-4">
          <div className="text-xs text-muted-foreground">累计收入</div>
          <div className="mt-1.5 font-display text-2xl font-bold tracking-tight text-brand">
            ¥{((stats.total_income ?? 0) / 100).toFixed(2)}
          </div>
        </div>
      </div>
    </AdminShell>
  )
}
