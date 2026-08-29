import { createApp, h } from 'vue';
import { createRouter, createWebHistory, RouterView } from 'vue-router';
import { NConfigProvider, NMessageProvider, NDialogProvider, zhCN, dateZhCN } from 'naive-ui';
import AdminLayout from './layouts/AdminLayout.vue';
import Login from './pages/Login.vue';
import Dashboard from './pages/Dashboard.vue';
import Settings from './pages/Settings.vue';
import Resource from './pages/Resource.vue';
import './index.css';

const routes = [
  { path: '/login', component: Login, meta: { public: true } },
  {
    path: '/',
    component: AdminLayout,
    children: [
      { path: '', redirect: '/dashboard' },
      { path: 'dashboard', component: Dashboard },
      { path: 'r/:resource', component: Resource, props: true },
      { path: 'settings', component: Settings },
    ],
  },
  { path: '/:pathMatch(.*)*', redirect: '/dashboard' },
];

const router = createRouter({ history: createWebHistory('/admin/'), routes });

router.beforeEach((to) => {
  if (to.meta.public) return true;
  if (localStorage.getItem('admin_logged_in') === '1') return true;
  return '/login';
});

const app = createApp(() =>
  h(NConfigProvider, { locale: zhCN, dateLocale: dateZhCN }, () =>
    h(NMessageProvider, {}, () =>
      h(NDialogProvider, {}, () =>
        h(RouterView),
      ),
    ),
  ),
);
app.use(router);
app.mount('#admin');
