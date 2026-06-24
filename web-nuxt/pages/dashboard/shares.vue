<script setup lang="ts">
// 分享管理
definePageMeta({ layout: 'dashboard', middleware: 'auth' })

import { Plus, Copy, ExternalLink, Trash2 } from '@lucide/vue'

const api = useApi()

const rawData = ref<any>(null)
const loading = ref(false)

async function fetchShares() {
  loading.value = true
  try {
    rawData.value = await api.get<any[]>('/api/v1/shares').catch(() => [] as any[])
  } catch {
    rawData.value = []
  } finally {
    loading.value = false
  }
}

const shares = computed<any[]>(() => {
  const d = rawData.value
  return Array.isArray(d) ? d : ((d as any)?.data ?? [])
})

onMounted(() => fetchShares())

const showCreate = ref(false)
const form = reactive({
  type: 'photo' as 'photo' | 'album',
  ids: '',
  password: '',
  expire_minutes: 0,
})
const creating = ref(false)
const msg = ref('')

async function create() {
  const ids = form.ids.split(/[,\s]+/).map((s) => Number(s.trim())).filter((n) => n > 0)
  if (ids.length === 0) {
    msg.value = '请填写至少一个资源ID'
    return
  }
  creating.value = true
  msg.value = ''
  try {
    const body: any = { type: form.type, ids }
    if (form.password) body.password = form.password
    if (form.expire_minutes > 0) body.expire_minutes = form.expire_minutes
    await api.post('/api/v1/shares', body)
    msg.value = '创建成功'
    form.ids = ''
    form.password = ''
    form.expire_minutes = 0
    showCreate.value = false
    fetchShares()
  } catch (err: any) {
    msg.value = err?.statusMessage || '创建失败'
  } finally {
    creating.value = false
  }
}

async function remove(id: number) {
  if (!confirm('确定删除该分享？')) return
  await api.del(`/api/v1/shares/${id}`)
  fetchShares()
}

function copyUrl(slug: string) {
  const url = `${window.location.origin}/s/${slug}`
  navigator.clipboard?.writeText(url).then(
    () => (msg.value = '链接已复制'),
    () => (msg.value = '复制失败，请手动复制')
  )
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-4">
      <h1 class="text-2xl font-bold text-foreground">分享管理</h1>
      <Button @click="showCreate = !showCreate">
        <Plus v-if="!showCreate" class="mr-2 h-4 w-4" />
        {{ showCreate ? '取消' : '新建分享' }}
      </Button>
    </div>

    <Card v-if="showCreate" class="mb-6">
      <CardContent class="pt-6 space-y-3">
        <div>
          <Label>资源类型</Label>
          <Select v-model="form.type">
            <SelectTrigger class="mt-1">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="photo">图片</SelectItem>
              <SelectItem value="album">相册</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div>
          <Label>资源 ID（多个用逗号或空格分隔）</Label>
          <Input v-model="form.ids" placeholder="如 1,2,3" class="mt-1" />
        </div>
        <div>
          <Label>访问密码（可选）</Label>
          <Input v-model="form.password" class="mt-1" />
        </div>
        <div>
          <Label>过期分钟数（0 = 永不过期）</Label>
          <Input v-model.number="form.expire_minutes" type="number" min="0" class="mt-1" />
        </div>
        <Button :loading="creating" @click="create">创建</Button>
        <p v-if="msg" class="text-sm" :class="msg.includes('失败') ? 'text-destructive' : 'text-green-600'">{{ msg }}</p>
      </CardContent>
    </Card>

    <AppEmpty v-if="shares.length === 0" title="还没有分享" description="把图片或相册生成可分享链接" />
    <Card v-else>
      <CardContent class="p-0 divide-y divide-border">
        <div v-for="s in shares" :key="s.id" class="flex items-center justify-between p-4">
          <div class="min-w-0 flex-1">
            <div class="flex items-center gap-2">
              <span class="text-sm font-medium text-foreground">{{ s.type === 'photo' ? '图片' : '相册' }} 分享</span>
              <Badge variant="secondary">#{{ s.id }}</Badge>
            </div>
            <div class="mt-1 text-xs text-muted-foreground truncate">
              /s/{{ s.slug }} · 浏览 {{ s.view_count ?? 0 }} 次              <span v-if="s.expired_at"> · 过期 {{ new Date(s.expired_at * 1000).toLocaleString() }}</span>
            </div>
          </div>
          <div class="flex gap-2">
            <Button variant="outline" size="sm" @click="copyUrl(s.slug)">
              <Copy class="mr-1 h-3 w-3" />
              复制链接
            </Button>
            <a :href="`/s/${s.slug}`" target="_blank">
              <Button variant="outline" size="sm">
                <ExternalLink class="mr-1 h-3 w-3" />
                查看
              </Button>
            </a>
            <Button variant="ghost" size="sm" class="text-destructive hover:text-destructive" @click="remove(s.id)">
              <Trash2 class="mr-1 h-3 w-3" />
              删除
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>
  </div>
</template>
