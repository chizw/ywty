// 统一 API 客户端：自动注入 access token、统一错误处理、统一返回 data
import { useAuthStore } from '~/stores/auth'
import type { ApiResponse } from '~/types/api'

interface RequestOptions {
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH'
  body?: any
  query?: Record<string, any>
  headers?: Record<string, string>
  /** 是否跳过统一解包（默认 false：返回 res.data） */
  raw?: boolean
}

export function useApi() {
  const config = useRuntimeConfig()
  const baseURL = config.apiBase as string

  async function request<T = any>(url: string, opts: RequestOptions = {}): Promise<T> {
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
      // 401 时尝试 refresh
      if (err?.statusCode === 401 && auth.refreshToken) {
        try {
          const refreshed = await $fetch<ApiResponse<TokenPair>>('/api/v1/auth/refresh', {
            baseURL,
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: { refresh_token: auth.refreshToken },
          })
          if (refreshed.code === 0 && refreshed.data) {
            auth.setTokens(refreshed.data)
            // 重试原请求
            return await request<T>(url, opts)
          }
        } catch {
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
