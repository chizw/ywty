<script setup lang="ts">
// 我的相册
definePageMeta({ layout: 'dashboard', middleware: 'auth' })

import { Plus, Trash2, Eye } from '@lucide/vue'

const api = useApi()

const rawData = ref<any>(null)
const loading = ref(false)

async function fetchAlbums() {
  loading.value = true
  try {
    rawData.value = await api.get<any[]>('/api/v1/albums').catch(() => [] as any[])
  } catch {
    rawData.value = []
  } finally {
    loading.value = false
  }
}

const albums = computed<any[]>(() => {
  const d = rawData.value
  return Array.isArray(d) ? d : ((d as any)?.data ?? [])
})

onMounted(() => fetchAlbums())

const showCreate = ref(false)
const newAlbum = reactive({ name: '', intro: '', is_public: false })

async function create() {
  if (!newAlbum.name) return
  await api.post('/api/v1/albums', newAlbum)
  newAlbum.name = ''
  newAlbum.intro = ''
  newAlbum.is_public = false
  showCreate.value = false
  fetchAlbums()
}

async function remove(id: number) {
  if (!confirm('确定删除该相册？')) return
  await api.del(`/api/v1/albums/${id}`)
  fetchAlbums()
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-4">
      <h1 class="text-2xl font-bold text-foreground">我的相册</h1>
      <Button @click="showCreate = !showCreate">
        <Plus v-if="!showCreate" class="mr-2 h-4 w-4" />
        {{ showCreate ? '取消' : '新建相册' }}
      </Button>
    </div>

    <Card v-if="showCreate" class="mb-6">
      <CardContent class="pt-6 space-y-3">
        <div>
          <Label>相册名</Label>
          <Input v-model="newAlbum.name" placeholder="相册名" class="mt-1" />
        </div>
        <div>
          <Label>介绍（可选）</Label>
          <Textarea v-model="newAlbum.intro" placeholder="介绍（可选）" rows="2" class="mt-1" />
        </div>
        <div class="flex items-center gap-2">
          <Checkbox v-model:checked="newAlbum.is_public" id="album-public" />
          <Label for="album-public">公开相册</Label>
        </div>
        <Button @click="create">创建</Button>
      </CardContent>
    </Card>

    <AppEmpty v-if="albums.length === 0" title="还没有相册" description="点击右上角新建一个相册，把图片归类管理" />
    <div v-else class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4">
      <Card
        v-for="a in albums"
        :key="a.id"
      >
        <CardHeader class="pb-3">
          <div class="flex items-center justify-between">
            <CardTitle class="text-base">{{ a.name }}</CardTitle>
            <Badge v-if="a.is_public" variant="secondary">公开</Badge>
          </div>
          <CardDescription v-if="a.intro" class="line-clamp-2">{{ a.intro }}</CardDescription>
        </CardHeader>
        <CardContent class="pt-0">
          <p class="text-xs text-muted-foreground">{{ a.photo_count }} 张图片</p>
        </CardContent>
        <CardFooter class="gap-2">
          <NuxtLink :to="`/dashboard/albums/${a.id}`">
            <Button variant="outline" size="sm">
              <Eye class="mr-1 h-3 w-3" />
              查看
            </Button>
          </NuxtLink>
          <Button variant="ghost" size="sm" class="text-destructive hover:text-destructive" @click="remove(a.id)">
            <Trash2 class="mr-1 h-3 w-3" />
            删除
          </Button>
        </CardFooter>
      </Card>
    </div>
  </div>
</template>
