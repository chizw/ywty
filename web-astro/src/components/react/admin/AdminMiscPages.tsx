// 违规记录 / 单页管理 / 角色组 / 通知管理 / 工单管理 / 存储 / 驱动
// 用一个文件组织后台的列表型页面，减少样板。

import { useEffect, useState, type ReactNode } from 'react'
import { Plus } from 'lucide-react'
import { AdminShell } from './AdminShell'
import { AdminPageHeader } from './AdminPageHeader'
import { useApi } from '@/lib/api'
import { toast } from '@/lib/react-store'
import { formatDate } from '@/lib/utils'
import { Badge } from '../ui/badge'
import { Button } from '../ui/button'
import { Input } from '../ui/input'
import { Label } from '../ui/label'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '../ui/dialog'
import { Textarea } from '../ui/textarea'

// ---------- 违规记录 ----------
export function AdminViolationsPage() {
  const api = useApi()
  const [items, setItems] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)

  const load = () => {
    setLoading(true)
    api.get<any>('/api/v1/admin/violations', { query: { page, per_page: 20 }, raw: true })
      .then((r) => {
        setItems(Array.isArray(r?.data) ? r.data : [])
        setTotal(Number(r?.meta?.total ?? 0))
      })
      .catch(() => setItems([]))
      .finally(() => setLoading(false))
  }
  useEffect(() => { load() }, [page]) // eslint-disable-line react-hooks/exhaustive-deps

  return (
    <AdminShell>
      <AdminPageHeader title="违规记录" description={`共 ${total} 条`} />
      <ListTable
        loading={loading}
        empty="暂无违规记录"
        head={['用户', '图片', '原因', '状态', '时间']}
        rows={items.map((v) => [String(v.user_id ?? '-'), String(v.photo_id ?? '-'), v.reason, <Badge key="s" variant={v.status === 'pending' ? 'warning' : 'success'}>{v.status || '-'}</Badge>, formatDate(v.created_at)])}
      />
      <Pager page={page} total={total} perPage={20} onChange={setPage} />
    </AdminShell>
  )
}

// ---------- 单页管理 ----------
const EMPTY_PAGE_FORM = { page_type: 'internal', name: '', icon: '', title: '', slug: '', url: '', sort: 0, is_show: 1, keywords: '', description: '', content: '' }

