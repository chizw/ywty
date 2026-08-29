<script setup>
import { ref } from 'vue';
import { useMessage } from 'naive-ui';

const username = ref('');
const password = ref('');
const loading = ref(false);
const msg = useMessage();

async function submit() {
  if (!username.value || !password.value) {
    msg.error('请输入用户名和密码');
    return;
  }
  loading.value = true;
  try {
    await fetch('/api/admin/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({ username: username.value, password: password.value }),
    }).then(async (r) => {
      const body = await r.json().catch(() => ({}));
      if (!r.ok || body.status === 'error') throw new Error(body.message || '登录失败');
    });
    localStorage.setItem('admin_logged_in', '1');
    window.location.href = '/admin/dashboard';
  } catch (e) {
    msg.error(e.message);
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <div class="flex min-h-screen items-center justify-center bg-gray-100 px-4">
    <div class="w-full max-w-sm">
      <div class="fi-card px-8 py-10">
        <div class="mb-8 flex flex-col items-center gap-3">
          <div class="flex h-12 w-12 items-center justify-center rounded-xl bg-primary-600 text-xl font-bold text-white">y</div>
          <h1 class="text-lg font-bold text-gray-900">登录到管理后台</h1>
        </div>
        <form @submit.prevent="submit">
          <div class="mb-4">
            <label class="fi-label">用户名</label>
            <input v-model="username" class="fi-input" placeholder="请输入用户名" autocomplete="username" />
          </div>
          <div class="mb-6">
            <label class="fi-label">密码</label>
            <input v-model="password" type="password" class="fi-input" placeholder="请输入密码" autocomplete="current-password" />
          </div>
          <button type="submit" class="fi-btn-primary w-full !py-2.5" :disabled="loading">
            {{ loading ? '登录中…' : '登录' }}
          </button>
        </form>
      </div>
    </div>
  </div>
</template>
