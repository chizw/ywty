// 图片管理：上传 + 筛选 + 框选网格 + 批量操作 + 分页 + 灯箱
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { MousePointerClick, X } from 'lucide-react'
import { AppShell } from './AppShell'
import { PageHeader } from './PageHeader'
import { PhotoUploader } from './PhotoUploader'
import { Lightbox } from './Lightbox'
import { useConfirm } from './ConfirmDialog'
import { ShareCreateDialog } from './ShareCreateDialog'
import { TagAttachDialog } from './TagAttachDialog'
import { MoveToAlbumDialog } from './MoveToAlbumDialog'
import { useApi } from '@/lib/api'
import { useStatsStore, toast } from '@/lib/react-store'
import type { MyPhoto, PagedData } from '@/lib/types'
import { cn } from '@/lib/utils'
import { Button } from '../ui/button'
import { Badge } from '../ui/badge'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '../ui/select'
import { Label } from '../ui/label'

const PER_PAGE = 24

interface Rect {
  left: number
  top: number
  width: number
  height: number
}

export function PhotosPage() {
  const api = useApi()
  const stats = useStatsStore()
  const { confirm, node: confirmNode } = useConfirm()

  const [photos, setPhotos] = useState<MyPhoto[]>([])
  const [total, setTotal] = useState(0)
  const [lastPage, setLastPage] = useState(1)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(true)

  // 筛选 / 排序
  const [filterAlbum, setFilterAlbum] = useState<string>('')
  const [filterTag, setFilterTag] = useState('')
  const [sortBy, setSortBy] = useState<'created_at' | 'size' | 'name'>('created_at')
  const [sortOrder, setSortOrder] = useState<'desc' | 'asc'>('desc')

  // 选项
  const [albums, setAlbums] = useState<{ id: number; name: string }[]>([])
  const [tags, setTags] = useState<string[]>([])

  // 多选 + 框选
  const [selected, setSelected] = useState<number[]>([])
  const [dragStart, setDragStart] = useState<{ x: number; y: number } | null>(null)
  const [dragEnd, setDragEnd] = useState<{ x: number; y: number } | null>(null)
  const [dragging, setDragging] = useState(false)
  const gridRef = useRef<HTMLDivElement>(null)

  // 灯箱
  const [lightbox, setLightbox] = useState<number | null>(null)

  // 批量操作对话框
  const [shareOpen, setShareOpen] = useState(false)
  const [tagOpen, setTagOpen] = useState(false)
  const [moveOpen, setMoveOpen] = useState(false)

  const fetchPhotos = useCallback(async () => {
    setLoading(true)
    try {
      const q: Record<string, unknown> = { page, per_page: PER_PAGE, sort: sortBy, order: sortOrder }
      if (filterAlbum) q.album_id = Number(filterAlbum)
      if (filterTag) q.tag = filterTag
      const res = await api.get<PagedData<MyPhoto>>('/api/v1/photos', { query: q, raw: true })
      const data = (res as any).data ?? []
      setPhotos(Array.isArray(data) ? data : [])
      setTotal(Number((res as any).meta?.total ?? 0))
      setLastPage(Number((res as any).meta?.last_page ?? 1))
    } catch {
      setPhotos([])
      setTotal(0)
    } finally {
      setLoading(false)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [page, sortBy, sortOrder, filterAlbum, filterTag])

  const reloadTagNames = useCallback(() => {
    api.get<any>('/api/v1/tags').then((r) => {
      const list = Array.isArray(r) ? r : ((r as any)?.data ?? [])
      setTags(list.map((t: any) => t.name))
    }).catch(() => {})
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  // 初次加载
  useEffect(() => {
    fetchPhotos()
    api.get<any>('/api/v1/albums').then((r) => setAlbums(Array.isArray(r) ? r : ((r as any)?.data ?? []))).catch(() => {})
    reloadTagNames()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const selectMode = selected.length > 0 || dragging

  // 框选几何
  const dragBox: Rect | null = useMemo(() => {
    if (!dragging || !dragStart || !dragEnd) return null
    const x1 = Math.min(dragStart.x, dragEnd.x)
    const y1 = Math.min(dragStart.y, dragEnd.y)
    const x2 = Math.max(dragStart.x, dragEnd.x)
    const y2 = Math.max(dragStart.y, dragEnd.y)
    return { left: x1, top: y1, width: x2 - x1, height: y2 - y1 }
  }, [dragging, dragStart, dragEnd])

  const onPointerDown = (e: React.PointerEvent) => {
    if (e.button !== 0) return
    if ((e.target as HTMLElement).closest('[data-photo]')) return
    setDragStart({ x: e.clientX, y: e.clientY })
    setDragEnd({ x: e.clientX, y: e.clientY })
  }
  const onPointerMove = (e: React.PointerEvent) => {
    if (!dragStart) return
    setDragEnd({ x: e.clientX, y: e.clientY })
    const dx = e.clientX - dragStart.x
    const dy = e.clientY - dragStart.y
    if (!dragging && (Math.abs(dx) > 5 || Math.abs(dy) > 5)) setDragging(true)
  }
  const onPointerUp = () => {
    if (dragging && dragBox) {
      const grid = gridRef.current
      if (grid) {
        const gridRect = grid.getBoundingClientRect()
        const inBox = (id: number) => {
          const el = grid.querySelector<HTMLElement>(`[data-photo="${id}"]`)
          if (!el) return false
          const r = el.getBoundingClientRect()
          const boxLeft = dragBox.left - gridRect.left
          const boxTop = dragBox.top - gridRect.top
          return !(r.right < boxLeft || r.bottom < boxTop || r.left > boxLeft + dragBox.width || r.top > boxTop + dragBox.height)
        }
        setSelected((prev) => {
          const set = new Set(prev)
          photos.forEach((p) => inBox(p.id) && set.add(p.id))
          return Array.from(set)
        })
      }
    }
    setDragStart(null)
    setDragEnd(null)
    setDragging(false)
  }

  const toggleSelect = (id: number) =>
    setSelected((prev) => (prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id]))

  const onUploaded = () => {
    setPage(1)
    fetchPhotos()
    stats.refresh()
  }

  const resetFilters = () => {
    setFilterAlbum('')
    setFilterTag('')
    setSortBy('created_at')
    setSortOrder('desc')
    setPage(1)
  }

  // 单张操作
  const removeOne = async (id: number) => {
    const ok = await confirm({ title: '删除图片', message: '确定删除这张图片？此操作不可撤销。', okText: '删除', danger: true })
    if (!ok) return
    try {
      await api.del(`/api/v1/photos/${id}`)
      toast.success('已删除')
      fetchPhotos()
      stats.refresh()
    } catch (e: any) {
      toast.error(e?.message || '删除失败')
    }
  }

  const copyOne = async (id: number) => {
    try {
      await api.post(`/api/v1/photos/${id}/copy`, {})
      toast.success('已复制')
      fetchPhotos()
      stats.refresh()
    } catch (e: any) {
      toast.error(e?.message || '复制失败')
    }
  }

  const togglePublic = async (p: MyPhoto) => {
    try {
      await api.patch(`/api/v1/photos/${p.id}`, { is_public: !p.is_public })
      toast.success(p.is_public ? '已转为私有' : '已转为公开')
      fetchPhotos()
    } catch (e: any) {
      toast.error(e?.message || '操作失败')
    }
  }

  // 批量操作
  const batchDelete = async () => {
    const ok = await confirm({ title: '批量删除', message: `确定删除选中的 ${selected.length} 张图片？此操作不可撤销。`, okText: '全部删除', danger: true })
    if (!ok) return
    try {
      await api.post('/api/v1/photos/batch-delete', { ids: selected })
      toast.success(`已删除 ${selected.length} 张图片`)
      setSelected([])
      fetchPhotos()
      stats.refresh()
    } catch (e: any) {
      toast.error(e?.message || '批量删除失败')
    }
  }

  const batchPublic = async (isPublic: boolean) => {
    try {
      await api.patch('/api/v1/photos/batch-update', { ids: selected, is_public: isPublic })
      toast.success(isPublic ? `已公开 ${selected.length} 张` : `已转私有 ${selected.length} 张`)
      setSelected([])
      fetchPhotos()
    } catch (e: any) {
      toast.error(e?.message || '操作失败')
    }
  }

  return (
    <AppShell>
      <PageHeader title="我的图片" description={`共 ${total} 张`}>
        {selectMode && (
          <Button variant="outline" size="sm" onClick={() => setSelected([])}>
            <X className="h-3 w-3" /> 取消选择
          </Button>
        )}
      </PageHeader>

      {!selectMode && (
        <p className="mb-3 flex items-center gap-1.5 text-xs text-muted-foreground">
          <MousePointerClick className="h-3.5 w-3.5" />
          在网格空白处按住左键拖动，可框选多张图片
        </p>
      )}

      <div className="mb-6">
        <PhotoUploader onUploaded={onUploaded} />
      </div>

      {/* 批量操作栏 */}
      {selectMode && (
        <div className="mb-4 flex flex-wrap items-center gap-2 rounded-md border border-brand/20 bg-brand/5 p-2">
          <span className="px-2 text-sm">
            已选 <b className="text-brand">{selected.length}</b> 项
          </span>
          <div className="flex-1" />
          {selected.length === 1 ? (
            <Button variant="outline" size="sm" onClick={() => setShareOpen(true)}>分享</Button>
          ) : (
            <Button variant="outline" size="sm" disabled title="仅支持对单张选中图片发起分享">分享</Button>
          )}
          <Button variant="outline" size="sm" onClick={() => setTagOpen(true)}>打标签</Button>
          <Button variant="outline" size="sm" onClick={() => setMoveOpen(true)}>移入相册</Button>
          <Button variant="outline" size="sm" onClick={() => batchPublic(true)}>批量公开</Button>
          <Button variant="outline" size="sm" onClick={() => batchPublic(false)}>批量私有</Button>
          <Button variant="destructive" size="sm" onClick={batchDelete}>批量删除</Button>
        </div>
      )}

      {/* 筛选栏 */}
      <div className="mb-4 flex flex-wrap items-end gap-3 rounded-md border border-border bg-card p-3">
        <div>
          <Label className="mb-1 text-xs text-muted-foreground">相册</Label>
          <Select value={filterAlbum} onValueChange={(v) => { setFilterAlbum(v); setPage(1) }}>
            <SelectTrigger className="w-[140px]">
              <SelectValue placeholder="全部" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="">全部</SelectItem>
              {albums.map((a) => (
                <SelectItem key={a.id} value={String(a.id)}>{a.name}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <div>
          <Label className="mb-1 text-xs text-muted-foreground">标签</Label>
          <Select value={filterTag} onValueChange={(v) => { setFilterTag(v); setPage(1) }}>
            <SelectTrigger className="w-[140px]">
              <SelectValue placeholder="全部" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="">全部</SelectItem>
              {tags.map((t) => (
                <SelectItem key={t} value={t}>{t}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <div>
          <Label className="mb-1 text-xs text-muted-foreground">排序</Label>
          <Select value={sortBy} onValueChange={(v) => setSortBy(v as any)}>
            <SelectTrigger className="w-[120px]"><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="created_at">按时间</SelectItem>
              <SelectItem value="size">按大小</SelectItem>
              <SelectItem value="name">按名称</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div>
          <Label className="mb-1 text-xs text-muted-foreground">方向</Label>
          <Select value={sortOrder} onValueChange={(v) => setSortOrder(v as any)}>
            <SelectTrigger className="w-[100px]"><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="desc">降序</SelectItem>
              <SelectItem value="asc">升序</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <Button variant="outline" size="sm" onClick={resetFilters}>重置</Button>
      </div>

      {/* 网格 */}
      {loading ? (
        <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6">
          {Array.from({ length: 12 }).map((_, i) => (
            <div key={i} className="skeleton aspect-square rounded-md" />
          ))}
        </div>
      ) : photos.length === 0 ? (
        <div className="rounded-md border border-dashed border-border py-20 text-center">
          <p className="text-sm text-muted-foreground">还没有图片，拖拽或点击上方上传第一张。</p>
        </div>
      ) : (
        <div
          ref={gridRef}
          className={cn(
            'relative grid select-none grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6',
            selectMode ? 'cursor-pointer' : 'cursor-crosshair'
          )}
          onPointerDown={onPointerDown}
          onPointerMove={onPointerMove}
          onPointerUp={onPointerUp}
        >
          {photos.map((p) => {
            const isSel = selected.includes(p.id)
            return (
              <div
                key={p.id}
                data-photo={p.id}
                className={cn(
                  'group relative aspect-square overflow-hidden rounded-md border border-border bg-muted',
                  selectMode && isSel && 'ring-2 ring-brand'
                )}
                onClick={(e) => {
                  if (selectMode) {
                    e.stopPropagation()
                    toggleSelect(p.id)
                  } else {
                    setLightbox(photos.findIndex((x) => x.id === p.id))
                  }
                }}
              >
                <img src={p.thumbnail_url || p.url} alt={p.original_name || ''} loading="lazy" className="h-full w-full object-cover transition-transform duration-300 group-hover:scale-105" />

                {selectMode && (
                  <span className={cn('absolute left-1.5 top-1.5 z-10 grid h-5 w-5 place-items-center rounded border-2 text-xs text-white', isSel ? 'border-brand bg-brand' : 'border-white bg-black/30')}>
                    {isSel ? '✓' : ''}
                  </span>
                )}

                {!selectMode && (
                  <div className="absolute inset-0 flex flex-col justify-between bg-black/0 p-2 opacity-0 transition-all group-hover:bg-black/50 group-hover:opacity-100">
                    <div className="flex justify-end">
                      {p.is_public && <Badge variant="brand" className="text-[10px]">公开</Badge>}
                    </div>
                    <div className="flex flex-wrap justify-end gap-1">
                      <Button variant="secondary" size="sm" className="h-6 px-2 text-[10px]" onClick={(e) => { e.stopPropagation(); togglePublic(p) }}>
                        {p.is_public ? '转私有' : '转公开'}
                      </Button>
                      <Button variant="secondary" size="sm" className="h-6 px-2 text-[10px]" onClick={(e) => { e.stopPropagation(); copyOne(p.id) }}>复制</Button>
                      <Button variant="destructive" size="sm" className="h-6 px-2 text-[10px]" onClick={(e) => { e.stopPropagation(); removeOne(p.id) }}>删除</Button>
                    </div>
                  </div>
                )}
              </div>
            )
          })}

          {/* 框选矩形 */}
          {dragBox && (
            <div
              className="pointer-events-none fixed z-50 border-2 border-brand bg-brand/10"
              style={{ left: dragBox.left, top: dragBox.top, width: dragBox.width, height: dragBox.height }}
            />
          )}
        </div>
      )}

      {/* 分页 */}
      {total > PER_PAGE && (
        <div className="mt-6 flex items-center justify-center gap-3 text-sm">
          <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>上一页</Button>
          <span className="text-muted-foreground">第 {page} / {lastPage} 页</span>
          <Button variant="outline" size="sm" disabled={page >= lastPage} onClick={() => setPage((p) => p + 1)}>下一页</Button>
        </div>
      )}

      {lightbox !== null && (
        <Lightbox photos={photos as any} index={lightbox} onClose={() => setLightbox(null)} onIndexChange={setLightbox} />
      )}

      <ShareCreateDialog open={shareOpen} onOpenChange={setShareOpen} photoId={selected.length === 1 ? selected[0] : null} />
      <TagAttachDialog
        open={tagOpen}
        onOpenChange={setTagOpen}
        photoIds={selected}
        onSaved={() => {
          reloadTagNames()
          fetchPhotos()
        }}
      />
      <MoveToAlbumDialog
        open={moveOpen}
        onOpenChange={setMoveOpen}
        photoIds={selected}
        onDone={() => {
          setSelected([])
          fetchPhotos()
        }}
      />

      {confirmNode}
    </AppShell>
  )
}
