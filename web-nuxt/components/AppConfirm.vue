<script setup lang="ts">
// 通用确认弹窗：替代浏览器原生 confirm()
// 通过 v-model 打开关闭，resolve() 返回 true/false
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

function onOpenChange(val: boolean) {
  emit('update:show', val)
  if (!val) emit('cancel')
}

function close() { emit('update:show', false); emit('cancel') }
function ok() { emit('update:show', false); emit('confirm') }
</script>

<template>
  <Dialog :open="show" @update:open="onOpenChange">
    <DialogContent :style="{ maxWidth: width + 'px' }">
      <DialogHeader>
        <DialogTitle>{{ title }}</DialogTitle>
      </DialogHeader>
      <DialogDescription class="whitespace-pre-line">
        {{ message }}
      </DialogDescription>
      <DialogFooter>
        <Button variant="outline" @click="close">{{ cancelText }}</Button>
        <Button :variant="danger ? 'destructive' : 'default'" @click="ok">{{ okText }}</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
