# =========================================
# ywty Web (Nuxt 3) - Multi-stage Dockerfile
# Stage 1: deps    - 安装依赖（含 devDependencies）
# Stage 2: builder - 构建（.output）
# Stage 3: runner  - 仅 .output（更小更快）
# =========================================

# ---------- Stage 1: deps ----------
FROM node:22-alpine AS deps
WORKDIR /app
COPY package.json package-lock.json* .npmrc* ./
RUN --mount=type=cache,target=/root/.npm \
    npm ci --no-audit --no-fund --prefer-offline

# ---------- Stage 2: builder ----------
FROM node:22-alpine AS builder
WORKDIR /app

ENV NODE_ENV=production \
    NUXT_TELEMETRY_DISABLED=1

COPY --from=deps /app/node_modules ./node_modules
COPY . .

RUN --mount=type=cache,target=/root/.npm \
    --mount=type=cache,target=/app/.nuxt \
    --mount=type=cache,target=/app/node_modules/.cache \
    npm run build

# ---------- Stage 3: runner ----------
FROM node:22-alpine AS runner
WORKDIR /app

ENV NODE_ENV=production \
    PORT=3000 \
    HOST=0.0.0.0 \
    NUXT_TELEMETRY_DISABLED=1

# 创建非 root 用户
RUN addgroup -S app && adduser -S app -G app

# 复制构建产物
COPY --from=builder --chown=app:app /app/.output ./.output
COPY --from=builder --chown=app:app /app/package.json ./package.json
COPY --from=builder --chown=app:app /app/public ./public

USER app
EXPOSE 3000

HEALTHCHECK --interval=15s --timeout=3s --start-period=15s --retries=5 \
  CMD wget -qO- http://127.0.0.1:3000/ >/dev/null 2>&1 || exit 1

CMD ["node", ".output/server/index.mjs"]
