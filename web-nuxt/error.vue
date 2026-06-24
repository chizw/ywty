<script setup lang="ts">
import { Button } from '~/components/ui/button'

const props = defineProps<{ error: { statusCode: number; message: string } }>()
const handleError = () => clearError({ redirect: '/' })
</script>

<template>
  <div class="min-h-screen flex flex-col items-center justify-center bg-background px-4">
    <h1 class="text-6xl font-bold text-foreground">{{ props.error.statusCode }}</h1>
    <p class="mt-4 text-lg text-muted-foreground">{{ props.error.message }}</p>
    <Button class="mt-8" @click="handleError">
      返回首页
    </Button>
  </div>
</template>
