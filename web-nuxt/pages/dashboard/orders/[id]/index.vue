<script setup lang="ts">
// 订单详情 / 支付页：展示订单信息、发起支付、刷新状态、取消订单
definePageMeta({ layout: 'dashboard', middleware: 'auth' })

import { ArrowLeft, CreditCard, RefreshCw, XCircle, Copy, Loader2, CheckCircle, QrCode } from 'lucide-vue-next'

const api = useApi()
const message = useMessage()
const route = useRoute()
const router = useRouter()

const orderId = computed(() => Number(route.params.id))
const formRef = ref<HTMLDivElement | null>(null)
const qrcodeUrl = ref('')
const appParams = ref('')
const paying = ref(false)
const canceling = ref(false)
const refreshing = ref(false)

const rawData = ref<any>(null)

async function fetchOrder() {
  rawData.value = await api.get<any>(`/api/v1/orders/${orderId.value}`).catch((err: any) => {
    message.error(err?.statusMessage || '获取订单详情失败')
    return null
  })
}

const order = computed<any>(() => rawData.value || {})

onMounted(() => fetchOrder())

const statusMap: Record<string, { label: string; variant: 'warning' | 'success' | 'secondary' | 'destructive' }> = {
  unpaid: { label: '待支付', variant: 'warning' },
  paid: { label: '已支付', variant: 'success' },
  canceled: { label: '已取消', variant: 'secondary' },
  refunded: { label: '已退款', variant: 'destructive' },
}

function fmtTime(t: any) {
  if (!t) return '-'
  const ts = typeof t === 'number' ? t * 1000 : Date.parse(t)
  return isNaN(ts) ? '-' : new Date(ts).toLocaleString()
}

function formatPrice(price?: number) {
  if (typeof price !== 'number') return '-'
  return `¥${(price / 100).toFixed(2)}`
}

function planName() {
  return order.value.snapshot?.plan?.name || order.value.snapshot?.name || '套餐订单'
}

async function pay() {
  paying.value = true
  qrcodeUrl.value = ''
  appParams.value = ''
  try {
    const res = await api.post<any>(`/api/v1/orders/${orderId.value}/pay`, {})
    const { type, data: payload } = res || {}

    if (type === 'url') {
      window.location.href = payload
      return
    }

    if (type === 'form') {
      if (formRef.value) {
        formRef.value.innerHTML = payload
        await nextTick()
        const form = formRef.value.querySelector('form')
        if (form) form.submit()
        else message.error('支付表单渲染失败')
      }
      return
    }

    if (type === 'qrcode') {
      qrcodeUrl.value = payload
      return
    }

    if (type === 'app') {
      appParams.value = payload
      return
    }

    message.warning('未知支付类型')
  } catch (err: any) {
    message.error(err?.statusMessage || '发起支付失败')
  } finally {
    paying.value = false
  }
}

async function refreshOrder() {
  refreshing.value = true
  try {
    await fetchOrder()
    message.success('订单状态已刷新')
  } catch (err: any) {
    message.error(err?.statusMessage || '刷新失败')
  } finally {
    refreshing.value = false
  }
}

async function cancelOrder() {
  if (!confirm('确定取消该订单？')) return
  canceling.value = true
  try {
    await api.post(`/api/v1/orders/${orderId.value}/cancel`, {})
    message.success('订单已取消')
    await fetchOrder()
  } catch (err: any) {
    message.error(err?.statusMessage || '取消失败')
  } finally {
    canceling.value = false
  }
}

function copyAppParams() {
  if (!appParams.value) return
  navigator.clipboard.writeText(appParams.value).then(() => {
    message.success('已复制到剪贴板')
  }).catch(() => {
    message.error('复制失败，请手动复制')
  })
}
</script>