export function AdminPagesPage() {
  const api = useApi()
  const [items, setItems] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [editing, setEditing] = useState<any | null>(null)
  const [form, setForm] = useState({ ...EMPTY_PAGE_FORM })
  const [saving, setSaving] = useState(false)

  const load = () => {
    setLoading(true)
    api.get<any>('/api/v1/admin/pages', { raw: true })
      .then((r) => setItems(Array.isArray(r?.data) ? r.data : []))
      .catch(() => setItems([]))
      .finally(() => setLoading(false))
  }
  useEffect(() => { load() }, []) // eslint-disable-line react-hooks/exhaustive-deps

  const openCreate = () => {
    setEditing({})
    setForm({ ...EMPTY_PAGE_FORM })
  }

  const openEdit = (p: any) => {
    setEditing(p)
    setForm({
      page_type: p.type === 'external' ? 'external' : 'internal',
      name: p.name ?? '',
      icon: p.icon ?? '',
      title: p.title ?? '',
      slug: p.slug ?? '',
      url: p.url ?? '',
      sort: Number(p.sort ?? 0),
      is_show: Number(p.is_show ?? 0),
      keywords: p.keywords ?? '',
      description: p.description ?? '',
      content: p.content ?? '',
    })
  }

  const save = async () => {
    if (!form.name.trim()) return
    setSaving(true)
    try {
      if (editing?.id) {
        await api.patch(`/api/v1/admin/pages/${editing.id}`, {
          page_type: form.page_type,
          name: form.name.trim(),
          icon: form.icon,
          title: form.title,
          slug: form.slug.trim(),
          url: form.url.trim(),
          sort: Number(form.sort) || 0,
          is_show: Number(form.is_show) ? 1 : 0,
          keywords: form.keywords,
          description: form.description,
          content: form.content,
        })
        toast.success('单页已更新')
      } else {
        await api.post('/api/v1/admin/pages', {
          page_type: form.page_type,
          name: form.name.trim(),
          icon: form.icon,
          title: form.title,
          slug: form.slug.trim(),
          url: form.url.trim(),
          sort: Number(form.sort) || 0,
          is_show: Number(form.is_show) ? 1 : 0,
          keywords: form.keywords,
          description: form.description,
          content: form.content,
        })
        toast.success('单页已创建')
      }
      setEditing(null)
      load()
    } catch (e: any) {
      toast.error(e?.message || '保存失败')
    } finally {
      setSaving(false)
    }
  }

  const remove = async (p: any) => {
    if (!window.confirm(`确认删除单页「${p.title || p.name}」？此操作不可恢复。`)) return
    try {
      await api.del(`/api/v1/admin/pages/${p.id}`)
      toast.success('已删除')
      load()
    } catch (e: any) {
      toast.error(e?.message || '删除失败')
    }
  }

  return (
    <AdminShell>
      <AdminPageHeader title="单页管理" description={`共 ${items.length} 个`}>
        <Button size="sm" onClick={openCreate}><Plus className="h-4 w-4" /> 新建页面</Button>
      </AdminPageHeader>
      <ListTable
        loading={loading}
        empty="暂无单页"
        head={['标题', 'Slug', '浏览', '显示', '时间', '操作']}
        rows={items.map((p) => [p.title, <span key="s" className="font-mono text-xs">{p.slug}</span>, p.view_count, <Badge key="b" variant={p.is_show === 1 ? 'success' : 'secondary'}>{p.is_show === 1 ? '显示' : '隐藏'}</Badge>, formatDate(p.created_at), <span key="a" className="flex gap-2"><Button variant="outline" size="sm" onClick={() => openEdit(p)}>编辑</Button><Button variant="outline" size="sm" className="text-destructive" onClick={() => remove(p)}>删除</Button></span>])}
      />

      <Dialog open={!!editing} onOpenChange={(o) => { if (!o) setEditing(null) }}>
        <DialogContent className="max-w-lg max-h-[85vh] overflow-y-auto">
          <DialogHeader><DialogTitle>{editing?.id ? '编辑单页' : '新建单页'}</DialogTitle></DialogHeader>
          <div className="space-y-4">
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-2">
                <Label>名称</Label>
                <Input value={form.name} onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))} />
              </div>
              <div className="space-y-2">
                <Label>标题</Label>
                <Input value={form.title} onChange={(e) => setForm((f) => ({ ...f, title: e.target.value }))} />
              </div>
              <div className="space-y-2">
                <Label>Slug</Label>
                <Input value={form.slug} onChange={(e) => setForm((f) => ({ ...f, slug: e.target.value }))} placeholder="about" />
              </div>
              <div className="space-y-2">
                <Label>外链 URL（type=external）</Label>
                <Input value={form.url} onChange={(e) => setForm((f) => ({ ...f, url: e.target.value }))} placeholder="https://" />
              </div>
              <div className="space-y-2">
                <Label>图标</Label>
                <Input value={form.icon} onChange={(e) => setForm((f) => ({ ...f, icon: e.target.value }))} />
              </div>
              <div className="space-y-2">
                <Label>排序</Label>
                <Input type="number" value={form.sort} onChange={(e) => setForm((f) => ({ ...f, sort: Number(e.target.value) }))} />
              </div>
              <div className="space-y-2">
                <Label>类型</Label>
                <select
                  className="h-9 w-full rounded-md border border-border bg-background px-3 text-sm"
                  value={form.page_type}
                  onChange={(e) => setForm((f) => ({ ...f, page_type: e.target.value }))}
                >
                  <option value="internal">内部页面</option>
                  <option value="external">外部链接</option>
                </select>
              </div>
              <div className="space-y-2">
                <Label>是否显示</Label>
                <select
                  className="h-9 w-full rounded-md border border-border bg-background px-3 text-sm"
                  value={String(form.is_show)}
                  onChange={(e) => setForm((f) => ({ ...f, is_show: Number(e.target.value) }))}
                >
                  <option value="1">显示</option>
                  <option value="0">隐藏</option>
                </select>
              </div>
            </div>
            <div className="space-y-2">
              <Label>SEO 关键词</Label>
              <Input value={form.keywords} onChange={(e) => setForm((f) => ({ ...f, keywords: e.target.value }))} />
            </div>
            <div className="space-y-2">
              <Label>SEO 描述</Label>
              <Input value={form.description} onChange={(e) => setForm((f) => ({ ...f, description: e.target.value }))} />
            </div>
            <div className="space-y-2">
              <Label>内容</Label>
              <Textarea rows={6} value={form.content} onChange={(e) => setForm((f) => ({ ...f, content: e.target.value }))} />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setEditing(null)}>取消</Button>
            <Button onClick={save} loading={saving} disabled={!form.name.trim()}>保存</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </AdminShell>
  )
}

