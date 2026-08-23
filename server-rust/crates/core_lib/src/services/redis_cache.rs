//! Redis 缓存服务
//!
//! 为应用提供可选的 Redis 缓存层。当配置中存在 `redis` 节点时连接到 Redis，
//! 否则以 `None` 形式存在，调用方自动降级到 DB/内存实现。

use redis::AsyncCommands;

/// Redis 缓存连接（可选）
///
/// 使用 `redis::aio::ConnectionManager` 支持自动重连与多/复用连接。
#[derive(Clone)]
pub struct RedisCache {
    conn: redis::aio::ConnectionManager,
}

impl RedisCache {
    /// 根据配置创建 Redis 连接
    ///
    /// 返回 `Ok(None)` 表示未配置 Redis（静默降级）；
    /// 返回 `Ok(Some(cache))` 表示连接成功。
    pub async fn new(
        addr: &str,
        password: Option<&str>,
        db: Option<i64>,
    ) -> crate::AppResult<Option<Self>> {
        if addr.is_empty() {
            return Ok(None);
        }

        let url = match (password, db) {
            (Some(p), Some(d)) if !p.is_empty() => {
                format!("redis://:{}@{}/{}", p, addr, d)
            }
            (Some(p), _) if !p.is_empty() => {
                format!("redis://:{}@{}", p, addr)
            }
            (_, Some(d)) => format!("redis://{}/{}", addr, d),
            _ => format!("redis://{}", addr),
        };

        let client = match redis::Client::open(url) {
            Ok(c) => c,
            Err(e) => {
                tracing::warn!("Redis 客户端创建失败: {}", e);
                return Ok(None);
            }
        };

        match redis::aio::ConnectionManager::new(client).await {
            Ok(conn) => {
                tracing::info!("✅ Redis 连接成功: {}", addr);
                Ok(Some(Self { conn }))
            }
            Err(e) => {
                tracing::warn!("⚠️ Redis 连接失败，将降级为无缓存模式: {}", e);
                Ok(None)
            }
        }
    }

    /// 创建未初始化的缓存（始终返回 None，用于默认/测试场景）
    pub fn none() -> Option<Self> {
        None
    }

    /// 是否已连接
    pub fn is_connected(&self) -> bool {
        true
    }

    /// 获取字符串值
    pub async fn get(&mut self, key: &str) -> crate::AppResult<Option<String>> {
        let val: Option<String> = self.conn.get(key).await?;
        Ok(val)
    }

    /// 设置字符串值（无过期）
    pub async fn set(&mut self, key: &str, value: &str) -> crate::AppResult<()> {
        self.conn.set::<_, _, ()>(key, value).await?;
        Ok(())
    }

    /// 设置字符串值并指定过期时间（秒）
    pub async fn set_ex(&mut self, key: &str, value: &str, seconds: u64) -> crate::AppResult<()> {
        self.conn.set_ex::<_, _, ()>(key, value, seconds).await?;
        Ok(())
    }

    /// 删除键
    pub async fn del(&mut self, key: &str) -> crate::AppResult<bool> {
        let n: i64 = self.conn.del(key).await?;
        Ok(n > 0)
    }

    /// 检查键是否存在
    pub async fn exists(&mut self, key: &str) -> crate::AppResult<bool> {
        let n: bool = self.conn.exists(key).await?;
        Ok(n)
    }

    /// 设置键的过期时间（秒）
    pub async fn expire(&mut self, key: &str, seconds: u64) -> crate::AppResult<bool> {
        let n: bool = self.conn.expire(key, seconds as i64).await?;
        Ok(n)
    }

    /// 原子递增
    pub async fn incr(&mut self, key: &str, delta: i64) -> crate::AppResult<i64> {
        let n: i64 = self.conn.incr(key, delta).await?;
        Ok(n)
    }

    /// 原子递增并设置过期时间（用于限流计数器等）
    pub async fn incr_ex(&mut self, key: &str, delta: i64, seconds: u64) -> crate::AppResult<i64> {
        let pipe_result: (i64,) = redis::pipe()
            .atomic()
            .incr(key, delta)
            .expire(key, seconds as i64)
            .query_async(&mut self.conn)
            .await?;
        Ok(pipe_result.0)
    }

    /// 将值添加到集合
    pub async fn sadd(&mut self, key: &str, member: &str) -> crate::AppResult<bool> {
        let n: i64 = self.conn.sadd(key, member).await?;
        Ok(n > 0)
    }

    /// 检查值是否在集合中
    pub async fn sismember(&mut self, key: &str, member: &str) -> crate::AppResult<bool> {
        let n: bool = self.conn.sismember(key, member).await?;
        Ok(n)
    }
}
