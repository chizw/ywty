// 创建分享对话框：对单张图片发起分享（密码/过期时间可选），成功后展示短链并复制
import { useEffect, useState } from 'react'
import { Copy } from 'lucide-react'
import { useApi } from '@/lib/api'
import { toast } from '@/lib/react-store'
import { Button } from '../ui/button'
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

const EXPIRY_OPTIONS = [
  { value: 'never', label: '永不过期', days: 0 },
  { value: '1d', label: '1 天', days: 1 },
  { value: '7d', label: '7 天', days: 7 },
  { value: '30d', label: '30 天', days: 30 },
]

interface ShareCreateDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  photoId: number | null
}

async function copyText(text: string): Promise<boolean> {
  try {
    await navigator.clipboard.writeText(text)
    return true
  } catch {
    return false
  }
}

export function ShareCreateDialog({ open, onOpenChange, photoId }: ShareCreateDialogProps) {
  const api = useApi()

  const [password, setPassword] = useState('')
  const [expiry, setExpiry] = useState('never')
  const [submitting, setSubmitting] = useState(false)
  const [createdUrl, setCreatedUrl] = useState('')
  const [copied, setCopied] = useState(false)

  useEffect(() => {
    if (!open) {
      setPassword('')
      setExpiry('never')
      setCreatedUrl('')
      setCopied(false)
      setSubmitting(false)
    }
  }, [open])

  const submit = async () => {
    if (!photoId) return
    setSubmitting(true)
    try {
      const body: Record<string, unknown> = { shareable_type: 'photo', shareable_id: photoId }
      if (password.trim()) body.password = password.trim()
      const days = EXPIRY_OPTIONS.find((o) => o.value === expiry)?.days ?? 0
      if (days > 0) body.expires_at = new Date(Date.now() + days * 86400_000).toISOString()

      const res = await api.post<any>('/api/v1/shares', body)
      const slug = res?.data?.slug ?? res?.slug
      if (!slug) throw new Error('创建失败')
      const url = `${window.location.origin}/s/${slug}`
      setCreatedUrl(url)
      if (await copyText(url)) {
        setCopied(true)
        toast.success('短链接已复制到剪贴板')
      } else {
        toast.error('复制失败，请手动复制')
      }
    } catch (e: any) {
      toast.error(e?.message || '创建分享失败')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>分享图片</DialogTitle>
          <DialogDescription>生成一个公开短链接，任何人都可以通过它查看这张图片。</DialogDescription>
        </DialogHeader>

        {createdUrl ? (
          <div className="space-y-3">
            <div className="space-y-2">
              <Label>分享短链</Label>
              <div className="flex gap-2">
                <Input readOnly value={createdUrl} onFocus={(e) => e.currentTarget.select()} />
                <Button variant="outline" size="icon" aria-label="复制链接" onClick={async () => {
                  if (await copyText(createdUrl)) {
                    setCopied(true)
                    toast.success('链接已复制')
                  } else {
                    toast.error('复制失败，请手动复制')
                  }
                }}>
                  <Copy className="h-4 w-4" />
                </Button>
              </div>
            </div>
            <p className="text-xs text-muted-foreground">
              {copied ? '已复制到剪贴板。' : ''}可在「分享管理」中随时删除该分享。
            </p>
          </div>
        ) : (
          <div className="space-y-4">
            <div className="space-y-2">
              <Label>访问密码（可选）</Label>
              <Input
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="留空则无需密码"
                maxLength={64}
              />
            </div>
            <div className="space-y-2">
              <Label>有效期</Label>
              <Select value={expiry} onValueChange={setExpiry}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {EXPIRY_OPTIONS.map((o) => (
                    <SelectItem key={o.value} value={o.value}>{o.label}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>
        )}

        <DialogFooter>
          {createdUrl ? (
            <Button onClick={() => onOpenChange(false)}>完成</Button>
          ) : (
            <>
              <Button variant="outline" onClick={() => onOpenChange(false)} disabled={submitting}>取消</Button>
              <Button onClick={submit} loading={submitting} disabled={!photoId}>创建分享</Button>
            </>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
