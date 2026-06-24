<script setup lang="ts">
// 管理后台：图片管理
definePageMeta({ layout: 'admin', middleware: 'admin' })

const api = useApi()
const page = ref(1)
const perPage = 24
const keyword = ref('')

const { data, refresh } = await useAsyncData('admin-photos', () =>
  api.get<any>('/api/v1/admin/photos', { query: { page, per_page: perPage, keyword } })
)
const photos = computed<any[]>(() => (data.value as any)?.data ?? [])
const meta = computed(() => (data.value as any)?.meta)

async function remove(id: number) {
  if (!confirm('管理员删除操作：确定删除该图片？此操作不可恢复')) return
  try {
    await api.del(`/api/v1/photos/${id}`)
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
      <h1 class="text-2xl font-bold text-gray-900">图片管理</h1>
      <span class="text-sm text-gray-500">共 {{ meta?.total ?? photos.length }} 张</span>
    </div>

    <div class="mb-4 flex gap-2">
      <input
        v-model="keyword"
        placeholder="搜索文件名/原始名"
        class="flex-1 max-w-sm px-3 py-2 border border-gray-300 rounded-md"
        @keyup.enter="() => refresh()"
      />
      <AppButton @click="refresh">搜索</AppButton>
    </div>

    <div v-if="photos.length === 0" class="text-sm text-gray-500">暂无图片</div>
    <div v-else class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-3">
      <div
        v-for="p in photos"
        :key="p.id"
        class="group relative bg-gray-100 rounded overflow-hidden aspect-square"
      >
        <img :src="`/uploads/${p.pathname}`" :alt="p.name" class="w-full h-full object-cover" loading="lazy" />
        <div class="absolute inset-0 bg-black/0 group-hover:bg-black/50 transition opacity-0 group-hover:opacity-100 flex flex-col justify-between p-2">
          <div class="text-white text-[10px] truncate">
            #{{ p.id }} · {{ p.name }}
          </div>
          <div class="flex justify-end">
            <button class="px-2 py-1 bg-red-500 text-white text-xs rounded" @click="remove(p.id)">删除</button>
          </div>
        </div>
      </div>
    </div>

    <div v-if="meta && meta.last_page > 1" class="mt-4 flex items-center justify-end gap-2">
      <button :disabled="page <= 1" class="px-3 py-1 text-sm border border-gray-300 rounded disabled:opacity-50" @click="page--; refresh()">上一页</button>
      <span class="text-sm text-gray-500">第 {{ meta.current_page }} / {{ meta.last_page }} 页</span>
      <button :disabled="page >= meta.last_page" class="px-3 py-1 text-sm border border-gray-300 rounded disabled:opacity-50" @click="page++; refresh()">下一页</button>
    </div>
  </div>
</template>
