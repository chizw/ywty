# ---------- 运行时镜像（Caddy 静态托管 + 反代，仅打包预编译产物） ----------
#
# 构建上下文需由流水线/脚本预先组装：
#   linux-amd64/api   — x64 后端二进制
#   dist/             — Astro 静态导出产物（含 index.html / 404.html 等）
#   Caddyfile         — 站点配置（静态托管 + /api 反代 + 参数路由回退）
#   config.yaml       — 默认配置（来源 server-rust/config.yaml）
#
# 本地组装请使用 deploy/deploy.sh；CI 由 .github/workflows 自动完成。

FROM caddy:2-alpine AS runtime

RUN apk add --no-cache tzdata \
 && addgroup -S app && adduser -S app -G app

WORKDIR /app

COPY Caddyfile /etc/caddy/Caddyfile
COPY dist/ /srv/
COPY linux-amd64/api /app/api
COPY config.yaml /app/config.yaml

RUN mkdir -p /app/data /app/uploads /config /data \
 && chown -R app:app /app /srv /config /data

# 入口脚本：后台启动 API，前台运行 Caddy（静态托管 + 反代）
COPY --chmod=0755 <<'EOF' /app/entrypoint.sh
#!/bin/sh
set -e

# JWT 密钥解析顺序：环境变量 > 数据卷持久文件 > 自动生成并持久化
# （零配置可用；持久化保证重启后密钥不变，登录态不失效）
if [ -z "$JWT_SECRET" ]; then
  if [ -f /app/data/.jwt_secret ]; then
    JWT_SECRET="$(cat /app/data/.jwt_secret)"
  else
    JWT_SECRET="$(head -c 32 /dev/urandom | od -An -tx1 | tr -d ' \n' | head -c 64)"
    printf '%s' "$JWT_SECRET" > /app/data/.jwt_secret
    chmod 600 /app/data/.jwt_secret
    echo "[entrypoint] JWT secret generated and persisted to /app/data/.jwt_secret"
  fi
fi
export JWT_SECRET

APP_URL="${APP_URL:-http://localhost:3000}"
export APP_URL

# 后台启动 API（独立端口，由 Caddy 反代对外）
PORT="${API_PORT:-3001}" /app/api &

# 前台运行 Caddy：静态资源 + /api 反代，监听 3000
exec caddy run --config /etc/caddy/Caddyfile
EOF

ENV TZ=Asia/Shanghai \
    XDG_CONFIG_HOME=/config \
    XDG_DATA_HOME=/data \
    API_PORT=3001 \
    APP_URL="http://localhost:3000"

USER app
EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD wget -qO- http://127.0.0.1:3000/api/v1/healthz || exit 1

ENTRYPOINT ["/app/entrypoint.sh"]
