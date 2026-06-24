<script setup lang="ts">
// 工单详情 + 回复（前端骨架，后端 P7 完成后对接）
definePageMeta({ layout: 'dashboard', middleware: 'auth' })

const route = useRoute()
const ticketId = Number(route.params.id)
const api = useApi()
const message = useMessage()

// 后端未就绪时容错
const ticket = ref<any>(null)
const repliesRaw = ref<any>(null)

async function fetchTicket() {
  try { ticket.value = await api.get<any>(`/api/v1/tickets/${ticketId}`).catch(() => null) } catch { ticket.value = null }
}
async function fetchReplies() {
  try { repliesRaw.value = await api.get<any>(`/api/v1/tickets/${ticketId}/replies`).catch(() => ({ data: [] })) } catch { repliesRaw.value = { data: [] } }
}

const replyList = computed<any[]>(() => {
  const d = repliesRaw.value
  if (Array.isArray(d)) return d
  if (d && Array.isArray(d.data)) return d.data
  return []
})

const replyContent = ref('')
const replying = ref(false)
const notFound = computed(() => !ticket.value)

onMounted(() => { fetchTicket(); fetchReplies() })

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

async function submitReply() {
  if (!replyContent.value.trim()) {
    message.warning('请输入回复内容')
    return
  }
  replying.value = true
  try {
    await api.post(`/api/v1/tickets/${ticketId}/replies`, { content: replyContent.value })
    replyContent.value = ''
    message.success('回复成功')
    fetchReplies()
  } catch (err: any) {
    message.error(err?.statusMessage || '回复失败')
  } finally {
    replying.value = false
  }
}

async function closeTicket() {
  if (!confirm('确定关闭该工单？')) return
  try {
    await api.request(`/api/v1/tickets/${ticketId}`, { method: 'PATCH', body: { status: 'closed' } })
    message.success('工单已关闭')
    fetchTicket()
  } catch (err: any) {
    message.error(err?.statusMessage || '操作失败')
  }
}

function fmtTime(t: any) {
  if (!t) return ''
  const ts = typeof t === 'number' ? t * 1000 : Date.parse(t)
  return isNaN(ts) ? '' : new Date(ts).toLocaleString()
}
</script>

<template>
  <div>
    <div class="mb-4">
      <NuxtLink to="/dashboard/tickets" class="text-xs text-gray-500 hover:text-primary-600">← 返回工单列表</NuxtLink>
    </div>

    <AppEmpty v-if="notFound" title="工单不存在或尚未加载" description="该工单可能已被删除，或工单服务尚未上线">
      <NuxtLink to="/dashboard/tickets" class="px-3 py-1.5 bg-primary-600 text-white text-sm rounded-md hover:bg-primary-700">返回列表</NuxtLink>
    </AppEmpty>

    <template v-else>
      <div class="bg-white border border-gray-200 rounded-lg p-6 mb-6">
        <div class="flex items-start justify-between gap-4">
          <div class="min-w-0">
            <h1 class="text-xl font-bold text-gray-900">{{ ticket.title }}</h1>
            <div class="mt-2 flex flex-wrap items-center gap-2 text-xs text-gray-500">
              <span>#{{ ticket.id }}</span>
              <span>·</span>
              <span>{{ typeMap[ticket.type] || ticket.type }}</span>
              <span>·</span>
              <span :class="priorityMap[ticket.priority]?.cls">{{ priorityMap[ticket.priority]?.label || ticket.priority }}</span>
              <span>·</span>
              <span>{{ fmtTime(ticket.created_at) }}</span>
            </div>
          </div>
          <span
            class="px-2 py-1 text-xs rounded whitespace-nowrap"
            :class="statusMap[ticket.status]?.cls || 'bg-gray-100 text-gray-700'"
          >{{ statusMap[ticket.status]?.label || ticket.status }}</span>
        </div>
        <div class="mt-4 text-sm text-gray-700 whitespace-pre-wrap">{{ ticket.content }}</div>
        <div class="mt-4 flex gap-2" v-if="ticket.status !== 'closed'">
          <button class="px-3 py-1.5 text-sm border border-gray-300 rounded-md text-gray-600 hover:bg-gray-50" @click="closeTicket">关闭工单</button>
        </div>
      </div>

      <!-- 回复列表 -->
      <div class="mb-6">
        <h3 class="text-sm font-medium text-gray-700 mb-3">回复（{{ replyList.length }}）</h3>
        <div v-if="replyList.length === 0" class="text-sm text-gray-400 py-4">暂无回复</div>
        <div v-else class="space-y-3">
          <div v-for="r in replyList" :key="r.id" class="bg-white border border-gray-200 rounded-lg p-4">
            <div class="flex items-center justify-between text-xs text-gray-500 mb-2">
              <span>{{ r.user?.name || r.user?.username || '用户' }}{{ r.is_staff ? '（客服）' : '' }}</span>
              <span>{{ fmtTime(r.created_at) }}</span>
            </div>
            <div class="text-sm text-gray-700 whitespace-pre-wrap">{{ r.content }}</div>
          </div>
        </div>
      </div>

      <!-- 回复输入 -->
      <div class="bg-white border border-gray-200 rounded-lg p-4">
        <h3 class="text-sm font-medium text-gray-700 mb-2">添加回复</h3>
        <textarea v-model="replyContent" rows="4" placeholder="输入回复内容..." class="w-full px-3 py-2 border border-gray-300 rounded-md"></textarea>
        <div class="mt-2 flex justify-end">
          <AppButton :loading="replying" @click="submitReply">回复</AppButton>
        </div>
      </div>
    </template>
  </div>
</template>
