// 编辑分享对话框：修改访问密码 / 过期时间
// PATCH /api/v1/shares/:id 三态语义：字段省略 = 不修改；password: null = 清除密码；
// expires_at: null = 取消过期；非空字符串 = 设置（expires_at 为 RFC3339）
import { useEffect, useState } from 'react'
import { useApi } from '@/lib/api'
import { toast } from '@/lib/react-store'
import { Button } from '../ui/button'
import { Checkbox } from '../ui/checkbox'
import { Input } from '../ui/input'
import { Label } from '../ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '../ui/select'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '../ui/dialog'

interface ShareRow {
  id: number
  slug: string
  has_password: boolean
  expires_at: string | null
}

interface ShareEditDialogProps {
  share: ShareRow | null
  open: boolean
  onOpenChange: (open: boolean) => void
  onSaved?: () => void
}

function toLocalInput(d: Date): string {
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

export function ShareEditDialog({ share, open, onOpenChange, onSaved }: ShareEditDialogProps) {
  const api = useApi()

  const [newPassword, setNewPassword] = useState('')
  const [clearPassword, setClearPassword] = useState(false)
  const [expiresChoice, setExpiresChoice] = useState<'keep' | 'clear' | 'custom'>('keep')
  const [customExpires, setCustomExpires] = useState('')
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    if (open && share) {
      setNewPassword('')
      setClearPassword(false)
      setExpiresChoice('keep')
      setCustomExpires(share.expires_at ? toLocalInput(new Date(share.expires_at)) : '')
      setSaving(false)
    }
  }, [open, share])

  const submit = async () => {
    if (!share) return
    if (expiresChoice === 'custom') {
      const d = new Date(customExpires)
      if (!customExpires || Number.isNaN(d.getTime())) {
        toast.error('请选择有效的过期时间')
        return
      }
    }

    // 省略 = 不修改，因此只带上用户明确要改的字段
    const body: Record<string, unknown> = {}
    if (newPassword.trim()) body.password = newPassword.trim()
    else if (clearPassword) body.password = null
    if (expiresChoice === 'clear') body.expires_at = null
    else if (expiresChoice === 'custom') body.expires_at = new Date(customExpires).toISOString()

    if (Object.keys(body).length === 0) {
      onOpenChange(false)
      return
    }

    setSaving(true)
    try {
      await api.patch(`/api/v1/shares/${share.id}`, body)
      toast.success('分享已更新')
      onOpenChange(false)
      onSaved?.()
    } catch (e: any) {
      toast.error(e?.message || '更新失败')
    } finally {
      setSaving(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>编辑分享</DialogTitle>
          <DialogDescription className="font-mono">/s/{share?.slug ?? ''}</DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          <div className="space-y-2">
            <Label>访问密码</Label>
            {share?.has_password && (
              <label className="flex items-center gap-2 text-sm text-muted-foreground">
                <Checkbox
                  checked={clearPassword}
                  disabled={!!newPassword.trim()}
                  onCheckedChange={(v) => setClearPassword(v === true)}
                />
                清除现有密码（恢复免密访问）
              </label>
            )}
            <Input
              value={newPassword}
              onChange={(e) => {
                setNewPassword(e.target.value)
                if (e.target.value.trim()) setClearPassword(false)
              }}
              placeholder={share?.has_password ? '输入新密码可覆盖，留空则不修改' : '留空则不设置密码'}
              maxLength={64}
            />
          </div>

          <div className="space-y-2">
            <Label>过期时间{share?.expires_at ? `（当前：${new Date(share.expires_at).toLocaleString()}）` : '（当前：永久有效）'}</Label>
            <Select value={expiresChoice} onValueChange={(v) => setExpiresChoice(v as 'keep' | 'clear' | 'custom')}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="keep">保持不变</SelectItem>
                <SelectItem value="clear">取消过期（永久有效）</SelectItem>
                <SelectItem value="custom">自定义过期时间</SelectItem>
              </SelectContent>
            </Select>
            {expiresChoice === 'custom' && (
              <Input type="datetime-local" value={customExpires} onChange={(e) => setCustomExpires(e.target.value)} />
            )}
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={saving}>取消</Button>
          <Button onClick={submit} loading={saving}>保存</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
