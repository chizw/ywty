// 三方账号绑定
import { useEffect, useState } from 'react'
import { Link2 } from 'lucide-react'
import { AppShell } from './AppShell'
import { PageHeader } from './PageHeader'
import { useApi } from '@/lib/api'
import { toast } from '@/lib/react-store'
import { Button } from '../ui/button'

interface OAuthItem {
  id: number
  provider: string
  created_at?: string
}

const PROVIDER_LABEL: Record<string, string> = {
  github: 'GitHub',
  google: 'Google',
  gitee: 'Gitee',
  wechat: '微信',
  qq: 'QQ',
}

export function OAuthPage() {
  const api = useApi()
  const [bound, setBound] = useState<OAuthItem[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    api.get<any>('/api/v1/oauth', { raw: true })
      .then((r) => {
        const data = r?.data?.data ?? r?.data ?? []
        setBound(Array.isArray(data) ? data : [])
      })
      .catch(() => setBound([]))
      .finally(() => setLoading(false))
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const unbind = async (o: OAuthItem) => {
    try {
      await api.del(`/api/v1/oauth/${o.id}`)
      toast.success('已解绑')
      setBound((prev) => prev.filter((x) => x.id !== o.id))
    } catch (e: any) {
      toast.error(e?.message || '解绑失败')
    }
  }

  return (
    <AppShell>
      <PageHeader title="三方账号" description="绑定第三方账号，一键登录" />

      {loading ? (
        <div className="space-y-3">
          {Array.from({ length: 2 }).map((_, i) => (
            <div key={i} className="skeleton h-14 rounded-md" />
          ))}
        </div>
      ) : (
        <div className="max-w-md space-y-3">
          {bound.length === 0 && (
            <div className="rounded-md border border-dashed border-border py-14 text-center">
              <Link2 className="mx-auto mb-3 h-8 w-8 text-muted-foreground" />
              <p className="text-sm text-muted-foreground">暂未绑定任何三方账号。</p>
            </div>
          )}
          {bound.map((o) => (
            <div key={o.id} className="flex items-center gap-3 rounded-md border border-border bg-card p-4">
              <div className="grid h-9 w-9 place-items-center rounded-md bg-accent">
                <Link2 className="h-4 w-4 text-accent-foreground" />
              </div>
              <div className="flex-1">
                <p className="text-sm font-medium">{PROVIDER_LABEL[o.provider] || o.provider}</p>
              </div>
              <Button variant="outline" size="sm" onClick={() => unbind(o)}>解绑</Button>
            </div>
          ))}
          <p className="pt-2 text-xs text-muted-foreground">三方登录功能即将上线，当前版本为占位实现。</p>
        </div>
      )}
    </AppShell>
  )
}
