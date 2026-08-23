#!/usr/bin/env bash
# 停止本地开发环境
set -euo pipefail
cd "$(dirname "$0")/.."
docker compose down
echo "✓ 已停止所有容器"
