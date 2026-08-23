//! HTTP 处理器层

pub mod admin;
pub mod album;
pub mod auth;
pub mod capacity;
pub mod coupon;
pub mod drivers;
pub mod feedback;
pub mod group;
pub mod like;
pub mod notice;
pub mod oauth;
pub mod order;
pub mod page;
pub mod photo;
pub mod plan;
pub mod settings;
pub mod share;
pub mod site;
pub mod storage;
pub mod storage_admin;
pub mod tag;
pub mod ticket;
pub mod token;
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

/// 管理员提取器（从 admin_guard 注入的 AdminContext 中提取）
/// 仅用于 /admin/* 路由，已通过角色校验
pub struct AdminUser {
    pub user_id: i64,
    pub role: String,
    pub is_super_admin: bool,
}

#[async_trait::async_trait]
impl<S> FromRequestParts<S> for AdminUser {
    type Rejection = AppError;

    async fn from_request_parts(parts: &mut Parts, _state: &S) -> Result<Self, Self::Rejection> {
        let ctx = parts
            .extensions
            .get::<crate::middleware::AdminContext>()
            .cloned()
            .ok_or(AppError::Forbidden)?;

        Ok(AdminUser {
            user_id: ctx.user_id,
            role: ctx.role,
            is_super_admin: ctx.is_super_admin,
        })
    }
}
