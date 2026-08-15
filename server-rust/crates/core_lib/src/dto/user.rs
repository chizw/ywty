//! 用户域 DTO

use chrono::{DateTime, Utc};
use serde::Serialize;
use sqlx::FromRow;

/// 用户公开资料响应
#[derive(Debug, Clone, Serialize, FromRow)]
pub struct UserProfile {
    pub id: i64,
    pub uuid: String,
    pub username: String,
    pub email: String,
    pub avatar: Option<String>,
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
#[derive(Debug, Clone, serde::Deserialize, validator::Validate)]
pub struct ChangeEmailRequest {
    #[validate(email(message = "邮箱格式不正确"))]
    pub new_email: String,
    pub verify_code: String,
}
