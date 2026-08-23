# ============================================================================
# ywty-server-rust Dockerfile
# 多阶段构建：编译阶段用 Rust 镜像，运行阶段用 Debian slim
# ============================================================================

# ----------------------------------------------------------------------------
# 阶段 1：编译
# ----------------------------------------------------------------------------
FROM rust:1.85-slim-bookworm AS builder

# 安装系统依赖（SQLite、SSL、pkg-config）
RUN apt-get update && apt-get install -y \
    pkg-config \
    libssl-dev \
    libsqlite3-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# 先复制 Cargo 清单，利用 Docker 缓存层
COPY Cargo.toml Cargo.lock ./
COPY crates/api/Cargo.toml ./crates/api/
COPY crates/core_lib/Cargo.toml ./crates/core_lib/

# 创建空的 main.rs 以通过依赖编译
RUN mkdir -p crates/api/src crates/core_lib/src && \
    echo "fn main() {}" > crates/api/src/main.rs && \
    echo "fn main() {}" > crates/core_lib/src/main.rs

# 编译依赖（缓存层）
RUN cargo build --release 2>/dev/null || true

# 复制实际源代码
COPY crates/ ./crates/
COPY config.yaml.example ./config.yaml.example

# 强制重新编译（因为源码变了）
RUN touch crates/api/src/main.rs crates/core_lib/src/main.rs && \
    cargo build --release --bin api

# ----------------------------------------------------------------------------
# 阶段 2：运行
# ----------------------------------------------------------------------------
FROM debian:bookworm-slim AS runtime

# 安装运行时依赖
RUN apt-get update && apt-get install -y \
    ca-certificates \
    libsqlite3-0 \
    libssl3 \
    && rm -rf /var/lib/apt/lists/*

# 创建非 root 用户
RUN groupadd -r ywty && useradd -r -g ywty -d /app -s /sbin/nologin ywty

WORKDIR /app

# 从编译阶段复制二进制文件
COPY --from=builder /app/target/release/api /app/api

# 创建数据目录
RUN mkdir -p /app/data /app/uploads && chown -R ywty:yuwty /app

# 切换到非 root 用户
USER ywty

# 环境变量（短名覆盖 config.yaml，见 crates/core_lib/src/config.rs）
ENV HOST=0.0.0.0
ENV PORT=3000
ENV DB_PATH=/app/data/ywty.db
ENV STORAGE_ROOT=/app/uploads

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:3000/api/v1/healthz || exit 1

EXPOSE 3000

ENTRYPOINT ["/app/api"]
