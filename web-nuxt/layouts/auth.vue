<script setup lang="ts">
// 鉴权页（登录 / 注册）布局：极简居中
</script>

<template>
  <div class="min-h-screen flex flex-col items-center justify-center bg-gradient-to-br from-background to-muted px-4 py-12">
    <NuxtLink to="/" class="mb-8 text-2xl font-bold text-primary">云雾图驿</NuxtLink>
    <Card class="w-full max-w-md">
      <CardContent class="p-8">
        <slot />
      </CardContent>
    </Card>
    <p class="mt-8 text-sm text-muted-foreground">
      &copy; {{ new Date().getFullYear() }} 云雾图驿
    </p>
  </div>
</template>
