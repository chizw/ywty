<script setup>
import { onMounted, ref } from 'vue';
import { NGrid, NGridItem, NStatistic, useMessage } from 'naive-ui';
import { get, errMsg } from '../api';

const msg = useMessage();
const stats = ref(null);

onMounted(async () => {
  try {
    stats.value = (await get('/dashboard')).data;
  } catch (e) {
    msg.error(errMsg(e));
  }
});

const widgets = [
  ['用户', 'user_count'],
  ['图片', 'photo_count'],
  ['相册', 'album_count'],
  ['分享', 'share_count'],
  ['已用空间(KB)', 'photo_size_kb'],
  ['已支付订单', 'order_count'],
  ['收入(分)', 'revenue_fen'],
  ['待处理工单', 'ticket_open'],
  ['待处理举报', 'report_open'],
];
</script>

<template>
  <div>
    <h1 class="fi-page-title">仪表盘</h1>
    <div class="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5">
      <div v-for="[label, key] in widgets" :key="key" class="fi-card flex items-center gap-3 px-4 py-4">
        <div class="flex h-10 w-10 items-center justify-center rounded-full bg-primary-50 text-primary-600 font-bold">
          {{ (stats?.[key] ?? 0) > 999 ? '99+' : (stats?.[key] ?? 0) }}
        </div>
        <div>
          <div class="text-sm font-semibold text-gray-900">{{ stats?.[key] ?? 0 }}</div>
          <div class="text-xs text-gray-500">{{ label }}</div>
        </div>
      </div>
    </div>
  </div>
</template>
