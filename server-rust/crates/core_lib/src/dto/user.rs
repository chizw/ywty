//! 用户域 DTO

use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use sqlx::FromRow;
use utoipa::ToSchema;

/// 用户公开资料响应
#[derive(Debug, Clone, Serialize, FromRow, ToSchema)]
pub struct UserProfile {
    pub id: i64,
    pub uuid: String,
    pub username: String,
    pub email: String,
    pub avatar: Option<String>,
    /// WeAvatar 兜底头像（avatar 为空时由后端按邮箱生成）
    #[sqlx(default)]
    pub avatar_url: Option<String>,
    pub bio: Option<String>,
    pub role: String,
    pub capacity_used: i64,
    pub capacity_max: i64,
    pub plan_id: Option<i64>,
    pub plan_expires_at: Option<DateTime<Utc>>,
    pub email_verified_at: Option<DateTime<Utc>>,
    pub last_login_at: Option<DateTime<Utc>>,
    pub created_at: DateTime<Utc>,
}

/// 修改邮箱请求（复用验证）
#[derive(Debug, Clone, Deserialize, validator::Validate, ToSchema)]
pub struct ChangeEmailRequest {
    #[validate(email(message = "邮箱格式不正确"))]
    pub new_email: String,
    pub verify_code: String,
}

/// 后台用户列表项（不含密码等敏感字段）
#[derive(Debug, Clone, Serialize, FromRow, ToSchema)]
pub struct AdminUserResponse {
    pub id: i64,
    pub uuid: String,
    pub username: String,
    pub email: String,
    pub avatar: Option<String>,
    pub role: String,
    pub is_super_admin: bool,
    pub status: i32,
    pub capacity_used: i64,
    pub capacity_max: i64,
    pub created_at: DateTime<Utc>,
}
