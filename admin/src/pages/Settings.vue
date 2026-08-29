<script setup>
import { onMounted, ref } from 'vue';
import { NButton, NInput, useMessage } from 'naive-ui';
import { get, put, errMsg } from '../api';

const msg = useMessage();
const groups = ref([]); // [{key, label, items: [{group, name, payloadText}]}]
const edits = ref({});  // key: "group.name" → textarea 文本

const groupOrder = ['app', 'site', 'admin', 'user'];
const groupLabels = { app: '应用', site: '站点', admin: '后台', user: '用户' };

onMounted(async () => {
  try {
    const res = await get('/settings');
    const all = res.data.settings.map((s) => ({ ...s, payloadText: JSON.stringify(s.payload ?? null) }));
    const byKey = { app: [], site: [], admin: [], user: [] };
    for (const s of all) {
      if (!byKey[s.group]) byKey[s.group] = [];
      byKey[s.group].push(s);
    }
    groups.value = groupOrder
      .filter((g) => byKey[g].length)
      .map((g) => ({ key: g, label: groupLabels[g] || g, items: byKey[g] }));
  } catch (e) {
    msg.error(errMsg(e));
  }
});

async function save() {
  try {
    const updates = Object.entries(edits.value).map(([key, text]) => {
      const dot = key.indexOf('.');
      return {
        group: key.slice(0, dot),
        name: key.slice(dot + 1),
        payload: JSON.parse(text),
      };
    });
    if (!updates.length) {
      msg.info('没有修改');
      return;
    }
    await put('/settings', { updates });
    msg.success('设置已保存');
    edits.value = {};
  } catch (e) {
    msg.error(errMsg(e));
  }
}
</script>

<template>
  <div>
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <h1 class="fi-page-title">设置</h1>
        <p class="mt-1 text-sm text-gray-500">JSON 值直接编辑，支持字符串 / 布尔 / 数字 / 数组</p>
      </div>
      <NButton type="primary" @click="save">保存修改</NButton>
    </div>

    <div v-for="g in groups" :key="g.key" class="fi-card mt-4 p-5">
      <h2 class="text-base font-bold text-gray-900">{{ g.label }}设置</h2>
      <div class="mt-4 grid grid-cols-1 gap-x-6 gap-y-4 lg:grid-cols-2">
        <div v-for="s in g.items" :key="g.key + '.' + s.name">
          <label class="fi-label">{{ s.name }}</label>
          <NInput
            :default-value="s.payloadText"
            type="textarea"
            :autosize="{ minRows: 1, maxRows: 4 }"
            @update:value="(v) => { edits[g.key + '.' + s.name] = v }"
          />
        </div>
      </div>
    </div>
  </div>
</template>
