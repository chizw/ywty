// Nitro BFF：转发到 Go 后端，避免 SSR 阶段 CORS / 鉴权头泄漏到浏览器
export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const path = event.path.replace(/^\/api/, '')
  const target = `${config.apiBase}${path}`

  try {
    return await $fetch(target, {
      method: event.method,
      headers: {
        // 仅透传必要头
        authorization: getHeader(event, 'authorization') || '',
        'x-token': getHeader(event, 'x-token') || '',
      },
      query: getQuery(event),
      body: ['GET', 'HEAD'].includes(event.method || '')
        ? undefined
        : await readBody(event),
    })
  } catch (err: any) {
    throw createError({
      statusCode: err?.statusCode || 500,
      statusMessage: err?.message || 'BFF proxy error',
    })
  }
})
