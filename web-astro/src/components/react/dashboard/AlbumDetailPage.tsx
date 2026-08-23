// 相册详情：相册信息 + 图片网格（支持移出相册）
import { useEffect, useState } from 'react'
import { ArrowLeft, ImageOff } from 'lucide-react'
import { AppShell } from './AppShell'
import { Lightbox } from './Lightbox'
import { useConfirm } from './ConfirmDialog'
import { useApi } from '@/lib/api'
import { toast } from '@/lib/react-store'
import { timeAgo } from '@/lib/utils'
import { Button } from '../ui/button'
import type { MyPhoto } from '@/lib/types'

interface Album {
  id: number
  name: string
  description: string | null
  is_public: boolean
  photo_count: number
  views: number
  created_at: string
}

export function AlbumDetailPage({ id }: { id: number }) {
  const api = useApi()
  const { confirm, node } = useConfirm()

  const [album, setAlbum] = useState<Album | null>(null)
  const [photos, setPhotos] = useState<MyPhoto[]>([])
  const [lightbox, setLightbox] = useState<number | null>(null)
  const [loading, setLoading] = useState(true)

  const load = () => {
    setLoading(true)
    Promise.all([
      api.get<Album>(`/api/v1/albums/${id}`).catch(() => null),
      api.get<any>(`/api/v1/albums/${id}/photos`, { raw: true }).catch(() => ({ data: [] })),
    ])
      .then(([a, p]) => {
        setAlbum(a)
        setPhotos(Array.isArray(p?.data) ? p.data : [])
      })
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id])

  const removeFromAlbum = async (photoId: number) => {
    const ok = await confirm({ title: '移出相册', message: '将这张图片移出当前相册？图片本身不会被删除。', okText: '移出' })
    if (!ok) return
    try {
      await api.post(`/api/v1/photos/${photoId}/move-to-album`, { album_id: 0 })
      toast.success('已移出相册')
      load()
    } catch (e: any) {
      toast.error(e?.message || '操作失败')
    }
  }

  return (
    <AppShell>
      <a href="/dashboard/albums" className="mb-4 inline-flex items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground">
        <ArrowLeft className="h-4 w-4" /> 返回相册
      </a>

      <div className="mb-6 border-b border-border pb-4">
        <h1 className="font-display text-2xl font-bold tracking-tight">{album?.name || '相册'}</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          {album?.description || `${photos.length} 张图片`}
          {album && ` · ${album.photo_count} 张 · ${album.views} 浏览 · 创建于 ${timeAgo(album.created_at)}`}
        </p>
      </div>

      {loading ? (
        <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6">
          {Array.from({ length: 6 }).map((_, i) => (
            <div key={i} className="skeleton aspect-square rounded-md" />
          ))}
        </div>
      ) : photos.length === 0 ? (
        <div className="rounded-md border border-dashed border-border py-20 text-center">
          <ImageOff className="mx-auto mb-3 h-8 w-8 text-muted-foreground" />
          <p className="text-sm text-muted-foreground">这个相册还是空的。</p>
          <a href="/dashboard/photos" className="mt-2 inline-block text-sm text-brand hover:underline">去图片页上传 →</a>
        </div>
      ) : (
        <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6">
          {photos.map((p, i) => (
            <div key={p.id} className="group relative aspect-square overflow-hidden rounded-md border border-border bg-muted" onClick={() => setLightbox(i)}>
              <img src={p.thumbnail_url || p.url} alt={p.original_name || ''} loading="lazy" className="h-full w-full cursor-zoom-in object-cover transition-transform duration-300 group-hover:scale-105" />
              <div className="absolute inset-0 flex items-end justify-end bg-black/0 p-2 opacity-0 transition-all group-hover:bg-black/40 group-hover:opacity-100">
                <Button variant="secondary" size="sm" className="h-6 px-2 text-[10px]" onClick={(e) => { e.stopPropagation(); removeFromAlbum(p.id) }}>
                  移出
                </Button>
              </div>
            </div>
          ))}
        </div>
      )}

      {lightbox !== null && (
        <Lightbox photos={photos as any} index={lightbox} onClose={() => setLightbox(null)} onIndexChange={setLightbox} />
      )}

      {node}
    </AppShell>
  )
}
