// 用户中心外壳：左侧导航 + 顶栏 + 内容 + 认证门
import { useEffect, useState, type ReactNode } from 'react'
import {
  Home,
  Image,
  BookImage,
  Crown,
  Share2,
  ShoppingCart,
  Tag,
  Ticket,
  Bell,
  Key,
  Link as LinkIcon,
  Settings,
  LogOut,
  ExternalLink,
} from 'lucide-react'
import { useAuthStore, isAdminUser } from '@/lib/react-store'
import { useSiteInfo } from '@/lib/use-site-info'
import { weavatarUrl } from '@/lib/weavatar'
import { Toaster } from './Toaster'
import { cn } from '@/lib/utils'

interface NavItem {
  to: string
  label: string
  icon: React.ComponentType<{ className?: string }>
}

const NAV: NavItem[] = [
  { to: '/dashboard', label: '概览', icon: Home },
  { to: '/dashboard/photos', label: '我的图片', icon: Image },
  { to: '/dashboard/albums', label: '我的相册', icon: BookImage },
  { to: '/dashboard/plans', label: '套餐', icon: Crown },
  { to: '/dashboard/shares', label: '分享管理', icon: Share2 },
  { to: '/dashboard/orders', label: '订单', icon: ShoppingCart },
  { to: '/dashboard/tags', label: '标签', icon: Tag },
  { to: '/dashboard/tickets', label: '工单', icon: Ticket },
  { to: '/dashboard/notices', label: '通知', icon: Bell },
  { to: '/dashboard/tokens', label: 'API Token', icon: Key },
  { to: '/dashboard/oauth', label: '三方账号', icon: LinkIcon },
  { to: '/dashboard/settings', label: '设置', icon: Settings },
]

export function AppShell({ children }: { children: ReactNode }) {
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

    // 静态部署无服务端守卫：客户端校验登录
    if (!useAuthStore.getState().user) {
      window.location.assign(`/auth/login?redirect=${encodeURIComponent(window.location.pathname)}`)
      return
    }

    fetchMe().then((me) => {
      if (!me) {
        // token 失效且刷新失败 → 回登录页
        const still = useAuthStore.getState().user
        if (!still) window.location.assign('/auth/login')
      }
    })
  }, [hydrate, fetchMe])

  return (
    <div className="flex min-h-screen bg-background">
      {/* 侧边栏 */}
      <aside className="sticky top-0 hidden h-screen w-56 flex-shrink-0 flex-col border-r border-border bg-card lg:flex">
        <a href="/" className="flex h-16 items-center gap-2 border-b border-border px-5">
          <span className="font-display text-lg font-bold tracking-tight">{site.name}</span>
          <span className="seal h-[1.15rem] w-[1.15rem] text-[0.55rem]">{sealChar}</span>
        </a>
        <nav className="flex-1 overflow-y-auto px-2 py-4">
          {NAV.map((item) => {
            const active = path === item.to || (item.to !== '/dashboard' && path.startsWith(item.to))
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
            {(() => {
              // 首帧即可渲染：本地同步计算 WeAvatar 地址
              const src =
                user?.avatar_url ||
                user?.avatar ||
                weavatarUrl(user?.email || user?.username || 'user')
              return (
                <img
                  src={src}
                  alt={user?.name || user?.username || 'avatar'}
                  className="h-7 w-7 flex-shrink-0 rounded-full object-cover"
                />
              )
            })()}
            <div className="min-w-0 flex-1">
              <p className="truncate text-sm text-foreground">{user?.name || user?.username}</p>
              {isAdminUser(user) && <p className="text-[0.65rem] text-muted-foreground">管理员</p>}
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
            <a href="/dashboard" className="font-display text-lg font-bold tracking-tight">{site.name}</a>
          </div>
          <div className="hidden items-center gap-3 lg:flex">
            <span className="eyebrow !tracking-[0.12em]">用户中心</span>
          </div>
          <div className="flex items-center gap-2">
            {isAdminUser(user) && (
              <a href="/admin" className="inline-flex items-center gap-1.5 rounded-md px-3 py-1.5 text-sm text-muted-foreground transition-colors hover:bg-muted hover:text-foreground">
                <ExternalLink className="h-3.5 w-3.5" />
                管理后台
              </a>
            )}
            <a href="/" className="inline-flex items-center gap-1.5 rounded-md px-3 py-1.5 text-sm text-muted-foreground transition-colors hover:bg-muted hover:text-foreground">
              前台
            </a>
            <span className="mx-1 h-5 w-px bg-border" />
            <span className="max-w-[8rem] truncate text-sm text-muted-foreground">
              {user?.name || user?.username || '…'}
            </span>
          </div>
        </header>

        <main className="min-w-0 flex-1 px-5 py-6 sm:px-8">{children}</main>
      </div>

      <Toaster />
    </div>
  )
}
