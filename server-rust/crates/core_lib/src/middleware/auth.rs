use axum::{
    extract::{Request, State},
    middleware::Next,
    response::Response,
};
use axum_extra::extract::CookieJar;

use crate::app::AppState;
use crate::auth::jwt::Claims;
use crate::error::AppError;

/// 认证中间件 - 从 Authorization Header 或 Cookie 中提取并验证 JWT
/// 额外校验用户仍存在、未被删除、状态正常（使删除/禁用立即生效，防止失效 token 继续使用）
pub async fn auth_middleware(
    State(state): State<AppState>,
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
    let claims = state.jwt.verify_token(token)?;

    // DB 校验：用户必须存在、未被删除、状态正常
    let row: Option<(i32, Option<String>)> =
        sqlx::query_as("SELECT status, deleted_at FROM users WHERE id = ?")
            .bind(claims.sub)
            .fetch_optional(&state.db)
            .await?;
    let (status, deleted_at) =
        row.ok_or_else(|| AppError::Auth("账号不存在或已被删除".to_string()))?;
    if deleted_at.is_some() || status != 1 {
        return Err(AppError::Auth("账号已失效".to_string()));
    }

    // 将 claims 添加到 request extensions
    request.extensions_mut().insert(claims);

    Ok(next.run(request).await)
}

/// 管理员上下文（admin_guard 注入，供后台 handler 细粒度鉴权）
#[derive(Debug, Clone)]
pub struct AdminContext {
    pub user_id: i64,
    pub role: String,
    pub is_super_admin: bool,
}

/// 管理员守卫 - 仅对 /admin/* 路径生效：
/// - 从 DB 查最新角色与超管标记（防止陈旧 JWT 越权）
/// - 非管理员 → 403
/// - 通过后注入 AdminContext 供 handler 使用
pub async fn admin_guard(
    State(state): State<AppState>,
    mut request: Request,
    next: Next,
) -> Result<Response, AppError> {
    let path = request.uri().path();
    if !path.starts_with("/admin") {
        return Ok(next.run(request).await);
    }

    let claims = request
        .extensions()
        .get::<Claims>()
        .cloned()
        .ok_or_else(|| AppError::Auth("未登录".to_string()))?;

    let row: Option<(String, bool)> = sqlx::query_as(
        "SELECT role, is_super_admin FROM users WHERE id = ? AND deleted_at IS NULL AND status = 1",
    )
    .bind(claims.sub)
    .fetch_optional(&state.db)
    .await?;
    let (role, is_super_admin) = row.ok_or_else(|| AppError::Auth("账号已失效".to_string()))?;

    if !is_super_admin && role != "admin" && role != "super_admin" {
        return Err(AppError::Forbidden);
    }

    request.extensions_mut().insert(AdminContext {
        user_id: claims.sub,
        role,
        is_super_admin,
    });

    Ok(next.run(request).await)
}

/// 可选认证中间件 - 不强制要求登录，但如果有 token 会解析
pub async fn optional_auth_middleware(
    State(jwt_auth): State<crate::auth::jwt::JwtAuth>,
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
