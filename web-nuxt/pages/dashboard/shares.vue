<script setup lang="ts">
// 分享管理
definePageMeta({ layout: 'dashboard', middleware: 'auth' })

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
const loading = ref(false)
const msg = ref('')

async function create() {
  const ids = form.ids.split(/[,\s]+/).map((s) => Number(s.trim())).filter((n) => n > 0)
  if (ids.length === 0) {
    msg.value = '请填写至少一个资源 ID'
    return
  }
  loading.value = true
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
    loading.value = false
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
      <h1 class="text-2xl font-bold text-gray-900">分享管理</h1>
      <AppButton @click="showCreate = !showCreate">{{ showCreate ? '取消' : '新建分享' }}</AppButton>
    </div>

    <div v-if="showCreate" class="mb-6 p-4 bg-white border border-gray-200 rounded-lg space-y-3">
      <div>
        <label class="block text-sm text-gray-700 mb-1">资源类型</label>
        <select v-model="form.type" class="w-full px-3 py-2 border border-gray-300 rounded-md">
          <option value="photo">图片</option>
          <option value="album">相册</option>
        </select>
      </div>
      <div>
        <label class="block text-sm text-gray-700 mb-1">资源 ID（多个用逗号或空格分隔）</label>
        <input v-model="form.ids" placeholder="如 1,2,3" class="w-full px-3 py-2 border border-gray-300 rounded-md" />
      </div>
      <div>
        <label class="block text-sm text-gray-700 mb-1">访问密码（可选）</label>
        <input v-model="form.password" class="w-full px-3 py-2 border border-gray-300 rounded-md" />
      </div>
      <div>
        <label class="block text-sm text-gray-700 mb-1">过期分钟数（0 = 永不过期）</label>
        <input v-model.number="form.expire_minutes" type="number" min="0" class="w-full px-3 py-2 border border-gray-300 rounded-md" />
      </div>
      <AppButton :loading="loading" @click="create">创建</AppButton>
      <p v-if="msg" class="text-sm" :class="msg.includes('失败') ? 'text-red-500' : 'text-primary-600'">{{ msg }}</p>
    </div>

    <AppEmpty v-if="shares.length === 0" title="还没有分享" description="把图片或相册生成可分享链接" />
    <div v-else class="bg-white border border-gray-200 rounded-lg divide-y">
      <div v-for="s in shares" :key="s.id" class="flex items-center justify-between p-4">
        <div class="min-w-0 flex-1">
          <div class="flex items-center gap-2">
            <span class="text-sm font-medium text-gray-900">{{ s.type === 'photo' ? '图片' : '相册' }} 分享</span>
            <span class="text-xs text-gray-500">#{{ s.id }}</span>
          </div>
          <div class="mt-1 text-xs text-gray-500 truncate">
            /s/{{ s.slug }} · 浏览 {{ s.view_count ?? 0 }} 次
            <span v-if="s.expired_at"> · 过期 {{ new Date(s.expired_at * 1000).toLocaleString() }}</span>
          </div>
        </div>
        <div class="flex gap-2">
          <button class="px-2 py-1 text-xs border border-gray-300 rounded hover:bg-gray-50" @click="copyUrl(s.slug)">复制链接</button>
          <a :href="`/s/${s.slug}`" target="_blank" class="px-2 py-1 text-xs border border-gray-300 rounded hover:bg-gray-50">查看</a>
          <button class="px-2 py-1 text-xs text-red-500" @click="remove(s.id)">删除</button>
        </div>
      </div>
    </div>
  </div>
</template>
