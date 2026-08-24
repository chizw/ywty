// 图片管理（后台）：表格 + 搜索 + 处置
import { useEffect, useState } from 'react'
import { Search } from 'lucide-react'
import { AdminShell } from './AdminShell'
import { AdminPageHeader } from './AdminPageHeader'
import { useApi } from '@/lib/api'
import { formatBytes, formatDate } from '@/lib/utils'
import { toast } from '@/lib/react-store'
import { Badge } from '../ui/badge'
import { Button } from '../ui/button'
import { Input } from '../ui/input'

interface AdminPhoto {
  id: number
  original_name: string
  url: string
  thumbnail_url: string | null
  size: number
  is_public: boolean
  status: number
  views: number
  created_at: string
  username?: string
}

export function AdminPhotosPage() {
  const api = useApi()
  const [photos, setPhotos] = useState<AdminPhoto[]>([])
  const [total, setTotal] = useState(0)
  const [keyword, setKeyword] = useState('')
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(true)

  const load = () => {
    setLoading(true)
    api.get<any>('/api/v1/admin/photos', { query: { page, per_page: 24, keyword }, raw: true })
      .then((r) => {
        setPhotos(Array.isArray(r?.data) ? r.data : [])
        setTotal(Number(r?.meta?.total ?? 0))
      })
      .catch(() => setPhotos([]))
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [page])

  const removePhoto = async (p: AdminPhoto) => {
    if (!window.confirm(`确认删除图片「${p.original_name}」？该图片将立即从列表移除。`)) return
    try {
      await api.del(`/api/v1/admin/photos/${p.id}`)
      toast.success('已删除')
      load()
    } catch (e: any) {
      toast.error(e?.message || '删除失败')
    }
  }

  const lastPage = Math.max(1, Math.ceil(total / 24))

  return (
    <AdminShell>
      <AdminPageHeader title="图片管理" description={`共 ${total} 张`} />

      <div className="mb-4 flex max-w-sm gap-2">
        <Input value={keyword} onChange={(e) => setKeyword(e.target.value)} placeholder="搜索文件名" onKeyDown={(e) => { if (e.key === 'Enter') { setPage(1); load() } }} />
        <Button variant="outline" onClick={() => { setPage(1); load() }}>
          <Search className="h-4 w-4" /> 搜索
        </Button>
      </div>

      <div className="overflow-hidden rounded-md border border-border">
        <table className="w-full text-sm">
          <thead className="border-b border-border bg-muted/50">
            <tr className="text-left text-xs text-muted-foreground">
              <th className="px-4 py-2.5 font-medium">图片</th>
              <th className="px-4 py-2.5 font-medium">上传者</th>
              <th className="px-4 py-2.5 font-medium">大小</th>
              <th className="px-4 py-2.5 font-medium">可见性</th>
              <th className="px-4 py-2.5 font-medium">状态</th>
              <th className="px-4 py-2.5 font-medium">浏览</th>
              <th className="px-4 py-2.5 font-medium">上传时间</th>
              <th className="px-4 py-2.5 font-medium">操作</th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr><td colSpan={8} className="px-4 py-8 text-center text-muted-foreground">加载中…</td></tr>
            ) : photos.length === 0 ? (
              <tr><td colSpan={8} className="px-4 py-8 text-center text-muted-foreground">没有图片</td></tr>
            ) : photos.map((p) => (
              <tr key={p.id} className="border-b border-border last:border-0 hover:bg-muted/30">
                <td className="px-4 py-3">
                  <div className="flex items-center gap-2.5">
                    <div className="h-9 w-9 flex-shrink-0 overflow-hidden rounded bg-muted">
                      <img src={p.thumbnail_url || p.url} alt={p.original_name} loading="lazy" className="h-full w-full object-cover" />
                    </div>
                    <span className="max-w-[12rem] truncate">{p.original_name}</span>
                  </div>
                </td>
                <td className="px-4 py-3 text-xs text-muted-foreground">{p.username || '—'}</td>
                <td className="px-4 py-3 tabular-nums text-xs text-muted-foreground">{formatBytes(p.size)}</td>
                <td className="px-4 py-3">
                  <Badge variant={p.is_public ? 'brand' : 'secondary'}>{p.is_public ? '公开' : '私有'}</Badge>
                </td>
                <td className="px-4 py-3">
                  <Badge variant={p.status === 1 ? 'success' : 'warning'}>{p.status === 1 ? '正常' : '待审'}</Badge>
                </td>
                <td className="px-4 py-3 tabular-nums text-xs text-muted-foreground">{p.views}</td>
                <td className="px-4 py-3 text-xs text-muted-foreground">{formatDate(p.created_at)}</td>
                <td className="px-4 py-3">
                  <Button variant="outline" size="sm" className="text-destructive" onClick={() => removePhoto(p)}>删除</Button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {total > 24 && (
        <div className="mt-4 flex items-center justify-center gap-3 text-sm">
          <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>上一页</Button>
          <span className="text-muted-foreground">第 {page} / {lastPage} 页</span>
          <Button variant="outline" size="sm" disabled={page >= lastPage} onClick={() => setPage((p) => p + 1)}>下一页</Button>
        </div>
      )}
    </AdminShell>
  )
}
