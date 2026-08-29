<script setup>
import { computed } from 'vue';
import { useRoute } from 'vue-router';
import { resources } from '../resources';
import FilamentTable from '../components/FilamentTable.vue';

const route = useRoute();

const cfg = computed(() => {
  const r = resources[route.params.resource];
  return r || { title: '未找到', listURL: '/none', columns: [], searchable: false };
});
</script>

<template>
  <div>
    <FilamentTable :key="route.params.resource" v-bind="cfg" />
  </div>
</template>
