// 套餐列表：用户视角，可开通
import { useEffect, useState } from 'react'
import { Crown, Loader2, ShieldCheck } from 'lucide-react'
import { AppShell } from './AppShell'
import { PageHeader } from './PageHeader'
import { useApi } from '@/lib/api'
import { toast } from '@/lib/react-store'
import { Button } from '../ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '../ui/dialog'

interface Plan {
  id: number
  type: string
  name: string
  intro: string | null
  features: string | null
  badge: string
}

interface PlanPrice {
  id: number
  name: string
  duration: number
  price: number
}

interface Order {
  id: number
  trade_no: string
  amount: number
}

/** 发起支付返回的支付参数（Mock 模式附带预签名回调） */
interface MockNotifyRequest {
  method: string
  url: string
  headers: Record<string, string>
  body: string
}

interface PaymentParams {
  driver: string
  pay_url: string
  mock_notify?: MockNotifyRequest
}

/** 正在收银的订单 */
interface CashierSession {
  planName: string
  tradeNo: string
  amount: number
  notify: MockNotifyRequest
}

/** 原样发送服务端预签名的回调请求（签名不可伪造，前端只透传） */
async function confirmMockPay(notify: MockNotifyRequest) {
  const res = await fetch(notify.url, {
    method: notify.method || 'POST',
    headers: { 'Content-Type': 'application/json', ...(notify.headers || {}) },
    body: notify.body,
    credentials: 'include',
  })
  const json = await res.json().catch(() => null)
  if (!res.ok || !json || json.code !== 0) {
    throw new Error(json?.message || `支付失败（HTTP ${res.status}）`)
  }
}

