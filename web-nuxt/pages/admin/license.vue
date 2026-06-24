<script setup lang="ts">
// 管理后台：License 管理
definePageMeta({ layout: 'admin', middleware: 'admin' })

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
    msg.value = '请输入 License Key'
    return
  }
  loading.value = true
  msg.value = ''
  try {
    await api.post('/api/v1/admin/license/activate', form)
    msg.value = '激活成功'
    form.key = ''
    showActivate.value = false
    refresh()
  } catch (err: any) {
    msg.value = err?.statusMessage || '激活失败'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-900 mb-4">License 管理</h1>

    <div class="mb-6 p-4 bg-white border border-gray-200 rounded-lg">
      <div class="grid grid-cols-2 gap-4">
        <div>
          <div class="text-sm text-gray-500">当前版本</div>
          <div class="mt-1 text-lg font-semibold">{{ license.type || '免费版' }}</div>
        </div>
        <div>
          <div class="text-sm text-gray-500">状态</div>
          <div class="mt-1">
            <span class="px-2 py-1 text-sm rounded" :class="{
              'bg-green-100 text-green-700': license.status === 'active',
              'bg-red-100 text-red-700': license.status === 'expired',
              'bg-gray-100 text-gray-700': !license.status,
            }">{{ license.status === 'active' ? '已激活' : license.status === 'expired' ? '已过期' : '未激活' }}</span>
          </div>
        </div>
        <div>
          <div class="text-sm text-gray-500">最大用户数</div>
          <div class="mt-1 text-lg font-semibold">{{ license.max_users || '无限制' }}</div>
        </div>
        <div>
          <div class="text-sm text-gray-500">最大存储空间</div>
          <div class="mt-1 text-lg font-semibold">{{ license.max_storage ? `${(license.max_storage / 1024 / 1024 / 1024).toFixed(2)} GB` : '无限制' }}</div>
        </div>
        <div>
          <div class="text-sm text-gray-500">过期时间</div>
          <div class="mt-1 text-lg font-semibold">{{ license.expires_at ? new Date(license.expires_at).toLocaleDateString() : '永不过期' }}</div>
        </div>
      </div>

      <div v-if="license.features && license.features.length > 0" class="mt-4">
        <div class="text-sm text-gray-500 mb-2">已启用功能</div>
        <div class="flex flex-wrap gap-2">
          <span v-for="f in license.features" :key="f" class="px-2 py-1 text-xs bg-primary-100 text-primary-700 rounded">
            {{ f }}
          </span>
        </div>
      </div>
    </div>

    <div class="flex gap-2">
      <AppButton @click="showActivate = !showActivate">{{ showActivate ? '取消' : '激活 License' }}</AppButton>
    </div>

    <div v-if="showActivate" class="mt-6 p-4 bg-white border border-gray-200 rounded-lg space-y-3">
      <div>
        <label class="block text-sm text-gray-700 mb-1">License Key</label>
        <input v-model="form.key" class="w-full px-3 py-2 border border-gray-300 rounded-md font-mono" placeholder="请输入 License Key" />
      </div>
      <AppButton :loading="loading" @click="activate">激活</AppButton>
      <p v-if="msg" class="text-sm" :class="msg.includes('失败') ? 'text-red-500' : 'text-primary-600'">{{ msg }}</p>
    </div>
  </div>
</template>
