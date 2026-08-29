<script setup>
import { h, ref, watch } from 'vue';
import { NPagination, useMessage } from 'naive-ui';
import { get, errMsg } from '../api';

const props = defineProps({
  title: { type: String, required: true },
  description: { type: String, default: '' },
  listURL: { type: String, required: true },
  columns: { type: Array, required: true },
  actions: { type: Array, default: () => [] },
  searchable: { type: Boolean, default: true },
  searchPlaceholder: { type: String, default: '搜索…' },
  rowKey: { type: Function, default: (r) => r.id },
  extraHeader: { type: Function, default: null },
});

const msg = useMessage();
const rows = ref([]);
const loading = ref(false);
const q = ref('');
const page = ref(1);
const pageCount = ref(1);
const total = ref(0);

async function load() {
  loading.value = true;
  try {
    const sep = props.listURL.includes('?') ? '&' : '?';
    const res = await get(
      `${props.listURL}${sep}page=${page.value}&per_page=20${q.value ? '&q=' + encodeURIComponent(q.value) : ''}`,
    );
    const payload = res.data ?? res;
    rows.value = Array.isArray(payload) ? payload : (payload.data ?? []);
    total.value = payload.meta?.total ?? rows.value.length;
    pageCount.value = payload.meta?.last_page ?? 1;
  } catch (e) {
    msg.error(errMsg(e));
  } finally {
    loading.value = false;
  }
}

function reload() { load(); }
defineExpose({ reload });

watch(() => props.listURL, () => { page.value = 1; load(); });
load();

function runAction(a, row) {
  Promise.resolve(a.onClick(row, reload)).catch((e) => msg.error(errMsg(e)));
}
</script>

<template>
  <div>
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <h1 class="fi-page-title">{{ title }}</h1>
        <p v-if="description" class="mt-1 text-sm text-gray-500">{{ description }}</p>
      </div>
      <component :is="extraHeader" v-if="extraHeader" />
    </div>

    <div class="fi-card mt-4 overflow-hidden">
      <div v-if="searchable" class="flex flex-wrap items-center gap-2 border-b border-gray-200 px-4 py-3">
        <input v-model="q" class="fi-input !w-64" :placeholder="searchPlaceholder"
               @keyup.enter="page = 1; load()" />
        <button class="fi-btn-secondary" @click="page = 1; load()">搜索</button>
      </div>

      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th v-for="col in columns" :key="col.key" class="fi-th" :style="col.width ? 'width:' + col.width : ''">
                {{ col.label }}
              </th>
              <th v-if="actions.length" class="fi-th">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100">
            <tr v-if="loading">
              <td :colspan="columns.length + (actions.length ? 1 : 0)" class="fi-td text-center text-gray-400 py-8">
                加载中…
              </td>
            </tr>
            <tr v-else-if="!rows.length">
              <td :colspan="columns.length + (actions.length ? 1 : 0)" class="fi-td text-center text-gray-400 py-8">
                暂无数据
              </td>
            </tr>
            <template v-else>
              <tr v-for="row in rows" :key="rowKey(row)" class="hover:bg-gray-50">
                <td v-for="col in columns" :key="col.key" class="fi-td" :class="col.ellipsis ? 'max-w-[22rem] truncate' : ''">
                  <template v-if="col.render">
                    <component :is="() => col.render(row)" />
                  </template>
                  <template v-else>{{ row[col.key] ?? '' }}</template>
                </td>
                <td v-if="actions.length" class="fi-td">
                  <div class="flex gap-1.5">
                    <button v-for="a in actions" :key="a.label"
                            :class="a.type === 'danger' ? 'fi-btn-danger' : a.type === 'warning' ? 'fi-btn-secondary !bg-amber-100 !text-amber-700 hover:!bg-amber-200' : 'fi-btn-secondary'"
                            class="!px-2 !py-1 text-xs"
                            @click="runAction(a, row)">
                      {{ a.label }}
                    </button>
                  </div>
                </td>
              </tr>
            </template>
          </tbody>
        </table>
      </div>

      <div class="flex items-center justify-between border-t border-gray-200 px-4 py-3">
        <span class="text-xs text-gray-500">共 {{ total }} 条</span>
        <NPagination :page="page" :page-count="pageCount" size="small"
                     @update:page="(v) => { page = v; load(); }" />
      </div>
    </div>
  </div>
</template>
