// React 共享状态（Zustand）：auth / toast / stats
// 供用户中心与后台的 React Islands 使用（模块单例，同页多个 island 共享）
import { create } from 'zustand'
import { clearAuthPair, readAuthPairFromStorage, writeAuthPair } from './auth'
import type { TokenPair, UserInfo } from './auth'
import { useApi } from './api'

// ---------- 认证 ----------
interface AuthState {
  user: UserInfo | null
  accessToken: string
  refreshToken: string
  hydrated: boolean
  init: (pair: TokenPair | null) => void
  setPair: (pair: TokenPair) => void
  setUser: (user: UserInfo) => void
  clear: () => void
  logout: () => Promise<void>
  fetchMe: () => Promise<UserInfo | null>
  hydrate: () => void
}

export const useAuthStore = create<AuthState>((set, get) => ({
  user: null,
  accessToken: '',
  refreshToken: '',
  hydrated: false,

  init: (pair) => {
    if (get().hydrated) return
    if (pair) {
      set({
        user: pair.user,
        accessToken: pair.access_token,
        refreshToken: pair.refresh_token,
      })
    }
    set({ hydrated: true })
  },

  setPair: (pair) => {
    set({ user: pair.user, accessToken: pair.access_token, refreshToken: pair.refresh_token })
    writeAuthPair(pair)
  },

  setUser: (user) => set({ user }),

  clear: () => {
    clearAuthPair()
    set({ user: null, accessToken: '', refreshToken: '' })
  },

  logout: async () => {
    try {
      await useApi().post('/api/v1/auth/logout', {})
    } catch {
      /* 后端报错也清空本地 */
    }
    get().clear()
  },

  fetchMe: async () => {
    try {
      const me = await useApi().get<UserInfo>('/api/v1/auth/me')
      get().setUser(me)
      return me
    } catch {
      return null
    }
  },

  hydrate: () => {
    const pair = readAuthPairFromStorage()
    if (pair) {
      set({
        user: pair.user,
        accessToken: pair.access_token,
        refreshToken: pair.refresh_token,
        hydrated: true,
      })
    } else {
      set({ hydrated: true })
    }
  },
}))

export function isAdminUser(user: UserInfo | null | undefined): boolean {
  return !!user && (user.role === 'admin' || user.role === 'super_admin' || user.is_admin === true)
}

export function isSuperAdmin(user: UserInfo | null | undefined): boolean {
  return !!user && user.is_super_admin === true
}

// ---------- Toast ----------
export type ToastKind = 'success' | 'error' | 'warning' | 'info'
export interface ToastItem {
  id: number
  kind: ToastKind
  text: string
}

interface ToastState {
  toasts: ToastItem[]
  push: (kind: ToastKind, text: string) => void
  dismiss: (id: number) => void
}

let toastSeq = 0

export const useToastStore = create<ToastState>((set, get) => ({
  toasts: [],
  push: (kind, text) => {
    const id = ++toastSeq
    set({ toasts: [...get().toasts, { id, kind, text }] })
    setTimeout(() => get().dismiss(id), 3200)
  },
  dismiss: (id) => set({ toasts: get().toasts.filter((t) => t.id !== id) }),
}))

export const toast = {
  success: (text: string) => useToastStore.getState().push('success', text),
  error: (text: string) => useToastStore.getState().push('error', text),
  warning: (text: string) => useToastStore.getState().push('warning', text),
  info: (text: string) => useToastStore.getState().push('info', text),
}

// ---------- 统计 ----------
interface StatsState {
  photos: number
  albums: number
  usedBytes: number
  capacityBytes: number
  loading: boolean
  refresh: () => Promise<void>
  bumpPhotos: (delta: number, deltaBytes?: number) => void
}

export const useStatsStore = create<StatsState>((set, get) => ({
  photos: 0,
  albums: 0,
  usedBytes: 0,
  capacityBytes: 0,
  loading: false,

  refresh: async () => {
    if (get().loading) return
    set({ loading: true })
    try {
      const api = useApi()
      const [photosRes, albumsRes, cap] = await Promise.all([
        api.get<any>('/api/v1/photos', { query: { page: 1, per_page: 1 }, raw: true }).catch(() => null),
        api.get<any>('/api/v1/albums', { raw: true }).catch(() => null),
        api.get<any>('/api/v1/capacity').catch(() => null),
      ])
      const photoCount = Number(photosRes?.meta?.total ?? 0)
      const albumList: unknown[] = Array.isArray(albumsRes?.data) ? albumsRes.data : []
      set({
        photos: photoCount,
        albums: albumList.length,
        usedBytes: Number(cap?.used ?? 0),
        capacityBytes: Number(cap?.capacity ?? 0),
      })
    } catch {
      /* ignore */
    } finally {
      set({ loading: false })
    }
  },

  bumpPhotos: (delta, deltaBytes = 0) =>
    set((s) => ({
      photos: Math.max(0, s.photos + delta),
      usedBytes: Math.max(0, s.usedBytes + deltaBytes),
    })),
}))
