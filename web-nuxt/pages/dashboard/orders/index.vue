<script setup lang="ts">
// 订单列表页：展示我的订单，支持去支付、取消和分页
definePageMeta({ layout: 'dashboard', middleware: 'auth' })

import { NCard, NTag, NButton, NSpace } from 'naive-ui'

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

const statusMap: Record<string, { label: string; type: 'warning' | 'success' | 'default' | 'error' }> = {
  unpaid: { label: '待支付', type: 'warning' },
  paid: { label: '已支付', type: 'success' },
  canceled: { label: '已取消', type: 'default' },
  refunded: { label: '已退款', type: 'error' },
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
    message.success('订单已取消')
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
      <h1 class="text-2xl font-bold text-gray-900">我的订单</h1>
      <span class="text-sm text-gray-500">共 {{ total }} 条</span>
    </div>

    <AppEmpty v-if="orders.length === 0" title="还没有订单" description="去套餐列表选购一个套餐吧">
      <NuxtLink to="/dashboard/plans">
        <NButton type="primary">浏览套餐</NButton>
      </NuxtLink>
    </AppEmpty>

    <div v-else class="space-y-3">
      <NCard v-for="order in orders" :key="order.id" class="relative">
        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
          <div class="min-w-0 flex-1 space-y-1">
            <div class="flex items-center gap-2 flex-wrap">
              <span class="font-medium text-gray-900">{{ planName(order) }}</span>
              <NTag :type="statusMap[order.status]?.type || 'default'" size="small">
                {{ statusMap[order.status]?.label || order.status }}
              </NTag>
            </div>
            <div class="text-xs text-gray-500 space-x-2">
              <span>订单号：{{ order.trade_no }}</span>
              <span>·</span>
              <span>创建时间：{{ fmtTime(order.created_at) }}</span>
            </div>
            <div class="text-sm font-semibold text-red-500">
              {{ formatPrice(order.amount) }}
              <span v-if="order.deduct_amount > 0" class="text-xs text-green-600 font-normal">
                （已抵扣 {{ formatPrice(order.deduct_amount) }}）
              </span>
            </div>
          </div>

          <NSpace>
            <NButton v-if="order.status === 'unpaid'" type="primary" size="small" @click="router.push(`/dashboard/orders/${order.id}`)">
              去支付
            </NButton>
            <NButton v-if="order.status === 'unpaid'" type="error" size="small" :loading="loading" @click="cancelOrder(order.id)">
              取消
            </NButton>
            <NButton v-else size="small" @click="router.push(`/dashboard/orders/${order.id}`)">
              详情
            </NButton>
          </NSpace>
        </div>
      </NCard>

      <!-- 分页 -->
      <div v-if="lastPage > 1" class="mt-6 flex items-center justify-center gap-3 text-sm">
        <button
          class="px-3 py-1.5 border border-gray-300 rounded-md disabled:opacity-40"
          :disabled="page <= 1"
          @click="goPrev"
        >上一页</button>
        <span class="text-gray-600">第 {{ page }} / {{ lastPage }} 页</span>
        <button
          class="px-3 py-1.5 border border-gray-300 rounded-md disabled:opacity-40"
          :disabled="page >= lastPage"
          @click="goNext"
        >下一页</button>
      </div>
    </div>
  </div>
</template>
