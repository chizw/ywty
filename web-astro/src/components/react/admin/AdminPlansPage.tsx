// 套餐管理：列表 + 新建/编辑（含多档价格）/ 上下架 / 删除
// 字段对应后端 models::plan：type/name/intro/features/badge/sort/is_up + plan_prices{name,duration,price}
import { useEffect, useState } from 'react'
import { Plus, Trash2 } from 'lucide-react'
import { AdminShell } from './AdminShell'
import { AdminPageHeader } from './AdminPageHeader'
import { useConfirm } from '../dashboard/ConfirmDialog'
import { useApi } from '@/lib/api'
import { toast } from '@/lib/react-store'
import { Badge } from '../ui/badge'
import { Button } from '../ui/button'
import { Input } from '../ui/input'
import { Label } from '../ui/label'
import { Textarea } from '../ui/textarea'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '../ui/dialog'

interface PlanRecord {
  id: number
  type: string
  name: string
  intro: string | null
  features: string | null
  badge: string
  sort: number
  is_up: number
}

/** 后端价格档：duration=天、price=分 */
interface PriceTier {
  name: string
  duration: number
  price: number
}

/** 表单里的价格档行（时长用天、价格用元，提交时换算） */
interface PriceRow {
  name: string
  days: string
  yuan: string
}

const EMPTY_FORM = {
  plan_type: 'vip',
  name: '',
  intro: '',
  features: '',
  badge: '',
  sort: '0',
  is_up: '1',
}

const tierToRow = (t: PriceTier): PriceRow => ({
  name: t.name ?? '',
  days: String(t.duration ?? 0),
  yuan: ((Number(t.price) || 0) / 100).toFixed(2),
})

function priceRangeText(tiers: PriceTier[]): string {
  if (!tiers.length) return '—'
  const cents = tiers.map((t) => Number(t.price) || 0)
  const min = Math.min(...cents)
  const max = Math.max(...cents)
  return min === max ? `¥${(min / 100).toFixed(2)}` : `¥${(min / 100).toFixed(2)} ~ ¥${(max / 100).toFixed(2)}`
}

