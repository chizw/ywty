<script setup lang="ts">
// 鉴权页（登录 / 注册）布局：极简居中
</script>

<template>
  <div class="min-h-screen flex flex-col items-center justify-center bg-gradient-to-br from-gray-50 to-blue-50 px-4 py-12">
    <NuxtLink to="/" class="mb-8 text-2xl font-bold text-primary-600">ywty</NuxtLink>
    <div class="w-full max-w-md bg-white rounded-lg shadow-sm border border-gray-200 p-8">
      <slot />
    </div>
    <p class="mt-8 text-sm text-gray-500">
      &copy; {{ new Date().getFullYear() }} ywty
    </p>
  </div>
</template>
