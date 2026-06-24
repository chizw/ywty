<script setup lang="ts">
// 管理后台：存储策略管理（增删改查）
definePageMeta({ layout: 'admin', middleware: 'admin' })

const api = useApi()
const message = useMessage()

const { data: driversResp } = await useAsyncData('storage-drivers-list', () =>
  api.get<any>('/api/v1/admin/storage/drivers')
)
const drivers = computed<string[]>(() => (driversResp.value as any)?.data?.drivers ?? [])

const { data, refresh } = await useAsyncData('admin-storages', () =>
  api.get<any>('/api/v1/admin/storage/list')
)
const rows = computed<any[]>(() => Array.isArray(data.value) ? data.value : (data.value as any)?.data ?? [])

const showCreate = ref(false)
const editing = ref<any>(null)
const form = reactive({
  name: '',
  provider: 'local',
  intro: '',
  prefix: '',
  options_text: '{}',
})

function openCreate() {
  editing.value = null
  form.name = ''
  form.provider = drivers.value[0] || 'local'
  form.intro = ''
  form.prefix = ''
  form.options_text = '{}'
  showCreate.value = true
}

function openEdit(row: any) {
  editing.value = row
  form.name = row.name
  form.provider = row.provider
  form.intro = row.intro || ''
  form.prefix = row.prefix || ''
  form.options_text = JSON.stringify(row.options ?? {}, null, 2)
  showCreate.value = true
}

async function submit() {
  if (!form.name.trim()) {
    message.warning('请填写名称')
    return
  }
  let options: any
  try {
    options = JSON.parse(form.options_text || '{}')
  } catch (e: any) {
    message.error('options 必须是合法 JSON: ' + e.message)
    return
  }
  const body = {
    name: form.name,
    provider: form.provider,
    intro: form.intro,
    prefix: form.prefix,
    options,
  }
  try {
    if (editing.value) {
      await api.patch(`/api/v1/admin/storage/update/${editing.value.id}`, body)
      message.success('已更新')
    } else {
      await api.post('/api/v1/admin/storage/create', body)
      message.success('已创建')
    }
    showCreate.value = false
    refresh()
  } catch (err: any) {
    message.error(err?.statusMessage || '保存失败')
  }
}

async function remove(row: any) {
  if (!confirm(`确定删除存储策略「${row.name}」？`)) return
  try {
    await api.del(`/api/v1/admin/storage/delete/${row.id}`)
    message.success('已删除')
    refresh()
  } catch (err: any) {
    message.error(err?.statusMessage || '删除失败')
  }
}

function providerLabel(name: string) {
  const map: Record<string, string> = {
    local: '本地',
    s3: 'AWS S3',
    oss: '阿里云 OSS',
    cos: '腾讯云 COS',
    qiniu: '七牛云',
    upyun: '又拍云',
    ftp: 'FTP',
    sftp: 'SFTP',
    webdav: 'WebDAV',
  }
  return map[name] || name
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-4">
      <h1 class="text-2xl font-bold text-gray-900">存储策略</h1>
      <AppButton @click="openCreate">+ 新建策略</AppButton>
    </div>

    <p class="text-sm text-gray-500 mb-4">
      系统支持 {{ drivers.length }} 种存储驱动。配置后将保存到 <code>storages</code> 表，
      供用户上传、跨存储复制等场景按策略 ID 调用。
    </p>

    <div v-if="rows.length === 0" class="bg-white border border-dashed border-gray-300 rounded p-8 text-center text-sm text-gray-500">
      暂无存储策略，点击右上角"新建策略"开始。
    </div>
    <div v-else class="bg-white rounded-md border border-gray-200 overflow-hidden">
      <table class="w-full text-sm">
        <thead class="bg-gray-50 text-gray-500 text-xs uppercase">
          <tr>
            <th class="px-4 py-2 text-left w-16">ID</th>
            <th class="px-4 py-2 text-left">名称</th>
            <th class="px-4 py-2 text-left">驱动</th>
            <th class="px-4 py-2 text-left">前缀</th>
            <th class="px-4 py-2 text-left">说明</th>
            <th class="px-4 py-2 text-right w-32">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="row in rows" :key="row.id" class="border-t border-gray-100">
            <td class="px-4 py-2 text-gray-500">#{{ row.id }}</td>
            <td class="px-4 py-2 font-medium text-gray-800">{{ row.name }}</td>
            <td class="px-4 py-2">
              <span class="px-2 py-0.5 text-xs bg-blue-50 text-blue-700 rounded">{{ providerLabel(row.provider) }}</span>
            </td>
            <td class="px-4 py-2 text-gray-500 font-mono text-xs">{{ row.prefix || '-' }}</td>
            <td class="px-4 py-2 text-gray-500 text-xs truncate max-w-xs">{{ row.intro || '-' }}</td>
            <td class="px-4 py-2 text-right space-x-1">
              <button class="px-2 py-1 text-xs text-primary-600 hover:underline" @click="openEdit(row)">编辑</button>
              <button class="px-2 py-1 text-xs text-red-600 hover:underline" @click="remove(row)">删除</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="showCreate" class="fixed inset-0 z-50 flex items-center justify-center bg-black/40" @click.self="showCreate = false">
      <div class="bg-white rounded-lg shadow-xl w-full max-w-lg p-6">
        <h3 class="text-lg font-semibold text-gray-800 mb-4">
          {{ editing ? '编辑存储策略' : '新建存储策略' }}
        </h3>
        <div class="space-y-3 text-sm">
          <div>
            <label class="block text-gray-600 mb-1">名称 *</label>
            <input v-model="form.name" class="w-full px-3 py-2 border border-gray-300 rounded-md" placeholder="如：阿里云 OSS 主站" />
          </div>
          <div>
            <label class="block text-gray-600 mb-1">驱动 *</label>
            <select v-model="form.provider" class="w-full px-3 py-2 border border-gray-300 rounded-md">
              <option v-for="d in drivers" :key="d" :value="d">{{ providerLabel(d) }}（{{ d }}）</option>
            </select>
          </div>
          <div>
            <label class="block text-gray-600 mb-1">路径前缀</label>
            <input v-model="form.prefix" class="w-full px-3 py-2 border border-gray-300 rounded-md" placeholder="如：images/2026" />
          </div>
          <div>
            <label class="block text-gray-600 mb-1">说明</label>
            <input v-model="form.intro" class="w-full px-3 py-2 border border-gray-300 rounded-md" />
          </div>
          <div>
            <label class="block text-gray-600 mb-1">Options（JSON）</label>
            <textarea
              v-model="form.options_text"
              rows="6"
              class="w-full px-3 py-2 border border-gray-300 rounded-md font-mono text-xs"
              placeholder='{"endpoint": "https://oss-cn-hangzhou.aliyuncs.com", "bucket": "my-bucket", "access_key": "***", "secret_key": "***"}'
            />
            <p class="mt-1 text-xs text-gray-400">不同驱动的 options 字段不同，参考驱动文档填写。</p>
          </div>
        </div>
        <div class="mt-5 flex justify-end gap-2">
          <button class="px-3 py-1.5 text-sm text-gray-600 hover:text-gray-800" @click="showCreate = false">取消</button>
          <AppButton @click="submit">{{ editing ? '保存' : '创建' }}</AppButton>
        </div>
      </div>
    </div>
  </div>
</template>
