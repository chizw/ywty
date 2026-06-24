﻿﻿﻿﻿﻿# =========================================
# ywty 鍚堝苟闀滃儚锛圙o API + Nuxt SSR锛?# 鍗曞鍣?路 鍗曠鍙ｏ紙3000锛?# =========================================

# ---------- Stage 1: Go builder ----------
FROM golang:1.25-alpine AS go-builder
WORKDIR /src
COPY server/go.mod server/go.sum* ./
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go mod download
COPY server/ .
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api \
 && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate

# ---------- Stage 2: Nuxt builder ----------
FROM node:22-alpine AS web-builder
WORKDIR /app
ENV NODE_ENV=production NUXT_TELEMETRY_DISABLED=1
COPY web-nuxt/package.json web-nuxt/package-lock.json* .npmrc* ./
RUN --mount=type=cache,target=/root/.npm \
    npm ci --no-audit --no-fund --prefer-offline
COPY web-nuxt/ .
RUN --mount=type=cache,target=/root/.npm \
    npm run build

# ---------- Stage 3: runtime ----------
FROM node:22-alpine AS runtime

RUN apk add --no-cache ca-certificates tzdata wget \
 && addgroup -S app && adduser -S app -G app

WORKDIR /app

# Go 浜岃繘鍒?+ 閰嶇疆
COPY --from=go-builder /out/api /app/api
COPY --from=go-builder /out/migrate /app/migrate
COPY server/configs /app/configs

# Nuxt 鏋勫缓浜х墿
COPY --from=web-builder /app/.output /app/.output

# 鍏ュ彛鑴氭湰
COPY --chmod=0755 <<'EOF' /app/entrypoint.sh
#!/bin/sh
set -e

# 杩愯鏁版嵁搴撹縼绉?/app/migrate up

# 鍚庡彴鍚姩 API
/app/api &

# 鍓嶅彴鍚姩 Nuxt
exec node /app/.output/server/index.mjs
EOF

ENV TZ=Asia/Shanghai \
    NODE_ENV=production \
    PORT=3000 \
    HOST=0.0.0.0 \
    NUXT_TELEMETRY_DISABLED=1 \
    NUXT_API_BASE="" \
    NUXT_API_INTERNAL="http://127.0.0.1:8080"

USER app
EXPOSE 3000

HEALTHCHECK --interval=15s --timeout=3s --start-period=15s --retries=5 \
  CMD wget -qO- http://127.0.0.1:3000/healthz >/dev/null 2>&1 || exit 1

ENTRYPOINT ["/app/entrypoint.sh"]
