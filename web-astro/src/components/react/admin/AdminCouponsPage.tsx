// 优惠券管理：分页列表 + 新建/编辑/删除
// 字段对应后端 models::coupon：type/name/code/value/usage_limit/used_count/expired_at
import { useEffect, useState } from 'react'
import { Plus } from 'lucide-react'
import { AdminShell } from './AdminShell'
import { AdminPageHeader } from './AdminPageHeader'
import { useConfirm } from '../dashboard/ConfirmDialog'
import { useApi } from '@/lib/api'
import { toast } from '@/lib/react-store'
import { formatDate } from '@/lib/utils'
import { Badge } from '../ui/badge'
import { Button } from '../ui/button'
import { Input } from '../ui/input'
import { Label } from '../ui/label'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '../ui/dialog'

interface CouponRecord {
  id: number
  type: string
  name: string
  code: string
  value: number
  usage_limit: number
  used_count: number
  expired_at: string | null
  created_at: string
}

const EMPTY_FORM = {
  coupon_type: 'direct',
  name: '',
  code: '',
  value: '',
  usage_limit: '1',
  expired_at: '',
}

/** direct 面值以分为单位存储，percent 存百分比数值 */
function valueLabel(c: CouponRecord) {
  return c.type === 'percent' ? `${c.value}%` : `¥${(c.value / 100).toFixed(2)}`
}

