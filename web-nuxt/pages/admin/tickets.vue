<script setup lang="ts">
// 管理后台：工单管理
definePageMeta({ layout: 'admin', middleware: 'admin' })

const api = useApi()
const { data, refresh } = await useAsyncData('admin-tickets', () => api.get<any>('/api/v1/admin/tickets'))

const tickets = computed<any[]>(() => {
  const d = data.value
  return Array.isArray(d) ? d : ((d as any)?.data ?? [])
})

const statusMap: Record<string, string> = {
  pending: '待处理',
  processing: '处理中',
  resolved: '已解决',
  closed: '已关闭',
}

const levelMap: Record<string, string> = {
  low: '低',
  medium: '中',
  high: '高',
  urgent: '紧急',
}

async function updateStatus(id: number, status: string) {
  await api.put(`/api/v1/admin/tickets/${id}`, { status })
  refresh()
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-900 mb-4">工单管理</h1>

    <div v-if="tickets.length === 0" class="text-center py-12 text-gray-500">
      暂无工单
    </div>

    <div v-else class="bg-white border border-gray-200 rounded-lg overflow-hidden">
      <table class="w-full">
        <thead class="bg-gray-50 border-b border-gray-200">
          <tr>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">编号</th>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">标题</th>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">类型</th>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">优先级</th>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">状态</th>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">创建时间</th>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">操作</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-200">
          <tr v-for="t in tickets" :key="t.id" class="hover:bg-gray-50">
            <td class="px-4 py-3 text-sm text-gray-900">{{ t.issue_no }}</td>
            <td class="px-4 py-3 text-sm text-gray-900">{{ t.title }}</td>
            <td class="px-4 py-3 text-sm text-gray-600">{{ t.type }}</td>
            <td class="px-4 py-3 text-sm">
              <span class="px-2 py-1 text-xs rounded" :class="{
                'bg-gray-100 text-gray-700': t.level === 'low',
                'bg-blue-100 text-blue-700': t.level === 'medium',
                'bg-yellow-100 text-yellow-700': t.level === 'high',
                'bg-red-100 text-red-700': t.level === 'urgent',
              }">{{ levelMap[t.level] || t.level }}</span>
            </td>
            <td class="px-4 py-3 text-sm">
              <span class="px-2 py-1 text-xs rounded" :class="{
                'bg-yellow-100 text-yellow-700': t.status === 'pending',
                'bg-blue-100 text-blue-700': t.status === 'processing',
                'bg-green-100 text-green-700': t.status === 'resolved',
                'bg-gray-100 text-gray-700': t.status === 'closed',
              }">{{ statusMap[t.status] || t.status }}</span>
            </td>
            <td class="px-4 py-3 text-sm text-gray-500">{{ new Date(t.created_at).toLocaleString() }}</td>
            <td class="px-4 py-3 text-sm">
              <select
                :value="t.status"
                @change="updateStatus(t.id, ($event.target as HTMLSelectElement).value)"
                class="text-xs border border-gray-300 rounded px-2 py-1"
              >
                <option value="pending">待处理</option>
                <option value="processing">处理中</option>
                <option value="resolved">已解决</option>
                <option value="closed">已关闭</option>
              </select>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
