<script setup lang="ts">
// 公开端顶部导航
const { user, isLoggedIn, logout } = useAuth()

const navLinks = [
  { to: '/explore', label: '探索' },
  { to: '/plans', label: '套餐' },
]
</script>

<template>
  <header class="border-b border-gray-200 bg-white sticky top-0 z-30">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-14 flex items-center justify-between">
      <div class="flex items-center gap-8">
        <NuxtLink to="/" class="text-lg font-semibold text-primary-600">ywty</NuxtLink>
        <nav class="hidden md:flex items-center gap-6 text-sm text-gray-600">
          <NuxtLink
            v-for="link in navLinks"
            :key="link.to"
            :to="link.to"
            class="hover:text-primary-600"
            active-class="text-primary-600"
          >
            {{ link.label }}
          </NuxtLink>
        </nav>
      </div>

      <div class="flex items-center gap-3 text-sm">
        <template v-if="isLoggedIn">
          <NuxtLink to="/dashboard" class="text-gray-600 hover:text-primary-600">
            {{ user?.name || user?.username }}
          </NuxtLink>
          <button
            class="text-gray-500 hover:text-red-500"
            @click="logout"
          >
            退出
          </button>
        </template>
        <template v-else>
          <NuxtLink to="/auth/login" class="text-gray-600 hover:text-primary-600">登录</NuxtLink>
          <NuxtLink
            to="/auth/register"
            class="px-3 py-1.5 bg-primary-600 text-white rounded-md hover:bg-primary-700"
          >
            注册
          </NuxtLink>
        </template>
      </div>
    </div>
  </header>
</template>
