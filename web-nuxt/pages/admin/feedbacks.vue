<script setup lang="ts">
// 管理后台：意见反馈管理
definePageMeta({ layout: 'admin', middleware: 'admin' })

import { Trash2 } from '@lucide/vue'

const api = useApi()
const { data, refresh } = await useAsyncData('admin-feedbacks', () => api.get<any>('/api/v1/admin/feedbacks'))

const feedbacks = computed<any[]>(() => {
  const d = data.value
  return Array.isArray(d) ? d : ((d as any)?.data ?? [])
})

const confirmId = ref<number | null>(null)
function askRemove(id: number) {
  confirmId.value = id
}
function closeConfirm() {
  confirmId.value = null
}
async function doRemove() {
  if (confirmId.value == null) return
  await api.del(`/api/v1/admin/feedbacks/${confirmId.value}`)
  confirmId.value = null
  refresh()
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-foreground mb-4">意见反馈</h1>

    <div v-if="feedbacks.length === 0" class="text-center py-12 text-muted-foreground">
      暂无反馈
    </div>

    <div v-else class="space-y-3">
      <Card v-for="f in feedbacks" :key="f.id">
        <CardContent class="p-4">
          <div class="flex items-start justify-between">
            <div class="flex-1">
              <div class="text-sm font-medium text-foreground">{{ f.type }}</div>
              <div class="mt-2 text-sm text-muted-foreground whitespace-pre-line">{{ f.content }}</div>
              <div class="mt-2 text-xs text-muted-foreground">
                {{ f.email || f.phone || '匿名用户' }} · {{ new Date(f.created_at).toLocaleString() }}
              </div>
            </div>
            <Button variant="ghost" size="sm" class="text-destructive hover:text-destructive" @click="askRemove(f.id)">
              <Trash2 class="h-4 w-4" />
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>

    <!-- 删除确认弹窗 -->
    <Dialog :open="confirmId != null" @update:open="(val: boolean) => { if (!val) closeConfirm() }">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>确认删除</DialogTitle>
        </DialogHeader>
        <p class="text-sm text-muted-foreground">确定删除该反馈？</p>
        <DialogFooter>
          <Button variant="outline" @click="closeConfirm">取消</Button>
          <Button variant="destructive" @click="doRemove">删除</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
