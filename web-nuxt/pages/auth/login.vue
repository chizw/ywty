<script setup lang="ts">
definePageMeta({ layout: 'auth', middleware: 'guest' })

const { login } = useAuth()
const route = useRoute()
const router = useRouter()

const form = reactive({ account: '', password: '' })
const loading = ref(false)
const errorMsg = ref('')

async function onSubmit() {
  if (!form.account || !form.password) {
    errorMsg.value = '请输入账号和密码'
    return
  }
  loading.value = true
  errorMsg.value = ''
  try {
    await login(form.account, form.password)
    const redirect = (route.query.redirect as string) || '/dashboard'
    await router.push(redirect)
  } catch (err: any) {
    errorMsg.value = err?.statusMessage || err?.message || '登录失败'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <Card>
    <CardHeader>
      <CardTitle>登录</CardTitle>
      <CardDescription>使用账号密码登录�?云雾图驿</CardDescription>
    </CardHeader>
    <CardContent>
      <form class="space-y-4" @submit.prevent="onSubmit">
        <div class="space-y-2">
          <Label for="login-account">账号</Label>
          <Input
            id="login-account"
            v-model="form.account"
            type="text"
            required
            placeholder="用户�?/ 邮箱 / 手机�?
          />
        </div>
        <div class="space-y-2">
          <Label for="login-password">密码</Label>
          <Input
            id="login-password"
            v-model="form.password"
            type="password"
            required
            minlength="6"
            placeholder="至少 6 �?
          />
        </div>
        <Alert v-if="errorMsg" variant="destructive">
          <AlertDescription>{{ errorMsg }}</AlertDescription>
        </Alert>
        <Button type="submit" :loading="loading" class="w-full">登录</Button>
      </form>
    </CardContent>
    <CardFooter class="justify-center">
      <p class="text-sm text-muted-foreground">
        还没有账号？
        <NuxtLink to="/auth/register" class="text-primary hover:underline">立即注册</NuxtLink>
      </p>
    </CardFooter>
  </Card>
</template>
