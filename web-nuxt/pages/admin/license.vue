<script setup lang="ts">
// 管理后台：License 管理
definePageMeta({ layout: 'admin', middleware: 'admin' })

import { Key, Shield, Users, HardDrive, Clock, Sparkles } from '@lucide/vue'

const api = useApi()
const { data, refresh } = await useAsyncData('admin-license', () => api.get<any>('/api/v1/admin/license'))

const license = computed(() => (data.value as any) ?? {})

const showActivate = ref(false)
const form = reactive({
  key: '',
})
const loading = ref(false)
const msg = ref('')

async function activate() {
  if (!form.key.trim()) {
    msg.value = '请输�?License Key'
    return
  }
  loading.value = true
  msg.value = ''
  try {
    await api.post('/api/v1/admin/license/activate', form)
    msg.value = '激活成�?
    form.key = ''
    showActivate.value = false
    refresh()
  } catch (err: any) {
    msg.value = err?.statusMessage || '激活失�?
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-foreground mb-4">License 管理</h1>

    <Card class="mb-6">
      <CardContent class="p-6">
        <div class="grid grid-cols-2 gap-4">
          <div>
            <div class="flex items-center gap-2 text-sm text-muted-foreground mb-1">
              <Shield class="h-4 w-4" />
              当前版本
            </div>
            <div class="text-lg font-semibold">{{ license.type || '免费�? }}</div>
          </div>
          <div>
            <div class="flex items-center gap-2 text-sm text-muted-foreground mb-1">
              <Key class="h-4 w-4" />
              状�?            </div>
            <Badge :variant="license.status === 'active' ? 'success' : license.status === 'expired' ? 'destructive' : 'secondary'">
              {{ license.status === 'active' ? '已激�? : license.status === 'expired' ? '已过�? : '未激�? }}
            </Badge>
          </div>
          <div>
            <div class="flex items-center gap-2 text-sm text-muted-foreground mb-1">
              <Users class="h-4 w-4" />
              最大用户数
            </div>
            <div class="text-lg font-semibold">{{ license.max_users || '无限�? }}</div>
          </div>
          <div>
            <div class="flex items-center gap-2 text-sm text-muted-foreground mb-1">
              <HardDrive class="h-4 w-4" />
              最大存储空�?            </div>
            <div class="text-lg font-semibold">{{ license.max_storage ? `${(license.max_storage / 1024 / 1024 / 1024).toFixed(2)} GB` : '无限�? }}</div>
          </div>
          <div>
            <div class="flex items-center gap-2 text-sm text-muted-foreground mb-1">
              <Clock class="h-4 w-4" />
              过期时间
            </div>
            <div class="text-lg font-semibold">{{ license.expires_at ? new Date(license.expires_at).toLocaleDateString() : '永不过期' }}</div>
          </div>
        </div>

        <div v-if="license.features && license.features.length > 0" class="mt-4">
          <div class="flex items-center gap-2 text-sm text-muted-foreground mb-2">
            <Sparkles class="h-4 w-4" />
            已启用功�?          </div>
          <div class="flex flex-wrap gap-2">
            <Badge v-for="f in license.features" :key="f" variant="default">
              {{ f }}
            </Badge>
          </div>
        </div>
      </CardContent>
    </Card>

    <div class="flex gap-2">
      <Button variant="outline" @click="showActivate = !showActivate">
        <Key class="h-4 w-4 mr-2" />
        {{ showActivate ? '取消' : '激�?License' }}
      </Button>
    </div>

    <Card v-if="showActivate" class="mt-6">
      <CardContent class="p-4 space-y-3">
        <div>
          <Label class="mb-1.5 block">License Key</Label>
          <Input v-model="form.key" class="font-mono" placeholder="请输�?License Key" />
        </div>
        <Button :loading="loading" @click="activate">激�?/Button>
        <p v-if="msg" class="text-sm" :class="msg.includes('失败') ? 'text-destructive' : 'text-primary'">{{ msg }}</p>
      </CardContent>
    </Card>
  </div>
</template>