// ---------- 角色组 ----------
const GB = 1024 ** 3
export function AdminGroupsPage() {
  const api = useApi()
  const [items, setItems] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [editing, setEditing] = useState<any | null>(null)
  const [showNew, setShowNew] = useState(false)
  const [form, setForm] = useState({ name: '', intro: '', maxStorageGb: '' })

  const load = () => {
    setLoading(true)
    api.get<any>('/api/v1/admin/groups', { raw: true })
      .then((r) => setItems(Array.isArray(r?.data) ? r.data : []))
      .catch(() => setItems([]))
      .finally(() => setLoading(false))
  }
  useEffect(() => { load() }, []) // eslint-disable-line react-hooks/exhaustive-deps

  const quotaBytes = () =>
    form.maxStorageGb.trim() === '' ? null : Math.round(Number(form.maxStorageGb) * GB)

  const openEdit = (g: any) => {
    setEditing(g)
    setForm({
      name: g.name ?? '',
      intro: g.intro ?? '',
      maxStorageGb: g.max_storage != null ? String(g.max_storage / GB) : '',
    })
  }

  const submit = async () => {
    if (!form.name.trim()) return
    try {
      if (editing) {
        // 后端语义：缺省字段=不修改、显式 null=清除，因此总是携带 max_storage
        await api.put(`/api/v1/admin/groups/${editing.id}`, {
          name: form.name.trim(),
          intro: form.intro,
          max_storage: quotaBytes(),
        })
        toast.success('已保存角色组')
      } else {
        await api.post('/api/v1/admin/groups', {
          name: form.name.trim(),
          intro: form.intro,
          max_storage: quotaBytes(),
        })
        toast.success('已创建角色组')
      }
      setShowNew(false)
      setEditing(null)
      setForm({ name: '', intro: '', maxStorageGb: '' })
      load()
    } catch (e: any) {
      toast.error(e?.message || '保存失败')
    }
  }

  const removeGroup = async (g: any) => {
    if (!window.confirm(`确认删除角色组「${g.name}」？该组用户将失去组配额。`)) return
    try {
      await api.del(`/api/v1/admin/groups/${g.id}`)
      toast.success('已删除角色组')
      load()
    } catch (e: any) {
      toast.error(e?.message || '删除失败')
    }
  }

  return (
    <AdminShell>
      <AdminPageHeader title="角色组" description={`共 ${items.length} 个`}>
        <Button size="sm" onClick={() => setShowNew(true)}><Plus className="h-4 w-4" /> 新建角色组</Button>
      </AdminPageHeader>
      <ListTable
        loading={loading}
        empty="暂无角色组"
        head={['名称', '描述', '存储配额', '默认', '访客', '时间', '操作']}
        rows={items.map((g) => [
          g.name,
          g.intro || '-',
          g.max_storage != null ? `${(g.max_storage / GB).toFixed(1)} GB` : '不限',
          g.is_default === 1 ? '是' : '否',
          g.is_guest === 1 ? '是' : '否',
          formatDate(g.created_at),
          <span key="a" className="flex gap-2">
            <Button variant="outline" size="sm" onClick={() => openEdit(g)}>编辑</Button>
            {g.is_default !== 1 && (
              <Button variant="outline" size="sm" className="text-destructive" onClick={() => removeGroup(g)}>删除</Button>
            )}
          </span>,
        ])}
      />

      <Dialog open={showNew || !!editing} onOpenChange={(o) => { setShowNew(o); if (!o) setEditing(null) }}>
        <DialogContent className="max-w-md">
          <DialogHeader><DialogTitle>{editing ? '编辑角色组' : '新建角色组'}</DialogTitle></DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label>名称</Label>
              <Input value={form.name} onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))} autoFocus />
            </div>
            <div className="space-y-2">
              <Label>描述</Label>
              <Input value={form.intro} onChange={(e) => setForm((f) => ({ ...f, intro: e.target.value }))} />
            </div>
            <div className="space-y-2">
              <Label>存储配额（GB）</Label>
              <Input
                type="number"
                min="0"
                step="0.1"
                value={form.maxStorageGb}
                onChange={(e) => setForm((f) => ({ ...f, maxStorageGb: e.target.value }))}
                placeholder="留空 = 不限"
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => { setShowNew(false); setEditing(null) }}>取消</Button>
            <Button onClick={submit} disabled={!form.name.trim()}>{editing ? '保存' : '创建'}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </AdminShell>
  )
}

