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
      <h1 class="text-2xl font-bold text-foreground mt-1">{{ plan.name || '套餐详情' }}