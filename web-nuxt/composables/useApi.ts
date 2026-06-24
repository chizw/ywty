// 统一 API 客户端：自动注入 access token、统一错误处理、统一返回 data
import { useAuthStore } from '~/stores/auth'
import type { ApiResponse, TokenPair } from '~/types/api'

interface RequestOptions {
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH'
  body?: any
  query?: Record<string, any>
  headers?: Record<string, string>
  /** 是否跳过统一解包（默�?false：返�?res.data�?*/
  raw?: boolean
}

// 全局唯一�?refresh 任务：避�?401 时并发刷新把 refresh_token 旋转失效
let refreshPromise: Promise<TokenPair | null> | null = null

async function doRefresh(refreshToken: string): Promise<TokenPair | null> {
  const config = useRuntimeConfig()
  const baseURL = config.apiBase as string
  try {
    const refreshed = await $fetch<ApiResponse<TokenPair>>('/api/v1/auth/refresh', {
      baseURL,
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: { refresh_token: refreshToken },
    })
    if (refreshed.code === 0 && refreshed.data) {
      return refreshed.data
    }
    return null
  } catch {
    return null
  }
}

export function useApi() {
  const config = useRuntimeConfig()
  const baseURL = config.apiBase as string

  async function request<T = any>(url: string, opts: RequestOptions = {}, _retried = false): Promise<T> {
    const auth = useAuthStore()
    const headers: Record<string, string> = {
      Accept: 'application/json',
      ...(opts.headers || {}),
    }
    if (auth.accessToken) {
      headers.Authorization = `Bearer ${auth.accessToken}`
    }

    let body: any = undefined
    if (opts.body !== undefined && opts.body !== null) {
      if (opts.body instanceof FormData) {
        body = opts.body
      } else {
        body = JSON.stringify(opts.body)
        headers['Content-Type'] = 'application/json'
      }
    }

    try {
      const res = await $fetch<ApiResponse<T>>(url, {
        baseURL,
        method: opts.method || 'GET',
        headers,
        body,
        query: opts.query,
        credentials: 'include',
      })
      if (opts.raw) return res as unknown as T
      if (res && typeof res === 'object' && 'code' in res) {
        if (res.code !== 0) {
          throw createError({
            statusCode: res.code,
            statusMessage: res.message || 'request failed',
          })
        }
        return (res as ApiResponse<T>).data as T
      }
      return res as unknown as T
    } catch (err: any) {
      // 401 时尝�?refresh（单飞：所有并发请求共享同一�?refresh�?      if (err?.statusCode === 401 && auth.refreshToken && !_retried) {
        if (!refreshPromise) {
          refreshPromise = doRefresh(auth.refreshToken).finally(() => {
            // 立刻清空，让下一�?401 走新流程
            refreshPromise = null
          })
        }
        const pair = await refreshPromise
        if (pair) {
          auth.setTokens(pair)
          // 重试原请求（标记 _retried 避免再次 401 时无限循环）
          return await request<T>(url, opts, true)
        }
        // refresh 失败：仅在用户触发的真实请求上清空登录态；
        // 后台轮询 / hydrate 期间的偶发失败不立即踢人，由上层组件显式处理
        if (opts.method && opts.method !== 'GET') {
          auth.clear()
        }
      }
      throw err
    }
  }

  return {
    get: <T = any>(url: string, opts?: Omit<RequestOptions, 'method' | 'body'>) =>
      request<T>(url, { ...opts, method: 'GET' }),
    post: <T = any>(url: string, body?: any, opts?: Omit<RequestOptions, 'method' | 'body'>) =>
      request<T>(url, { ...opts, method: 'POST', body }),
    put: <T = any>(url: string, body?: any, opts?: Omit<RequestOptions, 'method' | 'body'>) =>
      request<T>(url, { ...opts, method: 'PUT', body }),
    del: <T = any>(url: string, opts?: Omit<RequestOptions, 'method' | 'body'>) =>
      request<T>(url, { ...opts, method: 'DELETE' }),
    patch: <T = any>(url: string, body?: any, opts?: Omit<RequestOptions, 'method' | 'body'>) =>
      request<T>(url, { ...opts, method: 'PATCH', body }),
    request,
  }
}
