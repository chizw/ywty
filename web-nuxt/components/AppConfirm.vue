<script setup lang="ts">
// 通用确认弹窗：替代浏览器原生 confirm()
// 通过 v-model 打开关闭，resolve() 返回 true/false
import { NModal, NButton } from 'naive-ui'

const props = withDefaults(defineProps<{
  show: boolean
  title?: string
  message?: string
  okText?: string
  cancelText?: string
  danger?: boolean
  width?: number
}>(), {
  title: '确认操作',
  message: '确定继续？',
  okText: '确定',
  cancelText: '取消',
  danger: false,
  width: 380,
})

const emit = defineEmits<{
  'update:show': [boolean]
  confirm: []
  cancel: []
}>()

function close() { emit('update:show', false); emit('cancel') }
function ok() { emit('update:show', false); emit('confirm') }
</script>

<template>
  <NModal
    :show="show"
    preset="card"
    :title="title"
    :style="{ maxWidth: width + 'px' }"
    :mask-closable="false"
    @update:show="(v) => emit('update:show', v)"
  >
    <div class="text-sm text-gray-700 whitespace-pre-line">{{ message }}</div>
    <template #footer>
      <div class="flex justify-end gap-2">
        <NButton @click="close">{{ cancelText }}</NButton>
        <NButton :type="danger ? 'error' : 'primary'" @click="ok">{{ okText }}</NButton>
      </div>
    </template>
  </NModal>
</template>
