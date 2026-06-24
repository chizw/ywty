<script setup lang="ts">
// 管理后台：驱动管理（只读列表 + 短信/邮件测试发送）
definePageMeta({ layout: 'admin', middleware: 'admin' })

const api = useApi()
const message = useMessage()

const { data, refresh } = await useAsyncData('admin-drivers', () =>
  api.get<any>('/api/v1/admin/drivers')
)
const drivers = computed(() => (data.value as any)?.data ?? {})

const providerLabel: Record<string, string> = {
  local: '本地',
  s3: 'AWS S3 / 兼容',
  oss: '阿里云 OSS',
  cos: '腾讯云 COS',
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
  tencent_ims: '腾讯云 IMS',
  custom_http: '自定义 HTTP',
  noop: '空操作',
  aliyun_sms: '阿里云短信(旧)',
}

// 短信测试发送
const smsTest = reactive({ to: '', body: '【测试】您的验证码是 123456' })
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
const mailTest = reactive({ to: '', subject: '【测试】邮件通知', text: '这是一封来自 ywty 的测试邮件。' })
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
      <h1 class="text-2xl font-bold text-gray-900">驱动管理</h1>
      <AppButton size="sm" @click="refresh">刷新</AppButton>
    </div>

    <p class="text-sm text-gray-500 mb-6">
      系统所有可用的扩展驱动列表。存储策略详情见 <NuxtLink to="/admin/storage" class="text-primary-600">存储策略</NuxtLink>。
      驱动实际配置来自 <code>configs/config.yaml</code> 或环境变量（<code>YWTY_*</code>）。
    </p>

    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <div class="bg-white rounded-md border border-gray-200 p-4">
        <h3 class="font-semibold text-gray-800 mb-3 flex items-center">
          <span class="w-2 h-2 rounded-full bg-blue-500 mr-2"></span>
          存储驱动
          <span class="ml-auto text-xs text-gray-400">{{ (drivers.storage ?? []).length }}</span>
        </h3>
        <div class="flex flex-wrap gap-2">
          <span
            v-for="d in (drivers.storage ?? [])"
            :key="`s-${d}`"
            class="px-2 py-1 text-xs bg-blue-50 text-blue-700 border border-blue-200 rounded"
          >
            {{ prettyName(d) }}
          </span>
          <span v-if="!(drivers.storage ?? []).length" class="text-xs text-gray-400">无</span>
        </div>
      </div>

      <div class="bg-white rounded-md border border-gray-200 p-4">
        <h3 class="font-semibold text-gray-800 mb-3 flex items-center">
          <span class="w-2 h-2 rounded-full bg-emerald-500 mr-2"></span>
          短信驱动
          <span class="ml-auto text-xs text-gray-400">{{ (drivers.sms ?? []).length }}</span>
        </h3>
        <div class="flex flex-wrap gap-2">
          <span
            v-for="d in (drivers.sms ?? [])"
            :key="`sms-${d}`"
            class="px-2 py-1 text-xs bg-emerald-50 text-emerald-700 border border-emerald-200 rounded"
          >
            {{ prettyName(d) }}
          </span>
          <span v-if="!(drivers.sms ?? []).length" class="text-xs text-gray-400">无</span>
        </div>
      </div>

      <div class="bg-white rounded-md border border-gray-200 p-4">
        <h3 class="font-semibold text-gray-800 mb-3 flex items-center">
          <span class="w-2 h-2 rounded-full bg-amber-500 mr-2"></span>
          邮件驱动
          <span class="ml-auto text-xs text-gray-400">{{ (drivers.mail ?? []).length }}</span>
        </h3>
        <div class="flex flex-wrap gap-2">
          <span
            v-for="d in (drivers.mail ?? [])"
            :key="`mail-${d}`"
            class="px-2 py-1 text-xs bg-amber-50 text-amber-700 border border-amber-200 rounded"
          >
            {{ prettyName(d) }}
          </span>
          <span v-if="!(drivers.mail ?? []).length" class="text-xs text-gray-400">无</span>
        </div>
      </div>

      <div class="bg-white rounded-md border border-gray-200 p-4">
        <h3 class="font-semibold text-gray-800 mb-3 flex items-center">
          <span class="w-2 h-2 rounded-full bg-purple-500 mr-2"></span>
          社交登录
          <span class="ml-auto text-xs text-gray-400">{{ (drivers.social ?? []).length }}</span>
        </h3>
        <div class="flex flex-wrap gap-2">
          <span
            v-for="d in (drivers.social ?? [])"
            :key="`so-${d}`"
            class="px-2 py-1 text-xs bg-purple-50 text-purple-700 border border-purple-200 rounded"
          >
            {{ prettyName(d) }}
          </span>
          <span v-if="!(drivers.social ?? []).length" class="text-xs text-gray-400">无</span>
        </div>
      </div>

      <div class="bg-white rounded-md border border-gray-200 p-4">
        <h3 class="font-semibold text-gray-800 mb-3 flex items-center">
          <span class="w-2 h-2 rounded-full bg-rose-500 mr-2"></span>
          图片扫描
          <span class="ml-auto text-xs text-gray-400">{{ (drivers.scan ?? []).length }}</span>
        </h3>
        <div class="flex flex-wrap gap-2">
          <span
            v-for="d in (drivers.scan ?? [])"
            :key="`sc-${d}`"
            class="px-2 py-1 text-xs bg-rose-50 text-rose-700 border border-rose-200 rounded"
          >
            {{ prettyName(d) }}
          </span>
          <span v-if="!(drivers.scan ?? []).length" class="text-xs text-gray-400">无</span>
        </div>
      </div>

      <div class="bg-white rounded-md border border-gray-200 p-4">
        <h3 class="font-semibold text-gray-800 mb-3 flex items-center">
          <span class="w-2 h-2 rounded-full bg-cyan-500 mr-2"></span>
          图片处理
          <span class="ml-auto text-xs text-gray-400">{{ (drivers.process ?? []).length }}</span>
        </h3>
        <div class="flex flex-wrap gap-2">
          <span
            v-for="d in (drivers.process ?? [])"
            :key="`pr-${d}`"
            class="px-2 py-1 text-xs bg-cyan-50 text-cyan-700 border border-cyan-200 rounded"
          >
            {{ prettyName(d) }}
          </span>
          <span v-if="!(drivers.process ?? []).length" class="text-xs text-gray-400">无</span>
        </div>
      </div>
    </div>

    <div class="mt-8 grid grid-cols-1 lg:grid-cols-2 gap-4">
      <div class="bg-white rounded-md border border-gray-200 p-4">
        <h3 class="font-semibold text-gray-800 mb-3">短信测试</h3>
        <p class="text-xs text-gray-500 mb-3">通过短信验证码接口触发一次发送（请确保已配置 SMS provider）</p>
        <div class="flex gap-2">
          <input
            v-model="smsTest.to"
            placeholder="手机号"
            class="flex-1 px-3 py-2 border border-gray-300 rounded-md text-sm"
          />
          <AppButton :loading="smsBusy" @click="testSMS">发送</AppButton>
        </div>
      </div>
      <div class="bg-white rounded-md border border-gray-200 p-4">
        <h3 class="font-semibold text-gray-800 mb-3">邮件测试</h3>
        <p class="text-xs text-gray-500 mb-3">通过邮件验证码接口触发一次发送</p>
        <div class="flex gap-2">
          <input
            v-model="mailTest.to"
            placeholder="收件人邮箱"
            class="flex-1 px-3 py-2 border border-gray-300 rounded-md text-sm"
          />
          <AppButton :loading="mailBusy" @click="testMail">发送</AppButton>
        </div>
      </div>
    </div>
  </div>
</template>
