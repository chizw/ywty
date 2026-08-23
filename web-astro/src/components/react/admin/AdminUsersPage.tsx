// 用户管理：表格 + 搜索 + 角色/状态/删除
// 权限：超级管理员可管理除自己/超管外的所有账号；普通管理员只能管理普通用户
import { useEffect, useState } from 'react'
import { HardDrive, Search, Trash2 } from 'lucide-react'
import { AdminShell } from './AdminShell'
import { AdminPageHeader } from './AdminPageHeader'
import { useConfirm } from '../dashboard/ConfirmDialog'
import { useApi } from '@/lib/api'
import { useAuthStore, isSuperAdmin, toast } from '@/lib/react-store'
import { formatDate } from '@/lib/utils'
import { Badge } from '../ui/badge'
import { Button } from '../ui/button'
import { Input } from '../ui/input'
import { Label } from '../ui/label'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '../ui/dialog'

interface AdminUser {
  id: number
  username: string
  email: string
  role: string
  is_super_admin: boolean
  status: number
  capacity_used: number
  capacity_max: number
  quota_override?: number | null
  created_at: string
}

export function AdminUsersPage() {
  const api = useApi()
  const actor = useAuthStore((s) => s.user)
  const actorSuper = isSuperAdmin(actor)
  const { confirm, node } = useConfirm()

  const [users, setUsers] = useState<AdminUser[]>([])
  const [total, setTotal] = useState(0)
  const [keyword, setKeyword] = useState('')
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(true)

  // 存储配额覆盖编辑弹窗（GB，留空 = 清除覆盖、跟随角色组）
  const [quotaUser, setQuotaUser] = useState<AdminUser | null>(null)
  const [quotaGb, setQuotaGb] = useState('')
  const [quotaSaving, setQuotaSaving] = useState(false)

  const openQuota = (u: AdminUser) => {
    setQuotaUser(u)
    setQuotaGb(u.quota_override != null ? String(u.quota_override / 1024 ** 3) : '')
  }

  const saveQuota = async () => {
    if (!quotaUser) return
    const trimmed = quotaGb.trim()
    let payload: number | null = null
    if (trimmed !== '') {
      const gb = Number(trimmed)
      if (!Number.isFinite(gb) || gb <= 0) {
        toast.error('请输入大于 0 的 GB 数值')
        return
      }
      payload = Math.round(gb * 1024 * 1024 * 1024)
    }
    setQuotaSaving(true)
    try {
      await api.patch(`/api/v1/admin/users/${quotaUser.id}`, { quota_override: payload })
      toast.success(payload == null ? '已清除配额覆盖' : '已更新存储配额')
      setQuotaUser(null)
      load()
    } catch (e: any) {
      toast.error(e?.message || '保存失败')
    } finally {
      setQuotaSaving(false)
    }
  }

  const load = () => {
    setLoading(true)
    api.get<any>('/api/v1/admin/users', { query: { page, per_page: 20, keyword }, raw: true })
      .then((r) => {
        setUsers(Array.isArray(r?.data) ? r.data : [])
        setTotal(Number(r?.meta?.total ?? 0))
      })
      .catch(() => setUsers([]))
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [page])

  // 当前管理员对该用户能做什么
  const canManage = (u: AdminUser) => {
    if (u.is_super_admin) return false // 超管账号不可操作
    if (u.id === actor?.id) return false // 不能操作自己
    if (actorSuper) return true // 超管可管理除自己/超管外的所有人
    return u.role !== 'admin' // 普通管理员只能管理普通用户
  }

  const toggleRole = async (u: AdminUser) => {
    const isAdmin = u.role === 'admin'
    try {
      await api.patch(`/api/v1/admin/users/${u.id}`, { is_admin: !isAdmin })
      toast.success(isAdmin ? '已取消管理员' : '已设为管理员')
      load()
    } catch (e: any) {
      toast.error(e?.message || '操作失败')
    }
  }

  const toggleStatus = async (u: AdminUser) => {
    try {
      await api.patch(`/api/v1/admin/users/${u.id}`, { status: u.status === 1 ? 0 : 1 })
      toast.success(u.status === 1 ? '已禁用' : '已启用')
      load()
    } catch (e: any) {
      toast.error(e?.message || '操作失败')
    }
  }

  const remove = async (u: AdminUser) => {
    const ok = await confirm({
      title: '删除用户',
      message: `确定删除用户「${u.username}」？其账号及所有图片、相册、分享都会被删除，且不可恢复。`,
      okText: '删除',
      danger: true,
    })
    if (!ok) return
    try {
      await api.del(`/api/v1/admin/users/${u.id}`)
      toast.success('已删除')
      load()
    } catch (e: any) {
      toast.error(e?.message || '删除失败')
    }
  }

  const roleLabel = (u: AdminUser) =>
    u.is_super_admin ? '超级管理员' : u.role === 'admin' ? '管理员' : '用户'

  const roleVariant = (u: AdminUser) =>
    u.is_super_admin ? ('brand' as const) : u.role === 'admin' ? ('brand' as const) : ('secondary' as const)

  const lastPage = Math.max(1, Math.ceil(total / 20))

  return (
    <AdminShell>
      <AdminPageHeader title="用户" description={`共 ${total} 个用户${actorSuper ? ' · 超级管理员' : ''}`} />

      <div className="mb-4 flex max-w-sm gap-2">
        <Input value={keyword} onChange={(e) => setKeyword(e.target.value)} placeholder="搜索用户名 / 邮箱" onKeyDown={(e) => { if (e.key === 'Enter') { setPage(1); load() } }} />
        <Button variant="outline" onClick={() => { setPage(1); load() }}>
          <Search className="h-4 w-4" /> 搜索
        </Button>
      </div>

      <div className="overflow-hidden rounded-md border border-border">
        <table className="w-full text-sm">
          <thead className="border-b border-border bg-muted/50">
            <tr className="text-left text-xs text-muted-foreground">
              <th className="px-4 py-2.5 font-medium">用户</th>
              <th className="px-4 py-2.5 font-medium">角色</th>
              <th className="px-4 py-2.5 font-medium">状态</th>
              <th className="px-4 py-2.5 font-medium">存储</th>
              <th className="px-4 py-2.5 font-medium">注册时间</th>
              <th className="px-4 py-2.5 font-medium">操作</th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr><td colSpan={6} className="px-4 py-8 text-center text-muted-foreground">加载中…</td></tr>
            ) : users.length === 0 ? (
              <tr><td colSpan={6} className="px-4 py-8 text-center text-muted-foreground">没有用户</td></tr>
            ) : users.map((u) => (
              <tr key={u.id} className="border-b border-border last:border-0 hover:bg-muted/30">
                <td className="px-4 py-3">
                  <div className="flex items-center gap-2">
                    <span className="font-medium">{u.username}</span>
                    {u.id === actor?.id && <span className="text-xs text-muted-foreground">（我）</span>}
                  </div>
                  <div className="text-xs text-muted-foreground">{u.email}</div>
                </td>
                <td className="px-4 py-3">
                  <Badge variant={roleVariant(u)}>{roleLabel(u)}</Badge>
                </td>
                <td className="px-4 py-3">
                  <Badge variant={u.status === 1 ? 'success' : 'destructive'}>{u.status === 1 ? '正常' : '禁用'}</Badge>
                </td>
                <td className="px-4 py-3 text-xs text-muted-foreground tabular-nums">
                  {Math.round(u.capacity_used / 1024 / 1024)} MB / {u.capacity_max ? `${Math.round(u.capacity_max / 1024 / 1024)} MB` : '∞'}
                </td>
                <td className="px-4 py-3 text-xs text-muted-foreground">{formatDate(u.created_at)}</td>
                <td className="px-4 py-3">
                  <div className="flex gap-1.5">
                    {actorSuper && !u.is_super_admin && u.id !== actor?.id && (
                      <Button variant="outline" size="sm" className="h-7 text-xs" onClick={() => toggleRole(u)}>
                        {u.role === 'admin' ? '取消管理员' : '设为管理员'}
                      </Button>
                    )}
                    {canManage(u) && (
                      <>
                        <Button variant="outline" size="sm" className="h-7 text-xs" onClick={() => openQuota(u)}>
                          <HardDrive className="h-3.5 w-3.5" /> 配额
                        </Button>
                        <Button variant="outline" size="sm" className="h-7 text-xs" onClick={() => toggleStatus(u)}>
                          {u.status === 1 ? '禁用' : '启用'}
                        </Button>
                        <Button variant="ghost" size="sm" className="h-7 text-xs text-destructive" onClick={() => remove(u)}>
                          <Trash2 className="h-3.5 w-3.5" /> 删除
                        </Button>
                      </>
                    )}
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {total > 20 && (
        <div className="mt-4 flex items-center justify-center gap-3 text-sm">
          <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>上一页</Button>
          <span className="text-muted-foreground">第 {page} / {lastPage} 页</span>
          <Button variant="outline" size="sm" disabled={page >= lastPage} onClick={() => setPage((p) => p + 1)}>下一页</Button>
        </div>
      )}

      <Dialog open={quotaUser != null} onOpenChange={(open) => { if (!open) setQuotaUser(null) }}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>存储配额覆盖</DialogTitle>
            <DialogDescription>为「{quotaUser?.username}」设置单独的存储配额，优先于角色组配额。</DialogDescription>
          </DialogHeader>
          <div className="grid gap-2">
            <Label htmlFor="quota-override">配额（GB）</Label>
            <Input
              id="quota-override"
              type="number"
              min="0"
              step="0.1"
              value={quotaGb}
              onChange={(e) => setQuotaGb(e.target.value)}
              placeholder="留空 = 跟随角色组"
            />
            <p className="text-xs text-muted-foreground">已用 {Math.round((quotaUser?.capacity_used ?? 0) / 1024 / 1024)} MB</p>
          </div>
          <DialogFooter>
            <Button variant="outline" size="sm" onClick={() => setQuotaUser(null)}>取消</Button>
            <Button size="sm" disabled={quotaSaving} onClick={saveQuota}>保存</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {node}
    </AdminShell>
  )
}
