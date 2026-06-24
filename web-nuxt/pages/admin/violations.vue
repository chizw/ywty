<script setup lang="ts">
// 管理后台：违规记录管理
definePageMeta({ layout: 'admin', middleware: 'admin' })

const api = useApi()
const { data, refresh } = await useAsyncData('admin-violations', () => api.get<any>('/api/v1/admin/violations'))

const violations = computed<any[]>(() => {
  const d = data.value
  return Array.isArray(d) ? d : ((d as any)?.data ?? [])
})

const statusMap: Record<string, string> = {
  pending: '待处理',
  processed: '已处理',
  ignored: '已忽略',
}

async function updateStatus(id: number, status: string) {
  await api.put(`/api/v1/admin/violations/${id}`, { status })
  refresh()
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-900 mb-4">违规记录</h1>

    <div v-if="violations.length === 0" class="text-center py-12 text-gray-500">
      暂无违规记录
    </div>

    <div v-else class="bg-white border border-gray-200 rounded-lg overflow-hidden">
      <table class="w-full">
        <thead class="bg-gray-50 border-b border-gray-200">
          <tr>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">资源类型</th>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">资源 ID</th>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">违规原因</th>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">状态</th>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">创建时间</th>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">操作</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-200">
          <tr v-for="v in violations" :key="v.id" class="hover:bg-gray-50">
            <td class="px-4 py-3 text-sm text-gray-900">{{ v.resource_type }}</td>
            <td class="px-4 py-3 text-sm text-gray-900">{{ v.resource_id }}</td>
            <td class="px-4 py-3 text-sm text-gray-600">{{ v.reason }}</td>
            <td class="px-4 py-3 text-sm">
              <span class="px-2 py-1 text-xs rounded" :class="{
                'bg-yellow-100 text-yellow-700': v.status === 'pending',
                'bg-green-100 text-green-700': v.status === 'processed',
                'bg-gray-100 text-gray-700': v.status === 'ignored',
              }">{{ statusMap[v.status] || v.status }}</span>
            </td>
            <td class="px-4 py-3 text-sm text-gray-500">{{ new Date(v.created_at).toLocaleString() }}</td>
            <td class="px-4 py-3 text-sm">
              <select
                :value="v.status"
                @change="updateStatus(v.id, ($event.target as HTMLSelectElement).value)"
                class="text-xs border border-gray-300 rounded px-2 py-1"
              >
                <option value="pending">待处理</option>
                <option value="processed">已处理</option>
                <option value="ignored">已忽略</option>
              </select>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
