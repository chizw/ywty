// listPage：配置驱动的通用列表页（表格 + 搜索 + 分页 + 行操作）。
import { h, ref, onMounted } from 'vue';
import { NCard, NDataTable, NButton, NInput, NSpace, NPopconfirm, NPagination, useMessage } from 'naive-ui';
import { get, errMsg } from './api';

export function listPage({ title, listURL, columns, searchable = true, actions = null, rowKey = (r) => r.id }) {
  return {
    setup() {
      const msg = useMessage();
      const data = ref([]);
      const loading = ref(false);
      const q = ref('');
      const page = ref(1);
      const pageCount = ref(1);

      async function load() {
        loading.value = true;
        try {
          const sep = listURL.includes('?') ? '&' : '?';
          const res = await get(`${listURL}${sep}page=${page.value}&per_page=20${q.value ? '&q=' + encodeURIComponent(q.value) : ''}`);
          const payload = res.data ?? res;
          data.value = Array.isArray(payload) ? payload : (payload.data ?? []);
          pageCount.value = payload.meta?.last_page ?? 1;
        } catch (e) {
          msg.error(errMsg(e));
        } finally {
          loading.value = false;
        }
      }
      onMounted(load);

      const cols = [...columns];
      if (actions && actions.length) {
        cols.push({
          title: '操作',
          key: '__actions',
          width: actions.length * 95,
          render(row) {
            return h(NSpace, {}, () => actions.map((a) =>
              h(NPopconfirm, { onPositiveClick: () => a.onClick(row, load).catch((e) => msg.error(errMsg(e))) }, {
                trigger: () => h(NButton, { size: 'small', type: a.type || 'default' }, () => a.label),
                default: () => a.confirm || '确认执行？',
              })));
          },
        });
      }

      return () => h(NCard, { title }, () => [
        searchable ? h(NSpace, { style: 'margin-bottom:12px' }, () => [
          h(NInput, { value: q.value, 'onUpdate:value': (v) => (q.value = v), placeholder: '搜索…', clearable: true, style: 'width:240px', 'on-keyup': (e) => { if (e.key === 'Enter') { page.value = 1; load(); } } }),
          h(NButton, { type: 'primary', onClick: () => { page.value = 1; load(); } }, () => '搜索'),
        ]) : null,
        h(NDataTable, { columns: cols, data: data.value, loading: loading.value, rowKey, size: 'small', scrollX: 'max-content' }),
        h(NPagination, { page: page.value, 'onUpdate:page': (v) => { page.value = v; load(); }, pageCount: pageCount.value, style: 'margin-top:12px;justify-content:flex-end' }),
      ]);
    },
  };
}
