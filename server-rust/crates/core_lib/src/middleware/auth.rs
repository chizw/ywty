use axum::{
    extract::{Request, State},
    middleware::Next,
    response::Response,
};
use axum_extra::extract::CookieJar;

use crate::auth::JwtAuth;
use crate::error::AppError;

/// 认证中间件 - 从 Authorization Header 或 Cookie 中提取并验证 JWT
pub async fn auth_middleware(
    State(jwt_auth): State<JwtAuth>,
    jar: CookieJar,
    mut request: Request,
    next: Next,
) -> Result<Response, AppError> {
    // 尝试从 Authorization Header 获取 token
    let token = request
        .headers()
        .get("authorization")
        .and_then(|v| v.to_str().ok())
        .and_then(|v| v.strip_prefix("Bearer "));

    // 如果 Header 中没有，尝试从 Cookie 获取
    let token = token
        .or_else(|| jar.get("access_token").map(|c| c.value()))
        .ok_or_else(|| AppError::Auth("缺少认证令牌".to_string()))?;

    // 验证 token
    let claims = jwt_auth.verify_token(token)?;

    // 将 claims 添加到 request extensions
    request.extensions_mut().insert(claims);

    Ok(next.run(request).await)
}

/// 可选认证中间件 - 不强制要求登录，但如果有 token 会解析
pub async fn optional_auth_middleware(
    State(jwt_auth): State<JwtAuth>,
    jar: CookieJar,
    mut request: Request,
    next: Next,
) -> Response {
    let token = request
        .headers()
        .get("authorization")
        .and_then(|v| v.to_str().ok())
        .and_then(|v| v.strip_prefix("Bearer "));

    let token = token.or_else(|| jar.get("access_token").map(|c| c.value()));

    if let Some(token) = token {
        if let Ok(claims) = jwt_auth.verify_token(token) {
            request.extensions_mut().insert(claims);
        }
    }

    next.run(request).await
}
