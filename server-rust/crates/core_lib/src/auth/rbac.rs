use axum::{
    extract::{Request, State},
    middleware::Next,
    response::Response,
};
use serde::Serialize;

use crate::auth::Claims;
use crate::error::AppError;

/// 用户角色
#[derive(Debug, Clone, PartialEq, Eq, Serialize)]
pub enum Role {
    User,
    Admin,
    SuperAdmin,
}

impl Role {
    pub fn parse(s: &str) -> Self {
        match s {
            "admin" => Role::Admin,
            "super_admin" => Role::SuperAdmin,
            _ => Role::User,
        }
    }

    pub fn as_str(&self) -> &'static str {
        match self {
            Role::User => "user",
            Role::Admin => "admin",
            Role::SuperAdmin => "super_admin",
        }
    }

    /// 检查角色是否有权限访问
    pub fn can_access(&self, required: &Role) -> bool {
        matches!(
            (self, required),
            (Role::SuperAdmin, _)
                | (Role::Admin, Role::Admin)
                | (Role::Admin, Role::User)
                | (Role::User, Role::User)
        )
    }
}

impl std::fmt::Display for Role {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.as_str())
    }
}

/// RBAC 守卫 - 用于中间件权限检查
#[derive(Debug, Clone)]
pub struct RbacGuard {
    pub required_role: Role,
}

impl RbacGuard {
    pub fn new(required_role: Role) -> Self {
        Self { required_role }
    }

    pub fn admin() -> Self {
        Self::new(Role::Admin)
    }

    pub fn user() -> Self {
        Self::new(Role::User)
    }

    pub fn super_admin() -> Self {
        Self::new(Role::SuperAdmin)
    }
}

/// RBAC 中间件
pub async fn rbac_middleware(
    State(guard): State<RbacGuard>,
    claims: Claims,
    request: Request,
    next: Next,
) -> Result<Response, AppError> {
    let user_role = Role::parse(&claims.role);

    if !user_role.can_access(&guard.required_role) {
        return Err(AppError::Forbidden);
    }

    Ok(next.run(request).await)
}

/// 检查是否为管理员
pub fn require_admin(claims: &Claims) -> Result<(), AppError> {
    let role = Role::parse(&claims.role);
    if !role.can_access(&Role::Admin) {
        return Err(AppError::Forbidden);
    }
    Ok(())
}

/// 检查是否为超级管理员
pub fn require_super_admin(claims: &Claims) -> Result<(), AppError> {
    let role = Role::parse(&claims.role);
    if !role.can_access(&Role::SuperAdmin) {
        return Err(AppError::Forbidden);
    }
    Ok(())
}

/// 检查是否为资源所有者或管理员
pub fn require_owner_or_admin(claims: &Claims, owner_id: i64) -> Result<(), AppError> {
    if claims.sub == owner_id {
        return Ok(());
    }
    require_admin(claims)
}