// ---------- 通知管理 ----------
export function AdminNoticesPage() {
  const api = useApi()
  const [items, setItems] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [showNew, setShowNew] = useState(false)
  const [editing, setEditing] = useState<any | null>(null)
  const [form, setForm] = useState({ title: '', content: '' })
  const [saving, setSaving] = useState(false)

  const load = () => {
    setLoading(true)
    api.get<any>('/api/v1/admin/notices', { raw: true })
      .then((r) => setItems(Array.isArray(r?.data) ? r.data : []))
      .catch(() => setItems([]))
      .finally(() => setLoading(false))
  }
  useEffect(() => { load() }, []) // eslint-disable-line react-hooks/exhaustive-deps

  const openEdit = (n: any) => {
    setEditing(n)
    setForm({ title: n.title ?? '', content: n.content ?? '' })
  }

  const closeDialog = () => { setShowNew(false); setEditing(null) }

  const submit = async () => {
    if (!form.title.trim()) return
    setSaving(true)
    try {
      if (editing) {
        await api.patch(`/api/v1/admin/notices/${editing.id}`, { title: form.title.trim(), content: form.content })
        toast.success('通知已更新')
      } else {
        await api.post('/api/v1/admin/notices', { title: form.title.trim(), content: form.content.trim() })
        toast.success('已发布通知')
      }
      closeDialog()
      setForm({ title: '', content: '' })
      load()
    } catch (e: any) {
      toast.error(e?.message || '保存失败')
    } finally {
      setSaving(false)
    }
  }

  const removeNotice = async (n: any) => {
    if (!window.confirm(`确认删除公告「${n.title}」？`)) return
    try {
      await api.del(`/api/v1/admin/notices/${n.id}`)
      toast.success('已删除')
      load()
    } catch (e: any) {
      toast.error(e?.message || '删除失败')
    }
  }

  return (
    <AdminShell>
      <AdminPageHeader title="通知管理" description={`共 ${items.length} 条`}>
        <Button size="sm" onClick={() => setShowNew(true)}><Plus className="h-4 w-4" /> 发布通知</Button>
      </AdminPageHeader>
      <ListTable
        loading={loading}
        empty="暂无通知"
        head={['标题', '阅读', '时间', '操作']}
        rows={items.map((n) => [n.title, n.view_count, formatDate(n.created_at), <span key="a" className="flex gap-2"><Button variant="outline" size="sm" onClick={() => openEdit(n)}>编辑</Button><Button variant="outline" size="sm" className="text-destructive" onClick={() => removeNotice(n)}>删除</Button></span>])}
      />

      <Dialog open={showNew || !!editing} onOpenChange={(o) => { if (!o) closeDialog() }}>
        <DialogContent className="max-w-md">
          <DialogHeader><DialogTitle>{editing ? '编辑通知' : '发布通知'}</DialogTitle></DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label>标题</Label>
              <Input value={form.title} onChange={(e) => setForm((f) => ({ ...f, title: e.target.value }))} autoFocus />
            </div>
            <div className="space-y-2">
              <Label>内容</Label>
              <Textarea rows={4} value={form.content} onChange={(e) => setForm((f) => ({ ...f, content: e.target.value }))} />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={closeDialog}>取消</Button>
            <Button onClick={submit} loading={saving} disabled={!form.title.trim()}>{editing ? '保存' : '发布'}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </AdminShell>
  )
}

