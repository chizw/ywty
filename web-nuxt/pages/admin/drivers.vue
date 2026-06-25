<script setup lang="ts">
// 管理后台：驱动管理（只读列表 + 短信/邮件测试发送）
definePageMeta({ layout: 'admin', middleware: 'admin' })

import { RefreshCw, HardDrive, MessageSquare, Mail, Users, ScanLine, Cpu } from '@lucide/vue'

const api = useApi()
const message = useMessage()

const { data, refresh } = await useAsyncData('admin-drivers', () =>
  api.get<any>('/api/v1/admin/drivers')
)
const drivers = computed(() => (data.value as any)?.data ?? {})

const providerLabel: Record<string, string> = {
  local: '本地',
  s3: 'AWS S3 / 兼容',
  oss: '阿里云OSS',
  cos: '腾讯云COS',
  qiniu: '七牛云',
  upyun: '又拍云',
  ftp: 'FTP',
  sftp: 'SFTP',
  webdav: 'WebDAV',
  smtp: 'SMTP 邮件',
  aliyun_directmail: '阿里云邮件推送',
  log: '日志占位',
  aliyun: '阿里云短信',
  tencent: '腾讯云短信',
  twilio: 'Twilio',
  github: 'GitHub',
  google: 'Google',
  wechat: '微信',
  qq: 'QQ',
  dingtalk: '钉钉',
  gitee: 'Gitee',
  weibo: '微博',
  aliyun_green: '阿里云内容安全',
  tencent_ims: '腾讯云IMS',
  custom_http: '自定义HTTP',
  noop: '空操作',
  aliyun_sms: '阿里云短信',
}

// 短信测试发送
const smsTest = reactive({ to: '', body: '【测试】您的验证码：123456' })
const smsBusy = ref(false)
async function testSMS() {
  if (!smsTest.to) {
    message.warning('请输入手机号')
    return
  }
  smsBusy.value = true
  try {
    await api.post('/api/v1/verify-codes/sms', { phone: smsTest.to, purpose: 'test' })
    message.success('短信发送请求已提交（请查看日志/上游）')
  } catch (err: any) {
    message.error(err?.statusMessage || '发送失败')
  } finally {
    smsBusy.value = false
  }
}

// 邮件测试发送
const mailTest = reactive({ to: '', subject: '【测试】邮件通知', text: '这是一封来自云雾图驿的测试邮件。' })
const mailBusy = ref(false)
async function testMail() {
  if (!mailTest.to) {
    message.warning('请输入收件人邮箱')
    return
  }
  mailBusy.value = true
  try {
    await api.post('/api/v1/verify-codes/email', { email: mailTest.to, purpose: 'test' })
    message.success('邮件发送请求已提交')
  } catch (err: any) {
    message.error(err?.statusMessage || '发送失败')
  } finally {
    mailBusy.value = false
  }
}

