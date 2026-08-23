#!/usr/bin/env bash
# 启动本地开发环境
# 依赖: docker / docker compose
set -euo pipefail
cd "$(dirname "$0")/.."

# 选择 profile
PROFILE=${1:-sqlite}

case "$PROFILE" in
  mysql)
    echo "启动 MySQL profile..."
    docker compose --profile mysql up -d
    ;;
  postgres)
    echo "启动 PostgreSQL profile..."
    docker compose --profile postgres up -d
    ;;
  sqlite)
    echo "启动 SQLite profile (无数据库容器)..."
    docker compose up -d redis ywty
    ;;
  *)
    echo "未知 profile: $PROFILE (可选: mysql/postgres/sqlite)"
    exit 1
    ;;
esac

echo "等待服务就绪..."
sleep 5
docker compose ps
echo "完成！访问 http://localhost:3000"
