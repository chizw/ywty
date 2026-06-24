<script setup lang="ts">
// 管理后台：存储策略管理（增删改查�?definePageMeta({ layout: 'admin', middleware: 'admin' })

import { Plus, Pencil, Trash2 } from '@lucide/vue'

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

function closeDialog() {
  showCreate.value = false
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
    message.error('options 必须是合法JSON: ' + e.message)
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
      message.success('已更�?)
    } else {
      await api.post('/api/v1/admin/storage/create', body)
      message.success('已创�?)
    }
    showCreate.value = false
    refresh()
  } catch (err: any) {
    message.error(err?.statusMessage || '保存失败')
  }
}

const confirmRow = ref<any>(null)
function askRemove(row: any) {
  confirmRow.value = row
}
function closeConfirm() {
  confirmRow.value = null
}
async function doRemove() {
  if (!confirmRow.value) return
  try {
    await api.del(`/api/v1/admin/storage/delete/${confirmRow.value.id}`)
    message.success('已删�?)
    confirmRow.value = null
    refresh()
  } catch (err: any) {
    message.error(err?.statusMessage || '删除失败')
  }
}

function providerLabel(name: string) {
  const map: Record<string, string> = {
    local: '本地',
    s3: 'AWS S3',
    oss: '阿里云OSS',
    cos: '腾讯云COS',
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
      <h1 class="text-2xl font-bold text-foreground">存储策略</h1>
      <Button @click="openCreate">
        <Plus class="h-4 w-4 mr-2" />
        新建策略
      </Button>
    </div>

    <p class="text-sm text-muted-foreground mb-4">
      系统支持 {{ drivers.length }} 种存储驱动。配置后将保存到 <code class="text-xs bg-muted px-1 py-0.5 rounded">storages</code> 表，
      供用户上传、跨存储复制等场景按策略 ID 调用�?    </p>

    <div v-if="rows.length === 0" class="border border-dashed border-border rounded-lg p-8 text-center text-sm text-muted-foreground">
      暂无存储策略，点击右上角"新建策略"开始�?    </div>
    <Card v-else>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead class="w-16">ID</TableHead>
            <TableHead>名称</TableHead>
            <TableHead>驱动</TableHead>
            <TableHead>前缀</TableHead>
            <TableHead>说明</TableHead>
            <TableHead class="text-right w-32">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-for="row in rows" :key="row.id">
            <TableCell class="text-muted-foreground">#{{ row.id }}</TableCell>
            <TableCell class="font-medium">{{ row.name }}</TableCell>
            <TableCell>
              <Badge variant="outline">{{ providerLabel(row.provider) }}</Badge>
            </TableCell>
            <TableCell class="text-muted-foreground font-mono text-xs">{{ row.prefix || '-' }}</TableCell>
            <TableCell class="text-muted-foreground text-xs truncate max-w-xs">{{ row.intro || '-' }}</TableCell>
            <TableCell class="text-right space-x-1">
              <Button variant="ghost" size="sm" @click="openEdit(row)">
                <Pencil class="h-4 w-4" />
              </Button>
              <Button variant="ghost" size="sm" class="text-destructive hover:text-destructive" @click="askRemove(row)">
                <Trash2 class="h-4 w-4" />
              </Button>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </Card>

    <!-- 创建/编辑弹窗 -->
    <Dialog :open="showCreate" @update:open="(val: boolean) => { if (!val) closeDialog() }">
      <DialogContent class="max-w-lg">
        <DialogHeader>
          <DialogTitle>{{ editing ? '编辑存储策略' : '新建存储策略' }}</DialogTitle>
        </DialogHeader>
        <div class="space-y-3">
          <div>
            <Label class="mb-1.5 block">名称 *</Label>
            <Input v-model="form.name" placeholder="如：阿里云 OSS 主站" />
          </div>
          <div>
            <Label class="mb-1.5 block">驱动 *</Label>
            <Select :modelValue="form.provider" @update:modelValue="(val: string) => form.provider = val">
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem v-for="d in drivers" :key="d" :value="d">{{ providerLabel(d) }}（{{ d }}�?/SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div>
            <Label class="mb-1.5 block">路径前缀</Label>
            <Input v-model="form.prefix" placeholder="如：images/2026" />
          </div>
          <div>
            <Label class="mb-1.5 block">说明</Label>
            <Input v-model="form.intro" />
          </div>
          <div>
            <Label class="mb-1.5 block">Options（JSON�?/Label>
            <Textarea
              v-model="form.options_text"
              :rows="6"
              class="font-mono text-xs"
              placeholder='{"endpoint": "https://oss-cn-hangzhou.aliyuncs.com", "bucket": "my-bucket", "access_key": "***", "secret_key": "***"}'
            />
            <p class="mt-1 text-xs text-muted-foreground">不同驱动的 options 字段不同，参考驱动文档填写�?/p>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" @click="closeDialog">取消</Button>
          <Button @click="submit">{{ editing ? '保存' : '创建' }}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- 删除确认弹窗 -->
    <Dialog :open="!!confirmRow" @update:open="(val: boolean) => { if (!val) closeConfirm() }">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>确认删除</DialogTitle>
        </DialogHeader>
        <p class="text-sm text-muted-foreground">确定删除存储策略「{{ confirmRow?.name }}」？</p>
        <DialogFooter>
          <Button variant="outline" @click="closeConfirm">取消</Button>
          <Button variant="destructive" @click="doRemove">删除</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
