# ---------- Stage 1: Rust builder ----------
FROM rust:1.85-slim-bookworm AS rust-builder

RUN apt-get update && apt-get install -y \
    pkg-config libssl-dev libsqlite3-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY Cargo.toml Cargo.lock ./
COPY crates/api/Cargo.toml ./crates/api/
COPY crates/core_lib/Cargo.toml ./crates/core_lib/

RUN mkdir -p crates/api/src crates/core_lib/src && \
    echo "fn main() {}" > crates/api/src/main.rs && \
    echo "fn main() {}" > crates/core_lib/src/main.rs

RUN cargo build --release 2>/dev/null || true

COPY crates/ ./crates/
COPY config.yaml ./config.yaml

# 同一镜像内置两种方言二进制，由入口脚本按 DB_DRIVER 运行时选择：
#   SQLite（默认零依赖） 与 MySQL/MariaDB
RUN touch crates/api/src/main.rs crates/core_lib/src/main.rs && \
    cargo build --release --bin api && \
    cp target/release/api /app/api-sqlite && \
    cargo build --release --bin api --features mysql && \
    cp target/release/api /app/api-mysql

# ---------- Stage 2: Astro builder ----------
FROM node:22-alpine AS web-builder
WORKDIR /app
ENV NODE_ENV=production ASTRO_TELEMETRY_DISABLED=1
COPY web-astro/package.json web-astro/package-lock.json* ./
RUN --mount=type=cache,target=/root/.npm \
    npm ci --no-audit --no-fund --prefer-offline
COPY web-astro/ .
RUN --mount=type=cache,target=/root/.npm \
    npm run build

# ---------- Stage 3: runtime ----------
FROM node:22-alpine AS runtime

RUN apk add --no-cache ca-certificates tzdata curl \
 && addgroup -S app && adduser -S app -G app

WORKDIR /app

# Rust 二进制（双方言）+ 配置
COPY --from=rust-builder /app/api-sqlite /app/api-sqlite
COPY --from=rust-builder /app/api-mysql /app/api-mysql
COPY --from=rust-builder /app/config.yaml /app/config.yaml

# Astro 构建产物
COPY --from=web-builder /app/dist /app/dist

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

# 按配置选择方言二进制：DB_DRIVER=mysql → MariaDB/MySQL；其余（默认）→ SQLite
if [ "$DB_DRIVER" = "mysql" ]; then
  echo "[entrypoint] database driver: mysql/mariadb"
  API_BIN=/app/api-mysql
else
  echo "[entrypoint] database driver: sqlite"
  API_BIN=/app/api-sqlite
fi

# 后台启动 API（独立端口，由 Astro 中间件反代对外）
PORT="${API_PORT:-3001}" "$API_BIN" &

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
