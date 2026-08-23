// 共享认证层：TokenPair 解析、cookie 读写（浏览器）、SSR 端解析（middleware 用）
import { AUTH_COOKIE_NAME } from './constants'

export interface UserInfo {
  id: number
  uuid?: string
  username: string
  name?: string
  email: string
  avatar?: string | null
  role?: string
  is_admin?: boolean
  is_super_admin?: boolean
  status?: string
  phone?: string
  created_at?: string
}

export interface TokenPair {
  access_token: string
  refresh_token: string
  token_type: string
  expires_at: string
  user: UserInfo
}

/** 是否服务端 */
export const isServer = () => import.meta.env.SSR

/** 管理员判定 */
export function isAdminUser(user: UserInfo | null | undefined): boolean {
  if (!user) return false
  return user.role === 'admin' || user.role === 'super_admin' || user.is_admin === true
}

/** 是否登录 */
export function isLoggedInPair(pair: TokenPair | null): boolean {
  return !!pair?.access_token && !!pair?.user
}

/**
 * 解析 cookie 原始值（写入时为 encodeURIComponent(JSON.stringify(...))）
 */
export function parseAuthPair(raw: string | undefined | null): TokenPair | null {
  if (!raw) return null
  let decoded = raw
  try {
    // 值被 encodeURIComponent 编码过；先试解码后的 JSON
    decoded = decodeURIComponent(raw)
  } catch {
    decoded = raw
  }
  // 解码后若仍有包裹引号/前缀，尝试提取 JSON 部分
  try {
    return JSON.parse(decoded) as TokenPair
  } catch {
    /* fallthrough */
  }
  // 原始未解码
  try {
    return JSON.parse(raw) as TokenPair
  } catch {
    return null
  }
}

function decodeCookieValue(v: string): string {
  try {
    return decodeURIComponent(v)
  } catch {
    return v
  }
}

/**
 * 读取认证 cookie 中的 TokenPair
 * 浏览器：解析 document.cookie；服务端：返回 null（SSR 数据由 Astro middleware 注入）
 */
export function readAuthPair(): TokenPair | null {
  if (isServer()) return null
  const m = document.cookie.match(new RegExp(`(?:^|;\\s*)${AUTH_COOKIE_NAME}=([^;]*)`))
  if (!m?.[1]) return null
  const raw = decodeCookieValue(m[1])
  return parseAuthPair(raw)
}

/** 仅浏览器端写入 cookie + localStorage 镜像 */
export function writeAuthPair(pair: TokenPair): void {
  if (isServer()) return
  const encoded = encodeURIComponent(JSON.stringify(pair))
  const days = 7
  const expires = new Date(Date.now() + days * 864e5).toUTCString()
  document.cookie = `${AUTH_COOKIE_NAME}=${encoded}; path=/; max-age=${days * 86400}; expires=${expires}; samesite=lax`
  try {
    window.localStorage.setItem(AUTH_COOKIE_NAME, JSON.stringify(pair))
  } catch {
    /* noop */
  }
}

/** 仅浏览器端清除 */
export function clearAuthPair(): void {
  if (isServer()) return
  document.cookie = `${AUTH_COOKIE_NAME}=; path=/; max-age=0; expires=${new Date(0).toUTCString()}; samesite=lax`
  try {
    window.localStorage.removeItem(AUTH_COOKIE_NAME)
  } catch {
    /* noop */
  }
}

/** 从 localStorage/cookie 兜底读取（浏览器端，供岛屿初始化） */
export function readAuthPairFromStorage(): TokenPair | null {
  if (isServer()) return null
  try {
    const ls = window.localStorage.getItem(AUTH_COOKIE_NAME)
    if (ls) {
      const pair = parseAuthPair(ls)
      if (pair) return pair
    }
  } catch {
    /* noop */
  }
  return readAuthPair()
}
