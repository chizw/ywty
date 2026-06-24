<script setup lang="ts">
// 公开端默认布局：顶部导航 + 内容 + 底部
</script>

<template>
  <div class="min-h-screen flex flex-col bg-background">
    <AppHeader />
    <main class="flex-1">
      <slot />
    </main>
    <AppFooter />
    <AppToaster />
  </div>
</template>
