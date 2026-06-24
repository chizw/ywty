<script setup lang="ts">
// 管理后台：角色组管理
definePageMeta({ layout: 'admin', middleware: 'admin' })

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
    msg.value = '已创建'
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
    msg.value = '已保存'
    closeEdit()
    refresh()
  } catch (err: any) {
    msg.value = err?.statusMessage || '保存失败'
  } finally {
    saving.value = false
  }
}

async function remove(id: number) {
  if (!confirm('确定删除该角色组？相关用户角色绑定也会被解除')) return
  try {
    await api.del(`/api/v1/admin/groups/${id}`)
    refresh()
  } catch (err: any) {
    alert(err?.statusMessage || '删除失败')
  }
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-4">
      <h1 class="text-2xl font-bold text-gray-900">角色组</h1>
      <AppButton @click="showCreate = !showCreate">{{ showCreate ? '取消' : '新建角色组' }}</AppButton>
    </div>

    <div v-if="showCreate" class="mb-6 p-4 bg-white border border-gray-200 rounded-lg space-y-3">
      <input v-model="newGroup.name" placeholder="角色组名" class="w-full px-3 py-2 border border-gray-300 rounded-md" />
      <textarea v-model="newGroup.intro" placeholder="介绍（可选）" rows="2" class="w-full px-3 py-2 border border-gray-300 rounded-md" />
      <label class="flex items-center gap-2 text-sm">
        <input v-model="newGroup.is_default" type="checkbox" />
        注册时默认使用
      </label>
      <AppButton :loading="saving" @click="create">创建</AppButton>
      <p v-if="msg" class="text-sm" :class="msg.includes('失败') ? 'text-red-500' : 'text-primary-600'">{{ msg }}</p>
    </div>

    <div v-if="groups.length === 0" class="text-sm text-gray-500">暂无角色组</div>
    <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
      <div
        v-for="g in groups"
        :key="g.id"
        class="bg-white border border-gray-200 rounded-lg p-4"
      >
        <div class="flex items-center justify-between">
          <h3 class="font-medium text-gray-900">{{ g.name }}</h3>
          <div class="flex gap-1">
            <span v-if="g.is_default" class="text-[10px] px-1.5 py-0.5 bg-primary-100 text-primary-700 rounded">默认</span>
            <span v-if="g.is_guest" class="text-[10px] px-1.5 py-0.5 bg-gray-100 text-gray-600 rounded">游客</span>
          </div>
        </div>
        <p v-if="g.intro" class="mt-1 text-sm text-gray-500 line-clamp-2">{{ g.intro }}</p>
        <div class="mt-3 flex gap-2">
          <button class="px-2 py-1 text-xs border border-gray-300 rounded hover:bg-gray-50" @click="openEdit(g)">编辑</button>
          <button class="px-2 py-1 text-xs text-red-500" @click="remove(g.id)">删除</button>
        </div>
      </div>
    </div>

    <!-- 编辑弹窗 -->
    <div
      v-if="editing"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4"
      @click.self="closeEdit"
    >
      <div class="w-full max-w-md bg-white rounded-lg p-5 space-y-3">
        <h3 class="text-lg font-semibold">编辑角色组 #{{ editing.id }}</h3>
        <input v-model="editForm.name" class="w-full px-3 py-2 border border-gray-300 rounded-md" />
        <textarea v-model="editForm.intro" rows="2" class="w-full px-3 py-2 border border-gray-300 rounded-md" />
        <label class="flex items-center gap-2 text-sm">
          <input v-model="editForm.is_default" type="checkbox" />
          注册时默认使用
        </label>
        <p v-if="msg" class="text-sm" :class="msg.includes('失败') ? 'text-red-500' : 'text-primary-600'">{{ msg }}</p>
        <div class="flex justify-end gap-2 pt-2">
          <button class="px-3 py-1.5 text-sm border border-gray-300 rounded-md" @click="closeEdit">取消</button>
          <AppButton :loading="saving" @click="saveEdit">保存</AppButton>
        </div>
      </div>
    </div>
  </div>
</template>
