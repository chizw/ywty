pub mod captcha;
pub mod pagination;
pub mod response;
pub mod time;

pub use captcha::{generate_captcha, generate_numeric_captcha, CaptchaResult};
pub use pagination::{PaginatedResponse, Pagination};
pub use response::ApiResponse;
pub use time::now;

/// 提取客户端 IP：优先 X-Forwarded-For（取第一个）/ X-Real-IP 头，回退连接地址
pub fn client_ip(
    headers: &axum::http::HeaderMap,
    connect_info: Option<std::net::SocketAddr>,
) -> String {
    if let Some(xff) = headers.get("x-forwarded-for").and_then(|v| v.to_str().ok()) {
        if let Some(first) = xff.split(',').next() {
            let ip = first.trim();
            if !ip.is_empty() {
                return ip.to_string();
            }
        }
    }
    if let Some(ip) = headers.get("x-real-ip").and_then(|v| v.to_str().ok()) {
        let ip = ip.trim();
        if !ip.is_empty() {
            return ip.to_string();
        }
    }
    connect_info
        .map(|addr| addr.ip().to_string())
        .unwrap_or_else(|| "0.0.0.0".to_string())
}

/// WeAvatar 头像地址：对邮箱（或任意稳定标识）做 MD5 后拼接
///
/// 用户设置了自定义 avatar 时应优先使用自定义值，本函数用于无头像时的兜底。
/// `d=mp` 表示无匹配头像时返回 WeAvatar 默认形象图。
pub fn weavatar_url(identifier: &str, size: u32) -> String {
    let digest = md5::compute(identifier.trim().to_lowercase());
    format!("https://weavatar.com/avatar/{:x}?s={size}&d=mp", digest)
}
