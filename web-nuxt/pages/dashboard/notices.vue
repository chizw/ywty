<script setup lang="ts">
// 用户中心：我的通知
definePageMeta({ layout: 'dashboard', middleware: 'auth' })

import { Bell, CheckCheck } from '@lucide/vue'

const api = useApi()

const rawData = ref<any>(null)
const loading = ref(false)

async function fetchNotices() {
  loading.value = true
  try {
    rawData.value = await api.get<any>('/api/v1/notices').catch(() => ({ data: [] }))
  } catch {
    rawData.value = { data: [] }
  } finally {
    loading.value = false
  }
}

const notices = computed<any[]>(() => {
  const d = rawData.value
  return Array.isArray(d) ? d : ((d as any)?.data ?? [])
})

onMounted(() => fetchNotices())

async function markAsRead(id: number) {
  await api.put(`/api/v1/notices/${id}/read`)
  fetchNotices()
}

async function markAllAsRead() {
  await api.put('/api/v1/notices/read-all')
  fetchNotices()
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-4">
      <h1 class="text-2xl font-bold text-foreground">我的通知</h1>
      <Button variant="outline" @click="markAllAsRead">
        <CheckCheck class="mr-2 h-4 w-4" />
        全部已读
      </Button>
    </div>

    <div v-if="notices.length === 0" class="text-center py-12 text-muted-foreground">
      暂无通知
    </div>

    <Card v-else>
      <CardContent class="p-0 divide-y divide-border">
        <div v-for="n in notices" :key="n.id" class="p-4 flex items-start gap-3" :class="{ 'bg-primary/5': !n.read_at }">
          <Bell class="h-4 w-4 mt-0.5 text-muted-foreground shrink-0" />
          <div class="flex-1">
            <div class="text-sm font-medium text-foreground">{{ n.title }}</div>
            <div class="mt-1 text-sm text-muted-foreground whitespace-pre-line">{{ n.content }}</div>
            <div class="mt-2 text-xs text-muted-foreground">{{ new Date(n.created_at).toLocaleString() }}</div>
          </div>
          <Button v-if="!n.read_at" variant="ghost" size="sm" class="text-primary" @click="markAsRead(n.id)">
            标记已读
          </Button>
        </div>
      </CardContent>
    </Card>
  </div>
</template>
