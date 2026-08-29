import { createApp, h, ref, onMounted } from 'vue';
import { createRouter, createWebHistory, RouterView } from 'vue-router';
import {
  NConfigProvider, NLayout, NMenu, NCard, NDataTable, NButton, NInput,
  NForm, NFormItem, NSpace, NTag, NMessageProvider, useMessage,
  NGrid, NGridItem, NStatistic, NPopconfirm, NPagination, zhCN, dateZhCN,
} from 'naive-ui';
import { login as apiLogin, get, put, del, errMsg } from './api';
import { listPage } from './pages';

// ---------- 通用页面 ----------

const Dashboard = {
  setup() {
    const stats = ref(null);
    onMounted(async () => {
      try { stats.value = (await get('/dashboard')).data; } catch (e) { /* 首次加载失败留空 */ }
    });
    return () => h(NCard, { title: '仪表盘' }, () => stats.value
      ? h(NGrid, { cols: 4, 'x-gap': 12, 'y-gap': 12 }, () => [
          ['用户', stats.value.user_count], ['图片', stats.value.photo_count],
          ['相册', stats.value.album_count], ['已用空间(KB)', stats.value.photo_size_kb],
          ['已支付订单', stats.value.order_count], ['收入(分)', stats.value.revenue_fen],
          ['待处理工单', stats.value.ticket_open], ['待处理举报', stats.value.report_open],
        ].map(([label, v]) => h(NGridItem, {}, () => h(NStatistic, { label }, () => String(v ?? 0)))))
      : '加载中…');
  },
};

const Users = listPage({
  title: '用户', listURL: '/users',
  columns: [
    { title: 'ID', key: 'id', width: 70 },
    { title: '用户名', key: 'username' },
    { title: '昵称', key: 'name' },
    { title: '邮箱', key: 'email' },
    { title: '图片数', key: 'photo_count', width: 80 },
    { title: '管理员', key: 'is_admin', width: 80, render: (r) => (r.is_admin ? '是' : '否') },
    { title: '状态', key: 'status', width: 90, render: (r) => h(NTag, { type: r.status === 'normal' ? 'success' : 'error' }, () => r.status) },
    { title: '注册时间', key: 'created_at', width: 170, render: (r) => (r.created_at || '').replace('T', ' ').slice(0, 19) },
  ],
  actions: [
    { label: '冻结/解冻', type: 'warning', confirm: '切换该用户的冻结状态？', onClick: async (row, reload) => { await put('/users/' + row.id, { status: row.status === 'normal' ? 'frozen' : 'normal' }); reload(); } },
    { label: '删除', type: 'error', confirm: '确定删除该用户？', onClick: async (row, reload) => { await del('/users/' + row.id); reload(); } },
  ],
});

const Photos = listPage({
  title: '图片', listURL: '/photos',
  columns: [
    { title: 'ID', key: 'id', width: 70 },
    { title: '名称', key: 'name' },
    { title: '路径', key: 'pathname', ellipsis: { tooltip: true } },
    { title: '大小(KB)', key: 'size', width: 90, render: (r) => Number(r.size).toFixed(1) },
    { title: '尺寸', key: 'dims', width: 100, render: (r) => `${r.width}x${r.height}` },
    { title: '状态', key: 'status', width: 90 },
    { title: '时间', key: 'created_at', width: 170, render: (r) => (r.created_at || '').replace('T', ' ').slice(0, 19) },
  ],
  actions: [
    { label: '标记违规', type: 'warning', confirm: '将该图片标记为违规？', onClick: async (row, reload) => { await put(`/photos/${row.id}/status`, { status: 'violation' }); reload(); } },
    { label: '恢复正常', confirm: '恢复正常状态？', onClick: async (row, reload) => { await put(`/photos/${row.id}/status`, { status: 'normal' }); reload(); } },
    { label: '删除', type: 'error', confirm: '删除会同时清理物理文件，确认？', onClick: async (row, reload) => { await del('/photos/' + row.id); reload(); } },
  ],
});

const Notices = listPage({
  title: '公告', listURL: '/notices',
  columns: [
    { title: 'ID', key: 'id', width: 70 },
    { title: '标题', key: 'title' },
    { title: '内容', key: 'content', ellipsis: { tooltip: true } },
    { title: '排序', key: 'sort', width: 70 },
  ],
  actions: [
    { label: '删除', type: 'error', confirm: '确认删除公告？', onClick: async (row, reload) => { await del('/notices/' + row.id); reload(); } },
  ],
});