function prettyName(name: string) {
  return providerLabel[name] || name
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-4">
      <h1 class="text-2xl font-bold text-foreground">驱动管理</h1>
      <Button size="sm" variant="outline" @click="refresh">
        <RefreshCw class="h-4 w-4 mr-2" />
        刷新
      </Button>
    </div>

    <p class="text-sm text-muted-foreground mb-6">
      系统所有可用的扩展驱动列表。存储策略详情见 <NuxtLink to="/admin/storage" class="text-primary underline">存储策略</NuxtLink>。      驱动实际配置来自 <code class="text-xs bg-muted px-1 py-0.5 rounded">configs/config.yaml</code> 或环境变量（<code class="text-xs bg-muted px-1 py-0.5 rounded">YWTY_*</code>）    </p>

    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <Card>
        <CardContent class="p-4">
          <h3 class="font-semibold text-foreground mb-3 flex items-center">
            <HardDrive class="h-4 w-4 text-blue-500 mr-2" />
            存储驱动
            <Badge variant="secondary" class="ml-auto">{{ (drivers.storage ?? []).length }}</Badge>
          </h3>
          <div class="flex flex-wrap gap-2">
            <Badge v-for="d in (drivers.storage ?? [])" :key="`s-${d}`" variant="outline">
              {{ prettyName(d) }}
            </Badge>
            <span v-if="!(drivers.storage ?? []).length" class="text-xs text-muted-foreground">无</span>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardContent class="p-4">
          <h3 class="font-semibold text-foreground mb-3 flex items-center">
            <MessageSquare class="h-4 w-4 text-emerald-500 mr-2" />
            短信驱动
            <Badge variant="secondary" class="ml-auto">{{ (drivers.sms ?? []).length }}</Badge>
          </h3>
          <div class="flex flex-wrap gap-2">
            <Badge v-for="d in (drivers.sms ?? [])" :key="`sms-${d}`" variant="outline">
              {{ prettyName(d) }}
            </Badge>
            <span v-if="!(drivers.sms ?? []).length" class="text-xs text-muted-foreground">无</span>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardContent class="p-4">
          <h3 class="font-semibold text-foreground mb-3 flex items-center">
            <Mail class="h-4 w-4 text-amber-500 mr-2" />
            邮件驱动
            <Badge variant="secondary" class="ml-auto">{{ (drivers.mail ?? []).length }}</Badge>
          </h3>
          <div class="flex flex-wrap gap-2">
            <Badge v-for="d in (drivers.mail ?? [])" :key="`mail-${d}`" variant="outline">
              {{ prettyName(d) }}
            </Badge>
            <span v-if="!(drivers.mail ?? []).length" class="text-xs text-muted-foreground">无</span>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardContent class="p-4">
          <h3 class="font-semibold text-foreground mb-3 flex items-center">
            <Users class="h-4 w-4 text-purple-500 mr-2" />
            社交登录
            <Badge variant="secondary" class="ml-auto">{{ (drivers.social ?? []).length }}</Badge>
          </h3>
          <div class="flex flex-wrap gap-2">
            <Badge v-for="d in (drivers.social ?? [])" :key="`so-${d}`" variant="outline">
              {{ prettyName(d) }}
            </Badge>
            <span v-if="!(drivers.social ?? []).length" class="text-xs text-muted-foreground">无</span>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardContent class="p-4">
          <h3 class="font-semibold text-foreground mb-3 flex items-center">
            <ScanLine class="h-4 w-4 text-rose-500 mr-2" />
            图片扫描
            <Badge variant="secondary" class="ml-auto">{{ (drivers.scan ?? []).length }}</Badge>
          </h3>
          <div class="flex flex-wrap gap-2">
            <Badge v-for="d in (drivers.scan ?? [])" :key="`sc-${d}`" variant="outline">
              {{ prettyName(d) }}
            </Badge>
            <span v-if="!(drivers.scan ?? []).length" class="text-xs text-muted-foreground">无</span>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardContent class="p-4">
          <h3 class="font-semibold text-foreground mb-3 flex items-center">
            <Cpu class="h-4 w-4 text-cyan-500 mr-2" />
            图片处理
            <Badge variant="secondary" class="ml-auto">{{ (drivers.process ?? []).length }}</Badge>
          </h3>
          <div class="flex flex-wrap gap-2">
            <Badge v-for="d in (drivers.process ?? [])" :key="`pr-${d}`" variant="outline">
              {{ prettyName(d) }}
            </Badge>
            <span v-if="!(drivers.process ?? []).length" class="text-xs text-muted-foreground">无</span>
          </div>
        </CardContent>
      </Card>
    </div>

    <div class="mt-8 grid grid-cols-1 lg:grid-cols-2 gap-4">
      <Card>
        <CardContent class="p-4">
          <h3 class="font-semibold text-foreground mb-3">短信测试</h3>
          <p class="text-xs text-muted-foreground mb-3">通过短信验证码接口触发一次发送（请确保已配置 SMS provider）</p>
          <div class="flex gap-2">
            <Input v-model="smsTest.to" placeholder="手机号" class="flex-1" />
            <Button :loading="smsBusy" @click="testSMS">发送</Button>
          </div>
        </CardContent>
      </Card>
      <Card>
        <CardContent class="p-4">
          <h3 class="font-semibold text-foreground mb-3">邮件测试</h3>
          <p class="text-xs text-muted-foreground mb-3">通过邮件验证码接口触发一次发送</p>
          <div class="flex gap-2">
            <Input v-model="mailTest.to" placeholder="收件人邮箱" class="flex-1" />
            <Button :loading="mailBusy" @click="testMail">发送</Button>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>
