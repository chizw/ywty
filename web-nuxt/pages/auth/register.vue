<script setup lang="ts">
definePageMeta({ layout: 'auth', middleware: 'guest' })

const { register } = useAuth()
const router = useRouter()

const form = reactive({ username: '', email: '', password: '', phone: '' })
const loading = ref(false)
const errorMsg = ref('')

async function onSubmit() {
  loading.value = true
  errorMsg.value = ''
  try {
    await register({
      username: form.username,
      email: form.email,
      password: form.password,
      phone: form.phone || undefined,
    })
    await router.push('/dashboard')
  } catch (err: any) {
    errorMsg.value = err?.statusMessage || err?.message || '注册失败'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <Card>
    <CardHeader>
      <CardTitle>注册</CardTitle>
      <CardDescription>创建你的 云雾图驿 账号</CardDescription>
    </CardHeader>
    <CardContent>
      <form class="space-y-4" @submit.prevent="onSubmit">
        <div class="space-y-2">
          <Label for="reg-username">用户�?/Label>
          <Input id="reg-username" v-model="form.username" required minlength="3" maxlength="32" />
        </div>
        <div class="space-y-2">
          <Label for="reg-email">邮箱</Label>
          <Input id="reg-email" v-model="form.email" type="email" required />
        </div>
        <div class="space-y-2">
          <Label for="reg-phone">手机号（可选）</Label>
          <Input id="reg-phone" v-model="form.phone" />
        </div>
        <div class="space-y-2">
          <Label for="reg-password">密码</Label>
          <Input id="reg-password" v-model="form.password" type="password" required minlength="6" />
        </div>
        <Alert v-if="errorMsg" variant="destructive">
          <AlertDescription>{{ errorMsg }}</AlertDescription>
        </Alert>
        <Button type="submit" :loading="loading" class="w-full">注册</Button>
      </form>
    </CardContent>
    <CardFooter class="justify-center">
      <p class="text-sm text-muted-foreground">
        已有账号�?        <NuxtLink to="/auth/login" class="text-primary hover:underline">去登�?/NuxtLink>
      </p>
    </CardFooter>
  </Card>
</template>
