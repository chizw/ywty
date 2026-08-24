# ---------- 运行时镜像（仅打包预编译产物，秒级构建） ----------
#
# 构建上下文需由流水线/脚本预先组装：
#   linux-amd64/api   — x64 后端二进制
#   dist/             — Astro 构建产物（含 dist/server/entry.mjs）
#   config.yaml       — 默认配置（来源 server-rust/config.yaml）
#
# 本地组装请使用 deploy/deploy.sh；CI 由 .github/workflows 自动完成。

FROM node:22-alpine AS runtime

RUN apk add --no-cache ca-certificates tzdata curl \
 && addgroup -S app && adduser -S app -G app

WORKDIR /app

COPY config.yaml /app/config.yaml
COPY dist/ /app/dist/
COPY linux-amd64/api /app/api

# 数据目录
RUN mkdir -p /app/data /app/uploads && chown -R app:app /app/data /app/uploads

# 入口脚本
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
    mkdir -p /app/data
    printf '%s' "$JWT_SECRET" > /app/data/.jwt_secret
    chmod 600 /app/data/.jwt_secret
    echo "[entrypoint] JWT secret generated and persisted to /app/data/.jwt_secret"
  fi
fi
export JWT_SECRET

APP_URL="${APP_URL:-http://localhost:3000}"
export APP_URL

# 后台启动 API（独立端口，由 Astro 中间件反代对外）
PORT="${API_PORT:-3001}" /app/api &

# 前台启动 Astro SSR（PORT 即对外端口，默认 3000）
exec node /app/dist/server/entry.mjs
EOF

ENV TZ=Asia/Shanghai \
    NODE_ENV=production \
    HOST=0.0.0.0 \
    PORT=3000 \
    API_PORT=3001 \
    APP_URL="http://localhost:3000" \
    ASTRO_TELEMETRY_DISABLED=1 \
    API_INTERNAL="http://127.0.0.1:3001"

USER app
EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=3s --start-period=15s --retries=3 \
  CMD curl -fsS http://127.0.0.1:3000/api/v1/healthz || exit 1

ENTRYPOINT ["/app/entrypoint.sh"]
