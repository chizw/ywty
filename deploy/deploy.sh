#!/usr/bin/env bash
# 部署脚本：本地编译后端 + 前端，组装上下文并构建运行时镜像
# 用法: ./deploy.sh <environment> [tag]
#   environment: dev | staging | prod
#   tag:        镜像 tag (默认 latest)
set -euo pipefail

ENV=${1:-dev}
TAG=${2:-latest}
PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

echo "=========================================="
echo "  部署环境: $ENV"
echo "  镜像 tag:  $TAG"
echo "  项目根:    $PROJECT_ROOT"
echo "=========================================="

# 1. 编译后端（Rust）
echo "[1/4] 编译 Rust 后端..."
cd "$PROJECT_ROOT/server-rust"
cargo build --release --bin api
echo "  ✓ Rust 编译完成"

# 2. 编译前端
echo "[2/4] 构建 Astro 前端..."
cd "$PROJECT_ROOT/web-astro"
if [ ! -d node_modules ]; then
  npm ci
fi
npm run build
echo "  ✓ 前端构建完成"

# 3. 组装镜像构建上下文
echo "[3/4] 组装构建上下文..."
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)  PLATFORM_ARCH="amd64" ;;
  aarch64|arm64) PLATFORM_ARCH="arm64" ;;
  *) echo "不支持的架构: $ARCH"; exit 1 ;;
esac
CTX="$PROJECT_ROOT/.docker-ctx"
rm -rf "$CTX"
mkdir -p "$CTX/linux-$PLATFORM_ARCH"

cp "$PROJECT_ROOT/Caddyfile" "$CTX/Caddyfile"
cp "$PROJECT_ROOT/server-rust/target/release/api" "$CTX/linux-$PLATFORM_ARCH/api"
cp -r "$PROJECT_ROOT/web-astro/dist" "$CTX/dist"
cp "$PROJECT_ROOT/server-rust/config.yaml" "$CTX/config.yaml"
cd "$PROJECT_ROOT"
docker build -t "ywty:$TAG" -f Dockerfile "$CTX"
echo "  ✓ 镜像构建完成"

# 4. 推送（生产环境）
if [ "$ENV" = "prod" ] || [ "$ENV" = "staging" ]; then
  echo "[4/4] 推送镜像到仓库..."
  docker push "ywty:$TAG"
  echo "  ✓ 推送完成"
else
  echo "[4/4] 跳过镜像推送（开发环境）"
fi

echo ""
echo "部署完成！"
