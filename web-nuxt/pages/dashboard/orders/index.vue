<script setup lang="ts">
// 订单列表页：展示我的订单，支持去支付、取消和分页
definePageMeta({ layout: 'dashboard', middleware: 'auth' })

import { CreditCard } from '@lucide/vue'

const api = useApi()
const message = useMessage()
const router = useRouter()

const page = ref(1)
const perPage = 20
const loading = ref(false)

const rawData = ref<any>(null)

async function fetchOrders() {
  rawData.value = await api.get<any>('/api/v1/orders', { query: { page: page.value, per_page: perPage } }).catch((err: any) => {
    message.error(err?.statusMessage || '获取订单列表失败')
    return { data: [], meta: { total: 0, last_page: 1 } }
  })
}

const orders = computed<any[]>(() => {
  const d = rawData.value
  if (Array.isArray(d)) return d
  if (d && Array.isArray(d.data)) return d.data
  return []
})
const total = computed(() => (rawData.value as any)?.meta?.total ?? orders.value.length)
const lastPage = computed(() => (rawData.value as any)?.meta?.last_page ?? Math.max(1, Math.ceil(total.value / perPage)))

onMounted(() => fetchOrders())

const statusMap: Record<string, { label: string; variant: 'warning' | 'success' | 'secondary' | 'destructive' }> = {
  unpaid: { label: '待支�?, variant: 'warning' },
  paid: { label: '已支�?, variant: 'success' },
  canceled: { label: '已取�?, variant: 'secondary' },
  refunded: { label: '已退�?, variant: 'destructive' },
}

watch(page, () => fetchOrders())

function fmtTime(t: any) {
  if (!t) return '-'
  const ts = typeof t === 'number' ? t * 1000 : Date.parse(t)
  return isNaN(ts) ? '-' : new Date(ts).toLocaleString()
}

function formatPrice(price?: number) {
  if (typeof price !== 'number') return '-'
  return `¥${(price / 100).toFixed(2)}`
}

function planName(order: any) {
  return order.snapshot?.plan?.name || order.snapshot?.name || '套餐订单'
}

async function cancelOrder(id: number) {
  if (!confirm('确定取消该订单？')) return
  loading.value = true
  try {
    await api.post(`/api/v1/orders/${id}/cancel`, {})
    message.success('订单已取�?)
    fetchOrders()
  } catch (err: any) {
    message.error(err?.statusMessage || '取消失败')
  } finally {
    loading.value = false
  }
}

function goPrev() {
  if (page.value > 1) page.value--
}
function goNext() {
  if (page.value < lastPage.value) page.value++
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-4">
      <h1 class="text-2xl font-bold text-foreground">我的订单</h1>
      <span class="text-sm text-muted-foreground">�?{{ total }} �?/span>
    </div>

    <AppEmpty v-if="orders.length === 0" title="还没有订�? description="去套餐列表选购一个套餐吧">
      <NuxtLink to="/dashboard/plans">
        <Button>浏览套餐</Button>
      </NuxtLink>
    </AppEmpty>

    <div v-else class="space-y-3">
      <Card v-for="order in orders" :key="order.id">
        <CardContent class="p-4">
          <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
            <div class="min-w-0 flex-1 space-y-1">
              <div class="flex items-center gap-2 flex-wrap">
                <span class="font-medium text-foreground">{{ planName(order) }}</span>
                <Badge :variant="statusMap[order.status]?.variant || 'secondary'">
                  {{ statusMap[order.status]?.label || order.status }}
                </Badge>
              </div>
              <div class="text-xs text-muted-foreground space-x-2">
                <span>订单号：{{ order.trade_no }}</span>
                <span>·</span>
                <span>创建时间：{{ fmtTime(order.created_at) }}</span>
              </div>
              <div class="text-sm font-semibold text-destructive">
                {{ formatPrice(order.amount) }}
                <span v-if="order.deduct_amount > 0" class="text-xs text-green-600 font-normal">
                  （已抵扣 {{ formatPrice(order.deduct_amount) }}�?                </span>
              </div>
            </div>

            <div class="flex gap-2">
              <Button v-if="order.status === 'unpaid'" size="sm" @click="router.push(`/dashboard/orders/${order.id}`)">
                <CreditCard class="mr-1 h-3 w-3" />
                去支�?              </Button>
              <Button v-if="order.status === 'unpaid'" variant="destructive" size="sm" :loading="loading" @click="cancelOrder(order.id)">
                取消
              </Button>
              <Button v-else variant="outline" size="sm" @click="router.push(`/dashboard/orders/${order.id}`)">
                详情
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      <!-- 分页 -->
      <div v-if="lastPage > 1" class="mt-6 flex items-center justify-center gap-3 text-sm">
        <Button variant="outline" size="sm" :disabled="page <= 1" @click="goPrev">上一�?/Button>
        <span class="text-muted-foreground">�?{{ page }} / {{ lastPage }} �?/span>
        <Button variant="outline" size="sm" :disabled="page >= lastPage" @click="goNext">下一�?/Button>
      </div>
    </div>
  </div>
</template>
