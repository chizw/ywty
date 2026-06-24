<script setup lang="ts">
// 管理后台：违规记录管�?definePageMeta({ layout: 'admin', middleware: 'admin' })

import { ShieldAlert } from '@lucide/vue'

const api = useApi()
const { data, refresh } = await useAsyncData('admin-violations', () => api.get<any>('/api/v1/admin/violations'))

const violations = computed<any[]>(() => {
  const d = data.value
  return Array.isArray(d) ? d : ((d as any)?.data ?? [])
})

const statusMap: Record<string, string> = {
  pending: '待处�?,
  processed: '已处�?,
  ignored: '已忽�?,
}

function statusVariant(s: string) {
  if (s === 'pending') return 'warning'
  if (s === 'processed') return 'success'
  return 'secondary'
}

async function updateStatus(id: number, status: string) {
  await api.put(`/api/v1/admin/violations/${id}`, { status })
  refresh()
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-foreground mb-4">违规记录</h1>

    <div v-if="violations.length === 0" class="text-center py-12 text-muted-foreground">
      <ShieldAlert class="h-12 w-12 mx-auto mb-2 text-muted-foreground/50" />
      暂无违规记录
    </div>

    <Card v-else>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>资源类型</TableHead>
            <TableHead>资源 ID</TableHead>
            <TableHead>违规原因</TableHead>
            <TableHead>状�?/TableHead>
            <TableHead>创建时间</TableHead>
            <TableHead>操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-for="v in violations" :key="v.id">
            <TableCell class="font-medium">{{ v.resource_type }}</TableCell>
            <TableCell>{{ v.resource_id }}</TableCell>
            <TableCell class="text-muted-foreground">{{ v.reason }}</TableCell>
            <TableCell>
              <Badge :variant="statusVariant(v.status)">{{ statusMap[v.status] || v.status }}</Badge>
            </TableCell>
            <TableCell class="text-muted-foreground text-sm">{{ new Date(v.created_at).toLocaleString() }}</TableCell>
            <TableCell>
              <Select :modelValue="v.status" @update:modelValue="(val: string) => updateStatus(v.id, val)">
                <SelectTrigger class="w-[120px] h-8">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="pending">待处�?/SelectItem>
                  <SelectItem value="processed">已处�?/SelectItem>
                  <SelectItem value="ignored">已忽�?/SelectItem>
                </SelectContent>
              </Select>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </Card>
  </div>
</template>
