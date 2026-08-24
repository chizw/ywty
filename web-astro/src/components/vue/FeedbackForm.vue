<script setup lang="ts">
// 意见反馈表单（匿名可提交）
import { reactive, ref } from 'vue'
import Button from './ui/Button.vue'
import Input from './ui/Input.vue'
import Label from './ui/Label.vue'
import { useApi } from '../../lib/api'

const form = reactive({ nickname: '', contact: '', content: '' })
const loading = ref(false)
const done = ref(false)
const errorMsg = ref('')

async function submit() {
  if (form.content.trim().length < 5) {
    errorMsg.value = '请填写至少 5 个字的反馈内容'
    return
  }
  loading.value = true
  errorMsg.value = ''
  try {
    // 后端 feedback 表：content 必填，其余拼进内容便于运营查看
    const header = [
      form.nickname ? `昵称:${form.nickname}` : '',
      form.contact ? `联系方式:${form.contact}` : '',
    ]
      .filter(Boolean)
      .join(' | ')
    await useApi().post('/api/v1/feedback', {
      content: header ? `[${header}] ${form.content.trim()}` : form.content.trim(),
    })
    done.value = true
  } catch (err: any) {
    errorMsg.value = err?.message || '提交失败，请稍后重试'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="mx-auto max-w-lg px-6 py-16">
    <h1 class="font-display text-2xl font-bold tracking-tight">意见反馈</h1>
    <p class="mt-1.5 text-sm text-muted-foreground">遇到问题或有改进建议？欢迎告诉我们。</p>

    <div v-if="done" class="mt-8 rounded-md border border-border bg-card p-6 text-center">
      <p class="text-sm font-medium text-emerald-600 dark:text-emerald-400">反馈已提交，感谢你的支持！</p>
      <a href="/" class="mt-3 inline-block text-sm text-brand hover:underline">返回首页</a>
    </div>

    <form v-else class="mt-8 space-y-5" @submit.prevent="submit">
      <div class="grid grid-cols-2 gap-4">
        <div class="space-y-2">
          <Label for="fb-nickname">昵称（选填）</Label>
          <Input id="fb-nickname" v-model="form.nickname" placeholder="怎么称呼你" />
        </div>
        <div class="space-y-2">
          <Label for="fb-contact">联系方式（选填）</Label>
          <Input id="fb-contact" v-model="form.contact" placeholder="邮箱等，便于回复" />
        </div>
      </div>
      <div class="space-y-2">
        <Label for="fb-content">反馈内容</Label>
        <textarea
          id="fb-content"
          v-model="form.content"
          rows="6"
          required
          class="w-full rounded-md border border-border bg-background px-3 py-2 text-sm"
          placeholder="请描述问题或建议…"
        />
      </div>

      <p v-if="errorMsg" class="text-[13px] text-destructive">{{ errorMsg }}</p>

      <Button type="submit" :loading="loading" class="w-full">提交反馈</Button>
    </form>
  </div>
</template>