export function PlansPage() {
  const api = useApi()
  const [plans, setPlans] = useState<Plan[]>([])
  const [prices, setPrices] = useState<Record<number, PlanPrice[]>>({})
  const [loading, setLoading] = useState(true)
  const [ordering, setOrdering] = useState<number | null>(null)
  const [cashier, setCashier] = useState<CashierSession | null>(null)
  const [paying, setPaying] = useState(false)

  useEffect(() => {
    api.get<any>('/api/v1/plans', { raw: true })
      .then((r) => {
        // 列表接口已内联价格档：[{plan, prices}]
        const list: any[] = Array.isArray(r?.data?.data)
          ? r.data.data
          : Array.isArray(r?.data)
            ? r.data
            : []
        setPlans(list.map((d) => d.plan))
        const map: Record<number, PlanPrice[]> = {}
        for (const d of list) {
          if (d?.plan?.id != null) map[d.plan.id] = Array.isArray(d.prices) ? d.prices : []
        }
        setPrices(map)
      })
      .catch(() => setPlans([]))
      .finally(() => setLoading(false))
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  /** 立即订阅：创建订单 → 发起支付 → 弹出模拟收银台 */
  const createOrder = async (plan: Plan) => {
    const price = prices[plan.id]?.[0]
    if (!price) return toast.warning('该套餐暂不可开通')
    setOrdering(plan.id)
    try {
      // 创建订单
      const created: any = await api.post('/api/v1/orders', { plan_id: plan.id, price_id: price.id })
      const order: Order | undefined = created?.data ?? created
      if (!order?.id) throw new Error('下单失败')

      // 发起支付
      const payRes: any = await api.post(`/api/v1/orders/${order.id}/pay`)
      const payment: PaymentParams | undefined = payRes?.data?.payment ?? payRes?.payment

      if (payment?.mock_notify) {
        // Mock 收银台：本地弹窗确认后发送预签名回调
        setCashier({
          planName: plan.name,
          tradeNo: order.trade_no,
          amount: order.amount ?? price.price,
          notify: payment.mock_notify,
        })
      } else if (payment?.pay_url) {
        // 真实驱动：跳转第三方收银台
        window.location.assign(payment.pay_url)
      } else {
        throw new Error('未获取到支付参数')
      }
    } catch (e: any) {
      toast.error(e?.message || '下单失败')
    } finally {
      setOrdering(null)
    }
  }

  /** 确认支付：回调成功后订单即为已订阅 */
  const payNow = async () => {
    if (!cashier) return
    setPaying(true)
    try {
      await confirmMockPay(cashier.notify)
      toast.success('订阅成功')
      setCashier(null)
      window.location.assign('/dashboard/orders')
    } catch (e: any) {
      toast.error(e?.message || '支付失败')
    } finally {
      setPaying(false)
    }
  }

  return (
    <AppShell>
      <PageHeader title="套餐" description="升级存储空间，按需选择" />

      {loading ? (
        <div className="grid gap-6 md:grid-cols-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className="skeleton h-56 rounded-md" />
          ))}
        </div>
      ) : plans.length === 0 ? (
        <div className="rounded-md border border-dashed border-border py-20 text-center">
          <p className="text-sm text-muted-foreground">暂无可用套餐。</p>
        </div>
      ) : (
        <div className="grid gap-6 md:grid-cols-3">
          {plans.map((p) => {
            const price = prices[p.id]?.[0]
            const features = (p.features || '').split('\n').filter(Boolean)
            return (
              <div key={p.id} className={`relative flex flex-col rounded-md border bg-card p-6 ${p.badge === 'popular' ? 'border-brand/40 ring-2 ring-brand/15' : 'border-border'}`}>
                {p.badge === 'popular' && (
                  <span className="absolute -top-2.5 left-1/2 -translate-x-1/2 rounded-full bg-brand px-3 py-0.5 text-xs text-brand-foreground">推荐</span>
                )}
                <div className="flex items-center gap-2">
                  <Crown className="h-4 w-4 text-brand" />
                  <h3 className="font-display text-xl font-bold">{p.name}</h3>
                </div>
                <p className="mt-2 min-h-10 text-sm text-muted-foreground">{p.intro || ''}</p>
                <div className="mt-4 flex items-baseline gap-1">
                  <span className="font-display text-3xl font-bold">{price ? `¥${(price.price / 100).toFixed(2)}` : '—'}</span>
                  {price && <span className="text-sm text-muted-foreground">/ {price.duration} 天</span>}
                </div>
                <ul className="mt-5 flex-1 space-y-2 text-sm text-muted-foreground">
                  {features.map((f, i) => (
                    <li key={i} className="flex items-start gap-2">
                      <span className="mt-0.5 text-brand">·</span>
                      {f}
                    </li>
                  ))}
                </ul>
                <Button className="mt-6 w-full" disabled={ordering !== null} onClick={() => createOrder(p)}>
                  {ordering === p.id && <Loader2 className="mr-1.5 h-4 w-4 animate-spin" />}
                  立即订阅
                </Button>
              </div>
            )
          })}
        </div>
      )}

      {/* 模拟收银台（仅 Mock 支付模式） */}
      <Dialog open={cashier !== null} onOpenChange={(open) => !open && !paying && setCashier(null)}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>确认付款</DialogTitle>
            <DialogDescription>当前为模拟支付环境，不产生真实扣款。</DialogDescription>
          </DialogHeader>
          {cashier && (
            <div className="space-y-3 rounded-md border border-border bg-muted/30 p-4">
              <div className="flex items-center justify-between text-sm">
                <span className="text-muted-foreground">商品</span>
                <span>{cashier.planName}</span>
              </div>
              <div className="flex items-center justify-between text-sm">
                <span className="text-muted-foreground">订单号</span>
                <span className="font-mono text-xs">{cashier.tradeNo}</span>
              </div>
              <div className="flex items-baseline justify-between">
                <span className="text-sm text-muted-foreground">应付金额</span>
                <span className="font-display text-2xl font-bold">¥{(cashier.amount / 100).toFixed(2)}</span>
              </div>
            </div>
          )}
          <DialogFooter>
            <Button variant="outline" disabled={paying} onClick={() => setCashier(null)}>取消</Button>
            <Button disabled={paying} onClick={payNow}>
              {paying ? <Loader2 className="mr-1.5 h-4 w-4 animate-spin" /> : <ShieldCheck className="mr-1.5 h-4 w-4" />}
              确认支付
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </AppShell>
  )
}
