// 站点信息：SSR 直连内部后端 GET /api/v1/site/info，内存缓存 60s
// 任何失败（后端未启动 / 接口暂不可用 / 响应异常）都降级为默认值，不缓存失败结果
import { APP_NAME, APP_TAGLINE } from './constants'

export interface SiteInfo {
  name: string
  description: string
  keywords: string
  footer: string
  icp: string
  allow_register: boolean
  require_email_verify: boolean
}

export const DEFAULT_SITE_INFO: SiteInfo = {
  name: APP_NAME,
  description: APP_TAGLINE,
  keywords: '',
  footer: '',
  icp: '',
  allow_register: true,
  require_email_verify: true,
}

const CACHE_TTL_MS = 60_000

let cache: { value: SiteInfo; expiresAt: number } | null = null
let inflight: Promise<SiteInfo> | null = null

/** 宽容解析：兼容裸对象与 {code,message,data} 信封两种返回形态 */
export function normalizeSiteInfo(raw: unknown): SiteInfo {
  const o = (raw && typeof raw === 'object' ? raw : {}) as Record<string, unknown>
  if (typeof o.code === 'number' && o.data && typeof o.data === 'object') {
    return normalizeSiteInfo(o.data)
  }
  const str = (k: keyof SiteInfo, dflt: string) =>
    typeof o[k] === 'string' ? (o[k] as string) : dflt
  return {
    name: str('name', DEFAULT_SITE_INFO.name),
    description: str('description', DEFAULT_SITE_INFO.description),
    keywords: str('keywords', ''),
    footer: str('footer', ''),
    icp: str('icp', ''),
    allow_register: typeof o.allow_register === 'boolean' ? o.allow_register : true,
    require_email_verify:
      typeof o.require_email_verify === 'boolean' ? o.require_email_verify : true,
  }
}

async function fetchSiteInfo(): Promise<SiteInfo> {
  const base =
    typeof process !== 'undefined' && process.env?.API_INTERNAL
      ? String(process.env.API_INTERNAL)
      : ''
  const res = await fetch(`${base}/api/v1/site/info`, {
    headers: { Accept: 'application/json' },
    signal: AbortSignal.timeout(3000),
  })
  if (!res.ok) throw new Error(`site info HTTP ${res.status}`)
  return normalizeSiteInfo(await res.json())
}

export async function getSiteInfo(): Promise<SiteInfo> {
  const now = Date.now()
  if (cache && cache.expiresAt > now) return cache.value
  if (!inflight) {
    inflight = fetchSiteInfo()
      .then((info) => {
        cache = { value: info, expiresAt: Date.now() + CACHE_TTL_MS }
        return info
      })
      .catch(() => DEFAULT_SITE_INFO)
      .finally(() => {
        inflight = null
      })
  }
  return inflight
}
