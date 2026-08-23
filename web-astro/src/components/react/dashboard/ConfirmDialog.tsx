import { useCallback, useState } from 'react'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '../ui/dialog'
import { Button } from '../ui/button'

interface ConfirmOptions {
  title?: string
  message?: string
  okText?: string
  danger?: boolean
}

interface ConfirmState extends ConfirmOptions {
  open: boolean
  resolve: ((v: boolean) => void) | null
}

/**
 * 通用确认弹窗 hook：`const { confirm, node } = useConfirm()`，`confirm({...})` 返回 Promise<boolean>，
 * 把 `node` 渲染在组件末尾。
 */
export function useConfirm() {
  const [state, setState] = useState<ConfirmState>({
    open: false,
    title: '确认操作',
    message: '确定继续？',
    okText: '确定',
    danger: false,
    resolve: null,
  })

  const confirm = useCallback((opts?: ConfirmOptions) => {
    return new Promise<boolean>((resolve) => {
      setState({
        open: true,
        title: opts?.title || '确认操作',
        message: opts?.message || '确定继续？',
        okText: opts?.okText || '确定',
        danger: !!opts?.danger,
        resolve,
      })
    })
  }, [])

  const settle = useCallback(
    (v: boolean) => {
      state.resolve?.(v)
      setState((s) => ({ ...s, open: false, resolve: null }))
    },
    [state.resolve]
  )

  const node = (
    <Dialog open={state.open} onOpenChange={(o) => !o && settle(false)}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>{state.title}</DialogTitle>
        </DialogHeader>
        <DialogDescription className="whitespace-pre-line">{state.message}</DialogDescription>
        <DialogFooter>
          <Button variant="outline" onClick={() => settle(false)}>
            取消
          </Button>
          <Button variant={state.danger ? 'destructive' : 'default'} onClick={() => settle(true)}>
            {state.okText}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )

  return { confirm, node }
}
