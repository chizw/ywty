<script setup lang="ts">
// 管理后台：图片管�?definePageMeta({ layout: 'admin', middleware: 'admin' })

import { Search, Trash2 } from '@lucide/vue'

const api = useApi()
const page = ref(1)
const perPage = 24
const keyword = ref('')

const { data, refresh } = await useAsyncData('admin-photos', () =>
  api.get<any>('/api/v1/admin/photos', { query: { page, per_page: perPage, keyword } })
)
const photos = computed<any[]>(() => (data.value as any)?.data ?? [])
const meta = computed(() => (data.value as any)?.meta)

const confirmId = ref<number | null>(null)
function askRemove(id: number) {
  confirmId.value = id
}
function closeConfirm() {
  confirmId.value = null
}
async function doRemove() {
  if (confirmId.value == null) return
  try {
    await api.del(`/api/v1/photos/${confirmId.value}`)
    confirmId.value = null
    refresh()
  } catch (err: any) {
    alert(err?.statusMessage || '删除失败')
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
      <h1 class="text-2xl font-bold text-foreground">图片管理</h1>
      <span class="text-sm text-muted-foreground">共 {{ meta?.total ?? photos.length }} �?/span>
    </div>

    <div class="mb-4 flex gap-2">
      <div class="relative flex-1 max-w-sm">
        <Search class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
        <Input
          v-model="keyword"
          placeholder="搜索文件名/原始名
          class="pl-9"
          @keyup.enter="() => refresh()"
        />
      </div>
      <Button @click="refresh">搜索</Button>
    </div>

    <div v-if="photos.length === 0" class="text-sm text-muted-foreground">暂无图片</div>
    <div v-else class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-3">
      <div
        v-for="p in photos"
        :key="p.id"
        class="group relative bg-muted rounded-md overflow-hidden aspect-square"
      >
        <img :src="`/uploads/${p.pathname}`" :alt="p.name" class="w-full h-full object-cover" loading="lazy" />
        <div class="absolute inset-0 bg-black/0 group-hover:bg-black/50 transition opacity-0 group-hover:opacity-100 flex flex-col justify-between p-2">
          <div class="text-white text-[10px] truncate">
            #{{ p.id }} · {{ p.name }}
          </div>
          <div class="flex justify-end">
            <Button size="sm" variant="destructive" class="h-7 px-2" @click="askRemove(p.id)">
              <Trash2 class="h-3 w-3" />
            </Button>
          </div>
        </div>
      </div>
    </div>

    <div v-if="meta && meta.last_page > 1" class="mt-4 flex items-center justify-end gap-2">
      <Button variant="outline" size="sm" :disabled="page <= 1" @click="page--; refresh()">上一页</Button>
      <span class="text-sm text-muted-foreground">第 {{ meta.current_page }} / {{ meta.last_page }} 页</span>
      <Button variant="outline" size="sm" :disabled="page >= meta.last_page" @click="page++; refresh()">下一页</Button>
    </div>

    <!-- 删除确认弹窗 -->
    <Dialog :open="confirmId != null" @update:open="(val: boolean) => { if (!val) closeConfirm() }">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>确认删除</DialogTitle>
        </DialogHeader>
        <p class="text-sm text-muted-foreground">管理员删除操作：确定删除该图片？此操作不可恢复。</p>
        <DialogFooter>
          <Button variant="outline" @click="closeConfirm">取消</Button>
          <Button variant="destructive" @click="doRemove">删除</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
