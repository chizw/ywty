<script setup lang="ts">
// API Token 管理
definePageMeta({ layout: 'dashboard', middleware: 'auth' })

const api = useApi()

const rawData = ref<any>(null)
const loading = ref(false)

async function fetchTokens() {
  loading.value = true
  try {
    rawData.value = await api.get<any[]>('/api/v1/tokens').catch(() => [] as any[])
  } catch {
    rawData.value = []
  } finally {
    loading.value = false
  }
}

const tokens = computed<any[]>(() => {
  const d = rawData.value
  return Array.isArray(d) ? d : ((d as any)?.data ?? [])
})

onMounted(() => fetchTokens())

const showCreate = ref(false)
const form = reactive({ name: '', ttl_days: 0, abilities: '*' })
const loading = ref(false)
const created = ref<{ token: string; info: any } | null>(null)
const msg = ref('')

async function create() {
  if (!form.name) {
    msg.value = '请填写名称'
    return
  }
  loading.value = true
  msg.value = ''
  try {
    const abilities = form.abilities === '*' ? ['*'] : form.abilities.split(',').map((s) => s.trim()).filter(Boolean)
    const res: any = await api.post('/api/v1/tokens', {
      name: form.name,
      abilities,
      ttl_days: form.ttl_days,
    })
    created.value = { token: res.token || res.access_token, info: res.info }
    form.name = ''
    fetchTokens()
  } catch (err: any) {
    msg.value = err?.statusMessage || '创建失败'
  } finally {
    loading.value = false
  }
}

async function revoke(id: number) {
  if (!confirm('确定吊销该 Token？')) return
  await api.del(`/api/v1/tokens/${id}`)
  fetchTokens()
}

function copyToken() {
  if (created.value) {
    navigator.clipboard?.writeText(created.value.token).then(
      () => (msg.value = '已复制到剪贴板'),
      () => (msg.value = '复制失败')
    )
  }
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-4">
      <h1 class="text-2xl font-bold text-gray-900">API Token</h1>
      <AppButton @click="showCreate = !showCreate; created = null">{{ showCreate ? '取消' : '新建 Token' }}</AppButton>
    </div>

    <div v-if="showCreate" class="mb-6 p-4 bg-white border border-gray-200 rounded-lg space-y-3">
      <div>
        <label class="block text-sm text-gray-700 mb-1">名称</label>
        <input v-model="form.name" placeholder="如 ci-deploy" class="w-full px-3 py-2 border border-gray-300 rounded-md" />
      </div>
      <div>
        <label class="block text-sm text-gray-700 mb-1">能力（逗号分隔，* = 全部）</label>
        <input v-model="form.abilities" class="w-full px-3 py-2 border border-gray-300 rounded-md" />
      </div>
      <div>
        <label class="block text-sm text-gray-700 mb-1">过期天数（0 = 永不过期）</label>
        <input v-model.number="form.ttl_days" type="number" min="0" class="w-full px-3 py-2 border border-gray-300 rounded-md" />
      </div>
      <AppButton :loading="loading" @click="create">创建</AppButton>
      <p v-if="msg" class="text-sm" :class="msg.includes('失败') ? 'text-red-500' : 'text-primary-600'">{{ msg }}</p>

      <div v-if="created" class="mt-4 p-3 bg-yellow-50 border border-yellow-200 rounded-md">
        <p class="text-sm text-yellow-800 mb-2">请妥善保存，关闭后无法再次查看：</p>
        <div class="flex gap-2">
          <input :value="created.token" readonly class="flex-1 px-3 py-2 border border-gray-300 rounded-md bg-white text-xs font-mono" />
          <AppButton size="sm" @click="copyToken">复制</AppButton>
        </div>
      </div>
    </div>

    <AppEmpty v-if="tokens.length === 0" title="还没有 Token" description="创建 API Token 用于外部脚本访问" />
    <div v-else class="bg-white border border-gray-200 rounded-lg divide-y">
      <div v-for="t in tokens" :key="t.id" class="flex items-center justify-between p-4">
        <div class="min-w-0 flex-1">
          <div class="text-sm font-medium text-gray-900">{{ t.name }}</div>
          <div class="mt-1 text-xs text-gray-500">
            创建于 {{ new Date(t.created_at).toLocaleString() }}
            <span v-if="t.last_used_at"> · 最后使用 {{ new Date(t.last_used_at * 1000).toLocaleString() }}</span>
            <span v-if="t.expires_at"> · 过期 {{ new Date(t.expires_at * 1000).toLocaleString() }}</span>
          </div>
        </div>
        <button class="px-2 py-1 text-xs text-red-500" @click="revoke(t.id)">吊销</button>
      </div>
    </div>
  </div>
</template>
