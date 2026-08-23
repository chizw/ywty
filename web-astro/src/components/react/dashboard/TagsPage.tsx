// 标签管理：列表 + 新建 + 删除
import { useEffect, useState } from 'react'
import { Tag, Plus, Trash2 } from 'lucide-react'
import { AppShell } from './AppShell'
import { PageHeader } from './PageHeader'
import { useConfirm } from './ConfirmDialog'
import { useApi } from '@/lib/api'
import { toast } from '@/lib/react-store'
import { Button } from '../ui/button'
import { Input } from '../ui/input'

interface TagItem {
  id: number
  name: string
  slug: string
  photo_count: number
}

export function TagsPage() {
  const api = useApi()
  const { confirm, node } = useConfirm()

  const [tags, setTags] = useState<TagItem[]>([])
  const [name, setName] = useState('')
  const [loading, setLoading] = useState(true)

  const load = () => {
    setLoading(true)
    api.get<any>('/api/v1/tags', { raw: true })
      .then((r) => setTags(Array.isArray(r?.data) ? r.data : []))
      .catch(() => setTags([]))
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const create = async () => {
    if (!name.trim()) return
    try {
      await api.post('/api/v1/tags', { name: name.trim() })
      toast.success('已创建标签')
      setName('')
      load()
    } catch (e: any) {
      toast.error(e?.message || '创建失败')
    }
  }

  const remove = async (t: TagItem) => {
    const ok = await confirm({ title: '删除标签', message: `确定删除标签「${t.name}」？`, okText: '删除', danger: true })
    if (!ok) return
    try {
      await api.del(`/api/v1/tags/${t.id}`)
      toast.success('已删除')
      load()
    } catch (e: any) {
      toast.error(e?.message || '删除失败')
    }
  }

  return (
    <AppShell>
      <PageHeader title="标签" description="用标签整理图片" />

      <div className="mb-6 flex max-w-sm gap-2">
        <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="新标签名称" onKeyDown={(e) => e.key === 'Enter' && create()} />
        <Button onClick={create} disabled={!name.trim()}>
          <Plus className="h-4 w-4" /> 新建
        </Button>
      </div>

      {loading ? (
        <div className="flex flex-wrap gap-2">
          {Array.from({ length: 6 }).map((_, i) => (
            <div key={i} className="skeleton h-9 w-24 rounded-md" />
          ))}
        </div>
      ) : tags.length === 0 ? (
        <div className="rounded-md border border-dashed border-border py-16 text-center">
          <p className="text-sm text-muted-foreground">还没有标签。</p>
        </div>
      ) : (
        <div className="flex flex-wrap gap-2">
          {tags.map((t) => (
            <span key={t.id} className="group inline-flex items-center gap-1.5 rounded-md border border-border bg-card px-3 py-1.5 text-sm">
              <Tag className="h-3.5 w-3.5 text-muted-foreground" />
              {t.name}
              <span className="text-xs text-muted-foreground">{t.photo_count}</span>
              <button className="ml-0.5 text-muted-foreground opacity-0 transition-opacity hover:text-destructive group-hover:opacity-100" onClick={() => remove(t)} aria-label="删除">
                <Trash2 className="h-3 w-3" />
              </button>
            </span>
          ))}
        </div>
      )}

      {node}
    </AppShell>
  )
}
