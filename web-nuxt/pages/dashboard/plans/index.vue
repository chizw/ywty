<script setup lang="ts">
// 套餐列表页：展示公开套餐，点击进入下单页
definePageMeta({ layout: 'dashboard', middleware: 'auth' })

const api = useApi()
const message = useMessage()

const { data } = await useAsyncData('dashboard-plans', () =>
  api.get<any>('/api/v1/plans').catch((err: any) => {
    message.error(err?.statusMessage || '获取套餐列表失败')
    return { data: [] }
  })
)

const plans = computed<any[]>(() => {
  const d = data.value as any
  if (Array.isArray(d)) return d
  if (d && Array.isArray(d.data)) return d.data
  return []
})

function formatPrice(price?: number) {
  if (typeof price !== 'number') return '-'
  return `¥${(price / 100).toFixed(2)}`
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-900 mb-4">套餐列表</h1>

    <AppEmpty v-if="plans.length === 0" title="暂无可用套餐" description="当前没有上架的套餐，请稍后再来" />

    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <NCard
        v-for="plan in plans"
        :key="plan.id"
        :title="plan.name"
        class="flex flex-col"
        content-style="flex: 1"
      >
        <template #header-extra>
          <NTag v-if="plan.badge" type="success" size="small">{{ plan.badge }}</NTag>
        </template>

        <p v-if="plan.intro" class="text-sm text-gray-500 mb-4">{{ plan.intro }}</p>

        <div v-if="plan.features?.length" class="space-y-2 mb-4">
          <div
            v-for="feature in plan.features"
            :key="feature"
            class="flex items-start gap-2 text-sm text-gray-700"
          >
            <span class="text-green-500 mt-0.5">✓</span>
            <span>{{ feature }}</span>
          </div>
        </div>

        <div v-if="plan.prices?.length" class="space-y-2">
          <div
            v-for="price in plan.prices"
            :key="price.id"
            class="flex items-center justify-between text-sm"
          >
            <span class="text-gray-600">{{ price.name }}</span>
            <span class="font-semibold text-red-500">{{ formatPrice(price.price) }}</span>
          </div>
        </div>

        <template #footer>
          <NButton type="primary" block @click="navigateTo(`/dashboard/plans/${plan.id}`)">
            购买
          </NButton>
        </template>
      </NCard>
    </div>
  </div>
</template>
