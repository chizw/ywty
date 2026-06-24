<script setup lang="ts">
// 管理后台：角色组管理
definePageMeta({ layout: 'admin', middleware: 'admin' })

import { Plus, Trash2 } from '@lucide/vue'

const api = useApi()
const { data, refresh } = await useAsyncData('admin-groups', () => api.get<any>('/api/v1/admin/groups'))
const groups = computed<any[]>(() => {
  const d = data.value as any
  return Array.isArray(d) ? d : (d as any)?.data ?? []
})

const showCreate = ref(false)
const newGroup = reactive({ name: '', intro: '', is_default: false })
const msg = ref('')
const saving = ref(false)

async function create() {
  if (!newGroup.name) return
  saving.value = true
  msg.value = ''
  try {
    await api.post('/api/v1/admin/groups', newGroup)
    msg.value = '已创�?
    newGroup.name = ''
    newGroup.intro = ''
    newGroup.is_default = false
    showCreate.value = false
    refresh()
  } catch (err: any) {
    msg.value = err?.statusMessage || '创建失败'
  } finally {
    saving.value = false
  }
}

const editing = ref<any | null>(null)
const editForm = reactive({ name: '', intro: '', is_default: false })
function openEdit(g: any) {
  editing.value = g
  editForm.name = g.name
  editForm.intro = g.intro || ''
  editForm.is_default = !!g.is_default
}
function closeEdit() { editing.value = null }

async function saveEdit() {
  if (!editing.value) return
  saving.value = true
  msg.value = ''
  try {
    await api.request(`/api/v1/admin/groups/${editing.value.id}`, {
      method: 'PATCH',
      body: { name: editForm.name, intro: editForm.intro, is_default: editForm.is_default },
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

const confirmId = ref<number | null>(null)
function askRemove(id: number) {
  confirmId.value = id
}
function closeConfirm() {
  confirmId.value = null
}
async function doRemove() {
  if (confirmId.value == null) return
  try {
    await api.del(`/api/v1/admin/groups/${confirmId.value}`)
    confirmId.value = null
    refresh()
  } catch (err: any) {
    alert(err?.statusMessage || '删除失败')
  }
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-4">
      <h1 class="text-2xl font-bold text-foreground">角色�?/h1>
      <Button @click="showCreate = !showCreate">
        <Plus v-if="!showCreate" class="h-4 w-4 mr-2" />
        {{ showCreate ? '取消' : '新建角色�? }}
      </Button>
    </div>

    <Card v-if="showCreate" class="mb-6">
      <CardContent class="p-4 space-y-3">
        <Input v-model="newGroup.name" placeholder="角色组名" />
        <Textarea v-model="newGroup.intro" placeholder="介绍（可选）" :rows="2" />
        <div class="flex items-center gap-2">
          <Checkbox :checked="newGroup.is_default" @update:checked="(val: boolean) => newGroup.is_default = val" />
          <Label>注册时默认使�?/Label>
        </div>
        <Button :loading="saving" @click="create">创建</Button>
        <p v-if="msg" class="text-sm" :class="msg.includes('失败') ? 'text-destructive' : 'text-primary'">{{ msg }}</p>
      </CardContent>
    </Card>

    <div v-if="groups.length === 0" class="text-sm text-muted-foreground">暂无角色�?/div>
    <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
      <Card v-for="g in groups" :key="g.id">
        <CardContent class="p-4">
          <div class="flex items-center justify-between">
            <h3 class="font-medium text-foreground">{{ g.name }}</h3>
            <div class="flex gap-1">
              <Badge v-if="g.is_default" variant="default">默认</Badge>
              <Badge v-if="g.is_guest" variant="secondary">游客</Badge>
            </div>
          </div>
          <p v-if="g.intro" class="mt-1 text-sm text-muted-foreground line-clamp-2">{{ g.intro }}</p>
          <div class="mt-3 flex gap-2">
            <Button variant="outline" size="sm" @click="openEdit(g)">编辑</Button>
            <Button variant="ghost" size="sm" class="text-destructive hover:text-destructive" @click="askRemove(g.id)">
              <Trash2 class="h-4 w-4" />
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>

    <!-- 编辑弹窗 -->
    <Dialog :open="!!editing" @update:open="(val: boolean) => { if (!val) closeEdit() }">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>编辑角色�?#{{ editing?.id }}</DialogTitle>
        </DialogHeader>
        <div class="space-y-3">
          <Input v-model="editForm.name" />
          <Textarea v-model="editForm.intro" :rows="2" />
          <div class="flex items-center gap-2">
            <Checkbox :checked="editForm.is_default" @update:checked="(val: boolean) => editForm.is_default = val" />
            <Label>注册时默认使�?/Label>
          </div>
          <p v-if="msg" class="text-sm" :class="msg.includes('失败') ? 'text-destructive' : 'text-primary'">{{ msg }}</p>
        </div>
        <DialogFooter>
          <Button variant="outline" @click="closeEdit">取消</Button>
          <Button :loading="saving" @click="saveEdit">保存</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- 删除确认弹窗 -->
    <Dialog :open="confirmId != null" @update:open="(val: boolean) => { if (!val) closeConfirm() }">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>确认删除</DialogTitle>
        </DialogHeader>
        <p class="text-sm text-muted-foreground">确定删除该角色组？相关用户角色绑定也会被解除�?/p>
        <DialogFooter>
          <Button variant="outline" @click="closeConfirm">取消</Button>
          <Button variant="destructive" @click="doRemove">删除</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
