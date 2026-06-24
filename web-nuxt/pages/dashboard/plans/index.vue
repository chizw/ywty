<script setup lang="ts">
// 套餐列表页：展示公开套餐，点击进入下单页
definePageMeta({ layout: 'dashboard', middleware: 'auth' })

const api = useApi()
const message = useMessage()

const rawData = ref<any>(null)

async function fetchPlans() {
  rawData.value = await api.get<any>('/api/v1/plans').catch((err: any) => {
    message.error(err?.statusMessage || '获取套餐列表失败')
    return { data: [] }
  })
}

const plans = computed<any[]>(() => {
  const d = rawData.value
  if (Array.isArray(d)) return d
  if (d && Array.isArray(d.data)) return d.data
  return []
})

onMounted(() => fetchPlans())

function formatPrice(price?: number) {
  if (typeof price !== 'number') return '-'
  return `¥${(price / 100).toFixed(2)}`
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-foreground mb-4">套餐列表</h1>

    <AppEmpty v-if="plans.length === 0" title="暂无可用套餐" description="当前没有上架的套餐，请稍后再�? />

    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <Card
        v-for="plan in plans"
        :key="plan.id"
        class="flex flex-col"
      >
        <CardHeader>
          <div class="flex items-center justify-between">
            <CardTitle>{{ plan.name }}</CardTitle>
            <Badge v-if="plan.badge" variant="success">{{ plan.badge }}</Badge>
          </div>
          <CardDescription v-if="plan.intro">{{ plan.intro }}</CardDescription>
        </CardHeader>
        <CardContent class="flex-1">
          <div v-if="plan.features?.length" class="space-y-2 mb-4">
            <div
              v-for="feature in plan.features"
              :key="feature"
              class="flex items-start gap-2 text-sm text-foreground"
            >
              <span class="text-green-500 mt-0.5">�?/span>
              <span>{{ feature }}</span>
            </div>
          </div>

          <div v-if="plan.prices?.length" class="space-y-2">
            <div
              v-for="price in plan.prices"
              :key="price.id"
              class="flex items-center justify-between text-sm"
            >
              <span class="text-muted-foreground">{{ price.name }}</span>
              <span class="font-semibold text-destructive">{{ formatPrice(price.price) }}</span>
            </div>
          </div>
        </CardContent>
        <CardFooter>
          <Button class="w-full" @click="navigateTo(`/dashboard/plans/${plan.id}`)">
            购买
          </Button>
        </CardFooter>
      </Card>
    </div>
  </div>
</template>
