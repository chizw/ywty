// 设置：资料 / 修改密码 / 修改邮箱 / 修改手机（Tabs）
import { useEffect, useState } from 'react'
import { AppShell } from './AppShell'
import { PageHeader } from './PageHeader'
import { useApi } from '@/lib/api'
import { useAuthStore, toast } from '@/lib/react-store'
import { formatBytes } from '@/lib/utils'
import { Button } from '../ui/button'
import { Input } from '../ui/input'
import { Label } from '../ui/label'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '../ui/tabs'
import { Avatar, AvatarFallback, AvatarImage } from '../ui/avatar'

interface Profile {
  id: number
  username: string
  email: string
  avatar: string | null
  bio: string | null
  role: string
  capacity_used: number
  capacity_max: number
  phone?: string
}

export function SettingsPage() {
  const api = useApi()
  const user = useAuthStore((s) => s.user)
  const fetchMe = useAuthStore((s) => s.fetchMe)

  const [profile, setProfile] = useState<Profile | null>(null)

  // 资料
  const [username, setUsername] = useState('')
  const [bio, setBio] = useState('')
  const [savingProfile, setSavingProfile] = useState(false)

  // 密码
  const [pwd, setPwd] = useState({ old: '', next: '' })
  const [savingPwd, setSavingPwd] = useState(false)

  // 邮箱
  const [email, setEmail] = useState({ addr: '', code: '' })

  useEffect(() => {
    api.get<Profile>('/api/v1/user/profile').then((p) => {
      setProfile(p)
      setUsername(p.username)
      setBio(p.bio || '')
    }).catch(() => {})
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const saveProfile = async () => {
    setSavingProfile(true)
    try {
      await api.patch('/api/v1/user/profile', { username, bio: bio || null })
      toast.success('已保存')
      fetchMe()
    } catch (e: any) {
      toast.error(e?.message || '保存失败')
    } finally {
      setSavingProfile(false)
    }
  }

  const changePassword = async () => {
    if (!pwd.old || !pwd.next) return
    setSavingPwd(true)
    try {
      await api.post('/api/v1/user/change-password', { old_password: pwd.old, new_password: pwd.next })
      toast.success('密码已修改')
      setPwd({ old: '', next: '' })
    } catch (e: any) {
      toast.error(e?.message || '修改失败')
    } finally {
      setSavingPwd(false)
    }
  }

  const sendCode = async (emailAddr: string, event: string) => {
    if (!emailAddr) return toast.warning('请先填写账号')
    try {
      await api.post('/api/v1/verify-codes', { email: emailAddr, event })
      toast.success('验证码已发送')
    } catch (e: any) {
      toast.error(e?.message || '发送失败')
    }
  }

  const changeEmail = async () => {
    try {
      await api.post('/api/v1/user/change-email', { new_email: email.addr, verify_code: email.code })
      toast.success('邮箱已修改')
      setEmail({ addr: '', code: '' })
      fetchMe()
    } catch (e: any) {
      toast.error(e?.message || '修改失败')
    }
  }

  const labelCls = 'text-xs text-muted-foreground'
  const inputWrap = 'space-y-2'

  return (
    <AppShell>
      <PageHeader title="设置" description="管理你的账号资料" />

      <div className="mb-6 flex items-center gap-4 rounded-md border border-border bg-card p-5">
        <Avatar className="h-14 w-14">
          {profile?.avatar ? <AvatarImage src={profile.avatar} /> : null}
          <AvatarFallback className="text-lg">{(user?.name || user?.username || 'U').slice(0, 1).toUpperCase()}</AvatarFallback>
        </Avatar>
        <div className="min-w-0 flex-1">
          <p className="font-display text-lg font-semibold">{user?.name || user?.username}</p>
          <p className="text-sm text-muted-foreground">{profile?.email}</p>
        </div>
        <div className="text-right text-sm text-muted-foreground">
          <p className="tabular-nums">{formatBytes(profile?.capacity_used ?? 0)} / {profile?.capacity_max ? formatBytes(profile.capacity_max) : '—'}</p>
          <p className="mt-0.5 text-xs">{profile?.role === 'admin' ? '管理员' : '普通用户'}</p>
        </div>
      </div>

      <Tabs defaultValue="profile" className="w-full">
        <TabsList>
          <TabsTrigger value="profile">资料</TabsTrigger>
          <TabsTrigger value="password">修改密码</TabsTrigger>
          <TabsTrigger value="email">修改邮箱</TabsTrigger>
        </TabsList>

        <TabsContent value="profile">
          <div className="max-w-md space-y-4 rounded-md border border-border bg-card p-5">
            <div className={inputWrap}>
              <Label className={labelCls}>用户名</Label>
              <Input value={username} onChange={(e) => setUsername(e.target.value)} />
            </div>
            <div className={inputWrap}>
              <Label className={labelCls}>个人简介</Label>
              <Input value={bio} onChange={(e) => setBio(e.target.value)} placeholder="一句话介绍自己" />
            </div>
            <Button onClick={saveProfile} loading={savingProfile}>保存资料</Button>
          </div>
        </TabsContent>

        <TabsContent value="password">
          <div className="max-w-md space-y-4 rounded-md border border-border bg-card p-5">
            <div className={inputWrap}>
              <Label className={labelCls}>当前密码</Label>
              <Input type="password" value={pwd.old} onChange={(e) => setPwd((p) => ({ ...p, old: e.target.value }))} />
            </div>
            <div className={inputWrap}>
              <Label className={labelCls}>新密码（至少 6 位）</Label>
              <Input type="password" value={pwd.next} onChange={(e) => setPwd((p) => ({ ...p, next: e.target.value }))} />
            </div>
            <Button onClick={changePassword} loading={savingPwd} disabled={!pwd.old || !pwd.next}>修改密码</Button>
          </div>
        </TabsContent>

        <TabsContent value="email">
          <div className="max-w-md space-y-4 rounded-md border border-border bg-card p-5">
            <div className={inputWrap}>
              <Label className={labelCls}>新邮箱</Label>
              <Input type="email" value={email.addr} onChange={(e) => setEmail((p) => ({ ...p, addr: e.target.value }))} />
            </div>
            <div className="flex gap-2">
              <Input value={email.code} onChange={(e) => setEmail((p) => ({ ...p, code: e.target.value }))} placeholder="验证码" />
              <Button variant="outline" onClick={() => sendCode(email.addr, 'change_email')}>发送验证码</Button>
            </div>
            <Button onClick={changeEmail}>修改邮箱</Button>
          </div>
        </TabsContent>
      </Tabs>
    </AppShell>
  )
}