const Pages = listPage({
  title: '自定义页面', listURL: '/pages',
  columns: [
    { title: 'ID', key: 'id', width: 70 },
    { title: '名称', key: 'name' },
    { title: '标题', key: 'title' },
    { title: 'Slug', key: 'slug' },
    { title: '类型', key: 'type', width: 90 },
    { title: '显示', key: 'is_show', width: 70, render: (r) => (r.is_show ? '是' : '否') },
  ],
  actions: [
    { label: '删除', type: 'error', confirm: '确认删除页面？', onClick: async (row, reload) => { await del('/pages/' + row.id); reload(); } },
  ],
});

const Orders = listPage({
  title: '订单', listURL: '/orders',
  columns: [
    { title: '订单号', key: 'trade_no', width: 190 },
    { title: '金额(分)', key: 'amount', width: 90 },
    { title: '抵扣(分)', key: 'deduct_amount', width: 90 },
    { title: '支付方式', key: 'pay_method', width: 100 },
    { title: '状态', key: 'status', width: 100, render: (r) => h(NTag, { type: r.status === 'paid' ? 'success' : r.status === 'cancelled' ? 'default' : 'warning' }, () => r.status) },
    { title: '时间', key: 'created_at', width: 170, render: (r) => (r.created_at || '').replace('T', ' ').slice(0, 19) },
  ],
});

const Coupons = listPage({
  title: '优惠码', listURL: '/coupons',
  columns: [
    { title: 'ID', key: 'id', width: 70 },
    { title: '名称', key: 'name' },
    { title: '券码', key: 'code' },
    { title: '类型', key: 'type', width: 90 },
    { title: '面值/折扣', key: 'value', width: 100 },
    { title: '限用次数', key: 'usage_limit', width: 90 },
  ],
  actions: [
    { label: '删除', type: 'error', confirm: '确认删除优惠码？', onClick: async (row, reload) => { await del('/coupons/' + row.id); reload(); } },
  ],
});

const Tickets = listPage({
  title: '工单', listURL: '/tickets',
  columns: [
    { title: '工单号', key: 'issue_no', width: 190 },
    { title: '标题', key: 'title' },
    { title: '用户', key: 'username', width: 120 },
    { title: '级别', key: 'level', width: 80 },
    { title: '状态', key: 'status', width: 110, render: (r) => h(NTag, { type: r.status === 'completed' ? 'success' : 'warning' }, () => r.status) },
  ],
});

const Reports = listPage({
  title: '举报', listURL: '/reports',
  columns: [
    { title: 'ID', key: 'id', width: 70 },
    { title: '类型', key: 'reportable_type', width: 180 },
    { title: '目标ID', key: 'reportable_id', width: 80 },
    { title: '内容', key: 'content', ellipsis: { tooltip: true } },
    { title: '状态', key: 'status', width: 100 },
  ],
  actions: [
    { label: '标记已处理', confirm: '标记为已处理？', onClick: async (row, reload) => { await put('/reports/' + row.id, {}); reload(); } },
  ],
});

const Groups = listPage({
  title: '角色组', listURL: '/groups', searchable: false,
  columns: [
    { title: 'ID', key: 'id', width: 70 },
    { title: '名称', key: 'name' },
    { title: '默认组', key: 'is_default', width: 80, render: (r) => (r.is_default ? '是' : '否') },
    { title: '游客组', key: 'is_guest', width: 80, render: (r) => (r.is_guest ? '是' : '否') },
  ],
});

const Storages = listPage({
  title: '储存策略', listURL: '/storages',
  columns: [
    { title: 'ID', key: 'id', width: 70 },
    { title: '名称', key: 'name' },
    { title: '前缀', key: 'prefix' },
    { title: '提供者', key: 'provider', width: 110 },
    { title: '配置', key: 'options', ellipsis: { tooltip: true }, render: (r) => JSON.stringify(r.options) },
  ],
  actions: [
    { label: '删除', type: 'error', confirm: '确认删除储存策略？', onClick: async (row, reload) => { await del('/storages/' + row.id); reload(); } },
  ],
});

const Drivers = listPage({
  title: '驱动', listURL: '/drivers',
  columns: [
    { title: 'ID', key: 'id', width: 70 },
    { title: '类型', key: 'type', width: 110 },
    { title: '名称', key: 'name' },
    { title: '配置', key: 'options', ellipsis: { tooltip: true }, render: (r) => JSON.stringify(r.options) },
  ],
  actions: [
    { label: '删除', type: 'error', confirm: '确认删除驱动？', onClick: async (row, reload) => { await del('/drivers/' + row.id); reload(); } },
  ],
});

