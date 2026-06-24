<script setup lang="ts">
// 工单列表（前端骨架，后端 P7 完成后对接）
definePageMeta({ layout: 'dashboard', middleware: 'auth' })

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
const { data, refresh } = await useAsyncData('my-tickets', () =>
  api.get<any>('/api/v1/tickets', { query: query.value }).catch(() => ({ data: [], meta: { total: 0 } }))
)

const tickets = computed<any[]>(() => {
  const d = data.value as any
  if (Array.isArray(d)) return d
  if (d && Array.isArray(d.data)) return d.data
  return []
})
const total = computed(() => (data.value as any)?.meta?.total ?? tickets.value.length)

const statusMap: Record<string, { label: string; cls: string }> = {
  pending: { label: '待处理', cls: 'bg-gray-100 text-gray-700' },
  processing: { label: '处理中', cls: 'bg-blue-100 text-blue-700' },
  resolved: { label: '已解决', cls: 'bg-emerald-100 text-emerald-700' },
  closed: { label: '已关闭', cls: 'bg-gray-200 text-gray-500' },
}
const priorityMap: Record<string, { label: string; cls: string }> = {
  low: { label: '低', cls: 'text-gray-500' },
  medium: { label: '中', cls: 'text-blue-600' },
  high: { label: '高', cls: 'text-orange-600' },
  urgent: { label: '紧急', cls: 'text-red-600' },
}
const typeMap: Record<string, string> = {
  bug: 'Bug 反馈',
  feature: '功能建议',
  complaint: '投诉',
  other: '其他',
}

watch([statusFilter], () => {
  page.value = 1
  refresh()
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
      <h1 class="text-2xl font-bold text-gray-900">我的工单</h1>
      <NuxtLink to="/dashboard/tickets/new" class="px-3 py-1.5 bg-primary-600 text-white text-sm rounded-md hover:bg-primary-700">新建工单</NuxtLink>
    </div>

    <div class="bg-white border border-gray-200 rounded-lg p-3 mb-4 flex items-center gap-3">
      <label class="text-sm text-gray-600">状态</label>
      <select v-model="statusFilter" class="px-2 py-1.5 border border-gray-300 rounded-md text-sm">
        <option value="">全部</option>
        <option v-for="(v, k) in statusMap" :key="k" :value="k">{{ v.label }}</option>
      </select>
      <span class="text-sm text-gray-500 ml-auto">共 {{ total }} 条</span>
    </div>

    <AppEmpty v-if="tickets.length === 0" title="还没有工单" description="遇到问题或有好想法？新建一个工单告诉我们">
      <NuxtLink to="/dashboard/tickets/new" class="px-3 py-1.5 bg-primary-600 text-white text-sm rounded-md hover:bg-primary-700">新建工单</NuxtLink>
    </AppEmpty>

    <div v-else class="bg-white border border-gray-200 rounded-lg divide-y">
      <NuxtLink
        v-for="t in tickets"
        :key="t.id"
        :to="`/dashboard/tickets/${t.id}`"
        class="block flex items-center justify-between p-4 hover:bg-gray-50"
      >
        <div class="min-w-0 flex-1">
          <div class="flex items-center gap-2">
            <span class="text-sm font-medium text-gray-900 truncate">{{ t.title }}</span>
            <span class="text-xs text-gray-400">#{{ t.id }}</span>
          </div>
          <div class="mt-1 flex items-center gap-2 text-xs text-gray-500">
            <span>{{ typeMap[t.type] || t.type }}</span>
            <span>·</span>
            <span :class="priorityMap[t.priority]?.cls">{{ priorityMap[t.priority]?.label || t.priority }}</span>
            <span>·</span>
            <span>{{ fmtTime(t.created_at) }}</span>
          </div>
        </div>
        <span
          class="px-2 py-0.5 text-xs rounded"
          :class="statusMap[t.status]?.cls || 'bg-gray-100 text-gray-700'"
        >{{ statusMap[t.status]?.label || t.status }}</span>
      </NuxtLink>
    </div>
  </div>
</template>
