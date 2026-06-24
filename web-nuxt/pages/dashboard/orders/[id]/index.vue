<script setup lang="ts">
// 订单详情 / 支付页：展示订单信息、发起支付、刷新状态、取消订单
definePageMeta({ layout: 'dashboard', middleware: 'auth' })

import { NCard, NTag, NButton, NSpace, NAlert, NDescriptions, NDescriptionsItem, NSpin } from 'naive-ui'

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

const statusMap: Record<string, { label: string; type: 'warning' | 'success' | 'default' | 'error' }> = {
  unpaid: { label: '待支付', type: 'warning' },
  paid: { label: '已支付', type: 'success' },
  canceled: { label: '已取消', type: 'default' },
  refunded: { label: '已退款', type: 'error' },
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
      <NuxtLink to="/dashboard/orders" class="text-xs text-gray-500 hover:text-primary-600">← 返回订单列表</NuxtLink>
      <h1 class="text-2xl font-bold text-gray-900 mt-1">订单详情</h1>
    </div>

    <NSpin :show="refreshing">
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-4">
        <!-- 左侧：订单信息 -->
        <div class="lg:col-span-2 space-y-4">
          <NCard title="订单信息">
            <NDescriptions :columns="1" label-placement="left" bordered>
              <NDescriptionsItem label="订单号">{{ order.trade_no }}</NDescriptionsItem>
              <NDescriptionsItem label="套餐名称">{{ planName() }}</NDescriptionsItem>
              <NDescriptionsItem label="订单状态">
                <NTag :type="statusMap[order.status]?.type || 'default'" size="small">
                  {{ statusMap[order.status]?.label || order.status }}
                </NTag>
              </NDescriptionsItem>
              <NDescriptionsItem label="支付方式">{{ order.pay_method || '-' }}</NDescriptionsItem>
              <NDescriptionsItem label="应付金额">
                <span class="text-red-500 font-semibold">{{ formatPrice(order.amount) }}</span>
              </NDescriptionsItem>
              <NDescriptionsItem v-if="order.deduct_amount > 0" label="优惠抵扣">
                <span class="text-green-600">{{ formatPrice(order.deduct_amount) }}</span>
              </NDescriptionsItem>
              <NDescriptionsItem label="创建时间">{{ fmtTime(order.created_at) }}</NDescriptionsItem>
              <NDescriptionsItem v-if="order.paid_at" label="支付时间">{{ fmtTime(order.paid_at) }}</NDescriptionsItem>
              <NDescriptionsItem v-if="order.canceled_at" label="取消时间">{{ fmtTime(order.canceled_at) }}</NDescriptionsItem>
            </NDescriptions>
          </NCard>

          <NCard v-if="order.snapshot?.plan?.features?.length" title="套餐权益">
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div
                v-for="feature in order.snapshot.plan.features"
                :key="feature"
                class="flex items-start gap-2 text-sm text-gray-700"
              >
                <span class="text-green-500 mt-0.5">✓</span>
                <span>{{ feature }}</span>
              </div>
            </div>
          </NCard>
        </div>

        <!-- 右侧：支付操作 -->
        <div class="space-y-4">
          <NCard v-if="order.status === 'unpaid'" title="支付">
            <NAlert type="info" :show-icon="false" class="mb-4 text-xs">
              点击「立即支付」后将根据支付方式跳转到对应收银台或展示二维码。
            </NAlert>

            <NSpace vertical class="w-full">
              <NButton type="primary" block size="large" :loading="paying" @click="pay">
                立即支付
              </NButton>
              <NButton block :loading="refreshing" @click="refreshOrder">
                我已支付
              </NButton>
              <NButton type="error" block :loading="canceling" @click="cancelOrder">
                取消订单
              </NButton>
            </NSpace>
          </NCard>

          <NCard v-else title="操作">
            <NSpace vertical class="w-full">
              <NButton block @click="refreshOrder" :loading="refreshing">刷新状态</NButton>
              <NuxtLink to="/dashboard/orders">
                <NButton block>返回订单列表</NButton>
              </NuxtLink>
            </NSpace>
          </NCard>

          <!-- 二维码 -->
          <NCard v-if="qrcodeUrl" title="请扫码支付">
            <div class="flex justify-center">
              <img :src="qrcodeUrl" alt="支付二维码" class="max-w-[240px] rounded border border-gray-200" />
            </div>
            <p class="text-center text-xs text-gray-500 mt-3">支付完成后请点击「我已支付」刷新状态</p>
          </NCard>

          <!-- App 参数 -->
          <NCard v-if="appParams" title="请复制以下参数完成支付">
            <textarea
              v-model="appParams"
              readonly
              rows="6"
              class="w-full px-3 py-2 border border-gray-300 rounded-md text-xs font-mono bg-gray-50"
            />
            <NButton class="mt-3" block @click="copyAppParams">复制参数</NButton>
          </NCard>
        </div>
      </div>
    </NSpin>

    <!-- 隐藏容器用于 form 类型自动提交 -->
    <div ref="formRef" class="hidden" />
  </div>
</template>
