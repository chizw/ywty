//! 数据库连接层
//!
//! 提供 sqlx 连接池工厂。
//!
//! 注意：当前使用具体的 [`sqlx::SqlitePool`]，而非 [`sqlx::AnyPool`]。
//! 原因在于 sqlx 的 `Any` 驱动仅对基础原始类型（bool/int/float/str/blob）实现了
//! `Type<Any>` / `Decode<'_, Any>`，而 **不支持 `chrono::DateTime<Utc>` 与
//! `serde_json::Value`**，导致 `#[derive(FromRow)]` 无法通过。
//! 由于本项目迁移文件均为 SQLite 专用（AUTOINCREMENT / datetime('now') / PRAGMA），
//! 故直接使用 `SqlitePool`，完整支持 chrono + JSON。

use sqlx::sqlite::{SqlitePool, SqlitePoolOptions};
use std::time::Duration;

use crate::config::DatabaseConfig;
use crate::error::AppResult;

/// 创建 SQLite 数据库连接池。
///
/// 追加 `foreign_keys(ON)` + `busy_timeout(5000)` pragma。
pub async fn create_pool(cfg: &DatabaseConfig) -> AppResult<SqlitePool> {
    let path = cfg.path.clone().unwrap_or_else(|| "ywty.db".to_string());

    // 确保 SQLite 数据库父目录存在
    if let Some(parent) = std::path::Path::new(&path).parent() {
        if !parent.as_os_str().is_empty() {
            std::fs::create_dir_all(parent)?;
        }
    }

    let mut options = SqlitePoolOptions::new()
        .max_connections(cfg.max_open_conns.unwrap_or(100))
        .min_connections(cfg.max_idle_conns.unwrap_or(10))
        .acquire_timeout(Duration::from_secs(10))
        .idle_timeout(Duration::from_secs(600))
        .max_lifetime(Duration::from_secs(1800));

    options = options.after_connect(|conn, _meta| {
        Box::pin(async move {
            // 启用外键约束 + 忙等待超时
            sqlx::query("PRAGMA foreign_keys = ON;")
                .execute(&mut *conn)
                .await?;
            sqlx::query("PRAGMA busy_timeout = 5000;")
                .execute(&mut *conn)
                .await?;
            Ok(())
        })
    });

    let pool = options
        .connect(&format!("sqlite://{}?mode=rwc", path))
        .await
        .map_err(crate::error::AppError::Database)?;

    Ok(pool)
}

/// 检查数据库连接是否健康
pub async fn ping(pool: &SqlitePool) -> AppResult<()> {
    sqlx::query("SELECT 1")
        .fetch_one(pool)
        .await
        .map(|_| ())
        .map_err(crate::error::AppError::Database)
}
