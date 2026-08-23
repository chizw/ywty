<script setup lang="ts">
// 分享展示：大图 + 信息卡片 + 下载按钮（密码由后端强制校验，错误密码不返回内容）
import { ref, computed } from 'vue'
import { Download, Image as ImageIcon, Folder, Eye, Calendar, Lock } from '@lucide/vue'
import Button from '../ui/Button.vue'
import Input from '../ui/Input.vue'
import Label from '../ui/Label.vue'
import { formatBytes } from '../../../lib/utils'
import { useApi } from '../../../lib/api'
import type { ShareData } from '../../../lib/types'

const props = defineProps<{
  slug: string
  initial: ShareData | null
}>()

const share = ref<ShareData | null>(props.initial)
const password = ref('')
const errorMsg = ref('')
const unlocking = ref(false)
const unlocked = ref(!props.initial?.requires_password)

const isPhoto = computed(() => share.value?.type === 'photo')
const isAlbum = computed(() => share.value?.type === 'album')

async function unlock() {
  if (!password.value) {
    errorMsg.value = '请输入访问密码'
    return
  }
  unlocking.value = true
  errorMsg.value = ''
  try {
    // 后端校验访问密码，密码错误返回 403
    const res = await useApi().get(`/s/${props.slug}`, { query: { password: password.value } })
    if (res) {
      share.value = res as ShareData
      unlocked.value = true
    }
  } catch (err: any) {
    errorMsg.value = err?.message || '解锁失败'
  } finally {
    unlocking.value = false
  }
}

const infoRows = computed(() => {
  const s = share.value
  if (!s) return []
  const rows: { label: string; value: string }[] = []
  if (s.width && s.height) rows.push({ label: '尺寸', value: `${s.width} × ${s.height}` })
  if (s.size) rows.push({ label: '大小', value: formatBytes(s.size) })
  if (isAlbum.value) rows.push({ label: '图片数', value: `${s.photo_count ?? 0} 张` })
  return rows
})
</script>

<template>
  <div>
    <!-- 加载/错误 -->
    <div v-if="!share" class="py-24 text-center text-muted-foreground">
      <p class="text-lg">分享不存在或已过期</p>
      <a href="/" class="mt-4 inline-block text-sm text-primary hover:underline">返回首页</a>
    </div>

    <!-- 密码门 -->
    <div v-else-if="!unlocked" class="mx-auto max-w-md py-16">
      <div class="rounded-2xl border border-border bg-card p-8 shadow-sm">
        <div class="mx-auto grid h-12 w-12 place-items-center rounded-full bg-accent">
          <Lock class="h-5 w-5 text-accent-foreground" />
        </div>
        <h1 class="mt-4 text-center font-display text-xl font-bold">此分享需要密码</h1>
        <p class="mt-1 text-center text-sm text-muted-foreground">输入访问密码后查看内容</p>
        <form class="mt-6 space-y-5" @submit.prevent="unlock">
          <div class="space-y-2">
            <Label for="share-pwd">访问密码</Label>
            <Input id="share-pwd" v-model="password" type="password" placeholder="请输入密码" variant="underline" autofocus />
          </div>
          <p v-if="errorMsg" class="text-[13px] text-destructive">{{ errorMsg }}</p>
          <Button type="submit" :loading="unlocking" class="w-full">解锁</Button>
        </form>
      </div>
    </div>

    <!-- 照片分享：大图 + 信息卡 -->
    <div v-else-if="isPhoto && share.url">
      <div class="overflow-hidden rounded-2xl border border-border bg-card shadow-sm">
        <div class="relative bg-black/5">
          <img :src="share.url" :alt="share.uuid || '分享图片'" class="mx-auto max-h-[72vh] w-full object-contain" />
        </div>
        <div class="flex flex-wrap items-center justify-between gap-3 border-t border-border p-4 sm:p-6">
          <div class="flex flex-wrap items-center gap-4 text-sm text-muted-foreground">
            <span class="inline-flex items-center gap-1.5">
              <ImageIcon class="h-4 w-4" />
              图片分享
            </span>
            <span v-if="share.width && share.height" class="inline-flex items-center gap-1.5">
              {{ share.width }} × {{ share.height }}
            </span>
            <span v-if="share.size" class="inline-flex items-center gap-1.5">
              {{ formatBytes(share.size) }}
            </span>
          </div>
          <div class="flex gap-2">
            <a :href="share.url" target="_blank" rel="noopener">
              <Button variant="outline" size="sm">新窗口查看</Button>
            </a>
            <a :href="share.url" :download="`${share.uuid}.jpg`">
              <Button size="sm">
                <Download class="h-4 w-4" />
                下载原图
              </Button>
            </a>
          </div>
        </div>
      </div>
    </div>

    <!-- 相册分享：信息卡 -->
    <div v-else-if="isAlbum">
      <div class="mx-auto max-w-2xl rounded-2xl border border-border bg-card p-8 shadow-sm">
        <div class="mx-auto grid h-14 w-14 place-items-center rounded-xl bg-accent">
          <Folder class="h-7 w-7 text-accent-foreground" />
        </div>
        <h1 class="mt-5 text-center font-display text-2xl font-bold">{{ share.name || '相册分享' }}</h1>
        <p v-if="share.description" class="mt-2 text-center text-sm text-muted-foreground">{{ share.description }}</p>
        <div class="mt-6 flex flex-wrap items-center justify-center gap-4 text-sm text-muted-foreground">
          <span class="inline-flex items-center gap-1.5">
            <ImageIcon class="h-4 w-4" />
            共 {{ share.photo_count ?? 0 }} 张图片
          </span>
        </div>
        <p class="mt-6 text-center text-xs text-muted-foreground">
          相册内容请在登录后查看完整相册
        </p>
      </div>
    </div>

    <!-- 兜底信息卡 -->
    <div v-else class="mx-auto max-w-md py-16 text-center text-muted-foreground">
      <p>暂不支持预览的分享类型</p>
      <a href="/" class="mt-4 inline-block text-sm text-primary hover:underline">返回首页</a>
    </div>
  </div>
</template>
