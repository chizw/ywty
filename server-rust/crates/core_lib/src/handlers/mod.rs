//! HTTP 处理器层（对应 Go 的 `internal/handler/`）

pub mod album;
pub mod auth;
pub mod photo;
pub mod storage;
pub mod user;

use axum::extract::FromRequestParts;
use axum::http::request::Parts;

use crate::auth::Claims;
use crate::error::AppError;

/// 当前用户提取器（从 auth_middleware 注入的 Claims 中提取）
/// 仅用于需要认证的路由
pub struct CurrentUser {
    pub user_id: i64,
    pub username: String,
    pub role: String,
}

#[async_trait::async_trait]
impl<S> FromRequestParts<S> for CurrentUser {
    type Rejection = AppError;

    async fn from_request_parts(parts: &mut Parts, _state: &S) -> Result<Self, Self::Rejection> {
        let claims = parts
            .extensions
            .get::<Claims>()
            .ok_or_else(|| AppError::Auth("未登录".to_string()))?;

        Ok(CurrentUser {
            user_id: claims.sub,
            username: claims.username.clone(),
            role: claims.role.clone(),
        })
    }
}

/// 验证请求并返回验证错误
pub fn validate_req<T: validator::Validate>(req: &T) -> Result<(), AppError> {
    req.validate()
        .map_err(|e| AppError::Validation(e.to_string()))
}