export function AdminPlansPage() {
  const api = useApi()
  const { confirm, node } = useConfirm()
  const [items, setItems] = useState<PlanRecord[]>([])
  // 套餐 id → 价格档（来自 admin/plans/:id 详情）
  const [tiers, setTiers] = useState<Record<number, PriceTier[]>>({})
  const [loading, setLoading] = useState(true)
  const [showNew, setShowNew] = useState(false)
  const [editing, setEditing] = useState<PlanRecord | null>(null)
  const [form, setForm] = useState({ ...EMPTY_FORM })
  const [priceRows, setPriceRows] = useState<PriceRow[]>([])
  const [saving, setSaving] = useState(false)

  const load = () => {
    setLoading(true)
    api.get<any>('/api/v1/admin/plans', { raw: true })
      .then(async (r) => {
        const list: PlanRecord[] = Array.isArray(r?.data) ? r.data : []
        setItems(list)
        const map: Record<number, PriceTier[]> = {}
        await Promise.all(
          list.map((p) =>
            api.get<any>(`/api/v1/admin/plans/${p.id}`, { raw: true })
              .then((d) => {
                map[p.id] = Array.isArray(d?.data?.prices) ? d.data.prices : []
              })
              .catch(() => {})
          )
        )
        setTiers(map)
      })
      .catch(() => setItems([]))
      .finally(() => setLoading(false))
  }
  useEffect(() => { load() }, []) // eslint-disable-line react-hooks/exhaustive-deps

  const openCreate = () => {
    setEditing(null)
    setForm({ ...EMPTY_FORM })
    setPriceRows([{ name: '', days: '30', yuan: '' }])
    setShowNew(true)
  }

  const openEdit = (p: PlanRecord) => {
    setShowNew(false)
    setEditing(p)
    setForm({
      plan_type: p.type === 'default' ? 'default' : 'vip',
      name: p.name ?? '',
      intro: p.intro ?? '',
      features: p.features ?? '',
      badge: p.badge ?? '',
      sort: String(p.sort ?? 0),
      is_up: p.is_up === 1 ? '1' : '0',
    })
    const known = tiers[p.id]
    if (known) {
      setPriceRows(known.map(tierToRow))
    } else {
      setPriceRows([])
      api.get<any>(`/api/v1/admin/plans/${p.id}`, { raw: true })
        .then((d) => setPriceRows(((Array.isArray(d?.data?.prices) ? d.data.prices : []) as PriceTier[]).map(tierToRow)))
        .catch(() => {})
    }
  }

  const closeDialog = () => { setShowNew(false); setEditing(null); setPriceRows([]) }

  const save = async () => {
    if (!form.name.trim()) return
    // 价格档校验并换算：天 → duration，元 → 分
    const prices: { name: string; duration: number; price: number }[] = []
    for (const row of priceRows) {
      const days = Math.floor(Number(row.days))
      const yuan = Number(row.yuan)
      if (!Number.isFinite(days) || days <= 0 || !Number.isFinite(yuan) || yuan < 0) {
        toast.error('请为每个价格档填写有效时长（天）和价格（元）')
        return
      }
      prices.push({ name: row.name.trim() || `${days} 天`, duration: days, price: Math.round(yuan * 100) })
    }
    setSaving(true)
    try {
      const payload = {
        plan_type: form.plan_type,
        name: form.name.trim(),
        intro: form.intro,
        features: form.features,
        badge: form.badge,
        sort: Number(form.sort) || 0,
        is_up: Number(form.is_up) ? 1 : 0,
        prices,
      }
      if (editing) {
        await api.patch(`/api/v1/admin/plans/${editing.id}`, payload)
        toast.success('套餐已更新')
      } else {
        await api.post('/api/v1/admin/plans', payload)
        toast.success('套餐已创建')
      }
      closeDialog()
      load()
    } catch (e: any) {
      toast.error(e?.message || '保存失败')
    } finally {
      setSaving(false)
    }
  }

  const toggleUp = async (p: PlanRecord) => {
    try {
      await api.post(`/api/v1/admin/plans/${p.id}/toggle`)
      toast.success(p.is_up === 1 ? '已下架' : '已上架')
      load()
    } catch (e: any) {
      toast.error(e?.message || '操作失败')
    }
  }

  const remove = async (p: PlanRecord) => {
    const ok = await confirm({
      title: '删除套餐',
      message: `确定删除套餐「${p.name}」？删除后前台不再展示。`,
      okText: '删除',
      danger: true,
    })
    if (!ok) return
    try {
      await api.del(`/api/v1/admin/plans/${p.id}`)
      toast.success('已删除')
      load()
    } catch (e: any) {
      toast.error(e?.message || '删除失败')
    }
  }

  return (
    <AdminShell>
      <AdminPageHeader title="套餐管理" description={`共 ${items.length} 个套餐`}>
        <Button size="sm" onClick={openCreate}><Plus className="h-4 w-4" /> 新建套餐</Button>
      </AdminPageHeader>

      <div className="overflow-hidden rounded-md border border-border">
        <table className="w-full text-sm">
          <thead className="border-b border-border bg-muted/50">
            <tr className="text-left text-xs text-muted-foreground">
              {['名称', '类型', '价格区间', '排序', '状态', '操作'].map((h) => (
                <th key={h} className="px-4 py-2.5 font-medium">{h}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr><td colSpan={6} className="px-4 py-8 text-center text-muted-foreground">加载中…</td></tr>
            ) : items.length === 0 ? (
              <tr><td colSpan={6} className="px-4 py-8 text-center text-muted-foreground">暂无套餐</td></tr>
            ) : items.map((p) => (
              <tr key={p.id} className="border-b border-border last:border-0 hover:bg-muted/30">
                <td className="px-4 py-3 font-medium">{p.name}</td>
                <td className="px-4 py-3">
                  <Badge variant={p.type === 'vip' ? 'brand' : 'secondary'}>{p.type === 'vip' ? 'VIP' : '默认'}</Badge>
                </td>
                <td className="px-4 py-3 tabular-nums">{priceRangeText(tiers[p.id] || [])}</td>
                <td className="px-4 py-3 tabular-nums">{p.sort}</td>
                <td className="px-4 py-3">
                  <Badge variant={p.is_up === 1 ? 'success' : 'secondary'}>{p.is_up === 1 ? '上架中' : '已下架'}</Badge>
                </td>
                <td className="px-4 py-3">
                  <div className="flex flex-wrap gap-2">
                    <Button variant="outline" size="sm" onClick={() => openEdit(p)}>编辑</Button>
                    <Button variant="outline" size="sm" onClick={() => toggleUp(p)}>{p.is_up === 1 ? '下架' : '上架'}</Button>
                    <Button variant="outline" size="sm" className="text-destructive hover:text-destructive" onClick={() => remove(p)}>删除</Button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <Dialog open={showNew || !!editing} onOpenChange={(o) => { if (!o) closeDialog() }}>
        <DialogContent className="max-w-lg max-h-[85vh] overflow-y-auto">
          <DialogHeader><DialogTitle>{editing ? '编辑套餐' : '新建套餐'}</DialogTitle></DialogHeader>
          <div className="space-y-4">
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-2">
                <Label>名称</Label>
                <Input value={form.name} onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))} autoFocus />
              </div>
              <div className="space-y-2">
                <Label>类型</Label>
                <select
                  className="h-9 w-full rounded-md border border-border bg-background px-3 text-sm"
                  value={form.plan_type}
                  onChange={(e) => setForm((f) => ({ ...f, plan_type: e.target.value }))}
                >
                  <option value="vip">VIP</option>
                  <option value="default">默认</option>
                </select>
              </div>
              <div className="space-y-2">
                <Label>角标</Label>
                <select
                  className="h-9 w-full rounded-md border border-border bg-background px-3 text-sm"
                  value={form.badge}
                  onChange={(e) => setForm((f) => ({ ...f, badge: e.target.value }))}
                >
                  <option value="">无</option>
                  <option value="popular">推荐（popular）</option>
                </select>
              </div>
              <div className="space-y-2">
                <Label>排序（小的在前）</Label>
                <Input type="number" value={form.sort} onChange={(e) => setForm((f) => ({ ...f, sort: e.target.value }))} />
              </div>
              <div className="space-y-2">
                <Label>是否上架</Label>
                <select
                  className="h-9 w-full rounded-md border border-border bg-background px-3 text-sm"
                  value={form.is_up}
                  onChange={(e) => setForm((f) => ({ ...f, is_up: e.target.value }))}
                >
                  <option value="1">上架</option>
                  <option value="0">下架</option>
                </select>
              </div>
            </div>
            <div className="space-y-2">
              <Label>简介</Label>
              <Input value={form.intro} onChange={(e) => setForm((f) => ({ ...f, intro: e.target.value }))} placeholder="一句话介绍" />
            </div>
            <div className="space-y-2">
              <Label>特性（每行一条）</Label>
              <Textarea rows={4} value={form.features} onChange={(e) => setForm((f) => ({ ...f, features: e.target.value }))} placeholder={'100GB 存储空间\n单文件 100MB'} />
            </div>

            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <Label>价格档</Label>
                <span className="text-xs text-muted-foreground">保存时整体替换已有价格档</span>
              </div>
              {priceRows.length > 0 && (
                <div className="grid grid-cols-[1fr_5rem_6rem_2rem] gap-2 text-xs text-muted-foreground">
                  <span>档位名</span>
                  <span>时长（天）</span>
                  <span>价格（元）</span>
                  <span />
                </div>
              )}
              <div className="space-y-2">
                {priceRows.map((row, i) => (
                  <div key={i} className="grid grid-cols-[1fr_5rem_6rem_2rem] items-center gap-2">
                    <Input
                      value={row.name}
                      onChange={(e) => setPriceRows((rows) => rows.map((r, j) => (j === i ? { ...r, name: e.target.value } : r)))}
                      placeholder="如 月付"
                    />
                    <Input
                      type="number"
                      min="1"
                      value={row.days}
                      onChange={(e) => setPriceRows((rows) => rows.map((r, j) => (j === i ? { ...r, days: e.target.value } : r)))}
                      placeholder="30"
                    />
                    <Input
                      type="number"
                      min="0"
                      step="0.01"
                      value={row.yuan}
                      onChange={(e) => setPriceRows((rows) => rows.map((r, j) => (j === i ? { ...r, yuan: e.target.value } : r)))}
                      placeholder="9.90"
                    />
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon"
                      aria-label="移除该价格档"
                      onClick={() => setPriceRows((rows) => rows.filter((_, j) => j !== i))}
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </div>
                ))}
              </div>
              <Button type="button" variant="outline" size="sm" onClick={() => setPriceRows((rows) => [...rows, { name: '', days: '30', yuan: '' }])}>
                <Plus className="h-4 w-4" /> 添加价格档
              </Button>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={closeDialog}>取消</Button>
            <Button onClick={save} loading={saving} disabled={!form.name.trim()}>{editing ? '保存' : '创建'}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {node}
    </AdminShell>
  )
}
