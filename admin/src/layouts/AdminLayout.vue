<script setup>
import { h, ref } from 'vue';
import { useRoute, useRouter, RouterView } from 'vue-router';
import { NAvatar, NDropdown, NMenu, useMessage } from 'naive-ui';
import { get, put } from '../api';

const route = useRoute();
const router = useRouter();
const msg = useMessage();
const username = localStorage.getItem('admin_username') || 'admin';
const menuOpen = ref(false);

const groups = [
  {
    label: '常规',
    items: [
      { key: '/dashboard', label: '仪表盘', icon: 'M3 12l9-9 9 9M5 10v10a1 1 0 001 1h3m10-11v10a1 1 0 01-1 1h-3m-6 0h6' },
    ],
  },
  {
    label: '资源',
    items: [
      { key: '/r/users', label: '用户', icon: 'M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z' },
      { key: '/r/photos', label: '图片', icon: 'M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z' },
      { key: '/r/albums', label: '相册', icon: 'M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10' },
      { key: '/r/groups', label: '角色组', icon: 'M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z' },
      { key: '/r/shares', label: '分享', icon: 'M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.368 2.684 3 3 0 00-5.368-2.684z' },
    ],
  },
  {
    label: '交易',
    items: [
      { key: '/r/plans', label: '套餐', icon: 'M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4' },
      { key: '/r/orders', label: '订单', icon: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01' },
      { key: '/r/coupons', label: '优惠码', icon: 'M15 5v2m0 4v2m0 4v2M5 5a2 2 0 00-2 2v3a2 2 0 002 2 2 2 0 012 2 2 2 0 01-2 2 2 2 0 00-2 2v3a2 2 0 002 2h14a2 2 0 002-2v-3a2 2 0 00-2-2 2 2 0 01-2-2 2 2 0 012-2 2 2 0 002-2V7a2 2 0 00-2-2H5z' },
    ],
  },
  {
    label: '社区',
    items: [
      { key: '/r/notices', label: '公告', icon: 'M11 5.882V19.24a1.76 1.76 0 01-3.417.592l-2.147-6.15M18 13a3 3 0 100-6M5.436 13.683A4.001 4.001 0 017 6h1.832c4.1 0 7.625-1.234 9.168-3 .14.163.36.427.619.681C19.653 7.706 20 9.201 20 10c0 .96-.278 1.854-.785 2.61M5.436 13.683a4.001 4.001 0 01-.562-3.034m.562 3.04l1.425 4.081a1.76 1.76 0 003.417-.592' },
      { key: '/r/pages', label: '页面', icon: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z' },
      { key: '/r/tickets', label: '工单', icon: 'M18.364 5.636l-3.536 3.536m0 5.656l3.536 3.536M9.172 9.172L5.636 5.636m3.536 9.192l-3.536 3.536M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-5 0a4 4 0 11-8 0 4 4 0 018 0z' },
      { key: '/r/reports', label: '举报', icon: 'M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z' },
      { key: '/r/violations', label: '违规', icon: 'M18.364 18.364A9 9 0 005.636 5.636m12.728 12.728A9 9 0 015.636 5.636m12.728 12.728L5.636 5.636' },
      { key: '/r/feedbacks', label: '反馈', icon: 'M7 8h10M7 12h4m1 8l4-4H21a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v14a2 2 0 002 2h6z' },
    ],
  },
  {
    label: '系统',
    items: [
      { key: '/r/storages', label: '储存策略', icon: 'M5 12H3a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2h-2m-4-3v8m-4-3h8m-5-9h4l1 4H7l1-4z' },
      { key: '/r/drivers', label: '驱动', icon: 'M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z' },
      { key: '/settings', label: '设置', icon: 'M12 6V4m0 2a2 2 0 100 4m0-4a2 2 0 110 4m-6 8a2 2 0 100-4m0 4a2 2 0 100 4m0-4v2m0-6V4m6 6v10m6-2a2 2 0 100-4m0 4a2 2 0 100 4m0-4v2m0-6V4' },
    ],
  },
];

function renderIcon(path) {
  return h('svg', {
    xmlns: 'http://www.w3.org/2000/svg',
    fill: 'none',
    viewBox: '0 0 24 24',
    stroke: 'currentColor',
    'stroke-width': '1.5',
    style: 'width:1.1rem;height:1.1rem;flex-shrink:0',
  }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', d: path })]);
}

const menuOptions = groups.map((g) => ({
  type: 'group',
  label: g.label,
  key: 'g-' + g.label,
  children: g.items.map((i) => ({
    label: i.label,
    key: i.key,
    icon: () => renderIcon(i.icon),
  })),
}));

const userMenuOptions = [{ label: '退出登录', key: 'logout' }];

const onUserMenu = async (key) => {
  if (key === 'logout') {
    try { await put('/logout', {}); } catch (e) { /* ignore */ }
    localStorage.removeItem('admin_logged_in');
    msg.success('已退出登录');
    router.push('/admin/login');
  }
};

const onMenu = (key) => {
  router.push(key);
  menuOpen.value = false;
};

const appName = ref('ywty');
get('/settings').then((res) => {
  const name = (res.data.settings || []).find((s) => s.group === 'app' && s.name === 'name');
  if (name && name.payload) appName.value = name.payload;
}).catch(() => {});
</script>

<template>
  <div class="flex min-h-screen bg-gray-100">
    <aside
      class="fixed inset-y-0 left-0 z-30 w-64 flex-col justify-between bg-gray-950 px-3 py-4 flex
             transition-transform lg:sticky lg:top-0 lg:h-screen lg:translate-x-0"
      :class="menuOpen ? 'translate-x-0' : '-translate-x-full'"
    >
      <div class="overflow-y-auto flex-1">
        <div class="flex items-center gap-2.5 px-3 pb-3">
          <div class="flex h-9 w-9 items-center justify-center rounded-lg bg-primary-600 text-white font-bold">y</div>
          <div>
            <div class="text-sm font-bold text-white leading-tight">{{ appName || 'ywty' }}</div>
            <div class="text-xs text-gray-500">管理后台</div>
          </div>
        </div>
        <NMenu :value="route.fullPath" :options="menuOptions" @update:value="onMenu" />
      </div>
      <div class="border-t border-white/10 pt-2">
        <NDropdown trigger="click" :options="userMenuOptions" @select="onUserMenu">
          <button class="fi-sidebar-item w-full">
            <NAvatar round size="small" class="!bg-primary-600">{{ username.slice(0, 1).toUpperCase() }}</NAvatar>
            <span class="truncate">{{ username }}</span>
          </button>
        </NDropdown>
      </div>
    </aside>
    <div v-if="menuOpen" class="fixed inset-0 z-20 bg-black/50 lg:hidden" @click="menuOpen = false" />

    <div class="flex min-h-screen flex-1 flex-col lg:pl-64">
      <header class="sticky top-0 z-10 flex items-center gap-3 border-b border-gray-200 bg-white/90 px-4 py-3 backdrop-blur lg:hidden">
        <button class="fi-btn-secondary" @click="menuOpen = !menuOpen">☰</button>
        <span class="font-bold">管理后台</span>
      </header>
      <main class="mx-auto w-full max-w-7xl flex-1 px-4 py-6">
        <RouterView />
      </main>
    </div>
  </div>
</template>