// ---------- 工单管理 ----------
const TICKET_STATUS_LABEL: Record<string, string> = { in_progress: '处理中', resolved: '已解决', closed: '已关闭' }

export function AdminTicketsPage() {
  const api = useApi()
  const [items, setItems] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [viewingId, setViewingId] = useState<number | null>(null)
  const [detail, setDetail] = useState<{ ticket: any; replies: any[] } | null>(null)
  const [replyText, setReplyText] = useState('')
  const [sending, setSending] = useState(false)

  const load = () => {
    setLoading(true)
    api.get<any>('/api/v1/admin/tickets', { raw: true })
      .then((r) => setItems(Array.isArray(r?.data) ? r.data : []))
      .catch(() => setItems([]))
      .finally(() => setLoading(false))
  }
  useEffect(() => { load() }, []) // eslint-disable-line react-hooks/exhaustive-deps

  const openView = (id: number) => {
    setViewingId(id)
    setDetail(null)
    setReplyText('')
    api.get<any>(`/api/v1/admin/tickets/${id}`, { raw: true })
      .then((r) => setDetail(r?.data?.data ?? null))
      .catch(() => setDetail(null))
  }

  const reloadDetail = () => {
    if (viewingId == null) return
    api.get<any>(`/api/v1/admin/tickets/${viewingId}`, { raw: true })
      .then((r) => setDetail(r?.data?.data ?? null))
      .catch(() => {})
  }

  const sendReply = async () => {
    if (viewingId == null || !replyText.trim()) return
    setSending(true)
    try {
      await api.post(`/api/v1/admin/tickets/${viewingId}/replies`, { content: replyText.trim() })
      toast.success('已回复')
      setReplyText('')
      reloadDetail()
    } catch (e: any) {
      toast.error(e?.message || '回复失败')
    } finally {
      setSending(false)
    }
  }

  const setStatus = async (status: string) => {
    if (viewingId == null) return
    try {
      await api.patch(`/api/v1/admin/tickets/${viewingId}/status`, { status })
      toast.success('状态已更新')
      reloadDetail()
      load()
    } catch (e: any) {
      toast.error(e?.message || '更新失败')
    }
  }

  const ticket = detail?.ticket

  return (
    <AdminShell>
      <AdminPageHeader title="工单管理" description={`共 ${items.length} 条`} />
      <ListTable
        loading={loading}
        empty="暂无工单"
        head={['工单号', '标题', '状态', '时间', '操作']}
        rows={items.map((t) => [
          <span key="n" className="font-mono text-xs">{t.issue_no}</span>,
          t.title,
          <Badge key="s" variant={t.status === 'in_progress' ? 'warning' : 'secondary'}>{TICKET_STATUS_LABEL[t.status] || t.status}</Badge>,
          formatDate(t.created_at),
          <Button key="a" variant="outline" size="sm" onClick={() => openView(t.id)}>查看</Button>,
        ])}
      />

      <Dialog open={viewingId != null} onOpenChange={(o) => { if (!o) setViewingId(null) }}>
        <DialogContent className="max-w-lg max-h-[85vh] overflow-y-auto">
          <DialogHeader><DialogTitle>工单详情</DialogTitle></DialogHeader>
          {!detail || !ticket ? (
            <p className="py-8 text-center text-sm text-muted-foreground">加载中…</p>
          ) : (
            <div className="space-y-4">
              <div className="flex items-center justify-between gap-2">
                <div className="min-w-0">
                  <p className="truncate text-sm font-medium">{ticket.title}</p>
                  <p className="mt-0.5 font-mono text-xs text-muted-foreground">{ticket.issue_no} · 用户 #{ticket.user_id}</p>
                </div>
                <select
                  className="h-8 shrink-0 rounded-md border border-border bg-background px-2 text-xs"
                  value={ticket.status}
                  onChange={(e) => setStatus(e.target.value)}
                >
                  {Object.entries(TICKET_STATUS_LABEL).map(([v, label]) => (
                    <option key={v} value={v}>{label}</option>
                  ))}
                </select>
              </div>

              <div className="space-y-3">
                {(detail.replies ?? []).map((r) => (
                  <div key={r.id} className={`rounded-md border p-3 ${r.is_admin ? 'border-brand/30 bg-brand/5' : 'border-border bg-card'}`}>
                    <div className="mb-1 flex items-center gap-2 text-xs text-muted-foreground">
                      <span className={r.is_admin ? 'font-medium text-brand' : ''}>{r.is_admin ? '客服' : '用户'}</span>
                      <span>{formatDate(r.created_at)}</span>
                    </div>
                    <p className="whitespace-pre-line text-sm leading-relaxed">{r.content}</p>
                  </div>
                ))}
                {(detail.replies ?? []).length === 0 && (
                  <p className="py-4 text-center text-sm text-muted-foreground">暂无回复。</p>
                )}
              </div>

              <div className="space-y-2">
                <Textarea rows={3} value={replyText} onChange={(e) => setReplyText(e.target.value)} placeholder="以客服身份回复…" />
                <div className="flex justify-end">
                  <Button size="sm" onClick={sendReply} loading={sending} disabled={!replyText.trim()}>回复</Button>
                </div>
              </div>
            </div>
          )}
        </DialogContent>
      </Dialog>
    </AdminShell>
  )
}

