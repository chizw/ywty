<script setup lang="ts">
// 三方账号管理
definePageMeta({ layout: 'dashboard', middleware: 'auth' })

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
    <h1 class="text-2xl font-bold text-gray-900 mb-4">三方账号</h1>
    <p class="text-sm text-gray-500 mb-6">绑定第三方账号后可使用其登录（具体登录流程在「登录页」选择对应方式）。</p>

    <AppEmpty v-if="oauths.length === 0" title="尚未绑定三方账号" description="绑定后可使用三方账号快捷登录" />
    <div v-else class="bg-white border border-gray-200 rounded-lg divide-y">
      <div v-for="o in oauths" :key="o.id" class="flex items-center gap-3 p-4">
        <img v-if="o.avatar" :src="o.avatar" :alt="o.nickname || o.name" class="w-10 h-10 rounded-full object-cover" />
        <div v-else class="w-10 h-10 rounded-full bg-gray-200 flex items-center justify-center text-gray-500">
          {{ (o.nickname || o.name || '?').charAt(0).toUpperCase() }}
        </div>
        <div class="flex-1 min-w-0">
          <div class="text-sm font-medium text-gray-900">{{ o.nickname || o.name || '未命名' }}</div>
          <div class="text-xs text-gray-500">driver #{{ o.driver_id }} · openid {{ o.openid }}</div>
        </div>
        <button class="px-2 py-1 text-xs text-red-500" @click="unbind(o.id)">解绑</button>
      </div>
    </div>
  </div>
</template>