const Settings = {
  setup() {
    const msg = useMessage();
    const items = ref([]);
    const edits = ref({});
    onMounted(async () => {
      try {
        const res = await get('/settings');
        items.value = res.data.settings.map((s) => ({ ...s, payloadText: JSON.stringify(s.payload ?? null) }));
      } catch (e) { msg.error(errMsg(e)); }
    });
    async function save() {
      try {
        const updates = Object.entries(edits.value).map(([idx, text]) => {
          const it = items.value[Number(idx)];
          return { group: it.group, name: it.name, payload: JSON.parse(text) };
        });
        await put('/settings', { updates });
        msg.success('已保存');
        edits.value = {};
      } catch (e) { msg.error(errMsg(e)); }
    }
    return () => h(NCard, { title: '系统设置' }, () => [
      h(NDataTable, {
        columns: [
          { title: '组', key: 'group', width: 90 },
          { title: '名称', key: 'name', width: 240 },
          { title: '值', key: 'payload', render: (row) => h(NInput, {
              defaultValue: row.payloadText,
              'onUpdate:value': (v) => { edits.value[items.value.indexOf(row)] = v; },
              type: 'textarea', autosize: { minRows: 1, maxRows: 4 },
            }) },
        ],
        data: items.value,
        rowKey: (r) => r.group + '.' + r.name,
        size: 'small',
      }),
      h(NSpace, { style: 'margin-top:12px' }, () => h(NButton, { type: 'primary', onClick: save }, () => '保存修改')),
    ]);
  },
};

// ---------- 登录 ----------

const Login = {
  setup() {
    const msg = useMessage();
    const username = ref('');
    const password = ref('');
    const loading = ref(false);
    async function submit() {
      loading.value = true;
      try {
        await apiLogin(username.value, password.value);
        localStorage.setItem('admin_logged_in', '1');
        window.location.href = '/admin/dashboard';
      } catch (e) {
        msg.error(errMsg(e));
      } finally { loading.value = false; }
    }
    return () => h('div', { style: 'display:flex;justify-content:center;align-items:center;height:100vh' },
      [h(NCard, { title: 'ywty 管理后台', style: 'width:360px' }, () =>
        h(NForm, { onSubmit: (e) => { e.preventDefault(); submit(); } }, () => [
          h(NFormItem, { label: '用户名' }, () => h(NInput, { value: username.value, 'onUpdate:value': (v) => (username.value = v) })),
          h(NFormItem, { label: '密码' }, () => h(NInput, { value: password.value, type: 'password', 'onUpdate:value': (v) => (password.value = v), 'show-password-on': 'click' })),
          h(NButton, { type: 'primary', block: true, loading: loading.value, attrType: 'submit', onClick: submit }, () => '登录'),
        ]))]);
  },
};

// ---------- 路由与布局 ----------

const routes = [
  { path: '/login', component: Login },
  { path: '/dashboard', component: Dashboard },
  { path: '/users', component: Users },
  { path: '/photos', component: Photos },
  { path: '/notices', component: Notices },
  { path: '/pages', component: Pages },
  { path: '/orders', component: Orders },
  { path: '/coupons', component: Coupons },
  { path: '/tickets', component: Tickets },
  { path: '/reports', component: Reports },
  { path: '/groups', component: Groups },
  { path: '/storages', component: Storages },
  { path: '/drivers', component: Drivers },
  { path: '/settings', component: Settings },
  { path: '/', redirect: '/dashboard' },
  { path: '/:pathMatch(.*)*', redirect: '/dashboard' },
];

const router = createRouter({ history: createWebHistory('/admin/'), routes });

const menuOptions = [
  { label: '仪表盘', key: '/dashboard' },
  { label: '用户', key: '/users' },
  { label: '图片', key: '/photos' },
  { label: '角色组', key: '/groups' },
  { label: '订单', key: '/orders' },
  { label: '优惠码', key: '/coupons' },
  { label: '工单', key: '/tickets' },
  { label: '举报', key: '/reports' },
  { label: '公告', key: '/notices' },
  { label: '页面', key: '/pages' },
  { label: '储存策略', key: '/storages' },
  { label: '驱动', key: '/drivers' },
  { label: '设置', key: '/settings' },
];

const Layout = {
  setup() {
    return () => h(NLayout, { style: 'height:100vh' }, () =>
      h('div', { style: 'display:flex;height:100%' }, [
        h('aside', { style: 'width:200px;border-right:1px solid #efeff5;padding:12px' }, [
          h('div', { style: 'font-weight:600;margin-bottom:12px' }, 'ywty 管理后台'),
          h(NMenu, { value: router.currentRoute.value.path, options: menuOptions, 'onUpdate:value': (k) => router.push(k) }),
        ]),
        h('main', { style: 'flex:1;overflow:auto;padding:16px' }, [h(RouterView)]),
      ]));
  },
};

router.beforeEach((to) => {
  if (to.path === '/login') return true;
  if (localStorage.getItem('admin_logged_in') === '1') return true;
  return '/login';
});

const app = createApp(() =>
  h(NConfigProvider, { locale: zhCN, dateLocale: dateZhCN }, () =>
    h(NMessageProvider, {}, () => h(RouterView))));
app.use(router);
app.mount('#admin');
