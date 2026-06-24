<script setup lang="ts">
// 套餐详情 / 下单页：选择价格、优惠券、支付方式并下单
definePageMeta({ layout: 'dashboard', middleware: 'auth' })

import { ArrowLeft, Check, Tag, ShoppingCart, Info } from 'lucide-vue-next'

const api = useApi()
const message = useMessage()
const route = useRoute()
const router = useRouter()

const planId = computed(() => Number(route.params.id))

const rawData = ref<any>(null)

async function fetchPlan() {
  rawData.value = await api.get<any>(`/api/v1/plans/${planId.value}`).catch((err: any) => {
    message.error(err?.statusMessage || '获取套餐详情失败')
    return null
  })
}

const plan = computed<any>(() => (rawData.value as any)?.plan || {})
const prices = computed<any[]>(() => (rawData.value as any)?.prices || [])
const capacities = computed<number[]>(() => (rawData.value as any)?.capacities || [])

onMounted(() => { fetchPlan() })

const selectedPriceId = ref<number | null>(null)
const couponCode = ref('')
const couponInfo = ref<any>(null)
const selectedPayMethod = ref<string>('log')
const submitting = ref(false)
const validating = ref(false)

const payMethods = [
  { value: 'log', label: '日志（测试）' },
  { value: 'alipay', label: '支付宝' },
  { value: 'wechat', label: '微信支付' },
  { value: 'paypal', label: 'PayPal' },
  { value: 'epay', label: '易支付' },
  { value: 'stripe', label: 'Stripe' },
]

onMounted(() => {
  const initialPriceId = route.query.price_id ? Number(route.query.price_id) : null
  if (prices.value.length) {
    selectedPriceId.value = initialPriceId || prices.value[0].id
  }
})

// 数据异步加载完成后，自动选中第一个价格方案
watch(prices, (list) => {
  if (list.length && selectedPriceId.value == null) {
    const initialPriceId = route.query.price_id ? Number(route.query.price_id) : null
    selectedPriceId.value = initialPriceId || list[0].id
  }
}, { immediate: true })

const selectedPrice = computed(() => prices.value.find((p) => p.id === selectedPriceId.value))
const discountAmount = computed(() => couponInfo.value?.discount || 0)
const finalAmount = computed(() => {
  if (!selectedPrice.value) return 0
  return Math.max(0, selectedPrice.value.price - discountAmount.value)
})

watch(selectedPriceId, () => {
  couponInfo.value = null
})

async function validateCoupon() {
  const code = couponCode.value.trim()
  if (!code) {
    message.warning('请输入优惠券码')
    return
  }
  if (!selectedPrice.value) {
    message.warning('请先选择价格方案')
    return
  }
  validating.value = true
  try {
    const res = await api.post<any>('/api/v1/coupons/validate', {
      code,
      amount: selectedPrice.value.price,
    })
    couponInfo.value = res
    message.success('优惠券有效')
  } catch (err: any) {
    couponInfo.value = null
    message.error(err?.statusMessage || '优惠券校验失败')
  } finally {
    validating.value = false
  }
}

async function submitOrder() {
  if (!selectedPrice.value) {
    message.warning('请选择价格方案')
    return
  }
  if (!selectedPayMethod.value) {
    message.warning('请选择支付方式')
    return
  }
  submitting.value = true
  try {
    const res = await api.post<any>('/api/v1/orders', {
      plan_id: planId.value,
      price_id: selectedPrice.value.id,
      coupon_code: couponCode.value.trim() || undefined,
      pay_method: selectedPayMethod.value,
    })
    const orderId = res?.id
    if (orderId) {
      router.push(`/dashboard/orders/${orderId}`)
    } else {
      message.error('下单失败，未返回订单号')
    }
  } catch (err: any) {
    message.error(err?.statusMessage || '下单失败')
  } finally {
    submitting.value = false
  }
}

function formatPrice(price?: number) {
  if (typeof price !== 'number') return '-'
  return `¥${(price / 100).toFixed(2)}`
}

function formatCapacity(kb?: number) {
  if (typeof kb !== 'number') return '-'
  if (kb >= 1024 * 1024) return `${(kb / 1024 / 1024).toFixed(2)} GB`
  if (kb >= 1024) return `${(kb / 1024).toFixed(2)} MB`
  return `${kb.toFixed(0)} KB`
}
</script>

