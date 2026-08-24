// 站点元数据同步（客户端）：静态部署下构建期的 title/description 是默认值，
// 挂载后拉取 /api/v1/site/info 将真实站点信息同步到 document.title / meta。
import { useEffect } from 'react'

const SESSION_KEY = 'ywty.site.info'

function readSession(): { name?: string; description?: string; keywords?: string } | null {
  try {
    const raw = sessionStorage.getItem(SESSION_KEY)
    return raw ? JSON.parse(raw) : null
  } catch {
    return null
  }
}

export default function SiteMetaSync(): null {
  useEffect(() => {
    const cached = readSession()
    if (cached?.name) document.title = document.title.replace(/^.*?(?= · |$)/, cached.name)

    fetch('/api/v1/site/info', { headers: { Accept: 'application/json' } })
      .then((r) => (r.ok ? r.json() : Promise.reject()))
      .then((env) => {
        const info = env?.data ?? {}
        if (!info.name) return
        try {
          sessionStorage.setItem(SESSION_KEY, JSON.stringify(info))
        } catch {
          /* ignore */
        }
        // 页面 title 形如 "<页名> · <站名>" 或 "<站名> · <描述>"
        if (!document.title.includes(info.name)) {
          const sep = ' · '
          const parts = document.title.split(sep)
          parts[parts.length - 1] = info.name
          document.title = parts.join(sep)
        }
        const desc = document.querySelector('meta[name="description"]')
        if (desc && info.description) desc.setAttribute('content', info.description)
      })
      .catch(() => {})
  }, [])

  return null
}