// ---------- 存储策略 ----------
const EMPTY_STORAGE_FORM = { name: '', provider: 'local', prefix: '', intro: '', options: '' }
const STORAGE_PROVIDERS = [
  { value: 'local', label: '本地存储' },
  { value: 's3', label: 'AWS S3 / 兼容' },
  { value: 'oss', label: '阿里云 OSS' },
  { value: 'cos', label: '腾讯云 COS' },
  { value: 'qiniu', label: '七牛云' },
]
const STORAGE_OPTIONS_PLACEHOLDER = `{
  "endpoint": "https://s3.amazonaws.com",
  "bucket": "my-bucket",
  "access_key": "AKIA...",
  "secret_key": "...",
  "region": "us-east-1"
}`

export function AdminStoragePage() {
  const api = useApi()
  const [items, setItems] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [editing, setEditing] = useState<any | null>(null)
  const [showNew, setShowNew] = useState(false)
  const [form, setForm] = useState({ ...EMPTY_STORAGE_FORM })
  const [saving, setSaving] = useState(false)

  const load = () => {
    setLoading(true)
    api.get<any>('/api/v1/admin/storages', { raw: true })
      .then((r) => setItems(Array.isArray(r?.data) ? r.data : []))
      .catch(() => setItems([]))
      .finally(() => setLoading(false))
  }
  useEffect(() => { load() }, []) // eslint-disable-line react-hooks/exhaustive-deps

  const openCreate = () => {
    setEditing(null)
    setForm({ ...EMPTY_STORAGE_FORM })
    setShowNew(true)
  }

  const openEdit = (s: any) => {
    setEditing(s)
    let opts = ''
    if (s.options) {
      try { opts = JSON.stringify(JSON.parse(s.options), null, 2) } catch { opts = String(s.options) }
    }
    setForm({
      name: s.name ?? '',
      provider: s.driver ?? s.provider ?? 'local',
      prefix: s.prefix ?? '',
      intro: s.intro ?? '',
      options: opts,
    })
    setShowNew(true)
  }

  const closeDialog = () => { setShowNew(false); setEditing(null) }

  const submit = async () => {
    if (!form.name.trim()) return
    const options = form.options.trim()
    if (options) {
      try { JSON.parse(options) } catch {
        toast.error('options 不是合法的 JSON')
        return
      }
    }
    setSaving(true)
    try {
      if (editing?.id) {
        // 后端 update 不支持改 driver，仅 name/intro/prefix/options
        await api.patch(`/api/v1/admin/storages/update/${editing.id}`, {
          name: form.name.trim(),
          intro: form.intro,
          prefix: form.prefix,
          options: options || null,
        })
        toast.success('已保存')
      } else {
        await api.post('/api/v1/admin/storages/create', {
          name: form.name.trim(),
          provider: form.provider,
          intro: form.intro,
          prefix: form.prefix,
          options: options || null,
        })
        toast.success('已创建存储策略')
      }
      closeDialog()
      load()
    } catch (e: any) {
      toast.error(e?.message || '保存失败')
    } finally {
      setSaving(false)
    }
  }

  const removeStorage = async (s: any) => {
    if (!window.confirm(`确认删除存储策略「${s.name}」？`)) return
    try {
      await api.del(`/api/v1/admin/storages/delete/${s.id}`)
      toast.success('已删除')
      load()
    } catch (e: any) {
      toast.error(e?.message || '删除失败')
    }
  }

  return (
    <AdminShell>
      <AdminPageHeader title="存储策略" description={`共 ${items.length} 个`}>
        <Button size="sm" onClick={openCreate}><Plus className="h-4 w-4" /> 新建策略</Button>
      </AdminPageHeader>
      <ListTable
        loading={loading}
        empty="暂无存储策略"
        head={['名称', '驱动', '前缀', '介绍', '状态', '操作']}
        rows={items.map((s) => [
          s.name || '-',
          <span key="d" className="font-mono text-xs">{s.driver || '-'}</span>,
          s.prefix || '-',
          s.intro || '-',
          <Badge key="b" variant={s.status === 1 || s.enabled ? 'success' : 'secondary'}>{s.status === 1 || s.enabled ? '启用' : '停用'}</Badge>,
          <span key="a" className="flex gap-2">
            <Button variant="outline" size="sm" onClick={() => openEdit(s)}>编辑</Button>
            <Button variant="outline" size="sm" className="text-destructive" onClick={() => removeStorage(s)}>删除</Button>
          </span>,
        ])}
      />

      <Dialog open={showNew} onOpenChange={(o) => { if (!o) closeDialog() }}>
        <DialogContent className="max-w-lg max-h-[85vh] overflow-y-auto">
          <DialogHeader><DialogTitle>{editing?.id ? '编辑存储策略' : '新建存储策略'}</DialogTitle></DialogHeader>
          <div className="space-y-4">
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-2">
                <Label>名称</Label>
                <Input value={form.name} onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))} autoFocus />
              </div>
              <div className="space-y-2">
                <Label>驱动类型{editing?.id ? '（创建后不可修改）' : ''}</Label>
                <select
                  className="h-9 w-full rounded-md border border-border bg-background px-3 text-sm disabled:opacity-60"
                  value={form.provider}
                  disabled={!!editing?.id}
                  onChange={(e) => setForm((f) => ({ ...f, provider: e.target.value }))}
                >
                  {STORAGE_PROVIDERS.map((p) => <option key={p.value} value={p.value}>{p.label}</option>)}
                </select>
              </div>
              <div className="space-y-2">
                <Label>路径前缀</Label>
                <Input value={form.prefix} onChange={(e) => setForm((f) => ({ ...f, prefix: e.target.value }))} placeholder="留空 = 根路径" />
              </div>
              <div className="space-y-2">
                <Label>介绍</Label>
                <Input value={form.intro} onChange={(e) => setForm((f) => ({ ...f, intro: e.target.value }))} />
              </div>
            </div>
            <div className="space-y-2">
              <Label>连接配置（JSON）</Label>
              <Textarea rows={8} className="font-mono text-xs" value={form.options} onChange={(e) => setForm((f) => ({ ...f, options: e.target.value }))} placeholder={STORAGE_OPTIONS_PLACEHOLDER} />
              <p className="text-xs text-muted-foreground">local 驱动可留空；远端驱动按需填写 endpoint / bucket / access_key / secret_key 等字段。</p>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={closeDialog}>取消</Button>
            <Button onClick={submit} loading={saving} disabled={!form.name.trim()}>{editing?.id ? '保存' : '创建'}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </AdminShell>
  )
}

