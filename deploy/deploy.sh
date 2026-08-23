#!/usr/bin/env bash
# 部署脚本：编译 Rust 后端 + Astro 前端，构建 Docker 镜像
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

# 2. 构建镜像（前端在镜像内多阶段构建，无需本地 npm）
echo "[2/4] 构建 Docker 镜像..."
cd "$PROJECT_ROOT"
docker build -t "ywty:$TAG" .
echo "  ✓ 镜像构建完成"

# 3. 推送（生产环境）
if [ "$ENV" = "prod" ] || [ "$ENV" = "staging" ]; then
  echo "[3/3] 推送镜像到仓库..."
  docker push "ywty:$TAG"
  echo "  ✓ 推送完成"
else
  echo "[3/3] 跳过镜像推送（开发环境）"
fi

echo ""
echo "部署完成！"
