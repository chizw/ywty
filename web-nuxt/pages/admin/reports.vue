<script setup lang="ts">
// 管理后台：举报管理
definePageMeta({ layout: 'admin', middleware: 'admin' })

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

function fmtTime(s: any) {
  if (!s) return '-'
  const d = typeof s === 'number' ? new Date(s * 1000) : new Date(s)
  return d.toLocaleString()
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-4">
      <h1 class="text-2xl font-bold text-gray-900">举报管理</h1>
      <div class="flex gap-2">
        <select v-model="filterStatus" class="px-3 py-2 text-sm border border-gray-300 rounded-md" @change="() => refresh()">
          <option value="">全部</option>
          <option value="unhandled">未处理</option>
          <option value="handled">已处理</option>
          <option value="ignored">已忽略</option>
        </select>
      </div>
    </div>

    <p v-if="msg" class="text-sm mb-2 text-red-500">{{ msg }}</p>

    <div v-if="reports.length === 0" class="text-sm text-gray-500">暂无举报</div>
    <div v-else class="bg-white border border-gray-200 rounded-lg divide-y">
      <div v-for="r in reports" :key="r.id" class="p-4 flex items-start gap-4">
        <div class="flex-1 min-w-0">
          <div class="flex items-center gap-2 text-sm">
            <span class="font-medium">#{{ r.id }}</span>
            <span class="text-gray-500">·</span>
            <span class="text-gray-600">{{ r.reportable_type }} #{{ r.reportable_id }}</span>
            <span
              class="px-2 py-0.5 text-[10px] rounded-full"
              :class="r.status === 'unhandled' ? 'bg-yellow-100 text-yellow-700'
                : r.status === 'handled' ? 'bg-green-100 text-green-700'
                : 'bg-gray-100 text-gray-500'"
            >{{ r.status }}</span>
          </div>
          <div class="mt-1 text-sm text-gray-700">{{ r.content || '（无说明）' }}</div>
          <div class="mt-1 text-xs text-gray-400">
            举报人 #{{ r.report_user_id || '-' }} · {{ r.ip_address || '-' }} · {{ fmtTime(r.created_at) }}
            <span v-if="r.handled_at"> · 处理于 {{ fmtTime(r.handled_at) }}</span>
          </div>
        </div>
        <div v-if="r.status === 'unhandled'" class="flex flex-col gap-1">
          <button class="px-2 py-1 text-xs border border-green-300 text-green-700 rounded hover:bg-green-50" @click="updateStatus(r.id, 'handled')">已处理</button>
          <button class="px-2 py-1 text-xs border border-gray-300 text-gray-600 rounded hover:bg-gray-50" @click="updateStatus(r.id, 'ignored')">忽略</button>
        </div>
      </div>
    </div>
  </div>
</template>
