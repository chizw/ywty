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

function statusVariant(s: string) {
  if (s === 'pending') return 'warning'
  if (s === 'processing') return 'default'
  if (s === 'resolved') return 'success'
  return 'secondary'
}

function levelVariant(l: string) {
  if (l === 'low') return 'secondary'
  if (l === 'medium') return 'default'
  if (l === 'high') return 'warning'
  if (l === 'urgent') return 'destructive'
  return 'secondary'
}

async function updateStatus(id: number, status: string) {
  await api.put(`/api/v1/admin/tickets/${id}`, { status })
  refresh()
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-foreground mb-4">工单管理</h1>

    <div v-if="tickets.length === 0" class="text-center py-12 text-muted-foreground">
      暂无工单
    </div>

    <Card v-else>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>编号</TableHead>
            <TableHead>标题</TableHead>
            <TableHead>类型</TableHead>
            <TableHead>优先级</TableHead>
            <TableHead>状态</TableHead>
            <TableHead>创建时间</TableHead>
            <TableHead>操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-for="t in tickets" :key="t.id">
            <TableCell class="font-medium">{{ t.issue_no }}</TableCell>
            <TableCell>{{ t.title }}</TableCell>
            <TableCell class="text-muted-foreground">{{ t.type }}</TableCell>
            <TableCell>
              <Badge :variant="levelVariant(t.level)">{{ levelMap[t.level] || t.level }}</Badge>
            </TableCell>
            <TableCell>
              <Badge :variant="statusVariant(t.status)">{{ statusMap[t.status] || t.status }}</Badge>
            </TableCell>
            <TableCell class="text-muted-foreground text-sm">{{ new Date(t.created_at).toLocaleString() }}</TableCell>
            <TableCell>
              <Select :modelValue="t.status" @update:modelValue="(val: string) => updateStatus(t.id, val)">
                <SelectTrigger class="w-[120px] h-8">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="pending">待处理</SelectItem>
                  <SelectItem value="processing">处理中</SelectItem>
                  <SelectItem value="resolved">已解决</SelectItem>
                  <SelectItem value="closed">已关闭</SelectItem>
                </SelectContent>
              </Select>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </Card>
  </div>
</template>
