<script setup lang="ts">
// 公开端顶部导航
const { user, isLoggedIn, logout } = useAuth()

const navLinks = [
  { to: '/explore', label: '探索' },
  { to: '/plans', label: '套餐' },
]
</script>

<template>
  <header class="border-b border-border bg-card sticky top-0 z-30">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-14 flex items-center justify-between">
      <div class="flex items-center gap-8">
        <NuxtLink to="/" class="text-lg font-semibold text-primary">云雾图驿</NuxtLink>
        <nav class="hidden md:flex items-center gap-6 text-sm text-muted-foreground">
          <NuxtLink
            v-for="link in navLinks"
            :key="link.to"
            :to="link.to"
            class="hover:text-foreground transition-colors"
            active-class="text-foreground font-medium"
          >
            {{ link.label }}
          </NuxtLink>
        </nav>
      </div>

      <div class="flex items-center gap-3">
        <template v-if="isLoggedIn">
          <Button as-child variant="ghost" size="sm">
            <NuxtLink to="/dashboard">
              {{ user?.name || user?.username }}
            </NuxtLink>
          </Button>
          <Button variant="ghost" size="sm" class="text-muted-foreground hover:text-destructive" @click="logout">
            退出          </Button>
        </template>
        <template v-else>
          <Button as-child variant="ghost" size="sm">
            <NuxtLink to="/auth/login">登录</NuxtLink>
          </Button>
          <Button as-child size="sm">
            <NuxtLink to="/auth/register">注册</NuxtLink>
          </Button>
        </template>
      </div>
    </div>
  </header>
</template>