// ---------- 驱动管理 ----------
export function AdminDriversPage() {
  const api = useApi()
  const [data, setData] = useState<{ storage?: string[]; email?: string[] }>({})
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    api.get<any>('/api/v1/admin/drivers', { raw: true })
      .then((r) => setData(r?.data ?? {}))
      .catch(() => setData({}))
      .finally(() => setLoading(false))
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const groups = [
    { key: 'storage', label: '存储驱动', items: data.storage ?? [] },
    { key: 'email', label: '邮件驱动', items: data.email ?? [] },
    { key: 'oauth', label: 'OAuth 驱动', items: (data as any).oauth ?? [] },
  ].filter((g) => (g.items as string[]).length > 0)

  return (
    <AdminShell>
      <AdminPageHeader title="驱动管理" description="系统已注册的驱动" />
      {loading ? (
        <div className="skeleton h-24 rounded-md" />
      ) : groups.length === 0 ? (
        <div className="rounded-md border border-dashed border-border py-20 text-center">
          <p className="text-sm text-muted-foreground">暂无驱动。</p>
        </div>
      ) : (
        <div className="space-y-4">
          {groups.map((g) => (
            <div key={g.key}>
              <h3 className="mb-2 text-sm font-medium text-muted-foreground">{g.label}</h3>
              <div className="flex flex-wrap gap-2">
                {(g.items as string[]).map((name) => (
                  <span key={name} className="rounded-md border border-border bg-card px-3 py-1.5 font-mono text-xs">{name}</span>
                ))}
              </div>
            </div>
          ))}
        </div>
      )}
    </AdminShell>
  )
}

