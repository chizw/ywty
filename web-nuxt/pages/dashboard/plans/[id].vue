<script setup lang="ts">
// 套餐详情 / 下单页：选择价格、优惠券、支付方式并下单
definePageMeta({ layout: 'dashboard', middleware: 'auth' })

import { NCard, NRadio, NRadioGroup, NInput, NInputGroup, NButton, NTag, NSpace, NAlert } from 'naive-ui'

const api = useApi()
const message = useMessage()
const route = useRoute()
const router = useRouter()

const planId = computed(() => Number(route.params.id))

const { data } = await useAsyncData(`dashboard-plan-${planId.value}`, () =>
  api.get<any>(`/api/v1/plans/${planId.value}`).catch((err: any) => {
    message.error(err?.statusMessage || '获取套餐详情失败')
    return null
  })
)

const plan = computed<any>(() => (data.value as any)?.plan || {})
const prices = computed<any[]>(() => (data.value as any)?.prices || [])
const capacities = computed<number[]>(() => (data.value as any)?.capacities || [])

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
      <NuxtLink to="/dashboard/plans" class="text-xs text-gray-500 hover:text-primary-600">← 返回套餐列表</NuxtLink>
      <h1 class="text-2xl font-bold text-gray-900 mt-1">{{ plan.name || '套餐详情' }}</h1>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-4">
      <!-- 左侧：套餐信息 -->
      <div class="lg:col-span-2 space-y-4">
        <NCard title="套餐信息">
          <p v-if="plan.intro" class="text-sm text-gray-500 mb-4">{{ plan.intro }}</p>

          <div v-if="plan.features?.length" class="grid grid-cols-1 sm:grid-cols-2 gap-3 mb-4">
            <div
              v-for="feature in plan.features"
              :key="feature"
              class="flex items-start gap-2 text-sm text-gray-700"
            >
              <span class="text-green-500 mt-0.5">✓</span>
              <span>{{ feature }}</span>
            </div>
          </div>

          <div v-if="capacities.length" class="flex flex-wrap gap-2">
            <span class="text-sm text-gray-500">包含容量：</span>
            <NTag v-for="(c, idx) in capacities" :key="idx" type="info" size="small">{{ formatCapacity(c) }}</NTag>
          </div>
        </NCard>

        <NCard title="选择价格方案">
          <NRadioGroup v-if="prices.length" v-model:value="selectedPriceId" class="w-full">
            <NSpace vertical class="w-full">
              <div
                v-for="price in prices"
                :key="price.id"
                class="flex items-center justify-between p-3 border border-gray-200 rounded-lg cursor-pointer hover:bg-gray-50"
                :class="{ 'border-primary-500 bg-primary-50': selectedPriceId === price.id }"
                @click="selectedPriceId = price.id"
              >
                <div class="flex items-center gap-3">
                  <NRadio :value="price.id" />
                  <div>
                    <div class="text-sm font-medium text-gray-900">{{ price.name }}</div>
                    <div class="text-xs text-gray-500">时长 {{ price.duration }} 分钟</div>
                  </div>
                </div>
                <div class="text-lg font-bold text-red-500">{{ formatPrice(price.price) }}</div>
              </div>
            </NSpace>
          </NRadioGroup>
          <AppEmpty v-else title="暂无价格方案" description="该套餐暂时没有可购买的价格方案" />
        </NCard>
      </div>

      <!-- 右侧：订单汇总 -->
      <div class="space-y-4">
        <NCard title="优惠券">
          <NInputGroup>
            <NInput v-model:value="couponCode" placeholder="输入券码" />
            <NButton type="primary" :disabled="!couponCode.trim() || validating" :loading="validating" @click="validateCoupon">
              校验
            </NButton>
          </NInputGroup>
          <div v-if="couponInfo" class="mt-3 text-sm text-green-600">
            抵扣金额：{{ formatPrice(discountAmount) }}
          </div>
        </NCard>

        <NCard title="订单汇总">
          <div class="space-y-3 text-sm">
            <div class="flex justify-between">
              <span class="text-gray-600">套餐</span>
              <span>{{ plan.name }}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-gray-600">方案</span>
              <span>{{ selectedPrice?.name || '-' }}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-gray-600">原价</span>
              <span>{{ formatPrice(selectedPrice?.price) }}</span>
            </div>
            <div v-if="discountAmount > 0" class="flex justify-between text-green-600">
              <span>优惠抵扣</span>
              <span>-{{ formatPrice(discountAmount) }}</span>
            </div>
            <div class="border-t border-gray-200 pt-3 flex justify-between items-center">
              <span class="font-medium">应付金额</span>
              <span class="text-xl font-bold text-red-500">{{ formatPrice(finalAmount) }}</span>
            </div>
          </div>
        </NCard>

        <NCard title="支付方式">
          <NRadioGroup v-model:value="selectedPayMethod" class="w-full">
            <NSpace vertical>
              <NRadio v-for="m in payMethods" :key="m.value" :value="m.value">{{ m.label }}</NRadio>
            </NSpace>
          </NRadioGroup>
        </NCard>

        <NAlert v-if="selectedPayMethod === 'log'" type="info" :show-icon="false" class="text-xs">
          日志支付方式仅用于测试，不会产生真实扣款。
        </NAlert>

        <NButton type="primary" size="large" block :loading="submitting" :disabled="!selectedPrice" @click="submitOrder">
          立即下单
        </NButton>
      </div>
    </div>
  </div>
</template>
