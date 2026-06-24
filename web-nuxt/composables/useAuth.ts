// useAuth 组合�?API：方便组件以 useAuth() 方式调用
import { useAuthStore } from '~/stores/auth'
import { useApi } from '~/composables/useApi'
import type { TokenPair } from '~/types/api'
import type { UserInfo } from '~/types/api'

export function useAuth() {
  const store = useAuthStore()
  const api = useApi()

  async function login(account: string, password: string) {
    const res = await api.post<TokenPair>('/api/v1/auth/login', { account, password })
    store.setTokens(res)
    return res
  }

  async function register(payload: { username: string; email: string; password: string; phone?: string }) {
    const res = await api.post<TokenPair>('/api/v1/auth/register', payload)
    store.setTokens(res)
    return res
  }

  async function logout() {
    try {
      await api.post('/api/v1/auth/logout', {})
    } catch {
      // 即使后端报错也清空本�?    }
    store.clear()
    await navigateTo('/')
  }

  async function fetchMe() {
    const me = await api.get<UserInfo>('/api/v1/auth/me')
    store.setUser(me)
    return me
  }

  return {
    user: computed(() => store.user),
    isLoggedIn: computed(() => store.isLoggedIn),
    isAdmin: computed(() => store.isAdmin),
    accessToken: computed(() => store.accessToken),
    login,
    register,
    logout,
    fetchMe,
    hydrate: () => store.hydrate(),
  }
}
