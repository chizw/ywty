<script setup lang="ts">
// 公开端底部
const year = new Date().getFullYear()
</script>

<template>
  <footer class="border-t border-border py-6 bg-card">
    <div class="max-w-7xl mx-auto px-4 flex flex-col sm:flex-row items-center justify-between gap-2 text-sm text-muted-foreground">
      <div>&copy; {{ year }} 云雾图驿 — Refactor Edition</div>
      <div class="flex items-center gap-4">
        <NuxtLink to="/about" class="hover:text-foreground transition-colors">关于</NuxtLink>
        <NuxtLink to="/terms" class="hover:text-foreground transition-colors">条款</NuxtLink>
        <NuxtLink to="/privacy" class="hover:text-foreground transition-colors">隐私</NuxtLink>
      </div>
    </div>
  </footer>
</template>
