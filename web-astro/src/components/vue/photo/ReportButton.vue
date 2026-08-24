<script setup lang="ts">
// 举报按钮 + 弹窗：POST /api/v1/reports（需登录，401 时引导登录）
import { ref } from 'vue'
import { Flag } from '@lucide/vue'
import Button from '../ui/Button.vue'
import Label from '../ui/Label.vue'
import { useApi } from '../../../lib/api'
import { message } from '../../../lib/vue-store'

const props = withDefaults(
  defineProps<{
    targetType: 'photo' | 'album'
    targetId: number
    /** 深色背景（灯箱内）样式 */
    dark?: boolean
  }>(),
  { dark: false },
)

const msg = message
const open = ref(false)
const reason = ref('违规内容')
const detail = ref('')
const submitting = ref(false)

const REASONS = ['色情低俗', '暴力血腥', '侵权盗用', '垃圾广告', '政治敏感', '其他']

function onClick() {
  open.value = true
}

async function submit() {
  submitting.value = true
  try {
    const content = detail.value.trim()
      ? `【${reason.value}】${detail.value.trim()}`
      : reason.value
    await useApi().post('/api/v1/reports', {
      target_type: props.targetType,
      target_id: props.targetId,
      content,
    })
    msg.success('举报已提交，感谢你的反馈')
    open.value = false
    detail.value = ''
  } catch (err: any) {
    const status = err?.status ?? err?.response?.status
    if (status === 401) {
      window.location.assign('/auth/login?redirect=' + encodeURIComponent(window.location.pathname))
      return
    }
    msg.error(err?.message || '举报提交失败')
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div>
    <button
      type="button"
      class="inline-flex h-8 items-center gap-1.5 rounded-md px-2 text-xs transition-colors"
      :class="dark
        ? 'text-white/80 hover:bg-white/10 hover:text-white'
        : 'text-muted-foreground hover:bg-muted hover:text-foreground'"
      title="举报此内容"
      @click.stop="onClick"
    >
      <Flag class="h-4 w-4" />
      <span>举报</span>
    </button>

    <Teleport to="body">
      <div
        v-if="open"
        class="fixed inset-0 z-[110] flex items-center justify-center bg-black/50 p-4"
        @click.self="open = false"
      >
        <div class="w-full max-w-sm rounded-lg border border-border bg-card p-5 shadow-xl">
          <h3 class="font-display text-base font-semibold">举报内容</h3>
          <div class="mt-4 space-y-4">
            <div class="space-y-1.5">
              <Label for="report-reason">举报类型</Label>
              <select
                id="report-reason"
                v-model="reason"
                class="h-9 w-full rounded-md border border-border bg-background px-3 text-sm"
              >
                <option v-for="r in REASONS" :key="r" :value="r">{{ r }}</option>
              </select>
            </div>
            <div class="space-y-1.5">
              <Label for="report-detail">补充说明（选填）</Label>
              <textarea
                id="report-detail"
                v-model="detail"
                rows="3"
                class="w-full rounded-md border border-border bg-background px-3 py-2 text-sm"
                placeholder="补充具体情况…"
              />
            </div>
          </div>
          <div class="mt-5 flex justify-end gap-2">
            <Button variant="outline" size="sm" @click="open = false">取消</Button>
            <Button size="sm" :loading="submitting" @click="submit">提交举报</Button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
