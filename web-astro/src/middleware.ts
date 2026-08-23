import { defineMiddleware } from 'astro:middleware'
import { isAdminUser, parseAuthPair } from './lib/auth'

/**
 * 生产环境反向代理
 * vite.server.proxy 仅在 dev 生效；node standalone 下由这里转发。
 * 注意 '/s/' 带尾斜杠：避免前缀匹配吞掉 /share/*、/search 等页面路由。
 */
const BACKEND = process.env.API_INTERNAL || 'http://127.0.0.1:3000'
const PROXY_EXACT = new Set(['/healthz', '/api'])
const PROXY_PREFIXES = ['/api/', '/uploads/', '/i/', '/s/']

function shouldProxy(pathname: string): boolean {
  return PROXY_EXACT.has(pathname) || PROXY_PREFIXES.some((p) => pathname.startsWith(p))
}

async function proxyToBackend(request: Request): Promise<Response> {
  const url = new URL(request.url)
  const target = `${BACKEND}${url.pathname}${url.search}`

  const headers = new Headers(request.headers)
  headers.delete('host')
  headers.delete('connection')

  const hasBody = !['GET', 'HEAD'].includes(request.method)
  const body = hasBody ? await request.arrayBuffer() : undefined

  try {
    const res = await fetch(target, { method: request.method, headers, body, redirect: 'manual' })

    const outHeaders = new Headers()
    const skip = new Set(['transfer-encoding', 'content-encoding', 'content-length'])
    res.headers.forEach((value, key) => {
      if (!skip.has(key.toLowerCase())) outHeaders.set(key, value)
    })
    // 多个 Set-Cookie 不能用 Headers.set 合并
    for (const cookie of res.headers.getSetCookie()) {
      outHeaders.append('set-cookie', cookie)
    }

    return new Response(res.body, { status: res.status, statusText: res.statusText, headers: outHeaders })
  } catch {
    return new Response(JSON.stringify({ code: -1, message: 'BFF proxy error' }), {
      status: 502,
      headers: { 'Content-Type': 'application/json' },
    })
  }
}

/**
 * 全局路由守卫 + SSR 认证注入
 * - 解析 ywty.auth cookie → context.locals.auth（供服务端组件使用）
 * - /dashboard/**：未登录跳登录页
 * - /admin/**：未登录跳登录页；非管理员跳首页
 * - /auth/**：已登录访问登录/注册页 → 跳 /dashboard
 */
export const onRequest = defineMiddleware(async (context, next) => {
  const { pathname } = context.url

  // API / 静态资源直连后端
  if (shouldProxy(pathname)) {
    return proxyToBackend(context.request)
  }

  // 解析认证 cookie
  const raw = context.cookies.get('ywty.auth')?.value
  const pair = parseAuthPair(raw)
  context.locals.auth = pair

  const isAuthPage = pathname.startsWith('/auth')
  const isUserArea = pathname.startsWith('/dashboard')
  const isAdminArea = pathname.startsWith('/admin')

  if (isAdminArea) {
    if (!pair || !isLoggedIn(pair)) {
      return context.redirect(`/auth/login?redirect=${encodeURIComponent(pathname)}`)
    }
    if (!isAdminUser(pair.user)) {
      return context.redirect('/')
    }
  } else if (isUserArea) {
    if (!pair || !isLoggedIn(pair)) {
      return context.redirect(`/auth/login?redirect=${encodeURIComponent(pathname)}`)
    }
  } else if (isAuthPage && pair) {
    // 已登录用户访问 auth 页 → 回控制台
    return context.redirect('/dashboard')
  }

  return next()
})

function isLoggedIn(pair: { access_token?: string; user?: unknown } | null): boolean {
  return !!pair?.access_token && !!pair?.user
}
