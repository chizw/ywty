# syntax=docker/dockerfile:1
# ============================================================================
# ywty 合并镜像：前端主题(node) + Go 后端(golang) + 运行时(alpine + vips)
# CI 冒烟约定：/healthz 200；/api/v2/__nope__ 返回 {"status":"error"}；/ 返回 SPA
# ============================================================================

# ---------- Stage 1: 前端主题构建 ----------
FROM node:22-alpine AS web
WORKDIR /src/web
COPY web/package.json web/package-lock.json web/.npmrc ./
RUN npm ci --no-audit --no-fund
COPY web/ ./
# 只做 vite 构建（产物 dist/：index.html + config.json + themes/default/*），
# 不跑 post-build 的 PHP 目录拷贝步骤
RUN npm run build-only

# ---------- Stage 2: Go 后端构建 ----------
FROM golang:1.26-alpine AS build
WORKDIR /src/server-go
COPY server-go/ ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /out/ywty ./cmd/ywty

# ---------- Stage 3: 运行时 ----------
FROM alpine:3.21
# vips CLI：heic/avif/webp 等高级格式处理的可选外挂（检测到即启用）
RUN apk add --no-cache vips ca-certificates tzdata
WORKDIR /app
COPY --from=build /out/ywty /app/ywty
COPY --from=web /src/web/dist/ /app/public/
COPY LICENSE.md CHANGELOG.md /app/

ENV HOST=0.0.0.0 \
    PORT=3000 \
    STATIC_DIR=/app/public \
    DATA_DIR=/app/data \
    UPLOADS_DIR=/app/uploads
VOLUME ["/app/data", "/app/uploads"]
EXPOSE 3000

ENTRYPOINT ["/app/ywty"]
