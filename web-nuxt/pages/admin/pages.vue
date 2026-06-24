<script setup lang="ts">
// 管理后台：单页管�?definePageMeta({ layout: 'admin', middleware: 'admin' })

import { Plus, Trash2, FileText, ExternalLink } from '@lucide/vue'

const api = useApi()
const { data, refresh } = await useAsyncData('admin-pages', () => api.get<any>('/api/v1/admin/pages'))

const pages = computed<any[]>(() => {
  const d = data.value
  return Array.isArray(d) ? d : ((d as any)?.data ?? [])
})

const showCreate = ref(false)
const form = reactive({
  title: '',
  slug: '',
  content: '',
})
const loading = ref(false)
const msg = ref('')

async function create() {
  if (!form.title.trim() || !form.slug.trim()) {
    msg.value = '请输入标题和别名'
    return
  }
  loading.value = true
  msg.value = ''
  try {
    await api.post('/api/v1/admin/pages', form)
    msg.value = '创建成功'
    form.title = ''
    form.slug = ''
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
  await api.del(`/api/v1/admin/pages/${confirmId.value}`)
  confirmId.value = null
  refresh()
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-4">
      <h1 class="text-2xl font-bold text-foreground">单页管理</h1>
      <Button @click="showCreate = !showCreate">
        <Plus v-if="!showCreate" class="h-4 w-4 mr-2" />
        {{ showCreate ? '取消' : '新建单页' }}
      </Button>
    </div>

    <Card v-if="showCreate" class="mb-6">
      <CardContent class="p-4 space-y-3">
        <div>
          <Label class="mb-1.5 block">标题</Label>
          <Input v-model="form.title" />
        </div>
        <div>
          <Label class="mb-1.5 block">别名（URL 中使用）</Label>
          <Input v-model="form.slug" placeholder="�?about" />
        </div>
        <div>
          <Label class="mb-1.5 block">内容</Label>
          <Textarea v-model="form.content" :rows="8" />
        </div>
        <Button :loading="loading" @click="create">创建</Button>
        <p v-if="msg" class="text-sm" :class="msg.includes('失败') ? 'text-destructive' : 'text-primary'">{{ msg }}</p>
      </CardContent>
    </Card>

    <div v-if="pages.length === 0" class="text-center py-12 text-muted-foreground">
      暂无单页
    </div>

    <div v-else class="space-y-3">
      <Card v-for="p in pages" :key="p.id">
        <CardContent class="p-4">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-3 flex-1 min-w-0">
              <FileText class="h-4 w-4 text-muted-foreground shrink-0" />
              <div class="min-w-0">
                <div class="text-sm font-medium text-foreground">{{ p.title }}</div>
                <div class="text-xs text-muted-foreground">/page/{{ p.slug }}</div>
              </div>
            </div>
            <div class="flex gap-2 shrink-0">
              <Button variant="ghost" size="sm" as="a" :href="`/page/${p.slug}`" target="_blank">
                <ExternalLink class="h-4 w-4 mr-1" />
                查看
              </Button>
              <Button variant="ghost" size="sm" class="text-destructive hover:text-destructive" @click="askRemove(p.id)">
                <Trash2 class="h-4 w-4" />
              </Button>
            </div>
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
        <p class="text-sm text-muted-foreground">确定删除该单页？</p>
        <DialogFooter>
          <Button variant="outline" @click="closeConfirm">取消</Button>
          <Button variant="destructive" @click="doRemove">删除</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
