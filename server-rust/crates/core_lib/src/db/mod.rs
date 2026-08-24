//! 数据库连接层
//!
//! 提供跨方言（SQLite / MySQL·MariaDB）的连接池工厂、方言助手与迁移器。
//!
//! 方言通过**编译期 feature** 选择，而非运行时切换：
//!
//! - 默认构建：SQLite（`sqlx::SqlitePool`），零依赖；
//! - `--features mysql`：MySQL / MariaDB ≥10.6（`sqlx::MySqlPool`）。
//!
//! 原因：sqlx 的 `AnyPool` 对 `chrono::DateTime<Utc>` 等类型不支持 `Decode`，
//! 无法配合本项目的 `#[derive(FromRow)]`；编译期选择则让每种构建都持有
//! 具象连接池与纯方言 SQL，由编译器保证正确性。
//!
//! 时间写入约定（两方言统一）：使用 [`now_str()`] 生成 `YYYY-MM-DD HH:MM:SS`
//! （UTC）字符串。SQLite 以 TEXT 存储（字典序可比），MySQL 以 DATETIME 存储，
//! 排序与比较语义一致。

use std::time::Duration;

use crate::config::DatabaseConfig;
use crate::error::AppResult;

#[cfg(feature = "mysql")]
pub type DbPool = sqlx::MySqlPool;
#[cfg(not(feature = "mysql"))]
pub type DbPool = sqlx::SqlitePool;

/// 当前 UTC 时间的 SQL 标准格式字符串：`YYYY-MM-DD HH:MM:SS`
///
/// SQLite（TEXT 列，字典序可比）与 MySQL（DATETIME 列）语义一致。
pub fn now_str() -> String {
    chrono::Utc::now().format("%Y-%m-%d %H:%M:%S").to_string()
}

/// 提取 INSERT 结果的自增主键。
///
/// sqlx 两种 QueryResult 的 API 不一致（Sqlite 为方法，MySql 为字段），在此统一。
pub fn last_id(result: &impl DbLastId) -> i64 {
    result.db_last_id()
}

/// 方言相关的自增 ID 提取 trait（内部使用）
pub trait DbLastId {
    fn db_last_id(&self) -> i64;
}

#[cfg(not(feature = "mysql"))]
impl DbLastId for sqlx::sqlite::SqliteQueryResult {
    fn db_last_id(&self) -> i64 {
        self.last_insert_rowid()
    }
}

#[cfg(feature = "mysql")]
impl DbLastId for sqlx::mysql::MySqlQueryResult {
    fn db_last_id(&self) -> i64 {
        self.last_insert_id() as i64
    }
}

/// 嵌入式迁移器（按 feature 选择方言目录）
#[cfg(not(feature = "mysql"))]
pub static MIGRATOR: sqlx::migrate::Migrator = sqlx::migrate!("./src/db/migrations");
#[cfg(feature = "mysql")]
pub static MIGRATOR: sqlx::migrate::Migrator = sqlx::migrate!("./src/db/migrations-mysql");

/// 创建数据库连接池（按编译期 feature 选择驱动）
pub async fn create_pool(cfg: &DatabaseConfig) -> AppResult<DbPool> {
    #[cfg(not(feature = "mysql"))]
    {
        create_sqlite_pool(cfg).await
    }
    #[cfg(feature = "mysql")]
    {
        create_mysql_pool(cfg).await
    }
}

#[cfg(not(feature = "mysql"))]
async fn create_sqlite_pool(cfg: &DatabaseConfig) -> AppResult<DbPool> {
    let path = cfg.path.clone().unwrap_or_else(|| "ywty.db".to_string());

    // 确保 SQLite 数据库父目录存在
    if let Some(parent) = std::path::Path::new(&path).parent() {
        if !parent.as_os_str().is_empty() {
            std::fs::create_dir_all(parent)?;
        }
    }

    let options = sqlx::sqlite::SqlitePoolOptions::new()
        .max_connections(cfg.max_open_conns.unwrap_or(100))
        .min_connections(cfg.max_idle_conns.unwrap_or(10))
        .acquire_timeout(Duration::from_secs(10))
        .idle_timeout(Duration::from_secs(600))
        .max_lifetime(Duration::from_secs(1800))
        .after_connect(|conn, _meta| {
            Box::pin(async move {
                // 启用外键约束 + 忙等待超时（SQLite 专属 pragma）
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

#[cfg(feature = "mysql")]
async fn create_mysql_pool(cfg: &DatabaseConfig) -> AppResult<DbPool> {
    use sqlx::mysql::MySqlPoolOptions;

    let host = cfg.host.clone().unwrap_or_else(|| "127.0.0.1".to_string());
    let port = cfg.port.unwrap_or(3306);
    let user = cfg.username.clone().unwrap_or_else(|| "ywty".to_string());
    let password = cfg.password.clone().unwrap_or_default();
    let database = cfg.database.clone().unwrap_or_else(|| "ywty".to_string());

    let url = format!(
        "mysql://{}:{}@{}:{}/{}",
        user, password, host, port, database
    );

    // 连接获取超时：默认 10s；跨公网/服务端开启反向解析时可能需要更长，
    // 通过 DB_ACQUIRE_TIMEOUT_SECS 覆盖。
    let acquire_secs: u64 = std::env::var("DB_ACQUIRE_TIMEOUT_SECS")
        .ok()
        .and_then(|v| v.parse().ok())
        .unwrap_or(10);

    // TLS：默认禁用。公网直连某些链路会在 TLS 握手大包上丢包导致连接悬挂
    // （表现为 pool timed out）；如需加密传输可加 DB_SSL=true 覆盖。
    let ssl_disabled = std::env::var("DB_SSL")
        .map(|v| !v.eq_ignore_ascii_case("true"))
        .unwrap_or(true);
    let url = if ssl_disabled {
        format!("{url}?ssl_mode=disabled")
    } else {
        url
    };

    let options = MySqlPoolOptions::new()
        .max_connections(cfg.max_open_conns.unwrap_or(100))
        .min_connections(cfg.max_idle_conns.unwrap_or(10))
        .acquire_timeout(Duration::from_secs(acquire_secs))
        .idle_timeout(Duration::from_secs(600))
        .max_lifetime(Duration::from_secs(1800))
        .after_connect(|conn, _meta| {
            Box::pin(async move {
                // 统一会话时区为 UTC，保证 DATETIME 写读语义与 SQLite TEXT 一致
                sqlx::query("SET time_zone = '+00:00'")
                    .execute(&mut *conn)
                    .await?;
                Ok(())
            })
        });

    let pool = options
        .connect(&url)
        .await
        .map_err(crate::error::AppError::Database)?;

    Ok(pool)
}

/// 方言适配：忽略冲突的 INSERT 前缀
///
/// SQLite 语法为 `INSERT OR IGNORE`，MySQL/MariaDB 为 `INSERT IGNORE`。
/// 入参统一使用 SQLite 写法，由本函数按 feature 转换。
pub fn sql_insert_ignore(sql: &str) -> String {
    #[cfg(feature = "mysql")]
    {
        sql.replace("INSERT OR IGNORE", "INSERT IGNORE")
    }
    #[cfg(not(feature = "mysql"))]
    {
        sql.to_string()
    }
}

/// 检查数据库连接是否健康
pub async fn ping(pool: &DbPool) -> AppResult<()> {
    sqlx::query("SELECT 1")
        .fetch_one(pool)
        .await
        .map(|_| ())
        .map_err(crate::error::AppError::Database)
}
