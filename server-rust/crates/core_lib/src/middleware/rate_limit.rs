//! 限流中间件
//!
//! 基于内存的令牌桶算法，按 IP 限流。
//! 生产环境建议替换为 Redis 实现分布式限流。

use std::collections::HashMap;
use std::net::SocketAddr;
use std::sync::Arc;
use std::time::{Duration, Instant};

use axum::{
    extract::{ConnectInfo, Request},
    middleware::Next,
    response::Response,
};
use tokio::sync::Mutex;

/// 令牌桶状态
#[derive(Debug)]
pub struct Bucket {
    tokens: f64,
    last_refill: Instant,
}

impl Bucket {
    fn new(capacity: f64) -> Self {
        Self {
            tokens: capacity,
            last_refill: Instant::now(),
        }
    }

    /// 尝试消费一个令牌，返回是否成功
    fn try_consume(&mut self, rate: f64, capacity: f64) -> bool {
        let now = Instant::now();
        let elapsed = now.duration_since(self.last_refill).as_secs_f64();

        // 补充令牌
        self.tokens = (self.tokens + elapsed * rate).min(capacity);
        self.last_refill = now;

        if self.tokens >= 1.0 {
            self.tokens -= 1.0;
            true
        } else {
            false
        }
    }
}

/// 限流配置
#[derive(Debug, Clone)]
pub struct RateLimitConfig {
    /// 每秒生成的令牌数（即允许的请求速率）
    pub rate_per_second: f64,
    /// 桶容量（最大突发请求数）
    pub burst: f64,
}

impl Default for RateLimitConfig {
    fn default() -> Self {
        Self {
            rate_per_second: 5.0, // 每秒 5 个请求
            burst: 10.0,          // 最多突发 10 个
        }
    }
}

/// 限流状态（共享 across handlers）
pub type RateLimitState = Arc<Mutex<HashMap<String, Bucket>>>;

/// 创建限流状态
pub fn create_rate_limit_state() -> RateLimitState {
    Arc::new(Mutex::new(HashMap::new()))
}

/// 限流中间件（可选 ConnectInfo，测试环境下无 socket 信息时跳过限流）
pub async fn rate_limit_middleware(
    axum::extract::State(state): axum::extract::State<RateLimitState>,
    request: Request,
    next: Next,
) -> Response {
    // 尝试从 extensions 获取 ConnectInfo（由服务器层注入）
    let ip = request
        .extensions()
        .get::<ConnectInfo<SocketAddr>>()
        .map(|ci| ci.0.ip().to_string())
        .unwrap_or_else(|| "unknown".to_string());

    if ip == "unknown" {
        // 测试环境或无 socket 信息，跳过限流
        return next.run(request).await;
    }

    let mut buckets = state.lock().await;
    let bucket = buckets
        .entry(ip.clone())
        .or_insert_with(|| Bucket::new(10.0));

    // 每秒 5 个请求，突发 10 个
    if bucket.try_consume(5.0, 10.0) {
        drop(buckets); // 释放锁
        next.run(request).await
    } else {
        tracing::warn!(ip = %ip, "请求被限流");
        use axum::response::IntoResponse;
        (
            axum::http::StatusCode::TOO_MANY_REQUESTS,
            axum::Json(serde_json::json!({
                "code": "RATE_LIMITED",
                "message": "请求过于频繁，请稍后再试"
            })),
        )
            .into_response()
    }
}

/// 清理过期的限流桶（定期调用）
pub fn start_rate_limit_cleanup(state: RateLimitState) {
    tokio::spawn(async move {
        let mut interval = tokio::time::interval(Duration::from_secs(60));
        loop {
            interval.tick().await;
            let mut buckets = state.lock().await;
            // 清理 60 秒未活跃的桶
            let now = Instant::now();
            buckets.retain(|_, bucket| {
                now.duration_since(bucket.last_refill) < Duration::from_secs(60)
            });
            tracing::debug!("限流桶清理完成，剩余 {} 个 IP", buckets.len());
        }
    });
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_bucket_allows_within_burst() {
        let mut bucket = Bucket::new(10.0);
        // 应该允许 10 个突发请求
        for _ in 0..10 {
            assert!(bucket.try_consume(5.0, 10.0), "应在突发范围内允许");
        }
        // 第 11 个应该被拒绝（令牌耗尽）
        assert!(!bucket.try_consume(5.0, 10.0), "超出突发应被拒绝");
    }

    #[test]
    fn test_bucket_refills_over_time() {
        let mut bucket = Bucket::new(10.0);
        // 耗尽令牌
        for _ in 0..10 {
            let _ = bucket.try_consume(5.0, 10.0);
        }
        assert!(!bucket.try_consume(5.0, 10.0));

        // 等待补充（使用高速率）
        std::thread::sleep(Duration::from_millis(300));
        // 此时应该补充了约 1.5 个令牌（5.0 * 0.3）
        assert!(bucket.try_consume(5.0, 10.0), "等待后应有新令牌");
    }
}
