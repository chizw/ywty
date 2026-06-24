// useMessage 轻量级消息提示 composable
// 用法：
//   const message = useMessage()
//   message.success('保存成功')
//   message.error('保存失败')
//   message.warning('请填写完整')
//   message.info('提示')
// 实现：基于 useState 共享一条最新的提示，由 AppToaster 组件渲染。
type ToastKind = 'success' | 'error' | 'warning' | 'info'

export interface Toast {
  id: number
  kind: ToastKind
  text: string
  ttl: number
}

let counter = 0

export function useMessage() {
  const toasts = useState<Toast[]>('app-toasts', () => [])

  function push(kind: ToastKind, text: string, ttl = 3000) {
    const id = ++counter
    toasts.value = [...toasts.value, { id, kind, text, ttl }]
    if (import.meta.client) {
      setTimeout(() => {
        toasts.value = toasts.value.filter((t) => t.id !== id)
      }, ttl)
    }
  }

  return {
    toasts,
    success: (text: string) => push('success', text),
    error: (text: string) => push('error', text),
    warning: (text: string) => push('warning', text),
    info: (text: string) => push('info', text),
  }
}
