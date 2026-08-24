// 订单列表：订单号 / 金额 / 状态 / 时间，待支付订单可发起支付或取消
import { useEffect, useState } from 'react'
import { Loader2, ReceiptText, ShieldCheck } from 'lucide-react'
import { AppShell } from './AppShell'
import { PageHeader } from './PageHeader'
import { useConfirm } from './ConfirmDialog'
import { useApi } from '@/lib/api'
import { formatDate } from '@/lib/utils'
import { toast } from '@/lib/react-store'
import { Badge } from '../ui/badge'
import { Button } from '../ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '../ui/dialog'

interface Order {
  id: number
  trade_no: string
  order_type: string
  amount: number
  deduct_amount: number
  status: string
  product: string | null
  created_at: string
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
  tradeNo: string
  amount: number
  notify: MockNotifyRequest
}

const STATUS_LABEL: Record<string, string> = {
  unpaid: '待支付',
  pending: '待支付',
  paid: '已支付',
  canceled: '已取消',
  cancelled: '已取消',
  refunded: '已退款',
  completed: '已完成',
}

const STATUS_VARIANT: Record<string, 'default' | 'success' | 'warning' | 'destructive' | 'secondary'> = {
  unpaid: 'warning',
  pending: 'warning',
  paid: 'success',
  canceled: 'secondary',
  cancelled: 'secondary',
  refunded: 'secondary',
  completed: 'default',
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

export function OrdersPage() {
  const api = useApi()
  const [orders, setOrders] = useState<Order[]>([])
  const [loading, setLoading] = useState(true)
  const [payingId, setPayingId] = useState<number | null>(null)
  const [cashier, setCashier] = useState<CashierSession | null>(null)
  const [paying, setPaying] = useState(false)
  const [cancelingId, setCancelingId] = useState<number | null>(null)
  const { confirm, node } = useConfirm()

  const loadOrders = () => {
    api.get<any>('/api/v1/orders', { raw: true })
      .then((r) => {
        const data = r?.data?.data ?? r?.data ?? []
        setOrders(Array.isArray(data) ? data : [])
      })
      .catch(() => setOrders([]))
      .finally(() => setLoading(false))
  }

  // eslint-disable-next-line react-hooks/exhaustive-deps
  useEffect(loadOrders, [])

  /** 去支付：发起支付 → Mock 模式弹出模拟收银台，否则跳转收银台地址 */
  const startPay = async (order: Order) => {
    if (order.status !== 'unpaid') return
    setPayingId(order.id)
    try {
      const res: any = await api.post(`/api/v1/orders/${order.id}/pay`)
      const payment: PaymentParams | undefined = res?.data?.payment ?? res?.payment
      if (payment?.mock_notify) {
        setCashier({
          tradeNo: order.trade_no,
          amount: order.amount,
          notify: payment.mock_notify,
        })
      } else if (payment?.pay_url) {
        window.location.assign(payment.pay_url)
      } else {
        throw new Error('未获取到支付参数')
      }
    } catch (e: any) {
      toast.error(e?.message || '发起支付失败')
    } finally {
      setPayingId(null)
    }
  }

  /** 确认支付：回调成功后刷新列表为已订阅 */
  const payNow = async () => {
    if (!cashier) return
    setPaying(true)
    try {
      await confirmMockPay(cashier.notify)
      toast.success('支付成功')
      setCashier(null)
      setLoading(true)
      loadOrders()
    } catch (e: any) {
      toast.error(e?.message || '支付失败')
    } finally {
      setPaying(false)
    }
  }

  /** 取消待支付订单：确认后置为 canceled 并刷新列表 */
  const cancelOrder = async (order: Order) => {
    if (order.status !== 'unpaid') return
    const ok = await confirm({
      title: '取消订单',
      message: `确定取消订单 ${order.trade_no}？取消后不可恢复。`,
      okText: '取消订单',
      danger: true,
    })
    if (!ok) return
    setCancelingId(order.id)
    try {
      await api.post(`/api/v1/orders/${order.id}/cancel`)
      toast.success('订单已取消')
      setLoading(true)
      loadOrders()
    } catch (e: any) {
      toast.error(e?.message || '取消失败')
    } finally {
      setCancelingId(null)
    }
  }

  return (
    <AppShell>
      <PageHeader title="订单" description="查看你的套餐订单" />

      {loading ? (
        <div className="space-y-3">
          {Array.from({ length: 4 }).map((_, i) => (
            <div key={i} className="skeleton h-14 rounded-md" />
          ))}
        </div>
      ) : orders.length === 0 ? (
        <div className="rounded-md border border-dashed border-border py-20 text-center">
          <ReceiptText className="mx-auto mb-3 h-8 w-8 text-muted-foreground" />
          <p className="text-sm text-muted-foreground">还没有订单，去套餐页开通一个。</p>
        </div>
      ) : (
        <div className="overflow-hidden rounded-md border border-border">
          <table className="w-full text-sm">
            <thead className="border-b border-border bg-muted/50">
              <tr className="text-left text-xs text-muted-foreground">
                <th className="px-4 py-2.5 font-medium">订单号</th>
                <th className="px-4 py-2.5 font-medium">商品</th>
                <th className="px-4 py-2.5 font-medium">金额</th>
                <th className="px-4 py-2.5 font-medium">状态</th>
                <th className="px-4 py-2.5 font-medium">创建时间</th>
                <th className="px-4 py-2.5 font-medium">操作</th>
              </tr>
            </thead>
            <tbody>
              {orders.map((o) => (
                <tr key={o.id} className="border-b border-border last:border-0 hover:bg-muted/30">
                  <td className="px-4 py-3 font-mono text-xs">{o.trade_no}</td>
                  <td className="px-4 py-3">{o.product || o.order_type || '—'}</td>
                  <td className="px-4 py-3 tabular-nums">¥{(o.amount / 100).toFixed(2)}</td>
                  <td className="px-4 py-3">
                    <Badge variant={STATUS_VARIANT[o.status] || 'secondary'}>{STATUS_LABEL[o.status] || o.status}</Badge>
                  </td>
                  <td className="px-4 py-3 text-xs text-muted-foreground">{formatDate(o.created_at)}</td>
                  <td className="px-4 py-3">
                    {o.status === 'unpaid' ? (
                      <div className="flex items-center gap-2">
                        <Button size="sm" disabled={payingId !== null || cancelingId !== null} onClick={() => startPay(o)}>
                          {payingId === o.id && <Loader2 className="mr-1.5 h-3.5 w-3.5 animate-spin" />}
                          去支付
                        </Button>
                        <Button
                          size="sm"
                          variant="outline"
                          loading={cancelingId === o.id}
                          disabled={payingId !== null || (cancelingId !== null && cancelingId !== o.id)}
                          onClick={() => cancelOrder(o)}
                        >
                          取消
                        </Button>
                      </div>
                    ) : (
                      <span className="text-xs text-muted-foreground">—</span>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
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

      {node}
    </AppShell>
  )
}
