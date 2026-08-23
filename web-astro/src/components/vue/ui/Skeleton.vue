<script setup lang="ts">
import { cn } from '../../../lib/utils'

const props = withDefaults(defineProps<{ class?: string }>(), {})
</script>

<template>
  <div :class="cn('animate-pulse rounded-md bg-muted', props.class)" />
</template>
