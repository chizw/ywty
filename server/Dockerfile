# =========================================
# ywty API - Multi-stage Dockerfile
# Stage 1: builder  - 编译 Go 静态二进制
# Stage 2: runtime  - scratch + ca-certificates
# Stage 3: migrate  - 额外构建迁移工具（可选）
# =========================================

# ---------- Stage 1: builder ----------
FROM golang:1.25-alpine AS builder

# 国内构建可选
# RUN go env -w GOPROXY=https://goproxy.cn,https://proxy.golang.org,direct

WORKDIR /src

# 先复制依赖描述以利用缓存
COPY go.mod go.sum* ./
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go mod download

# 复制源码
COPY . .

# 静态构建（无 CGO、剥离调试信息、路径短）
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go build -trimpath -ldflags="-s -w -buildid=" -o /out/api ./cmd/api \
 && go build -trimpath -ldflags="-s -w -buildid=" -o /out/migrate ./cmd/migrate \
 && go build -trimpath -ldflags="-s -w -buildid=" -o /out/worker ./cmd/worker

# ---------- Stage 2: runtime ----------
FROM alpine:3.20 AS runtime

RUN apk add --no-cache ca-certificates tzdata curl \
 && addgroup -S app && adduser -S app -G app

WORKDIR /app
COPY --from=builder /out/api /app/api
COPY --from=builder /out/migrate /app/migrate
COPY --from=builder /out/worker /app/worker
COPY configs /app/configs

ENV TZ=Asia/Shanghai \
    PATH=/app:$PATH

USER app
EXPOSE 8080

HEALTHCHECK --interval=15s --timeout=3s --start-period=10s --retries=5 \
  CMD curl -fsS http://127.0.0.1:8080/healthz || exit 1

# 默认启动 API
ENTRYPOINT ["/app/api"]
CMD []

# ---------- Stage 3: distroless 备选 ----------
# 如果追求更小镜像（约 20MB），可使用 distroless
# FROM gcr.io/distroless/static-debian12:nonroot AS distroless
# COPY --from=builder /out/api /app/api
# COPY configs /app/configs
# USER nonroot:nonroot
# ENTRYPOINT ["/app/api"]
