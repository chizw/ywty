<script setup lang="ts">
// 管理后台：通知管理
definePageMeta({ layout: 'admin', middleware: 'admin' })

import { Plus, Trash2, Bell } from '@lucide/vue'

const api = useApi()
const { data, refresh } = await useAsyncData('admin-notices', () => api.get<any>('/api/v1/admin/notices'))

const notices = computed<any[]>(() => {
  const d = data.value
  return Array.isArray(d) ? d : ((d as any)?.data ?? [])
})

const showCreate = ref(false)
const form = reactive({
  title: '',
  content: '',
})
const loading = ref(false)
const msg = ref('')

async function create() {
  if (!form.title.trim()) {
    msg.value = '请输入标题
    return
  }
  loading.value = true
  msg.value = ''
  try {
    await api.post('/api/v1/admin/notices', form)
    msg.value = '创建成功'
    form.title = ''
    form.content = ''
    showCreate.value = false
    refresh()
  } catch (err: any) {
    msg.value = err?.statusMessage || '创建失败'
  } finally {
    loading.value = false
  }
}

const confirmId = ref<number | null>(null)
function askRemove(id: number) {
  confirmId.value = id
}
function closeConfirm() {
  confirmId.value = null
}
async function doRemove() {
  if (confirmId.value == null) return
  await api.del(`/api/v1/admin/notices/${confirmId.value}`)
  confirmId.value = null
  refresh()
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-4">
      <h1 class="text-2xl font-bold text-foreground">通知管理</h1>
      <Button @click="showCreate = !showCreate">
        <Plus v-if="!showCreate" class="h-4 w-4 mr-2" />
        {{ showCreate ? '取消' : '新建通知' }}
      </Button>
    </div>

    <Card v-if="showCreate" class="mb-6">
      <CardContent class="p-4 space-y-3">
        <div>
          <Label class="mb-1.5 block">标题</Label>
          <Input v-model="form.title" />
        </div>
        <div>
          <Label class="mb-1.5 block">内容</Label>
          <Textarea v-model="form.content" :rows="5" />
        </div>
        <Button :loading="loading" @click="create">创建</Button>
        <p v-if="msg" class="text-sm" :class="msg.includes('失败') ? 'text-destructive' : 'text-primary'">{{ msg }}</p>
      </CardContent>
    </Card>

    <div v-if="notices.length === 0" class="text-center py-12 text-muted-foreground">
      暂无通知
    </div>

    <div v-else class="space-y-3">
      <Card v-for="n in notices" :key="n.id">
        <CardContent class="p-4">
          <div class="flex items-start justify-between">
            <div class="flex-1">
              <div class="flex items-center gap-2 text-sm font-medium text-foreground">
                <Bell class="h-4 w-4 text-muted-foreground" />
                {{ n.title }}
              </div>
              <div class="mt-1 text-sm text-muted-foreground whitespace-pre-line">{{ n.content }}</div>
              <div class="mt-2 text-xs text-muted-foreground">{{ new Date(n.created_at).toLocaleString() }}</div>
            </div>
            <Button variant="ghost" size="sm" class="text-destructive hover:text-destructive" @click="askRemove(n.id)">
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
        <p class="text-sm text-muted-foreground">确定删除该通知？</p>
        <DialogFooter>
          <Button variant="outline" @click="closeConfirm">取消</Button>
          <Button variant="destructive" @click="doRemove">删除</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