<template>
  <div>
    <div class="mb-4">
      <NuxtLink to="/dashboard/plans" class="text-xs text-muted-foreground hover:text-primary flex items-center gap-1">
        <ArrowLeft class="h-3 w-3" />
        返回套餐列表
      </NuxtLink>
      <h1 class="text-2xl font-bold text-foreground mt-1">{{ plan.name || '套餐详情' }}</h1>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-4">
      <!-- 左侧：套餐信息 -->
      <div class="lg:col-span-2 space-y-4">
        <Card>
          <CardHeader>
            <CardTitle>套餐信息</CardTitle>
          </CardHeader>
          <CardContent>
            <p v-if="plan.intro" class="text-sm text-muted-foreground mb-4">{{ plan.intro }}</p>

            <div v-if="plan.features?.length" class="grid grid-cols-1 sm:grid-cols-2 gap-3 mb-4">
              <div
                v-for="feature in plan.features"
                :key="feature"
                class="flex items-start gap-2 text-sm text-foreground"
              >
                <Check class="h-4 w-4 text-green-500 mt-0.5 shrink-0" />
                <span>{{ feature }}</span>
              </div>
            </div>

            <div v-if="capacities.length" class="flex flex-wrap gap-2">
              <span class="text-sm text-muted-foreground">包含容量：</span>
              <Badge v-for="(c, idx) in capacities" :key="idx" variant="secondary">{{ formatCapacity(c) }}</Badge>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>选择价格方案</CardTitle>
          </CardHeader>
          <CardContent>
            <div v-if="prices.length" class="space-y-2">
              <div
                v-for="price in prices"
                :key="price.id"
                class="flex items-center justify-between p-3 border rounded-lg cursor-pointer transition-colors"
                :class="selectedPriceId === price.id ? 'border-primary bg-primary/5' : 'border-border hover:bg-muted/50'"
                @click="selectedPriceId = price.id"
              >
                <div class="flex items-center gap-3">
                  <div
                    class="h-4 w-4 rounded-full border-2 flex items-center justify-center"
                    :class="selectedPriceId === price.id ? 'border-primary' : 'border-muted-foreground'"
                  >
                    <div v-if="selectedPriceId === price.id" class="h-2 w-2 rounded-full bg-primary" />
                  </div>
                  <div>
                    <div class="text-sm font-medium text-foreground">{{ price.name }}</div>
                    <div class="text-xs text-muted-foreground">时长 {{ price.duration }} 分钟</div>
                  </div>
                </div>
                <div class="text-lg font-bold text-destructive">{{ formatPrice(price.price) }}</div>
              </div>
            </div>
            <AppEmpty v-else title="暂无价格方案" description="该套餐暂时没有可购买的价格方案" />
          </CardContent>
        </Card>
      </div>

      <!-- 右侧：订单汇总 -->
      <div class="space-y-4">
        <Card>
          <CardHeader>
            <CardTitle class="flex items-center gap-2">
              <Tag class="h-4 w-4" />
              优惠券
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div class="flex gap-2">
              <Input v-model="couponCode" placeholder="输入券码" class="flex-1" />
              <Button :disabled="!couponCode.trim() || validating" :loading="validating" @click="validateCoupon">
                校验
              </Button>
            </div>
            <div v-if="couponInfo" class="mt-3 text-sm text-green-600">
              抵扣金额：{{ formatPrice(discountAmount) }}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>订单汇总</CardTitle>
          </CardHeader>
          <CardContent>
            <div class="space-y-3 text-sm">
              <div class="flex justify-between">
                <span class="text-muted-foreground">套餐</span>
                <span>{{ plan.name }}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-muted-foreground">方案</span>
                <span>{{ selectedPrice?.name || '-' }}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-muted-foreground">原价</span>
                <span>{{ formatPrice(selectedPrice?.price) }}</span>
              </div>
              <div v-if="discountAmount > 0" class="flex justify-between text-green-600">
                <span>优惠抵扣</span>
                <span>-{{ formatPrice(discountAmount) }}</span>
              </div>
              <div class="border-t border-border pt-3 flex justify-between items-center">
                <span class="font-medium">应付金额</span>
                <span class="text-xl font-bold text-destructive">{{ formatPrice(finalAmount) }}</span>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>支付方式</CardTitle>
          </CardHeader>
          <CardContent>
            <div class="space-y-2">
              <div
                v-for="m in payMethods"
                :key="m.value"
                class="flex items-center gap-3 p-2 rounded cursor-pointer transition-colors"
                :class="selectedPayMethod === m.value ? 'bg-primary/5' : 'hover:bg-muted/50'"
                @click="selectedPayMethod = m.value"
              >
                <div
                  class="h-4 w-4 rounded-full border-2 flex items-center justify-center"
                  :class="selectedPayMethod === m.value ? 'border-primary' : 'border-muted-foreground'"
                >
                  <div v-if="selectedPayMethod === m.value" class="h-2 w-2 rounded-full bg-primary" />
                </div>
                <span class="text-sm">{{ m.label }}</span>
              </div>
            </div>
          </CardContent>
        </Card>

        <Alert v-if="selectedPayMethod === 'log'" variant="info">
          <Info class="h-4 w-4" />
          <AlertDescription class="text-xs">
            日志支付方式仅用于测试，不会产生真实扣款。
          </AlertDescription>
        </Alert>

        <Button size="lg" class="w-full" :loading="submitting" :disabled="!selectedPrice" @click="submitOrder">
          <ShoppingCart class="mr-2 h-4 w-4" />
          立即下单
        </Button>
      </div>
    </div>
  </div>
</template>
