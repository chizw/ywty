// 鉴权 store：token、user、登录/注册/登出
import { defineStore } from 'pinia'
import type { TokenPair, UserInfo } from '~/types/api'

// cookie 名（同时被 SSR + 客户端使用）
const COOKIE_NAME = 'ywty.auth'

// 通过 Nuxt 内置 useCookie 做 SSR + 客户端双端同步
// 注意：useCookie 默认会用 JSON.stringify 再 encodeURIComponent 编码；
// 读取时用 destr(decodeURIComponent(value))，destr 会先尝试 JSON.parse。
// 因此 cookie.value 可能已经是对象（destr 解析成功）也可能是字符串。
// 我们兼容两种情况。
function useAuthCookie() {
  return useCookie<string | TokenPair>(COOKIE_NAME, {
    default: () => '' as any,
    maxAge: 7 * 24 * 3600,
    sameSite: 'lax',
    path: '/',
  })
}

function coerceTokenPair(v: string | TokenPair | null | undefined): TokenPair | null {
  if (!v) return null
  if (typeof v === 'object') return v
  try {
    return JSON.parse(v) as TokenPair
  } catch {
    return null
  }
}

export const useAuthStore = defineStore('auth', {
  state: () => ({
    accessToken: '' as string,
    refreshToken: '' as string,
    expiresAt: '' as string,
    user: null as UserInfo | null,
    hydrated: false,
  }),
  getters: {
    isLoggedIn: (s) => !!s.accessToken && !!s.user,
    isAdmin: (s) => !!s.user?.is_admin,
  },
  actions: {
    // 通用 hydrate：SSR 和客户端都从 cookie 读
    hydrate() {
      if (this.hydrated) return
      this.hydrated = true
      const cookie = useAuthCookie()
      const data = coerceTokenPair(cookie.value as any)
      if (!data) return
      this.accessToken = data.access_token
      this.refreshToken = data.refresh_token
      this.expiresAt = data.expires_at
      this.user = data.user
    },
    setTokens(payload: TokenPair) {
      this.accessToken = payload.access_token
      this.refreshToken = payload.refresh_token
      this.expiresAt = payload.expires_at
      this.user = payload.user
      // 通过 Nuxt useCookie 同步到两侧
      // useCookie 写入时会再 JSON.stringify + encodeURIComponent 一次，
      // 写入时给 string，读取时 destr 解析后就是我们想要的 object
      const cookie = useAuthCookie()
      cookie.value = JSON.stringify(payload) as any
      // 兼容：同时写 localStorage
      if (typeof window !== 'undefined') {
        try { window.localStorage.setItem('ywty.auth', JSON.stringify(payload)) } catch { /* noop */ }
      }
    },
    setUser(user: UserInfo) {
      this.user = user
      const cookie = useAuthCookie()
      const data = coerceTokenPair(cookie.value as any)
      if (data) {
        data.user = user
        cookie.value = JSON.stringify(data) as any
      }
      if (typeof window !== 'undefined') {
        try {
          const lsRaw = window.localStorage.getItem('ywty.auth')
          if (lsRaw) {
            const lsData = JSON.parse(lsRaw)
            lsData.user = user
            window.localStorage.setItem('ywty.auth', JSON.stringify(lsData))
          }
        } catch { /* noop */ }
      }
    },
    clear() {
      this.accessToken = ''
      this.refreshToken = ''
      this.expiresAt = ''
      this.user = null
      const cookie = useAuthCookie()
      cookie.value = '' as any
      if (typeof window !== 'undefined') {
        try { window.localStorage.removeItem('ywty.auth') } catch { /* noop */ }
      }
    },
  },
})
