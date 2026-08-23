#!/usr/bin/env bash
# 数据库备份脚本
# 用法: ./backup-db.sh <env>
set -euo pipefail

ENV=${1:-dev}
BACKUP_DIR="${BACKUP_DIR:-./backups}"
mkdir -p "$BACKUP_DIR"
TS=$(date +%Y%m%d-%H%M%S)

case "$ENV" in
  mysql)
    echo "备份 MySQL..."
    docker compose exec -T mysql mysqldump \
      -u"${MYSQL_USER:-root}" \
      -p"${MYSQL_PASSWORD:-password}" \
      --single-transaction \
      --routines \
      --triggers \
      "${MYSQL_DATABASE:-ywty}" > "$BACKUP_DIR/mysql-$TS.sql"
    ;;
  postgres)
    echo "备份 PostgreSQL..."
    docker compose exec -T postgres pg_dump \
      -U "${POSTGRES_USER:-postgres}" \
      "${POSTGRES_DB:-ywty}" > "$BACKUP_DIR/postgres-$TS.sql"
    ;;
  sqlite)
    echo "备份 SQLite..."
    DB_PATH="${SQLITE_PATH:-./storage/dev.db}"
    if [ -f "$DB_PATH" ]; then
      cp "$DB_PATH" "$BACKUP_DIR/sqlite-$TS.db"
    else
      echo "未找到 SQLite 数据库: $DB_PATH"
      exit 1
    fi
    ;;
  *)
    echo "未知环境: $ENV (mysql/postgres/sqlite)"
    exit 1
    ;;
esac

# 压缩
gzip "$BACKUP_DIR/"*"-$TS.sql" 2>/dev/null || true
echo "✓ 备份完成: $BACKUP_DIR"
ls -lh "$BACKUP_DIR"/*-$TS* 2>/dev/null || true
