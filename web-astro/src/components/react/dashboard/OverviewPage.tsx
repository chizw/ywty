// 概览：统计卡 + 存储用量 + 最近上传
import { useEffect, useState } from 'react'
import { AppShell } from './AppShell'
import { PageHeader } from './PageHeader'
import { StatCard } from './StatCard'
import { useApi } from '@/lib/api'
import { useStatsStore } from '@/lib/react-store'
import { formatBytes, timeAgo } from '@/lib/utils'

interface PhotoItem {
  id: number
  url: string
  thumbnail_url?: string | null
  name?: string
  size?: number
  created_at?: string
}

export function OverviewPage() {
  const api = useApi()
  const stats = useStatsStore()
  const [recent, setRecent] = useState<PhotoItem[]>([])
  const [loaded, setLoaded] = useState(false)

  useEffect(() => {
    stats.refresh()
    api
      .get<PhotoItem[]>('/api/v1/photos', { query: { page: 1, per_page: 6 } })
      .then((d) => setRecent(Array.isArray(d) ? d : []))
      .catch(() => setRecent([]))
      .finally(() => setLoaded(true))
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const pct = stats.capacityBytes > 0 ? Math.min(100, Math.round((stats.usedBytes / stats.capacityBytes) * 100)) : 0

  return (
    <AppShell>
      <PageHeader title="概览" description="你的图库，一目了然。" />

      <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
        <StatCard label="图片" value={stats.photos} accent />
        <StatCard label="相册" value={stats.albums} />
        <StatCard label="已用空间" value={formatBytes(stats.usedBytes)} hint={stats.capacityBytes > 0 ? `共 ${formatBytes(stats.capacityBytes)}` : undefined} />
        <StatCard label="空间占比" value={`${pct}%`} />
      </div>

      {/* 存储用量条 */}
      <div className="mt-6 rounded-md border border-border bg-card p-4">
        <div className="flex items-center justify-between text-xs text-muted-foreground">
          <span>存储用量</span>
          <span className="tabular-nums">
            {formatBytes(stats.usedBytes)} / {stats.capacityBytes > 0 ? formatBytes(stats.capacityBytes) : '—'}
          </span>
        </div>
        <div className="mt-2 h-1.5 w-full overflow-hidden rounded-full bg-muted">
          <div className="h-full rounded-full bg-brand transition-all duration-500" style={{ width: `${pct}%` }} />
        </div>
      </div>

      {/* 最近上传 */}
      <div className="mt-6">
        <h2 className="mb-3 font-display text-lg font-semibold">最近上传</h2>
        {!loaded ? (
          <div className="grid grid-cols-3 gap-3 sm:grid-cols-6">
            {Array.from({ length: 6 }).map((_, i) => (
              <div key={i} className="skeleton aspect-square rounded-md" />
            ))}
          </div>
        ) : recent.length === 0 ? (
          <div className="rounded-md border border-dashed border-border py-14 text-center">
            <p className="text-sm text-muted-foreground">还没有图片。</p>
            <a href="/dashboard/photos" className="mt-2 inline-block text-sm text-brand hover:underline">
              去上传第一张 →
            </a>
          </div>
        ) : (
          <div className="grid grid-cols-3 gap-3 sm:grid-cols-6">
            {recent.map((p) => (
              <a key={p.id} href={`/dashboard/photos`} className="group block">
                <div className="overflow-hidden rounded-md border border-border bg-muted">
                  <img src={p.thumbnail_url || p.url} alt={p.name || ''} loading="lazy" className="aspect-square w-full object-cover transition-transform duration-300 group-hover:scale-105" />
                </div>
                <div className="mt-1.5 flex items-baseline justify-between gap-1 text-[0.68rem] text-muted-foreground">
                  <span className="truncate">{p.name || `#${p.id}`}</span>
                  {p.created_at && <span className="shrink-0">{timeAgo(p.created_at)}</span>}
                </div>
              </a>
            ))}
          </div>
        )}
      </div>
    </AppShell>
  )
}
