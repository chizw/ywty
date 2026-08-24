// 三方账号绑定
import { useEffect, useState } from 'react'
import { Link2, PlusCircle } from 'lucide-react'
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
}

export function OAuthPage() {
  const api = useApi()
  const [bound, setBound] = useState<OAuthItem[]>([])
  const [providers, setProviders] = useState<{ provider: string; name: string }[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    Promise.all([
      api.get<any>('/api/v1/oauth', { raw: true }),
      api.get<any>('/api/v1/oauth/providers', { raw: true }).catch(() => null),
    ])
      .then(([boundRes, providersRes]) => {
        const data = boundRes?.data?.data ?? boundRes?.data ?? []
        setBound(Array.isArray(data) ? data : [])
        const list = providersRes?.data?.data?.providers ?? []
        setProviders(Array.isArray(list) ? list : [])
      })
      .catch(() => setBound([]))
      .finally(() => setLoading(false))

    // 绑定成功后由后端重定向回来，提示结果
    const boundProvider = new URLSearchParams(window.location.search).get('bound')
    if (boundProvider) {
      toast.success(`已成功绑定 ${PROVIDER_LABEL[boundProvider] || boundProvider}`)
      window.history.replaceState({}, '', '/dashboard/oauth')
    }
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

  // 发起绑定：携带当前登录态请求授权地址，回调后绑定到本账号
  const startBind = async (provider: string) => {
    try {
      const r = await api.get<any>(`/api/v1/oauth/${provider}/authorize?mode=bind`, { raw: true })
      const url = r?.data?.data?.url ?? r?.data?.url
      if (!url) throw new Error('获取授权地址失败')
      window.location.href = url as string
    } catch (e: any) {
      toast.error(e?.message || '发起绑定失败')
    }
  }

  const isBound = (provider: string) => bound.some((b) => b.provider === provider)

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
                <p className="text-xs text-muted-foreground">已绑定</p>
              </div>
              <Button variant="outline" size="sm" onClick={() => unbind(o)}>解绑</Button>
            </div>
          ))}

          {providers.length > 0 && (
            <div className="space-y-2 pt-2">
              <p className="text-xs font-medium text-muted-foreground">可绑定</p>
              {providers
                .filter((p) => !isBound(p.provider))
                .map((p) => (
                  <div key={p.provider} className="flex items-center gap-3 rounded-md border border-dashed border-border p-4">
                    <div className="grid h-9 w-9 place-items-center rounded-md bg-muted">
                      <PlusCircle className="h-4 w-4 text-muted-foreground" />
                    </div>
                    <div className="flex-1">
                      <p className="text-sm font-medium">{PROVIDER_LABEL[p.provider] || p.name}</p>
                    </div>
                    <Button variant="outline" size="sm" onClick={() => startBind(p.provider)}>绑定</Button>
                  </div>
                ))}
            </div>
          )}

          {providers.length === 0 && (
            <p className="pt-2 text-xs text-muted-foreground">站点尚未配置第三方登录（需在服务端 config.yaml 配置 client_id/secret）。</p>
          )}
        </div>
      )}
    </AppShell>
  )
}
