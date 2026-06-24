<script setup lang="ts">
// 工单列表（前端骨架，后端 P7 完成后对接）
definePageMeta({ layout: 'dashboard', middleware: 'auth' })

import { Plus, MessageSquare } from '@lucide/vue'

const api = useApi()
const message = useMessage()

const statusFilter = ref<string>('')
const page = ref(1)
const perPage = 20

const query = computed(() => {
  const q: Record<string, any> = { page: page.value, per_page: perPage }
  if (statusFilter.value) q.status = statusFilter.value
  return q
})

// 后端未就绪时容错：返回空列表
const rawData = ref<any>(null)

async function fetchTickets() {
  rawData.value = await api.get<any>('/api/v1/tickets', { query: query.value }).catch(() => ({ data: [], meta: { total: 0 } }))
}

const tickets = computed<any[]>(() => {
  const d = rawData.value
  if (Array.isArray(d)) return d
  if (d && Array.isArray(d.data)) return d.data
  return []
})
const total = computed(() => (rawData.value as any)?.meta?.total ?? tickets.value.length)

onMounted(() => fetchTickets())

const statusMap: Record<string, { label: string; variant: 'secondary' | 'default' | 'success' | 'warning' }> = {
  pending: { label: '待处�?, variant: 'secondary' },
  processing: { label: '处理�?, variant: 'default' },
  resolved: { label: '已解�?, variant: 'success' },
  closed: { label: '已关�?, variant: 'secondary' },
}
const priorityMap: Record<string, { label: string; cls: string }> = {
  low: { label: '�?, cls: 'text-muted-foreground' },
  medium: { label: '�?, cls: 'text-blue-600' },
  high: { label: '�?, cls: 'text-orange-600' },
  urgent: { label: '紧�?, cls: 'text-destructive' },
}
const typeMap: Record<string, string> = {
  bug: 'Bug 反馈',
  feature: '功能建议',
  complaint: '投诉',
  other: '其他',
}

watch([statusFilter], () => {
  page.value = 1
  fetchTickets()
})

function fmtTime(t: any) {
  if (!t) return ''
  const ts = typeof t === 'number' ? t * 1000 : Date.parse(t)
  return isNaN(ts) ? '' : new Date(ts).toLocaleString()
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-4">
      <h1 class="text-2xl font-bold text-foreground">我的工单</h1>
      <NuxtLink to="/dashboard/tickets/new">
        <Button>
          <Plus class="mr-2 h-4 w-4" />
          新建工单
        </Button>
      </NuxtLink>
    </div>

    <Card class="mb-4">
      <CardContent class="p-3">
        <div class="flex items-center gap-3">
          <Label class="text-sm text-muted-foreground">状�?/Label>
          <Select v-model="statusFilter">
            <SelectTrigger class="w-[140px]">
              <SelectValue placeholder="全部" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="">全部</SelectItem>
              <SelectItem v-for="(v, k) in statusMap" :key="k" :value="k">{{ v.label }}</SelectItem>
            </SelectContent>
          </Select>
          <span class="text-sm text-muted-foreground ml-auto">�?{{ total }} �?/span>
        </div>
      </CardContent>
    </Card>

    <AppEmpty v-if="tickets.length === 0" title="还没有工�? description="遇到问题或有好想法？新建一个工单告诉我�?>
      <NuxtLink to="/dashboard/tickets/new">
        <Button>
          <Plus class="mr-2 h-4 w-4" />
          新建工单
        </Button>
      </NuxtLink>
    </AppEmpty>

    <Card v-else>
      <CardContent class="p-0 divide-y divide-border">
        <NuxtLink
          v-for="t in tickets"
          :key="t.id"
          :to="`/dashboard/tickets/${t.id}`"
          class="block flex items-center justify-between p-4 hover:bg-muted/50 transition-colors"
        >
          <div class="min-w-0 flex-1">
            <div class="flex items-center gap-2">
              <MessageSquare class="h-4 w-4 text-muted-foreground shrink-0" />
              <span class="text-sm font-medium text-foreground truncate">{{ t.title }}</span>
              <span class="text-xs text-muted-foreground">#{{ t.id }}</span>
            </div>
            <div class="mt-1 flex items-center gap-2 text-xs text-muted-foreground">
              <span>{{ typeMap[t.type] || t.type }}</span>
              <span>·</span>
              <span :class="priorityMap[t.priority]?.cls">{{ priorityMap[t.priority]?.label || t.priority }}</span>
              <span>·</span>
              <span>{{ fmtTime(t.created_at) }}</span>
            </div>
          </div>
          <Badge :variant="statusMap[t.status]?.variant || 'secondary'">
            {{ statusMap[t.status]?.label || t.status }}
          </Badge>
        </NuxtLink>
      </CardContent>
    </Card>
  </div>
</template>
