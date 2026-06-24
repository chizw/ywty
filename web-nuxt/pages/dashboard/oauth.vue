<script setup lang="ts">
// 三方账号管理
definePageMeta({ layout: 'dashboard', middleware: 'auth' })

import { Unlink } from '@lucide/vue'

const api = useApi()

const rawData = ref<any>(null)

async function fetchOauth() {
  rawData.value = await api.get<any[]>('/api/v1/oauth').catch(() => [] as any[])
}

const oauths = computed<any[]>(() => {
  const d = rawData.value
  return Array.isArray(d) ? d : ((d as any)?.data ?? [])
})

onMounted(() => fetchOauth())

async function unbind(id: number) {
  if (!confirm('确定解绑该三方账号？')) return
  await api.del(`/api/v1/oauth/${id}`)
  fetchOauth()
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-foreground mb-4">三方账号</h1>
    <p class="text-sm text-muted-foreground mb-6">绑定第三方账号后可使用其登录（具体登录流程在「登录页」选择对应方式）。</p>

    <AppEmpty v-if="oauths.length === 0" title="尚未绑定三方账号" description="绑定后可使用三方账号快捷登录" />
    <Card v-else>
      <CardContent class="p-0 divide-y divide-border">
        <div v-for="o in oauths" :key="o.id" class="flex items-center gap-3 p-4">
          <Avatar v-if="o.avatar" class="h-10 w-10">
            <AvatarImage :src="o.avatar" :alt="o.nickname || o.name" />
            <AvatarFallback>{{ (o.nickname || o.name || '?').charAt(0).toUpperCase() }}</AvatarFallback>
          </Avatar>
          <Avatar v-else class="h-10 w-10">
            <AvatarFallback>{{ (o.nickname || o.name || '?').charAt(0).toUpperCase() }}</AvatarFallback>
          </Avatar>
          <div class="flex-1 min-w-0">
            <div class="text-sm font-medium text-foreground">{{ o.nickname || o.name || '未命名' }}</div>
            <div class="text-xs text-muted-foreground">driver #{{ o.driver_id }} · openid {{ o.openid }}</div>
          </div>
          <Button variant="ghost" size="sm" class="text-destructive hover:text-destructive" @click="unbind(o.id)">
            <Unlink class="mr-1 h-3 w-3" />
            解绑
          </Button>
        </div>
      </CardContent>
    </Card>
  </div>
</template>
