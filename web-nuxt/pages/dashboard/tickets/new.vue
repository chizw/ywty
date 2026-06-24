<script setup lang="ts">
// 新建工单（前端骨架，后端 P7 完成后对接）
definePageMeta({ layout: 'dashboard', middleware: 'auth' })

const api = useApi()
const message = useMessage()
const router = useRouter()

const form = reactive({
  type: 'bug' as 'bug' | 'feature' | 'complaint' | 'other',
  priority: 'medium' as 'low' | 'medium' | 'high' | 'urgent',
  title: '',
  content: '',
})
const loading = ref(false)

const typeOptions = [
  { value: 'bug', label: 'Bug 反馈' },
  { value: 'feature', label: '功能建议' },
  { value: 'complaint', label: '投诉' },
  { value: 'other', label: '其他' },
]
const priorityOptions = [
  { value: 'low', label: '低' },
  { value: 'medium', label: '中' },
  { value: 'high', label: '高' },
  { value: 'urgent', label: '紧急' },
]

async function submit() {
  if (!form.title.trim() || !form.content.trim()) {
    message.warning('请填写标题和内容')
    return
  }
  loading.value = true
  try {
    const res = await api.post<any>('/api/v1/tickets', {
      type: form.type,
      priority: form.priority,
      title: form.title,
      content: form.content,
    })
    message.success('工单已提交')
    const id = (res as any)?.id || (res as any)?.data?.id
    if (id) {
      router.push(`/dashboard/tickets/${id}`)
    } else {
      router.push('/dashboard/tickets')
    }
  } catch (err: any) {
    message.error(err?.statusMessage || '提交失败，请稍后重试')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div>
    <div class="mb-4">
      <NuxtLink to="/dashboard/tickets" class="text-xs text-gray-500 hover:text-primary-600">← 返回工单列表</NuxtLink>
      <h1 class="text-2xl font-bold text-gray-900 mt-1">新建工单</h1>
    </div>

    <form class="max-w-2xl bg-white border border-gray-200 rounded-lg p-6 space-y-4" @submit.prevent="submit">
      <div class="grid grid-cols-2 gap-4">
        <div>
          <label class="block text-sm text-gray-700 mb-1">类型</label>
          <select v-model="form.type" class="w-full px-3 py-2 border border-gray-300 rounded-md">
            <option v-for="o in typeOptions" :key="o.value" :value="o.value">{{ o.label }}</option>
          </select>
        </div>
        <div>
          <label class="block text-sm text-gray-700 mb-1">优先级</label>
          <select v-model="form.priority" class="w-full px-3 py-2 border border-gray-300 rounded-md">
            <option v-for="o in priorityOptions" :key="o.value" :value="o.value">{{ o.label }}</option>
          </select>
        </div>
      </div>
      <div>
        <label class="block text-sm text-gray-700 mb-1">标题</label>
        <input v-model="form.title" maxlength="100" placeholder="一句话描述你的问题" class="w-full px-3 py-2 border border-gray-300 rounded-md" />
      </div>
      <div>
        <label class="block text-sm text-gray-700 mb-1">内容</label>
        <textarea v-model="form.content" rows="8" placeholder="详细描述问题、复现步骤或建议..." class="w-full px-3 py-2 border border-gray-300 rounded-md"></textarea>
      </div>
      <div class="flex items-center gap-3">
        <AppButton type="submit" :loading="loading">提交工单</AppButton>
        <NuxtLink to="/dashboard/tickets" class="px-4 py-2 text-sm text-gray-600 hover:text-gray-900">取消</NuxtLink>
      </div>
    </form>
  </div>
</template>
