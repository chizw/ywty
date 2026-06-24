<script setup lang="ts">
// 公开端底部
const year = new Date().getFullYear()
</script>

<template>
  <footer class="border-t border-gray-200 py-6 bg-white">
    <div class="max-w-7xl mx-auto px-4 flex flex-col sm:flex-row items-center justify-between gap-2 text-sm text-gray-500">
      <div>&copy; {{ year }} ywty — Refactor Edition</div>
      <div class="flex items-center gap-4">
        <NuxtLink to="/about" class="hover:text-primary-600">关于</NuxtLink>
        <NuxtLink to="/terms" class="hover:text-primary-600">条款</NuxtLink>
        <NuxtLink to="/privacy" class="hover:text-primary-600">隐私</NuxtLink>
      </div>
    </div>
  </footer>
</template>
