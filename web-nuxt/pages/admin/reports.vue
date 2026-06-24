<script setup lang="ts">
// 管理后台：举报管�?definePageMeta({ layout: 'admin', middleware: 'admin' })

import { Flag, Filter } from '@lucide/vue'

const api = useApi()
const filterStatus = ref<string>('')

const { data, refresh } = await useAsyncData('admin-reports', () =>
  api.get<any>('/api/v1/admin/reports', { query: { status: filterStatus.value } })
)
const reports = computed<any[]>(() => {
  const d = data.value as any
  return Array.isArray(d) ? d : (d as any)?.data ?? []
})

const msg = ref('')
async function updateStatus(id: number, status: 'handled' | 'ignored') {
  msg.value = ''
  try {
    await api.request(`/api/v1/admin/reports/${id}`, { method: 'PATCH', body: { status } })
    refresh()
  } catch (err: any) {
    msg.value = err?.statusMessage || '操作失败'
  }
}

function onFilterChange(val: string) {
  filterStatus.value = val
  refresh()
}

function fmtTime(s: any) {
  if (!s) return '-'
  const d = typeof s === 'number' ? new Date(s * 1000) : new Date(s)
  return d.toLocaleString()
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-4">
      <h1 class="text-2xl font-bold text-foreground">举报管理</h1>
      <div class="flex items-center gap-2">
        <Filter class="h-4 w-4 text-muted-foreground" />
        <Select :modelValue="filterStatus" @update:modelValue="onFilterChange">
          <SelectTrigger class="w-[140px]">
            <SelectValue placeholder="全部" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全部</SelectItem>
            <SelectItem value="unhandled">未处理</SelectItem>
            <SelectItem value="handled">已处理</SelectItem>
            <SelectItem value="ignored">已忽略</SelectItem>
          </SelectContent>
        </Select>
      </div>
    </div>

    <p v-if="msg" class="text-sm mb-2 text-destructive">{{ msg }}</p>

    <div v-if="reports.length === 0" class="text-sm text-muted-foreground">暂无举报</div>
    <div v-else class="space-y-3">
      <Card v-for="r in reports" :key="r.id">
        <CardContent class="p-4">
          <div class="flex items-start gap-4">
            <div class="flex-1 min-w-0">
              <div class="flex items-center gap-2 text-sm">
                <Flag class="h-4 w-4 text-muted-foreground" />
                <span class="font-medium">#{{ r.id }}</span>
                <span class="text-muted-foreground">·</span>
                <span class="text-muted-foreground">{{ r.reportable_type }} #{{ r.reportable_id }}</span>
                <Badge :variant="r.status === 'unhandled' ? 'warning' : r.status === 'handled' ? 'success' : 'secondary'">
                  {{ r.status }}
                </Badge>
              </div>
              <div class="mt-1 text-sm text-foreground">{{ r.content || '（无说明�? }}</div>
              <div class="mt-1 text-xs text-muted-foreground">
                举报人#{{ r.report_user_id || '-' }} · {{ r.ip_address || '-' }} · {{ fmtTime(r.created_at) }}
                <span v-if="r.handled_at"> · 处理于{{ fmtTime(r.handled_at) }}</span>
              </div>
            </div>
            <div v-if="r.status === 'unhandled'" class="flex flex-col gap-1 shrink-0">
              <Button variant="outline" size="sm" class="text-green-700 border-green-300 hover:bg-green-50" @click="updateStatus(r.id, 'handled')">已处理</Button>
              <Button variant="outline" size="sm" @click="updateStatus(r.id, 'ignored')">忽略</Button>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>
