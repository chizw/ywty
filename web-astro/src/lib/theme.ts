// 主题管理（框架无关）：读写 cookie + 切换 document class
import { THEME_COOKIE_NAME } from './constants'
import { isServer } from './auth'

export type Theme = 'light' | 'dark'

function readCookie(name: string): string | null {
  if (isServer()) return null
  const m = document.cookie.match(new RegExp(`(?:^|;\\s*)${name}=([^;]*)`))
  return m ? m[1] : null
}

function writeCookie(name: string, value: string, days = 365): void {
  if (isServer()) return
  const expires = new Date(Date.now() + days * 864e5).toUTCString()
  document.cookie = `${name}=${encodeURIComponent(value)}; path=/; max-age=${days * 86400}; expires=${expires}; samesite=lax`
}

/** 读取当前主题（未设置时跟随系统偏好） */
export function getTheme(): Theme {
  const stored = readCookie(THEME_COOKIE_NAME)
  if (stored) return stored === 'dark' ? 'dark' : 'light'
  if (!isServer() && window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) {
    return 'dark'
  }
  return 'light'
}

/** 应用主题到 <html>（同时持久化 cookie） */
export function applyTheme(theme: Theme, persist = true): void {
  if (isServer()) return
  const root = document.documentElement
  if (theme === 'dark') root.classList.add('dark')
  else root.classList.remove('dark')
  if (persist) writeCookie(THEME_COOKIE_NAME, theme)
}

export function toggleTheme(): Theme {
  const next: Theme = getTheme() === 'dark' ? 'light' : 'dark'
  applyTheme(next)
  return next
}

export function initTheme(): void {
  applyTheme(getTheme(), false)
}