// ============ 通用子组件 ============

function ListTable({ head, rows, loading, empty }: { head: string[]; rows: ReactNode[][]; loading: boolean; empty: string }) {
  return (
    <div className="overflow-hidden rounded-md border border-border">
      <table className="w-full text-sm">
        <thead className="border-b border-border bg-muted/50">
          <tr className="text-left text-xs text-muted-foreground">
            {head.map((h) => <th key={h} className="px-4 py-2.5 font-medium">{h}</th>)}
          </tr>
        </thead>
        <tbody>
          {loading ? (
            <tr><td colSpan={head.length} className="px-4 py-8 text-center text-muted-foreground">加载中…</td></tr>
          ) : rows.length === 0 ? (
            <tr><td colSpan={head.length} className="px-4 py-8 text-center text-muted-foreground">{empty}</td></tr>
          ) : rows.map((cells, i) => (
            <tr key={i} className="border-b border-border last:border-0 hover:bg-muted/30">
              {cells.map((c, j) => <td key={j} className="px-4 py-3">{c}</td>)}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

function Pager({ page, total, perPage, onChange }: { page: number; total: number; perPage: number; onChange: (p: number) => void }) {
  const lastPage = Math.max(1, Math.ceil(total / perPage))
  if (total <= perPage) return null
  return (
    <div className="mt-4 flex items-center justify-center gap-3 text-sm">
      <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => onChange(page - 1)}>上一页</Button>
      <span className="text-muted-foreground">第 {page} / {lastPage} 页</span>
      <Button variant="outline" size="sm" disabled={page >= lastPage} onClick={() => onChange(page + 1)}>下一页</Button>
    </div>
  )
}
