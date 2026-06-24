const http = require('http')
function req(opts, body) {
  return new Promise((resolve, reject) => {
    const r = http.request(opts, (res) => {
      const chunks = []
      res.on('data', (c) => chunks.push(c))
      res.on('end', () => {
        const buf = Buffer.concat(chunks)
        resolve({ status: res.statusCode, headers: res.headers, body: buf.toString('utf8') })
      })
    })
    r.on('error', reject)
    if (body) r.write(body)
    r.end()
  })
}

async function main() {
  const login = await req({
    host: '127.0.0.1', port: 8080, method: 'POST', path: '/api/v1/auth/login',
    headers: { 'Content-Type': 'application/json' },
  }, JSON.stringify({ account: 'admin', password: 'admin123456' }))
  const tok = JSON.parse(login.body).data
  const payload = JSON.stringify({
    access_token: tok.access_token,
    refresh_token: tok.refresh_token,
    expires_at: tok.expires_at,
    user: tok.user,
  })
  const doubleEncoded = encodeURIComponent(payload)
  const cookie = `ywty.auth=${doubleEncoded}`

  for (const path of ['/auth/login', '/dashboard', '/admin', '/']) {
    const r = await req({
      host: '127.0.0.1', port: 3000, method: 'GET', path, headers: { Cookie: cookie },
    })
    const b = r.body
    // 检测关键字符串
    const isLogin = b.includes('使用账号密码登录') || b.includes('立即注册')
    const isDash = b.includes('我的图片') || b.includes('通知') || b.includes('我的相册')
    const isAdmin = b.includes('仪表盘') || b.includes('驱动管理') || b.includes('系统')
    const isHome = b.includes('ywty') && b.includes('首页')
    console.log(`[${path}] HTTP ${r.status}  login=${isLogin}  dash=${isDash}  admin=${isAdmin}  home=${isHome}`)
    console.log(`     Location: ${r.headers.location || '-'}  body-len=${b.length}`)
  }

  // 测试无 cookie
  console.log('\n--- no cookie ---')
  for (const path of ['/auth/login', '/dashboard']) {
    const r = await req({ host: '127.0.0.1', port: 3000, method: 'GET', path })
    const b = r.body
    const isLogin = b.includes('使用账号密码登录') || b.includes('立即注册')
    const isDash = b.includes('我的图片') || b.includes('通知')
    console.log(`[${path}] HTTP ${r.status}  login=${isLogin}  dash=${isDash}`)
  }
}
main().catch(console.error)
