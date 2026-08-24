// Vue 共享状态（跨 Island 单例）：
// - authState：登录态（从 cookie 初始化，登录/登出时同步 cookie）
// - messageState：全局 toast 队列（由 AppToaster 渲染）
// - 供公开页 Vue Islands 使用；React 侧用各自 Zustand store
import { reactive } from 'vue'
import { clearAuthPair, readAuthPairFromStorage, writeAuthPair } from './auth'
import type { TokenPair, UserInfo } from './auth'
import { useApi } from './api'

// ---------- 认证状态 ----------
export const authState = reactive({
  user: null as UserInfo | null,
  accessToken: '',
  refreshToken: '',
  ready: false,
})

export function initAuth(): void {
  const pair = readAuthPairFromStorage()
  if (pair) {
    authState.user = pair.user
    authState.accessToken = pair.access_token
    authState.refreshToken = pair.refresh_token
  }
  authState.ready = true
}

export function isLoggedIn(): boolean {
  return !!authState.accessToken && !!authState.user
}

export function isAdmin(): boolean {
  return authState.user?.role === 'admin' || authState.user?.role === 'super_admin'
}

function applyPair(pair: TokenPair): void {
  authState.user = pair.user
  authState.accessToken = pair.access_token
  authState.refreshToken = pair.refresh_token
  writeAuthPair(pair)
}

export async function login(account: string, password: string): Promise<TokenPair> {
  const api = useApi()
  const pair = await api.post<TokenPair>('/api/v1/auth/login', { account, password })
  applyPair(pair)
  return pair
}

export async function register(payload: {
  username: string
  email: string
  password: string
  phone?: string
  captcha_id?: string
  captcha_code?: string
}): Promise<TokenPair> {
  const api = useApi()
  const pair = await api.post<TokenPair>('/api/v1/auth/register', payload)
  applyPair(pair)
  return pair
}

export async function logout(): Promise<void> {
  try {
    await useApi().post('/api/v1/auth/logout', {})
  } catch {
    /* 后端报错也清空本地 */
  }
  authState.user = null
  authState.accessToken = ''
  authState.refreshToken = ''
  clearAuthPair()
}

/** 供 SSR 端把 middleware 解析的 auth 传给岛屿初始化 */
export function setAuthFromServer(pair: TokenPair | null): void {
  if (pair) {
    authState.user = pair.user
    authState.accessToken = pair.access_token
    authState.refreshToken = pair.refresh_token
  }
  authState.ready = true
}

// ---------- 消息队列 ----------
export type ToastKind = 'success' | 'error' | 'warning' | 'info'
export interface ToastItem {
  id: number
  kind: ToastKind
  text: string
  ttl: number
}

let toastSeq = 0

export const messageState = reactive<{ toasts: ToastItem[] }>({
  toasts: [],
})

export function pushToast(kind: ToastKind, text: string, ttl = 3000): void {
  const id = ++toastSeq
  messageState.toasts.push({ id, kind, text, ttl })
  if (!import.meta.env.SSR) {
    setTimeout(() => {
      const idx = messageState.toasts.findIndex((t) => t.id === id)
      if (idx >= 0) messageState.toasts.splice(idx, 1)
    }, ttl)
  }
}

export const message = {
  success: (text: string) => pushToast('success', text),
  error: (text: string) => pushToast('error', text),
  warning: (text: string) => pushToast('warning', text),
  info: (text: string) => pushToast('info', text),
}

// 模块加载即初始化（浏览器端读到 cookie 登录态；SSR 端为空）
if (!import.meta.env.SSR) {
  initAuth()
}
