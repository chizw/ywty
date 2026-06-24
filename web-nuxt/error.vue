<script setup lang="ts">
const props = defineProps<{ error: { statusCode: number; message: string } }>()
const handleError = () => clearError({ redirect: '/' })
</script>

<template>
  <div class="min-h-screen flex flex-col items-center justify-center bg-gray-50 px-4">
    <h1 class="text-6xl font-bold text-gray-900">{{ props.error.statusCode }}</h1>
    <p class="mt-4 text-lg text-gray-600">{{ props.error.message }}</p>
    <button
      class="mt-8 px-6 py-2 bg-primary-600 text-white rounded-md hover:bg-primary-700"
      @click="handleError"
    >
      返回首页
    </button>
  </div>
</template>
