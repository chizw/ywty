// 鉴权 store：token、user、登录/注册/登出
import { defineStore } from 'pinia'
import type { TokenPair, UserInfo } from '~/types/api'

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
    // 客户端首屏：尝试从 cookie/localStorage 恢复登录态
    hydrate() {
      if (this.hydrated) return
      this.hydrated = true
      if (typeof window === 'undefined') return
      const raw = window.localStorage.getItem('ywty.auth')
      if (!raw) return
      try {
        const data = JSON.parse(raw) as TokenPair
        this.accessToken = data.access_token
        this.refreshToken = data.refresh_token
        this.expiresAt = data.expires_at
        this.user = data.user
      } catch {
        // 损坏数据清空
        window.localStorage.removeItem('ywty.auth')
      }
    },
    setTokens(payload: TokenPair) {
      this.accessToken = payload.access_token
      this.refreshToken = payload.refresh_token
      this.expiresAt = payload.expires_at
      this.user = payload.user
      if (typeof window !== 'undefined') {
        window.localStorage.setItem('ywty.auth', JSON.stringify(payload))
      }
    },
    setUser(user: UserInfo) {
      this.user = user
      if (typeof window !== 'undefined') {
        const raw = window.localStorage.getItem('ywty.auth')
        if (raw) {
          const data = JSON.parse(raw) as TokenPair
          data.user = user
          window.localStorage.setItem('ywty.auth', JSON.stringify(data))
        }
      }
    },
    clear() {
      this.accessToken = ''
      this.refreshToken = ''
      this.expiresAt = ''
      this.user = null
      if (typeof window !== 'undefined') {
        window.localStorage.removeItem('ywty.auth')
      }
    },
  },
})
