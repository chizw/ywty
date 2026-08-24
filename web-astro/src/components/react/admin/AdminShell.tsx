// 管理后台外壳：与用户中心一致的侧边栏 + 顶栏 + 内容
import { useEffect, useState, type ReactNode } from 'react'
import {
  LayoutDashboard,
  Image,
  AlertTriangle,
  Ticket,
  Bell,
  FileText,
  ShieldAlert,
  MessageSquare,
  Users,
  Shield,
  Package,
  Percent,
  HardDrive,
  Cpu,
  Settings,
  ExternalLink,
  LogOut,
} from 'lucide-react'
import { useAuthStore } from '@/lib/react-store'
import { useSiteInfo } from '@/lib/use-site-info'
import { Toaster } from '../dashboard/Toaster'
import { cn } from '@/lib/utils'

const NAV = [
  { to: '/admin', label: '仪表盘', icon: LayoutDashboard },
  { to: '/admin/photos', label: '图片管理', icon: Image },
  { to: '/admin/reports', label: '举报管理', icon: AlertTriangle },
  { to: '/admin/tickets', label: '工单管理', icon: Ticket },
  { to: '/admin/notices', label: '通知管理', icon: Bell },
  { to: '/admin/pages', label: '单页管理', icon: FileText },
  { to: '/admin/violations', label: '违规记录', icon: ShieldAlert },
  { to: '/admin/feedbacks', label: '意见反馈', icon: MessageSquare },
  { to: '/admin/users', label: '用户', icon: Users },
  { to: '/admin/groups', label: '角色组', icon: Shield },
  { to: '/admin/plans', label: '套餐管理', icon: Package },
  { to: '/admin/coupons', label: '优惠券管理', icon: Percent },
  { to: '/admin/storage', label: '存储策略', icon: HardDrive },
  { to: '/admin/drivers', label: '驱动管理', icon: Cpu },
  { to: '/admin/settings', label: '系统设置', icon: Settings },
]

export function AdminShell({ children }: { children: ReactNode }) {
  const user = useAuthStore((s) => s.user)
  const hydrate = useAuthStore((s) => s.hydrate)
  const fetchMe = useAuthStore((s) => s.fetchMe)
  const logout = useAuthStore((s) => s.logout)
  const site = useSiteInfo()
  const sealChar = site.name.slice(-1)

  const [path, setPath] = useState('')

  useEffect(() => {
    hydrate()
    setPath(window.location.pathname)
    // 以服务端为准刷新用户信息：旧 cookie 可能缺少 is_super_admin 等字段，
    // 否则角色相关操作按钮会因陈旧缓存而不渲染
    fetchMe()
  }, [hydrate, fetchMe])

  return (
    <div className="flex min-h-screen bg-background">
      {/* 侧边栏 */}
      <aside className="sticky top-0 hidden h-screen w-56 flex-shrink-0 flex-col border-r border-border bg-card lg:flex">
        <a href="/admin" className="flex h-16 items-center gap-2 border-b border-border px-5">
          <span className="font-display text-lg font-bold tracking-tight">云雾图驿</span>
          <span className="seal h-[1.15rem] w-[1.15rem] text-[0.55rem]">驿</span>
        </a>
        <nav className="flex-1 overflow-y-auto px-2 py-4">
          {NAV.map((item) => {
            const active = path === item.to
            return (
              <a
                key={item.to}
                href={item.to}
                className={cn(
                  'flex items-center gap-2.5 rounded-md px-3 py-2 text-sm transition-colors',
                  active
                    ? 'bg-accent font-medium text-accent-foreground'
                    : 'text-muted-foreground hover:bg-muted hover:text-foreground'
                )}
              >
                <item.icon className="h-4 w-4 shrink-0" />
                {item.label}
              </a>
            )
          })}
        </nav>
        <div className="border-t border-border p-3">
          <div className="flex items-center gap-2 px-1 py-1">
            <span className="grid h-7 w-7 flex-shrink-0 place-items-center rounded-full bg-brand text-xs font-medium text-primary-foreground">
              {(user?.name || user?.username || 'A').slice(0, 1).toUpperCase()}
            </span>
            <div className="min-w-0 flex-1">
              <p className="truncate text-sm text-foreground">{user?.name || user?.username}</p>
              <p className="text-[0.65rem] text-muted-foreground">管理员</p>
            </div>
          </div>
          <button
            className="mt-1 flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-sm text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
            onClick={async () => {
              await logout()
              window.location.assign('/auth/login')
            }}
          >
            <LogOut className="h-4 w-4" />
            退出登录
          </button>
        </div>
      </aside>

      {/* 主区域 */}
      <div className="flex min-w-0 flex-1 flex-col">
        <header className="sticky top-0 z-20 flex h-16 items-center justify-between border-b border-border bg-background/85 px-5 backdrop-blur-sm sm:px-8">
          <div className="flex items-center gap-3 lg:hidden">
            <a href="/admin" className="font-display text-lg font-bold tracking-tight">云雾图驿</a>
          </div>
          <div className="hidden items-center gap-3 lg:flex">
            <span className="eyebrow !tracking-[0.12em]">管理后台</span>
          </div>
          <div className="flex items-center gap-2">
            <a href="/" className="inline-flex items-center gap-1.5 rounded-md px-3 py-1.5 text-sm text-muted-foreground transition-colors hover:bg-muted hover:text-foreground">
              <ExternalLink className="h-3.5 w-3.5" />
              前台
            </a>
          </div>
        </header>

        <main className="min-w-0 flex-1 px-5 py-6 sm:px-8">{children}</main>
      </div>

      <Toaster />
    </div>
  )
}
