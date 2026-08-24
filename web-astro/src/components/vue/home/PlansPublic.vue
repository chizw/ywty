<script setup lang="ts">
// 套餐展示（静态模式客户端拉取）
import { onMounted, ref } from 'vue'
import Button from '../ui/Button.vue'
import { useApi } from '../../../lib/api'
import { isLoggedIn } from '../../../lib/vue-store'

interface PlanPrice {
  id: number
  name: string
  duration: number
  price: number
}
interface Plan {
  id: number
  type: string
  name: string
  intro?: string | null
  features?: string | null
  badge?: string | null
  is_up?: number
}
interface PlanDetail {
  plan: Plan
  prices: PlanPrice[]
}

const details = ref<PlanDetail[]>([])
const loading = ref(true)
const failed = ref(false)

onMounted(async () => {
  try {
    const res = await useApi().get<PlanDetail[]>('/api/v1/plans', { raw: true })
    const list = res?.data ?? []
    details.value = Array.isArray(list) ? list : []
  } catch {
    failed.value = true
  } finally {
    loading.value = false
  }
})

function ctaHref(id: number): string {
  // 登录态判断交给目标页/守卫，静态站无法在模板期得知
  return `/dashboard/plans/${id}`
}

function go(id: number) {
  window.location.assign(isLoggedIn() ? `/dashboard/plans/${id}` : `/auth/register?redirect=${encodeURIComponent(`/dashboard/plans/${id}`)}`)
}
</script>

<template>
  <section class="container-site py-12">
    <div class="mx-auto max-w-2xl text-center">
      <p class="eyebrow">Plans · 套餐</p>
      <h1 class="font-display mt-4 text-4xl font-bold tracking-tight">存储套餐</h1>
      <p class="mt-3 text-muted-foreground">自托管也从容，按需选一座够大的驿站。</p>
    </div>

    <div v-if="loading" class="mx-auto mt-12 grid max-w-5xl gap-6 md:grid-cols-2 lg:grid-cols-3">
      <div v-for="i in 3" :key="i" class="skeleton h-64 rounded-2xl" />
    </div>

    <div v-else-if="failed || details.length === 0"
      class="mx-auto mt-12 max-w-xl rounded-xl border border-dashed border-border py-24 text-center text-muted-foreground">
      {{ failed ? '套餐数据暂不可用' : '暂无上架套餐' }}
    </div>

    <div v-else class="mx-auto mt-12 grid max-w-5xl gap-6 md:grid-cols-2 lg:grid-cols-3">
      <div
        v-for="({ plan, prices }) in details"
        :key="plan.id"
        class="card-hover relative flex flex-col rounded-2xl border bg-card p-6 shadow-sm"
        :class="plan.badge === 'popular' ? 'border-brand/40 ring-2 ring-brand/20' : 'border-border'"
      >
        <span
          v-if="plan.badge === 'popular'"
          class="absolute -top-3 left-1/2 -translate-x-1/2 rounded-full bg-brand px-3 py-0.5 text-xs font-semibold text-brand-foreground"
        >推荐</span>

        <div class="flex items-center justify-between">
          <h3 class="text-xl font-bold text-foreground">{{ plan.name }}</h3>
        </div>
        <p class="mt-2 min-h-10 text-sm text-muted-foreground">{{ plan.intro || '为普通用户打造的轻量存储套餐' }}</p>

        <div class="mt-5 flex items-baseline gap-1">
          <span class="text-3xl font-extrabold text-foreground">
            {{ prices?.[0] ? `¥${(prices[0].price / 100).toFixed(2)}` : '联系了解' }}
          </span>
          <span class="text-sm text-muted-foreground">
            {{ prices?.[0] ? `/ ${prices[0].duration} 天` : '' }}
          </span>
        </div>

        <ul class="mt-6 flex-1 space-y-2.5 text-sm">
          <li v-for="(f, fi) in (plan.features || '').split('\n').filter(Boolean)" :key="fi"
            class="flex items-start gap-2 text-muted-foreground">
            <span class="mt-0.5 grid h-4 w-4 shrink-0 place-items-center rounded-full bg-brand/10 text-[10px] text-brand">✓</span>
            {{ f }}
          </li>
        </ul>

        <div class="mt-8 flex items-end">
          <Button class="w-full" @click="go(plan.id)">立即开通</Button>
        </div>
      </div>
    </div>
  </section>
</template>
