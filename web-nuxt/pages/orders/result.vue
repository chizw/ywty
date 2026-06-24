<script setup lang="ts">
// 支付结果页：展示支付成功 / 失败，并提供返回按钮
const route = useRoute()

const status = computed(() => (route.query.status as string) || 'unknown')
const tradeNo = computed(() => (route.query.trade_no as string) || '')

const isSuccess = computed(() => status.value === 'success')
</script>

<template>
  <div class="min-h-[calc(100vh-140px)] flex items-center justify-center p-4">
    <Card class="max-w-md w-full">
      <CardHeader class="items-center text-center">
        <div class="mx-auto mb-2">
          <Badge v-if="isSuccess" variant="success" class="text-base px-4 py-1">支付成功</Badge>
          <Badge v-else variant="destructive" class="text-base px-4 py-1">支付失败</Badge>
        </div>
        <CardTitle>{{ isSuccess ? '支付成功' : '支付失败' }}</CardTitle>
        <CardDescription>
          {{ isSuccess ? '感谢您的购买，订单已生效' : '支付未完成，请重新尝试或联系客服' }}
        </CardDescription>
      </CardHeader>
      <CardContent class="text-center">
        <p v-if="tradeNo" class="text-sm text-muted-foreground mb-4">
          订单号：{{ tradeNo }}
        </p>
      </CardContent>
      <CardFooter class="justify-center gap-3">
        <Button as="NuxtLink" to="/" variant="outline">返回首页</Button>
        <Button as="NuxtLink" to="/dashboard">用户中心</Button>
      </CardFooter>
    </Card>
  </div>
</template>
