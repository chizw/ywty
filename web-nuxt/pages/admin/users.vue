<script setup lang="ts">
// 管理后台：用户管�?definePageMeta({ layout: 'admin', middleware: 'admin' })

const api = useApi()
const page = ref(1)
const perPage = 20
const keyword = ref('')
const { data, refresh } = await useAsyncData('admin-users', () =>
  api.get<any>('/api/v1/admin/users', {
    query: { page, per_page: perPage, keyword },
  })
)
const users = computed<any[]>(() => (data.value as any)?.data ?? [])
const meta = computed(() => (data.value as any)?.meta)

const groups = ref<any[]>([])
async function loadGroups() {
  try {
    const r: any = await api.get('/api/v1/admin/groups')
    groups.value = Array.isArray(r) ? r : (r as any)?.data ?? []
  } catch {
    groups.value = []
  }
}
onMounted(loadGroups)

const editing = ref<any | null>(null)
const form = reactive({ is_admin: false, status: 'normal', group_id: 0 })
const saving = ref(false)
const msg = ref('')

function openEdit(u: any) {
  editing.value = u
  form.is_admin = !!u.is_admin
  form.status = u.status || 'normal'
  form.group_id = 0
  msg.value = ''
}
function closeEdit() {
  editing.value = null
}

async function save() {
  if (!editing.value) return
  saving.value = true
  msg.value = ''
  try {
    await api.request(`/api/v1/admin/users/${editing.value.id}`, {
      method: 'PATCH',
      body: {
        is_admin: form.is_admin,
        status: form.status,
        group_id: form.group_id || undefined,
      },
    })
    msg.value = '已保�?
    closeEdit()
    refresh()
  } catch (err: any) {
    msg.value = err?.statusMessage || '保存失败'
  } finally {
    saving.value = false
  }
}

function fmtTime(s: any) {
  if (!s) return '-'
  const d = typeof s === 'number' ? new Date(s * 1000) : new Date(s)
  return d.toLocaleString()
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-4">
      <h1 class="text-2xl font-bold text-foreground">用户管理</h1>
      <span class="text-sm text-muted-foreground">�?{{ meta?.total ?? users.length }} 个用�?/span>
    </div>

    <div class="mb-4 flex gap-2">
      <Input
        v-model="keyword"
        placeholder="搜索用户�?邮箱/姓名"
        class="max-w-sm"
        @keyup.enter="() => refresh()"
      />
      <Button @click="refresh">搜索</Button>
    </div>

    <Card>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>ID</TableHead>
            <TableHead>用户�?/TableHead>
            <TableHead>邮箱</TableHead>
            <TableHead>姓名</TableHead>
            <TableHead>状�?/TableHead>
            <TableHead>管理�?/TableHead>
            <TableHead>注册时间</TableHead>
            <TableHead class="text-right">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-for="u in users" :key="u.id">
            <TableCell class="text-muted-foreground">{{ u.id }}</TableCell>
            <TableCell class="font-medium">{{ u.username }}</TableCell>
            <TableCell class="text-muted-foreground">{{ u.email }}</TableCell>
            <TableCell>{{ u.name || '-' }}</TableCell>
            <TableCell>
              <Badge :variant="u.status === 'normal' ? 'success' : 'secondary'">
                {{ u.status || 'normal' }}
              </Badge>
            </TableCell>
            <TableCell>
              <Badge :variant="u.is_admin ? 'default' : 'secondary'">
                {{ u.is_admin ? '�? : '�? }}
              </Badge>
            </TableCell>
            <TableCell class="text-muted-foreground text-xs">{{ fmtTime(u.created_at) }}</TableCell>
            <TableCell class="text-right">
              <Button variant="ghost" size="sm" @click="openEdit(u)">编辑</Button>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </Card>

    <div v-if="meta && meta.last_page > 1" class="mt-4 flex items-center justify-end gap-2">
      <Button variant="outline" size="sm" :disabled="page <= 1" @click="page--; refresh()">上一�?/Button>
      <span class="text-sm text-muted-foreground">�?{{ meta.current_page }} / {{ meta.last_page }} �?/span>
      <Button variant="outline" size="sm" :disabled="page >= meta.last_page" @click="page++; refresh()">下一�?/Button>
    </div>

    <!-- 编辑弹窗 -->
    <Dialog :open="!!editing" @update:open="(val: boolean) => { if (!val) closeEdit() }">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>编辑用户 #{{ editing?.id }}</DialogTitle>
        </DialogHeader>
        <div class="space-y-4">
          <div>
            <Label class="mb-1.5 block">状�?/Label>
            <Select :modelValue="form.status" @update:modelValue="(val: string) => form.status = val">
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="normal">正常</SelectItem>
                <SelectItem value="disabled">禁用</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div class="flex items-center gap-2">
            <Checkbox :checked="form.is_admin" @update:checked="(val: boolean) => form.is_admin = val" />
            <Label>设为管理�?/Label>
          </div>
          <div>
            <Label class="mb-1.5 block">角色�?/Label>
            <Select :modelValue="String(form.group_id)" @update:modelValue="(val: string) => form.group_id = Number(val)">
              <SelectTrigger>
                <SelectValue placeholder="不修�? />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="0">不修�?/SelectItem>
                <SelectItem v-for="g in groups" :key="g.id" :value="String(g.id)">{{ g.name }}</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <p v-if="msg" class="text-sm" :class="msg.includes('失败') ? 'text-destructive' : 'text-primary'">{{ msg }}</p>
        </div>
        <DialogFooter>
          <Button variant="outline" @click="closeEdit">取消</Button>
          <Button :loading="saving" @click="save">保存</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
