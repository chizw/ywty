// 全屏图片查看器（React）：键盘控制 + 缩略图导航 + 下载
import { useEffect } from 'react'
import { X, ChevronLeft, ChevronRight, Download } from 'lucide-react'

interface LightboxPhoto {
  id: number | string
  url: string
  thumbnail_url?: string | null
  name?: string
  [key: string]: unknown
}

export function Lightbox({
  photos,
  index,
  onClose,
  onIndexChange,
}: {
  photos: LightboxPhoto[]
  index: number
  onClose: () => void
  onIndexChange: (i: number) => void
}) {
  const current = photos[index]

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose()
      else if (e.key === 'ArrowLeft' && photos.length > 1) onIndexChange((index - 1 + photos.length) % photos.length)
      else if (e.key === 'ArrowRight' && photos.length > 1) onIndexChange((index + 1) % photos.length)
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [index, photos.length, onClose, onIndexChange])

  useEffect(() => {
    document.body.style.overflow = 'hidden'
    return () => {
      document.body.style.overflow = ''
    }
  }, [])

  if (!current) return null

  return (
    <div className="fixed inset-0 z-[100] flex flex-col bg-black/95" onClick={onClose}>
      <div className="flex items-center justify-between px-4 py-3 text-sm text-white/80">
        <span>
          {index + 1} / {photos.length}
        </span>
        <div className="flex items-center gap-2">
          <a
            href={current.url}
            download={current.name || `photo-${index + 1}.jpg`}
            target="_blank"
            rel="noreferrer"
            className="inline-flex h-8 w-8 items-center justify-center rounded-md text-white/80 transition-colors hover:bg-white/10 hover:text-white"
            onClick={(e) => e.stopPropagation()}
          >
            <Download className="h-4 w-4" />
          </a>
          <button
            className="inline-flex h-8 w-8 items-center justify-center rounded-md text-white/80 transition-colors hover:bg-white/10 hover:text-white"
            onClick={onClose}
            aria-label="关闭"
          >
            <X className="h-4 w-4" />
          </button>
        </div>
      </div>

      <div className="relative flex min-h-0 flex-1 items-center justify-center px-12">
        {photos.length > 1 && (
          <button
            className="absolute left-2 top-1/2 inline-flex h-10 w-10 -translate-y-1/2 items-center justify-center rounded-full bg-white/10 text-white transition-colors hover:bg-white/20"
            onClick={(e) => {
              e.stopPropagation()
              onIndexChange((index - 1 + photos.length) % photos.length)
            }}
            aria-label="上一张"
          >
            <ChevronLeft className="h-6 w-6" />
          </button>
        )}
        <img
          src={current.url}
          alt={current.name || ''}
          className="max-h-full max-w-full object-contain"
          onClick={(e) => e.stopPropagation()}
        />
        {photos.length > 1 && (
          <button
            className="absolute right-2 top-1/2 inline-flex h-10 w-10 -translate-y-1/2 items-center justify-center rounded-full bg-white/10 text-white transition-colors hover:bg-white/20"
            onClick={(e) => {
              e.stopPropagation()
              onIndexChange((index + 1) % photos.length)
            }}
            aria-label="下一张"
          >
            <ChevronRight className="h-6 w-6" />
          </button>
        )}
      </div>

      {current.name && <div className="px-4 py-2 text-center text-sm text-white/70">{current.name}</div>}

      {photos.length > 1 && (
        <div className="flex justify-center gap-2 overflow-x-auto px-4 py-3" onClick={(e) => e.stopPropagation()}>
          {photos.map((p, i) => (
            <button
              key={p.id}
              className={`h-14 w-14 flex-shrink-0 overflow-hidden rounded border-2 transition ${
                i === index ? 'border-primary' : 'border-transparent opacity-60 hover:opacity-100'
              }`}
              onClick={() => onIndexChange(i)}
            >
              <img src={p.thumbnail_url || p.url} alt={p.name || ''} className="h-full w-full object-cover" />
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
