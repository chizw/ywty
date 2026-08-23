// 共享 API 客户端：token 注入、统一错误处理、401 单飞刷新 + 重试
// 双端可用：浏览器（islands 内）自动从 cookie 取 token；
// 服务端（SSR）只调公开接口，token 通过 opts.token 显式传入。
import { clearAuthPair, isServer, readAuthPairFromStorage, writeAuthPair } from './auth'
import type { TokenPair } from './auth'

export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data?: T
  meta?: {
    current_page: number
    per_page: number
    total: number
    last_page: number
  }
}

export interface RequestOptions {
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH'
  body?: unknown
  query?: Record<string, unknown>
  headers?: Record<string, string>
  /** 跳过统一解包，返回完整信封（默认 false：返回 res.data） */
  raw?: boolean
  /** 显式 token（SSR 场景），默认从 cookie/localStorage 读取 */
  token?: string
}

/** 业务错误：携带 apiCode 与 HTTP status */
export class ApiError extends Error {
  apiCode: number
  status?: number
  constructor(message: string, apiCode = -1, status?: number) {
    super(message)
    this.name = 'ApiError'
    this.apiCode = apiCode
    this.status = status
  }
}

// 全局唯一 refresh 任务：避免 401 并发刷新把 refresh_token 旋转失效
let refreshPromise: Promise<TokenPair | null> | null = null

async function doRefresh(refreshToken: string): Promise<TokenPair | null> {
  try {
    const res = await fetch('/api/v1/auth/refresh', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refreshToken }),
    })
    const json = (await res.json()) as ApiResponse<TokenPair>
    if (res.ok && json.code === 0 && json.data) return json.data
    return null
  } catch {
    return null
  }
}

function buildQuery(query?: Record<string, unknown>): string {
  if (!query) return ''
  const params = new URLSearchParams()
  for (const [k, v] of Object.entries(query)) {
    if (v === undefined || v === null || v === '') continue
    params.set(k, String(v))
  }
  const s = params.toString()
  return s ? `?${s}` : ''
}

export function useApi() {
  // 浏览器：PUBLIC_API_BASE（默认相对路径，走 dev 代理 / 生产反代）
  // 服务端：API_INTERNAL 直连后端（避免相对路径在 SSR 请求上打回自身/代理）
  const publicBase = (import.meta.env.PUBLIC_API_BASE as string) || ''
  const internalBase = isServer()
    ? (typeof process !== 'undefined' ? (process.env.API_INTERNAL as string | undefined) : undefined) || publicBase
    : publicBase
  const baseURL = isServer() ? internalBase : publicBase

  async function request<T = unknown>(path: string, opts: RequestOptions = {}, _retried = false): Promise<T> {
    const headers: Record<string, string> = {
      Accept: 'application/json',
      ...(opts.headers || {}),
    }
    // token：优先显式传入（SSR），否则浏览器从 cookie 读取
    const token = opts.token || (isServer() ? undefined : readAuthPairFromStorage()?.access_token)
    if (token) headers.Authorization = `Bearer ${token}`

    let body: BodyInit | undefined
    if (opts.body !== undefined && opts.body !== null) {
      if (opts.body instanceof FormData) {
        body = opts.body
      } else {
        body = JSON.stringify(opts.body)
        headers['Content-Type'] = 'application/json'
      }
    }

    const url = `${baseURL}${path}${buildQuery(opts.query)}`
    let res: Response
    try {
      res = await fetch(url, {
        method: opts.method || 'GET',
        headers,
        body,
        credentials: 'include',
      })
    } catch (err) {
      throw new ApiError(err instanceof Error ? err.message : '网络错误', -1, 0)
    }

    let json: unknown
    try {
      json = await res.json()
    } catch {
      throw new ApiError(`请求失败（HTTP ${res.status}）`, -1, res.status)
    }

    if (opts.raw) return json as T

    const env = json as ApiResponse<T>
    if (env && typeof env === 'object' && 'code' in env) {
      if (env.code !== 0) {
        const err = new ApiError(env.message || '请求失败', env.code, res.status)
        // 401xx 业务错误：尝试单飞刷新后重试（仅浏览器端）
        if (Math.floor(env.code / 100) === 401 && !isServer() && !_retried) {
          const pair = readAuthPairFromStorage()
          if (pair?.refresh_token) {
            if (!refreshPromise) {
              refreshPromise = doRefresh(pair.refresh_token).finally(() => {
                refreshPromise = null
              })
            }
            const newPair = await refreshPromise
            if (newPair) {
              writeAuthPair(newPair)
              return await request<T>(path, opts, true)
            }
            // refresh 失败：仅在用户触发的非 GET 请求上清空登录态
            if (opts.method && opts.method !== 'GET') {
              clearAuthPair()
            }
          }
        }
        throw err
      }
      return env.data as T
    }
    return json as T
  }

  return {
    get: <T = unknown>(path: string, opts?: Omit<RequestOptions, 'method' | 'body'>) =>
      request<T>(path, { ...opts, method: 'GET' }),
    post: <T = unknown>(path: string, body?: unknown, opts?: Omit<RequestOptions, 'method' | 'body'>) =>
      request<T>(path, { ...opts, method: 'POST', body }),
    put: <T = unknown>(path: string, body?: unknown, opts?: Omit<RequestOptions, 'method' | 'body'>) =>
      request<T>(path, { ...opts, method: 'PUT', body }),
    patch: <T = unknown>(path: string, body?: unknown, opts?: Omit<RequestOptions, 'method' | 'body'>) =>
      request<T>(path, { ...opts, method: 'PATCH', body }),
    del: <T = unknown>(path: string, opts?: Omit<RequestOptions, 'method' | 'body'>) =>
      request<T>(path, { ...opts, method: 'DELETE' }),
    request,
  }
}