<template>
  <div>
    <div class="mb-4">
      <NuxtLink to="/dashboard/orders" class="text-xs text-muted-foreground hover:text-primary flex items-center gap-1">
        <ArrowLeft class="h-3 w-3" />
        返回订单列表
      </NuxtLink>
      <h1 class="text-2xl font-bold text-foreground mt-1">订单详情</h1>
    </div>

    <div v-if="refreshing" class="space-y-4">
      <Skeleton class="h-48 w-full" />
      <Skeleton class="h-32 w-full" />
    </div>
    <div v-else class="grid grid-cols-1 lg:grid-cols-3 gap-4">
      <!-- 左侧：订单信息 -->
      <div class="lg:col-span-2 space-y-4">
        <Card>
          <CardHeader>
            <CardTitle>订单信息</CardTitle>
          </CardHeader>
          <CardContent>
            <dl class="space-y-3 text-sm">
              <div class="flex justify-between">
                <dt class="text-muted-foreground">订单号</dt>
                <dd class="font-medium">{{ order.trade_no }}</dd>
              </div>
              <div class="flex justify-between">
                <dt class="text-muted-foreground">套餐名称</dt>
                <dd class="font-medium">{{ planName() }}</dd>
              </div>
              <div class="flex justify-between items-center">
                <dt class="text-muted-foreground">订单状态</dt>
                <dd>
                  <Badge :variant="statusMap[order.status]?.variant || 'secondary'">
                    {{ statusMap[order.status]?.label || order.status }}
                  </Badge>
                </dd>
              </div>
              <div class="flex justify-between">
                <dt class="text-muted-foreground">支付方式</dt>
                <dd class="font-medium">{{ order.pay_method || '-' }}</dd>
              </div>
              <div class="flex justify-between">
                <dt class="text-muted-foreground">应付金额</dt>
                <dd class="text-destructive font-semibold">{{ formatPrice(order.amount) }}</dd>
              </div>
              <div v-if="order.deduct_amount > 0" class="flex justify-between">
                <dt class="text-muted-foreground">优惠抵扣</dt>
                <dd class="text-green-600">{{ formatPrice(order.deduct_amount) }}</dd>
              </div>
              <div class="flex justify-between">
                <dt class="text-muted-foreground">创建时间</dt>
                <dd>{{ fmtTime(order.created_at) }}</dd>
              </div>
              <div v-if="order.paid_at" class="flex justify-between">
                <dt class="text-muted-foreground">支付时间</dt>
                <dd>{{ fmtTime(order.paid_at) }}</dd>
              </div>
              <div v-if="order.canceled_at" class="flex justify-between">
                <dt class="text-muted-foreground">取消时间</dt>
                <dd>{{ fmtTime(order.canceled_at) }}</dd>
              </div>
            </dl>
          </CardContent>
        </Card>

        <Card v-if="order.snapshot?.plan?.features?.length">
          <CardHeader>
            <CardTitle>套餐权益</CardTitle>
          </CardHeader>
          <CardContent>
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div
                v-for="feature in order.snapshot.plan.features"
                :key="feature"
                class="flex items-start gap-2 text-sm text-foreground"
              >
                <CheckCircle class="h-4 w-4 text-green-500 mt-0.5 shrink-0" />
                <span>{{ feature }}</span>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      <!-- 右侧：支付操作 -->
      <div class="space-y-4">
        <Card v-if="order.status === 'unpaid'">
          <CardHeader>
            <CardTitle>支付</CardTitle>
          </CardHeader>
          <CardContent class="space-y-3">
            <Alert variant="info">
              <AlertDescription class="text-xs">
                点击「立即支付」后将根据支付方式跳转到对应收银台或展示二维码。
              </AlertDescription>
            </Alert>

            <Button class="w-full" size="lg" :loading="paying" @click="pay">
              <CreditCard class="mr-2 h-4 w-4" />
              立即支付
            </Button>
            <Button variant="outline" class="w-full" :loading="refreshing" @click="refreshOrder">
              <RefreshCw class="mr-2 h-4 w-4" />
              我已支付
            </Button>
            <Button variant="destructive" class="w-full" :loading="canceling" @click="cancelOrder">
              <XCircle class="mr-2 h-4 w-4" />
              取消订单
            </Button>
          </CardContent>
        </Card>

        <Card v-else>
          <CardHeader>
            <CardTitle>操作</CardTitle>
          </CardHeader>
          <CardContent class="space-y-3">
            <Button variant="outline" class="w-full" :loading="refreshing" @click="refreshOrder">
              <RefreshCw class="mr-2 h-4 w-4" />
              刷新状态
            </Button>
            <NuxtLink to="/dashboard/orders" class="block">
              <Button variant="outline" class="w-full">返回订单列表</Button>
            </NuxtLink>
          </CardContent>
        </Card>

        <!-- 二维码 -->
        <Card v-if="qrcodeUrl">
          <CardHeader>
            <CardTitle class="flex items-center gap-2">
              <QrCode class="h-4 w-4" />
              请扫码支付
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div class="flex justify-center">
              <img :src="qrcodeUrl" alt="支付二维码" class="max-w-[240px] rounded border border-border" />
            </div>
            <p class="text-center text-xs text-muted-foreground mt-3">支付完成后请点击「我已支付」刷新状态</p>
          </CardContent>
        </Card>

        <!-- App 参数 -->
        <Card v-if="appParams">
          <CardHeader>
            <CardTitle>请复制以下参数完成支付</CardTitle>
          </CardHeader>
          <CardContent class="space-y-3">
            <Textarea
              v-model="appParams"
              readonly
              rows="6"
              class="text-xs font-mono bg-muted"
            />
            <Button variant="outline" class="w-full" @click="copyAppParams">
              <Copy class="mr-2 h-4 w-4" />
              复制参数
            </Button>
          </CardContent>
        </Card>
      </div>
    </div>

    <!-- 隐藏容器用于 form 类型自动提交 -->
    <div ref="formRef" class="hidden" />
  </div>
</template>
