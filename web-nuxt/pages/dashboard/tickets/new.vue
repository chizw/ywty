<script setup lang="ts">
// 新建工单（前端骨架，后端 P7 完成后对接）
definePageMeta({ layout: 'dashboard', middleware: 'auth' })

import { ArrowLeft, Send } from '@lucide/vue'

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
      <NuxtLink to="/dashboard/tickets" class="text-xs text-muted-foreground hover:text-primary flex items-center gap-1">
        <ArrowLeft class="h-3 w-3" />
        返回工单列表
      </NuxtLink>
      <h1 class="text-2xl font-bold text-foreground mt-1">新建工单</h1>
    </div>

    <Card class="max-w-2xl">
      <form @submit.prevent="submit">
        <CardContent class="pt-6 space-y-4">
          <div class="grid grid-cols-2 gap-4">
            <div>
              <Label>类型</Label>
              <Select v-model="form.type">
                <SelectTrigger class="mt-1">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem v-for="o in typeOptions" :key="o.value" :value="o.value">{{ o.label }}</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div>
              <Label>优先级</Label>
              <Select v-model="form.priority">
                <SelectTrigger class="mt-1">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem v-for="o in priorityOptions" :key="o.value" :value="o.value">{{ o.label }}</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
          <div>
            <Label>标题</Label>
            <Input v-model="form.title" maxlength="100" placeholder="一句话描述你的问题" class="mt-1" />
          </div>
          <div>
            <Label>内容</Label>
            <Textarea v-model="form.content" rows="8" placeholder="详细描述问题、复现步骤或建议..." class="mt-1" />
          </div>
        </CardContent>
        <CardFooter class="gap-3">
          <Button type="submit" :loading="loading">
            <Send class="mr-2 h-4 w-4" />
            提交工单
          </Button>
          <NuxtLink to="/dashboard/tickets" class="px-4 py-2 text-sm text-muted-foreground hover:text-foreground">取消</NuxtLink>
        </CardFooter>
      </form>
    </Card>
  </div>
</template>
