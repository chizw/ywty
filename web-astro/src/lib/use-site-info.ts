// 客户端（React islands）站点信息 hook：单飞请求 + sessionStorage 缓存
// 首帧渲染固定用默认值，避免 SSR/注水不一致；随后异步更新为真实站点信息
import { useEffect, useState } from 'react'
import { DEFAULT_SITE_INFO, normalizeSiteInfo, type SiteInfo } from './site'

const SESSION_KEY = 'ywty.site.info'
let promise: Promise<SiteInfo> | null = null

function readSession(): SiteInfo | null {
  try {
    const parsed: unknown = JSON.parse(sessionStorage.getItem(SESSION_KEY) || '')
    return normalizeSiteInfo(parsed && typeof parsed === 'object' ? parsed : null)
  } catch {
    return null
  }
}

function request(): Promise<SiteInfo> {
  if (!promise) {
    promise = fetch('/api/v1/site/info', { headers: { Accept: 'application/json' } })
      .then(async (res) => (res.ok ? normalizeSiteInfo(await res.json()) : DEFAULT_SITE_INFO))
      .catch(() => DEFAULT_SITE_INFO)
  }
  return promise
}

export function useSiteInfo(): SiteInfo {
  const [info, setInfo] = useState<SiteInfo>(DEFAULT_SITE_INFO)

  useEffect(() => {
    let alive = true
    const cached = readSession()
    if (cached) setInfo(cached)
    request().then((v) => {
      if (!alive) return
      setInfo(v)
      try {
        sessionStorage.setItem(SESSION_KEY, JSON.stringify(v))
      } catch {
        /* ignore */
      }
    })
    return () => {
      alive = false
    }
  }, [])

  return info
}
