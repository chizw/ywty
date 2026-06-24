<script setup lang="ts">
// API Token 管理
definePageMeta({ layout: 'dashboard', middleware: 'auth' })

import { Plus, Copy, Trash2, Key, AlertTriangle } from '@lucide/vue'

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
const creating = ref(false)
const created = ref<{ token: string; info: any } | null>(null)
const msg = ref('')

async function create() {
  if (!form.name) {
    msg.value = '请填写名�?
    return
  }
  creating.value = true
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
    creating.value = false
  }
}

async function revoke(id: number) {
  if (!confirm('确定吊销�?Token�?)) return
  await api.del(`/api/v1/tokens/${id}`)
  fetchTokens()
}

function copyToken() {
  if (created.value) {
    navigator.clipboard?.writeText(created.value.token).then(
      () => (msg.value = '已复制到剪贴�?),
      () => (msg.value = '复制失败')
    )
  }
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-4">
      <h1 class="text-2xl font-bold text-foreground">API Token</h1>
      <Button @click="showCreate = !showCreate; created = null">
        <Plus v-if="!showCreate" class="mr-2 h-4 w-4" />
        {{ showCreate ? '取消' : '新建 Token' }}
      </Button>
    </div>

    <Card v-if="showCreate" class="mb-6">
      <CardContent class="pt-6 space-y-3">
        <div>
          <Label>名称</Label>
          <Input v-model="form.name" placeholder="�?ci-deploy" class="mt-1" />
        </div>
        <div>
          <Label>能力（逗号分隔�? = 全部�?/Label>
          <Input v-model="form.abilities" class="mt-1" />
        </div>
        <div>
          <Label>过期天数�? = 永不过期�?/Label>
          <Input v-model.number="form.ttl_days" type="number" min="0" class="mt-1" />
        </div>
        <Button :loading="creating" @click="create">创建</Button>
        <p v-if="msg" class="text-sm" :class="msg.includes('失败') ? 'text-destructive' : 'text-green-600'">{{ msg }}</p>

        <Alert v-if="created" variant="warning">
          <AlertTriangle class="h-4 w-4" />
          <AlertDescription>
            <p class="text-sm mb-2">请妥善保存，关闭后无法再次查看：</p>
            <div class="flex gap-2">
              <Input :model-value="created.token" readonly class="flex-1 text-xs font-mono bg-background" />
              <Button variant="outline" size="sm" @click="copyToken">
                <Copy class="mr-1 h-3 w-3" />
                复制
              </Button>
            </div>
          </AlertDescription>
        </Alert>
      </CardContent>
    </Card>

    <AppEmpty v-if="tokens.length === 0" title="还没�?Token" description="创建 API Token 用于外部脚本访问" />
    <Card v-else>
      <CardContent class="p-0 divide-y divide-border">
        <div v-for="t in tokens" :key="t.id" class="flex items-center justify-between p-4">
          <div class="min-w-0 flex-1">
            <div class="flex items-center gap-2">
              <Key class="h-4 w-4 text-muted-foreground" />
              <span class="text-sm font-medium text-foreground">{{ t.name }}</span>
            </div>
            <div class="mt-1 text-xs text-muted-foreground">
              创建�?{{ new Date(t.created_at).toLocaleString() }}
              <span v-if="t.last_used_at"> · 最后使�?{{ new Date(t.last_used_at * 1000).toLocaleString() }}</span>
              <span v-if="t.expires_at"> · 过期 {{ new Date(t.expires_at * 1000).toLocaleString() }}</span>
            </div>
          </div>
          <Button variant="ghost" size="sm" class="text-destructive hover:text-destructive" @click="revoke(t.id)">
            <Trash2 class="mr-1 h-3 w-3" />
            吊销
          </Button>
        </div>
      </CardContent>
    </Card>
  </div>
</template>
