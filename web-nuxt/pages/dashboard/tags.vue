<script setup lang="ts">
// 标签管理
definePageMeta({ layout: 'dashboard', middleware: 'auth' })

import { Plus, X, Tag } from '@lucide/vue'

const api = useApi()

const rawData = ref<any>(null)
const loading = ref(false)

async function fetchTags() {
  loading.value = true
  try {
    rawData.value = await api.get<any[]>('/api/v1/tags').catch(() => [] as any[])
  } catch {
    rawData.value = []
  } finally {
    loading.value = false
  }
}

const tags = computed<any[]>(() => {
  const d = rawData.value
  return Array.isArray(d) ? d : ((d as any)?.data ?? [])
})

onMounted(() => fetchTags())

const newName = ref('')
const msg = ref('')

async function create() {
  if (!newName.value) return
  try {
    await api.post('/api/v1/tags', { name: newName.value })
    newName.value = ''
    msg.value = '已添加'
    fetchTags()
  } catch (err: any) {
    msg.value = err?.statusMessage || '添加失败'
  }
}

async function remove(id: number) {
  if (!confirm('确定删除该标签？')) return
  await api.del(`/api/v1/tags/${id}`)
  fetchTags()
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-4">
      <h1 class="text-2xl font-bold text-foreground">标签</h1>
    </div>

    <Card class="mb-6">
      <CardContent class="pt-6">
        <div class="flex gap-2">
          <Input v-model="newName" placeholder="新标签名" class="flex-1" />
          <Button @click="create">
            <Plus class="mr-1 h-4 w-4" />
            添加
          </Button>
        </div>
        <p v-if="msg" class="text-sm mt-2" :class="msg.includes('失败') ? 'text-destructive' : 'text-green-600'">{{ msg }}</p>
      </CardContent>
    </Card>

    <AppEmpty v-if="tags.length === 0" title="还没有标签" description="添加标签后可以绑定到图片" />
    <div v-else class="flex flex-wrap gap-2">
      <Badge
        v-for="t in tags"
        :key="t.id"
        variant="secondary"
        class="gap-1.5 pr-1.5 cursor-pointer hover:bg-secondary/80"
      >
        <Tag class="h-3 w-3" />
        {{ t.name }}
        <button class="ml-1 text-destructive hover:text-destructive/80" @click="remove(t.id)">
          <X class="h-3 w-3" />
        </button>
      </Badge>
    </div>
  </div>
</template>
