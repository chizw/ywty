<script setup lang="ts">
// 管理后台：用户管理
definePageMeta({ layout: 'admin', middleware: 'admin' })

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
    msg.value = '已保存'
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
      <h1 class="text-2xl font-bold text-gray-900">用户管理</h1>
      <span class="text-sm text-gray-500">共 {{ meta?.total ?? users.length }} 个用户</span>
    </div>

    <div class="mb-4 flex gap-2">
      <input
        v-model="keyword"
        placeholder="搜索用户名/邮箱/姓名"
        class="flex-1 max-w-sm px-3 py-2 border border-gray-300 rounded-md"
        @keyup.enter="refresh"
      />
      <AppButton @click="refresh">搜索</AppButton>
    </div>

    <div class="bg-white border border-gray-200 rounded-lg overflow-hidden">
      <table class="w-full text-sm">
        <thead class="bg-gray-50 text-gray-600 text-left">
          <tr>
            <th class="px-4 py-2">ID</th>
            <th class="px-4 py-2">用户名</th>
            <th class="px-4 py-2">邮箱</th>
            <th class="px-4 py-2">姓名</th>
            <th class="px-4 py-2">状态</th>
            <th class="px-4 py-2">管理员</th>
            <th class="px-4 py-2">注册时间</th>
            <th class="px-4 py-2 text-right">操作</th>
          </tr>
        </thead>
        <tbody class="divide-y">
          <tr v-for="u in users" :key="u.id" class="hover:bg-gray-50">
            <td class="px-4 py-2 text-gray-500">{{ u.id }}</td>
            <td class="px-4 py-2 font-medium">{{ u.username }}</td>
            <td class="px-4 py-2 text-gray-600">{{ u.email }}</td>
            <td class="px-4 py-2">{{ u.name || '-' }}</td>
            <td class="px-4 py-2">
              <span
                class="px-2 py-0.5 text-xs rounded-full"
                :class="u.status === 'normal' ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-600'"
              >{{ u.status || 'normal' }}</span>
            </td>
            <td class="px-4 py-2">
              <span
                class="px-2 py-0.5 text-xs rounded-full"
                :class="u.is_admin ? 'bg-primary-100 text-primary-700' : 'bg-gray-100 text-gray-500'"
              >{{ u.is_admin ? '是' : '否' }}</span>
            </td>
            <td class="px-4 py-2 text-gray-500 text-xs">{{ fmtTime(u.created_at) }}</td>
            <td class="px-4 py-2 text-right">
              <button class="text-primary-600 hover:underline text-xs" @click="openEdit(u)">编辑</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="meta && meta.last_page > 1" class="mt-4 flex items-center justify-end gap-2">
      <button
        :disabled="page <= 1"
        class="px-3 py-1 text-sm border border-gray-300 rounded disabled:opacity-50"
        @click="page--; refresh()"
      >上一页</button>
      <span class="text-sm text-gray-500">第 {{ meta.current_page }} / {{ meta.last_page }} 页</span>
      <button
        :disabled="page >= meta.last_page"
        class="px-3 py-1 text-sm border border-gray-300 rounded disabled:opacity-50"
        @click="page++; refresh()"
      >下一页</button>
    </div>

    <!-- 编辑弹窗 -->
    <div
      v-if="editing"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4"
      @click.self="closeEdit"
    >
      <div class="w-full max-w-md bg-white rounded-lg p-5 space-y-3">
        <h3 class="text-lg font-semibold">编辑用户 #{{ editing.id }}</h3>
        <div>
          <label class="block text-sm text-gray-700 mb-1">状态</label>
          <select v-model="form.status" class="w-full px-3 py-2 border border-gray-300 rounded-md">
            <option value="normal">正常</option>
            <option value="disabled">禁用</option>
          </select>
        </div>
        <label class="flex items-center gap-2 text-sm">
          <input v-model="form.is_admin" type="checkbox" />
          设为管理员
        </label>
        <div>
          <label class="block text-sm text-gray-700 mb-1">角色组</label>
          <select v-model.number="form.group_id" class="w-full px-3 py-2 border border-gray-300 rounded-md">
            <option :value="0">不修改</option>
            <option v-for="g in groups" :key="g.id" :value="g.id">{{ g.name }}</option>
          </select>
        </div>
        <p v-if="msg" class="text-sm" :class="msg.includes('失败') ? 'text-red-500' : 'text-primary-600'">{{ msg }}</p>
        <div class="flex justify-end gap-2 pt-2">
          <button class="px-3 py-1.5 text-sm border border-gray-300 rounded-md" @click="closeEdit">取消</button>
          <AppButton :loading="saving" @click="save">保存</AppButton>
        </div>
      </div>
    </div>
  </div>
</template>
