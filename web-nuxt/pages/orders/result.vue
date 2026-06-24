<script setup lang="ts">
// 支付结果页：展示支付成功 / 失败，并提供返回按钮
import { NCard, NButton, NResult } from 'naive-ui'

const route = useRoute()

const status = computed(() => (route.query.status as string) || 'unknown')
const tradeNo = computed(() => (route.query.trade_no as string) || '')

const isSuccess = computed(() => status.value === 'success')
</script>

<template>
  <div class="min-h-[calc(100vh-140px)] flex items-center justify-center p-4">
    <NCard class="max-w-md w-full">
      <NResult
        :status="isSuccess ? 'success' : 'error'"
        :title="isSuccess ? '支付成功' : '支付失败'"
        :description="isSuccess ? '感谢您的购买，订单已生效' : '支付未完成，请重新尝试或联系客服'"
      >
        <template v-if="tradeNo" #default>
          <div class="text-center text-sm text-gray-500 mb-4">
            订单号：{{ tradeNo }}
          </div>
        </template>
        <template #footer>
          <div class="flex justify-center gap-3">
            <NuxtLink to="/">
              <NButton>返回首页</NButton>
            </NuxtLink>
            <NuxtLink to="/dashboard">
              <NButton type="primary">用户中心</NButton>
            </NuxtLink>
          </div>
        </template>
      </NResult>
    </NCard>
  </div>
</template>
