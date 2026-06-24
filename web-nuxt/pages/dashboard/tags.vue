<script setup lang="ts">
// 标签管理
definePageMeta({ layout: 'dashboard', middleware: 'auth' })

const api = useApi()
const { data, refresh } = await useAsyncData('my-tags', () =>
  api.get<any[]>('/api/v1/tags').catch(() => [] as any[])
)

const tags = computed<any[]>(() => {
  const d = data.value
  return Array.isArray(d) ? d : ((d as any)?.data ?? [])
})

const newName = ref('')
const msg = ref('')

async function create() {
  if (!newName.value) return
  try {
    await api.post('/api/v1/tags', { name: newName.value })
    newName.value = ''
    msg.value = '已添加'
    refresh()
  } catch (err: any) {
    msg.value = err?.statusMessage || '添加失败'
  }
}

async function remove(id: number) {
  if (!confirm('确定删除该标签？')) return
  await api.del(`/api/v1/tags/${id}`)
  refresh()
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-4">
      <h1 class="text-2xl font-bold text-gray-900">标签</h1>
    </div>

    <div class="mb-6 p-4 bg-white border border-gray-200 rounded-lg flex gap-2">
      <input v-model="newName" placeholder="新标签名" class="flex-1 px-3 py-2 border border-gray-300 rounded-md" />
      <AppButton @click="create">添加</AppButton>
    </div>
    <p v-if="msg" class="text-sm mb-2" :class="msg.includes('失败') ? 'text-red-500' : 'text-primary-600'">{{ msg }}</p>

    <AppEmpty v-if="tags.length === 0" title="还没有标签" description="添加标签后可以绑定到图片" />
    <div v-else class="flex flex-wrap gap-2">
      <div v-for="t in tags" :key="t.id" class="inline-flex items-center gap-2 px-3 py-1.5 bg-white border border-gray-200 rounded-full text-sm">
        <span>{{ t.name }}</span>
        <button class="text-red-500 text-xs" @click="remove(t.id)">×</button>
      </div>
    </div>
  </div>
</template>