export function AdminCouponsPage() {
  const api = useApi()
  const { confirm, node } = useConfirm()
  const [items, setItems] = useState<CouponRecord[]>([])
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [showNew, setShowNew] = useState(false)
  const [editing, setEditing] = useState<CouponRecord | null>(null)
  const [form, setForm] = useState({ ...EMPTY_FORM })
  const [saving, setSaving] = useState(false)

  const perPage = 20

  const load = () => {
    setLoading(true)
    api.get<any>('/api/v1/admin/coupons', { query: { page, per_page: perPage }, raw: true })
      .then((r) => {
        setItems(Array.isArray(r?.data) ? r.data : [])
        setTotal(Number(r?.meta?.total ?? 0))
      })
      .catch(() => setItems([]))
      .finally(() => setLoading(false))
  }
  useEffect(() => { load() }, [page]) // eslint-disable-line react-hooks/exhaustive-deps

  const openCreate = () => {
    setEditing(null)
    setForm({ ...EMPTY_FORM })
    setShowNew(true)
  }

  const openEdit = (c: CouponRecord) => {
    setShowNew(false)
    setEditing(c)
    setForm({
      coupon_type: c.type === 'percent' ? 'percent' : 'direct',
      name: c.name ?? '',
      code: c.code ?? '',
      value: c.type === 'percent' ? String(c.value ?? '') : ((Number(c.value) || 0) / 100).toFixed(2),
      usage_limit: String(c.usage_limit ?? 1),
      expired_at: c.expired_at ? String(c.expired_at).slice(0, 16) : '',
    })
  }

  const closeDialog = () => { setShowNew(false); setEditing(null) }

  const save = async () => {
    if (!form.name.trim() || !form.code.trim()) return
    const value = Number(form.value)
    if (!Number.isFinite(value) || value <= 0 || (form.coupon_type === 'percent' && value >= 100)) {
      toast.error(form.coupon_type === 'percent' ? '折扣比例需在 1 ~ 99 之间' : '请填写大于 0 的面值')
      return
    }
    const usageLimit = Math.floor(Number(form.usage_limit))
    if (!Number.isFinite(usageLimit) || usageLimit < 1) {
      toast.error('使用上限需为不小于 1 的整数')
      return
    }
    setSaving(true)
    try {
      const payload = {
        coupon_type: form.coupon_type,
        name: form.name.trim(),
        code: form.code.trim(),
        // direct：元 → 分；percent：百分比原样存储
        value: form.coupon_type === 'percent' ? value : Math.round(value * 100),
        usage_limit: usageLimit,
        // datetime-local → RFC3339，空串表示长期有效
        expired_at: form.expired_at ? new Date(form.expired_at).toISOString() : '',
      }
      if (editing) {
        await api.patch(`/api/v1/admin/coupons/${editing.id}`, payload)
        toast.success('优惠券已更新')
      } else {
        await api.post('/api/v1/admin/coupons', payload)
        toast.success('优惠券已创建')
      }
      closeDialog()
      load()
    } catch (e: any) {
      toast.error(e?.message || '保存失败')
    } finally {
      setSaving(false)
    }
  }

  const remove = async (c: CouponRecord) => {
    const ok = await confirm({
      title: '删除优惠券',
      message: `确定删除优惠券「${c.name}」（${c.code}）？`,
      okText: '删除',
      danger: true,
    })
    if (!ok) return
    try {
      await api.del(`/api/v1/admin/coupons/${c.id}`)
      toast.success('已删除')
      load()
    } catch (e: any) {
      toast.error(e?.message || '删除失败')
    }
  }

  const lastPage = Math.max(1, Math.ceil(total / perPage))

  return (
    <AdminShell>
      <AdminPageHeader title="优惠券管理" description={`共 ${total} 张`}>
        <Button size="sm" onClick={openCreate}><Plus className="h-4 w-4" /> 新建优惠券</Button>
      </AdminPageHeader>

      <div className="overflow-hidden rounded-md border border-border">
        <table className="w-full text-sm">
          <thead className="border-b border-border bg-muted/50">
            <tr className="text-left text-xs text-muted-foreground">
              {['名称', '兑换码', '类型', '面值', '使用', '过期时间', '操作'].map((h) => (
                <th key={h} className="px-4 py-2.5 font-medium">{h}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr><td colSpan={7} className="px-4 py-8 text-center text-muted-foreground">加载中…</td></tr>
            ) : items.length === 0 ? (
              <tr><td colSpan={7} className="px-4 py-8 text-center text-muted-foreground">暂无优惠券</td></tr>
            ) : items.map((c) => (
              <tr key={c.id} className="border-b border-border last:border-0 hover:bg-muted/30">
                <td className="px-4 py-3 font-medium">{c.name}</td>
                <td className="px-4 py-3"><span className="font-mono text-xs">{c.code}</span></td>
                <td className="px-4 py-3">
                  <Badge variant={c.type === 'percent' ? 'warning' : 'brand'}>{c.type === 'percent' ? '折扣' : '满减'}</Badge>
                </td>
                <td className="px-4 py-3 tabular-nums">{valueLabel(c)}</td>
                <td className="px-4 py-3 tabular-nums">{c.used_count}/{c.usage_limit}</td>
                <td className="px-4 py-3 text-xs text-muted-foreground">{c.expired_at ? formatDate(c.expired_at) : '长期有效'}</td>
                <td className="px-4 py-3">
                  <div className="flex flex-wrap gap-2">
                    <Button variant="outline" size="sm" onClick={() => openEdit(c)}>编辑</Button>
                    <Button variant="outline" size="sm" className="text-destructive hover:text-destructive" onClick={() => remove(c)}>删除</Button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {total > perPage && (
        <div className="mt-4 flex items-center justify-center gap-3 text-sm">
          <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage(page - 1)}>上一页</Button>
          <span className="text-muted-foreground">第 {page} / {lastPage} 页</span>
          <Button variant="outline" size="sm" disabled={page >= lastPage} onClick={() => setPage(page + 1)}>下一页</Button>
        </div>
      )}

      <Dialog open={showNew || !!editing} onOpenChange={(o) => { if (!o) closeDialog() }}>
        <DialogContent className="max-w-md">
          <DialogHeader><DialogTitle>{editing ? '编辑优惠券' : '新建优惠券'}</DialogTitle></DialogHeader>
          <div className="space-y-4">
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-2">
                <Label>类型</Label>
                <select
                  className="h-9 w-full rounded-md border border-border bg-background px-3 text-sm"
                  value={form.coupon_type}
                  onChange={(e) => setForm((f) => ({ ...f, coupon_type: e.target.value }))}
                >
                  <option value="direct">满减（元）</option>
                  <option value="percent">折扣（%）</option>
                </select>
              </div>
              <div className="space-y-2">
                <Label>{form.coupon_type === 'percent' ? '折扣比例（%）' : '面值（元）'}</Label>
                <Input
                  type="number"
                  min="0"
                  step={form.coupon_type === 'percent' ? '1' : '0.01'}
                  value={form.value}
                  onChange={(e) => setForm((f) => ({ ...f, value: e.target.value }))}
                  placeholder={form.coupon_type === 'percent' ? '10 = 打九折' : '5.00'}
                />
              </div>
              <div className="space-y-2">
                <Label>名称</Label>
                <Input value={form.name} onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))} autoFocus />
              </div>
              <div className="space-y-2">
                <Label>兑换码</Label>
                <Input value={form.code} onChange={(e) => setForm((f) => ({ ...f, code: e.target.value }))} placeholder="VIP10" />
              </div>
              <div className="space-y-2">
                <Label>使用上限</Label>
                <Input type="number" min="1" value={form.usage_limit} onChange={(e) => setForm((f) => ({ ...f, usage_limit: e.target.value }))} />
              </div>
              <div className="space-y-2">
                <Label>过期时间</Label>
                <Input type="datetime-local" value={form.expired_at} onChange={(e) => setForm((f) => ({ ...f, expired_at: e.target.value }))} />
              </div>
            </div>
            <p className="text-xs text-muted-foreground">过期时间留空表示长期有效。</p>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={closeDialog}>取消</Button>
            <Button onClick={save} loading={saving} disabled={!form.name.trim() || !form.code.trim()}>{editing ? '保存' : '创建'}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {node}
    </AdminShell>
  )
}
