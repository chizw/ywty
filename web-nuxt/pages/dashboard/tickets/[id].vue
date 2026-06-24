<script setup lang="ts">
// 工单详情 + 回复（前端骨架，后端 P7 完成后对接）
definePageMeta({ layout: 'dashboard', middleware: 'auth' })

import { ArrowLeft, Send, XCircle, MessageSquare, UserCircle } from 'lucide-vue-next'

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

const statusMap: Record<string, { label: string; variant: 'secondary' | 'default' | 'success' | 'warning' }> = {
  pending: { label: '待处理', variant: 'secondary' },
  processing: { label: '处理中', variant: 'default' },
  resolved: { label: '已解决', variant: 'success' },
  closed: { label: '已关闭', variant: 'secondary' },
}
const priorityMap: Record<string, { label: string; cls: string }> = {
  low: { label: '低', cls: 'text-muted-foreground' },
  medium: { label: '中', cls: 'text-blue-600' },
  high: { label: '高', cls: 'text-orange-600' },
  urgent: { label: '紧急', cls: 'text-destructive' },
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
      <NuxtLink to="/dashboard/tickets" class="text-xs text-muted-foreground hover:text-primary flex items-center gap-1">
        <ArrowLeft class="h-3 w-3" />
        返回工单列表
      </NuxtLink>
    </div>

    <AppEmpty v-if="notFound" title="工单不存在或尚未加载" description="该工单可能已被删除，或工单服务尚未上线">
      <NuxtLink to="/dashboard/tickets">
        <Button>返回列表</Button>
      </NuxtLink>
    </AppEmpty>

    <template v-else>
      <Card class="mb-6">
        <CardContent class="pt-6">
          <div class="flex items-start justify-between gap-4">
            <div class="min-w-0">
              <h1 class="text-xl font-bold text-foreground">{{ ticket.title }}</h1>
              <div class="mt-2 flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
                <span>#{{ ticket.id }}</span>
                <span>·</span>
                <span>{{ typeMap[ticket.type] || ticket.type }}</span>
                <span>·</span>
                <span :class="priorityMap[ticket.priority]?.cls">{{ priorityMap[ticket.priority]?.label || ticket.priority }}</span>
                <span>·</span>
                <span>{{ fmtTime(ticket.created_at) }}</span>
              </div>
            </div>
            <Badge :variant="statusMap[ticket.status]?.variant || 'secondary'">
              {{ statusMap[ticket.status]?.label || ticket.status }}
            </Badge>
          </div>
          <div class="mt-4 text-sm text-foreground whitespace-pre-wrap">{{ ticket.content }}</div>
          <div class="mt-4 flex gap-2" v-if="ticket.status !== 'closed'">
            <Button variant="outline" size="sm" @click="closeTicket">
              <XCircle class="mr-1 h-3 w-3" />
              关闭工单
            </Button>
          </div>
        </CardContent>
      </Card>

      <!-- 回复列表 -->
      <div class="mb-6">
        <h3 class="text-sm font-medium text-foreground mb-3">回复（{{ replyList.length }}）</h3>
        <div v-if="replyList.length === 0" class="text-sm text-muted-foreground py-4">暂无回复</div>
        <div v-else class="space-y-3">
          <Card v-for="r in replyList" :key="r.id">
            <CardContent class="pt-4">
              <div class="flex items-center justify-between text-xs text-muted-foreground mb-2">
                <div class="flex items-center gap-1.5">
                  <UserCircle class="h-3.5 w-3.5" />
                  <span>{{ r.user?.name || r.user?.username || '用户' }}{{ r.is_staff ? '（客服）' : '' }}</span>
                </div>
                <span>{{ fmtTime(r.created_at) }}</span>
              </div>
              <div class="text-sm text-foreground whitespace-pre-wrap">{{ r.content }}</div>
            </CardContent>
          </Card>
        </div>
      </div>

      <!-- 回复输入 -->
      <Card>
        <CardContent class="pt-6">
          <h3 class="text-sm font-medium text-foreground mb-2 flex items-center gap-1.5">
            <MessageSquare class="h-4 w-4" />
            添加回复
          </h3>
          <Textarea v-model="replyContent" rows="4" placeholder="输入回复内容..." class="mt-2" />
          <div class="mt-2 flex justify-end">
            <Button :loading="replying" @click="submitReply">
              <Send class="mr-1 h-4 w-4" />
              回复
            </Button>
          </div>
        </CardContent>
      </Card>
    </template>
  </div>
</template>
