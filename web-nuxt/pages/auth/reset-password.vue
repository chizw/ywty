<script setup lang="ts">
// 重置密码（公开页）
definePageMeta({ layout: 'auth', middleware: 'guest' })

const route = useRoute()
const router = useRouter()
const api = useApi()
const mode = ref<'email' | 'phone'>('email')
const form = reactive({ account: '', code: '', password: '' })
const loading = ref(false)
const msg = ref('')

async function submit() {
  loading.value = true
  msg.value = ''
  try {
    const path = mode.value === 'email'
      ? '/api/v1/auth/reset-password'
      : '/api/v1/auth/reset-password/phone'
    await api.post(path, {
      [mode.value === 'email' ? 'email' : 'phone']: form.account,
      code: form.code,
      new_password: form.password,
    })
    msg.value = '重置成功，即将跳到登录页'
    setTimeout(() => router.push('/auth/login'), 1000)
  } catch (err: any) {
    msg.value = err?.statusMessage || '重置失败'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <Card>
    <CardHeader>
      <CardTitle>找回密码</CardTitle>
      <CardDescription>通过 {{ mode === 'email' ? '邮箱' : '手机' }} 验证后重置</CardDescription>
    </CardHeader>
    <CardContent>
      <Tabs v-model="mode" class="w-full">
        <TabsList class="w-full">
          <TabsTrigger value="email" class="flex-1">邮箱验证</TabsTrigger>
          <TabsTrigger value="phone" class="flex-1">手机验证</TabsTrigger>
        </TabsList>
      </Tabs>

      <form class="mt-4 space-y-4" @submit.prevent="submit">
        <div class="space-y-2">
          <Label for="reset-account">{{ mode === 'email' ? '邮箱' : '手机号' }}</Label>
          <Input
            id="reset-account"
            v-model="form.account"
            :placeholder="mode === 'email' ? '邮箱' : '手机号'"
          />
        </div>
        <VerifyCodeInput
          :channel="mode"
          :account="form.account"
          :event="'reset_password'"
        />
        <div class="space-y-2">
          <Label for="reset-code">验证码</Label>
          <Input
            id="reset-code"
            v-model="form.code"
            placeholder="验证码（自动填充，或手动输入）"
          />
        </div>
        <div class="space-y-2">
          <Label for="reset-password">新密码</Label>
          <Input
            id="reset-password"
            v-model="form.password"
            type="password"
            minlength="6"
            placeholder="新密码（至少 6 位）"
          />
        </div>
        <Alert v-if="msg" :variant="msg.includes('成功') ? 'success' : 'destructive'">
          <AlertDescription>{{ msg }}</AlertDescription>
        </Alert>
        <Button type="submit" :loading="loading" class="w-full">重置密码</Button>
      </form>
    </CardContent>
    <CardFooter class="justify-center">
      <p class="text-sm text-muted-foreground">
        记起密码了？<NuxtLink to="/auth/login" class="text-primary hover:underline">去登录</NuxtLink>
      </p>
    </CardFooter>
  </Card>
</template>
