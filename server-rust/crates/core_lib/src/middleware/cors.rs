use std::time::Duration;
use tower_http::cors::{Any, CorsLayer};

/// 创建 CORS 中间件层
pub fn cors_layer() -> CorsLayer {
    CorsLayer::new()
        .allow_origin(Any)
        .allow_methods(Any)
        .allow_headers(Any)
        .expose_headers(Any)
        .max_age(Duration::from_secs(3600))
}
