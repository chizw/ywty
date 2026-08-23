// 系统设置：邮件服务器（SMTP）+ 安全开关（settings 表驱动）
import { useEffect, useState } from 'react'
import { Save } from 'lucide-react'
import { AdminShell } from './AdminShell'
import { AdminPageHeader } from './AdminPageHeader'
import { useApi } from '@/lib/api'
import { toast } from '@/lib/react-store'
import { Button } from '../ui/button'
import { Checkbox } from '../ui/checkbox'
import { Input } from '../ui/input'
import { Label } from '../ui/label'

const K = {
  host: 'mail.smtp.host',
  port: 'mail.smtp.port',
  username: 'mail.smtp.username',
  password: 'mail.smtp.password',
  from: 'mail.smtp.from',
  ssl: 'mail.smtp.ssl',
  requireEmailVerify: 'security.require_email_verify',
  allowPasswordReset: 'security.allow_password_reset',
  allowRegister: 'security.allow_register',
} as const

type TextForm = {
  host: string
  port: string
  username: string
  password: string
  from: string
}

type BoolForm = {
  ssl: boolean
  requireEmailVerify: boolean
  allowPasswordReset: boolean
  allowRegister: boolean
}

export function AdminSettingsPage() {
  const api = useApi()
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [passwordSet, setPasswordSet] = useState(false)
  const [text, setText] = useState<TextForm>({
    host: '',
    port: '465',
    username: '',
    password: '',
    from: '',
  })
  const [flags, setFlags] = useState<BoolForm>({
    ssl: true,
    requireEmailVerify: true,
    allowPasswordReset: true,
    allowRegister: true,
  })

  useEffect(() => {
    api
      .get<Record<string, unknown>>('/api/v1/admin/settings')
      .then((data) => {
        const s = (k: string) => {
          const v = data?.[k]
          return typeof v === 'string' ? v : ''
        }
        const b = (k: string, dflt: boolean) => {
          const v = data?.[k]
          if (typeof v === 'boolean') return v
          if (typeof v === 'string') return ['1', 'true', 'on', 'yes'].includes(v.toLowerCase())
          return dflt
        }
        setText((f) => ({
          ...f,
          host: s(K.host),
          port: s(K.port) || '465',
          username: s(K.username),
          from: s(K.from),
        }))
        setPasswordSet(data?.[K.password] === true)
        setFlags({
          ssl: b(K.ssl, true),
          requireEmailVerify: b(K.requireEmailVerify, true),
          allowPasswordReset: b(K.allowPasswordReset, true),
          allowRegister: b(K.allowRegister, true),
        })
      })
      .catch((e) => toast.error(e?.message || '加载设置失败'))
      .finally(() => setLoading(false))
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const save = async () => {
    setSaving(true)
    try {
      const payload: Record<string, unknown> = {
        [K.host]: text.host.trim(),
        [K.port]: text.port.trim() || '465',
        [K.username]: text.username.trim(),
        [K.from]: text.from.trim(),
        [K.ssl]: flags.ssl,
        [K.requireEmailVerify]: flags.requireEmailVerify,
        [K.allowPasswordReset]: flags.allowPasswordReset,
        [K.allowRegister]: flags.allowRegister,
      }
      // 密码留空 = 保持不变
      if (text.password) payload[K.password] = text.password
      await api.put('/api/v1/admin/settings', payload)
      toast.success('设置已保存')
      setText((f) => ({ ...f, password: '' }))
      setPasswordSet(passwordSet || !!text.password)
    } catch (e) {
      toast.error((e as Error)?.message || '保存失败')
    } finally {
      setSaving(false)
    }
  }

  const setText_ = (k: keyof TextForm) => (e: React.ChangeEvent<HTMLInputElement>) =>
    setText((f) => ({ ...f, [k]: e.target.value }))

  const setFlag = (k: keyof BoolForm) => (v: boolean | 'indeterminate') =>
    setFlags((f) => ({ ...f, [k]: v === true }))

  return (
    <AdminShell>
      <AdminPageHeader title="系统设置" description="邮件服务器与安全开关（即时生效）">
        <Button size="sm" onClick={save} disabled={loading || saving}>
          <Save className="h-4 w-4" /> {saving ? '保存中…' : '保存设置'}
        </Button>
      </AdminPageHeader>

      {loading ? (
        <div className="skeleton h-64 rounded-md" />
      ) : (
        <div className="max-w-2xl space-y-6">
          {/* 邮件服务器 */}
          <section className="rounded-md border border-border bg-card p-5">
            <h3 className="mb-1 text-sm font-medium">邮件服务器（SMTP）</h3>
            <p className="mb-4 text-xs text-muted-foreground">
              配置后用于发送注册验证码、找回密码邮件；留空主机则回退到启动配置。
            </p>
            <div className="grid gap-4 sm:grid-cols-2">
              <div className="space-y-2">
                <Label htmlFor="smtp-host">SMTP 主机</Label>
                <Input id="smtp-host" placeholder="smtp.example.com" value={text.host} onChange={setText_('host')} />
              </div>
              <div className="space-y-2">
                <Label htmlFor="smtp-port">端口</Label>
                <Input id="smtp-port" placeholder="465" value={text.port} onChange={setText_('port')} />
              </div>
              <div className="space-y-2">
                <Label htmlFor="smtp-username">账号</Label>
                <Input id="smtp-username" value={text.username} onChange={setText_('username')} />
              </div>
              <div className="space-y-2">
                <Label htmlFor="smtp-password">密码</Label>
                <Input
                  id="smtp-password"
                  type="password"
                  placeholder={passwordSet ? '已设置，留空保持不变' : ''}
                  value={text.password}
                  onChange={setText_('password')}
                />
              </div>
              <div className="space-y-2 sm:col-span-2">
                <Label htmlFor="smtp-from">发件人</Label>
                <Input id="smtp-from" placeholder='云雾图驿 <noreply@example.com>' value={text.from} onChange={setText_('from')} />
              </div>
              <label className="flex items-center gap-2 text-sm sm:col-span-2">
                <Checkbox checked={flags.ssl} onCheckedChange={setFlag('ssl')} id="smtp-ssl" />
                使用 SSL/TLS（隐式 TLS，通常端口 465）
              </label>
            </div>
          </section>

          {/* 安全开关 */}
          <section className="rounded-md border border-border bg-card p-5">
            <h3 className="mb-1 text-sm font-medium">安全开关</h3>
            <p className="mb-4 text-xs text-muted-foreground">控制注册与找回密码相关功能的开放状态。</p>
            <div className="space-y-4">
              <label className="flex items-start gap-2 text-sm">
                <Checkbox checked={flags.requireEmailVerify} onCheckedChange={setFlag('requireEmailVerify')} id="flag-verify" className="mt-0.5" />
                <span>
                  注册需要邮箱验证码
                  <span className="block text-xs text-muted-foreground">关闭后注册不再要求验证码，也不会下发注册验证码邮件。</span>
                </span>
              </label>
              <label className="flex items-start gap-2 text-sm">
                <Checkbox checked={flags.allowPasswordReset} onCheckedChange={setFlag('allowPasswordReset')} id="flag-reset" className="mt-0.5" />
                <span>
                  开放找回密码
                  <span className="block text-xs text-muted-foreground">关闭后用户无法通过验证码重置密码。</span>
                </span>
              </label>
              <label className="flex items-start gap-2 text-sm">
                <Checkbox checked={flags.allowRegister} onCheckedChange={setFlag('allowRegister')} id="flag-register" className="mt-0.5" />
                <span>
                  允许注册新账号
                  <span className="block text-xs text-muted-foreground">关闭后注册接口将拒绝新用户注册。</span>
                </span>
              </label>
            </div>
          </section>
        </div>
      )}
    </AdminShell>
  )
}
