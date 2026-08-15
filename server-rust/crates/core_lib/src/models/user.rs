use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use sqlx::FromRow;
use uuid::Uuid;
use validator::Validate;

/// 用户实体
#[derive(Debug, Clone, FromRow, Serialize)]
pub struct User {
    pub id: i64,
    pub uuid: String,
    pub username: String,
    pub email: String,
    pub password: String,
    pub avatar: Option<String>,
    pub role: String,
    pub status: i32,
    pub capacity_used: i64,
    pub capacity_max: i64,
    pub plan_id: Option<i64>,
    pub plan_expires_at: Option<DateTime<Utc>>,
    pub email_verified_at: Option<DateTime<Utc>>,
    pub last_login_at: Option<DateTime<Utc>>,
    pub last_login_ip: Option<String>,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
    pub deleted_at: Option<DateTime<Utc>>,
}

/// 用户注册请求
#[derive(Debug, Clone, Deserialize, Validate)]
pub struct RegisterRequest {
    #[validate(length(min = 3, max = 32, message = "用户名长度必须在 3-32 之间"))]
    pub username: String,
    #[validate(email(message = "邮箱格式不正确"))]
    pub email: String,
    #[validate(length(min = 6, max = 64, message = "密码长度必须在 6-64 之间"))]
    pub password: String,
    pub captcha_id: Option<String>,
    pub captcha_code: Option<String>,
}

/// 用户登录请求
#[derive(Debug, Clone, Deserialize, Validate)]
pub struct LoginRequest {
    #[validate(email(message = "邮箱格式不正确"))]
    pub email: String,
    pub password: String,
    pub captcha_id: Option<String>,
    pub captcha_code: Option<String>,
}

/// 更新用户资料请求
#[derive(Debug, Clone, Deserialize, Validate, Default)]
pub struct UpdateProfileRequest {
    #[validate(length(min = 3, max = 32, message = "用户名长度必须在 3-32 之间"))]
    pub username: Option<String>,
    pub avatar: Option<String>,
    pub bio: Option<String>,
}

/// 修改密码请求
#[derive(Debug, Clone, Deserialize, Validate)]
pub struct ChangePasswordRequest {
    pub old_password: String,
    #[validate(length(min = 6, max = 64, message = "密码长度必须在 6-64 之间"))]
    pub new_password: String,
}

/// 用户公开信息（对外展示）
#[derive(Debug, Clone, Serialize)]
pub struct UserPublic {
    pub id: i64,
    pub uuid: String,
    pub username: String,
    pub avatar: Option<String>,
    pub bio: Option<String>,
    pub role: String,
    pub created_at: DateTime<Utc>,
}

impl From<User> for UserPublic {
    fn from(user: User) -> Self {
        Self {
            id: user.id,
            uuid: user.uuid,
            username: user.username,
            avatar: user.avatar,
            bio: None,
            role: user.role,
            created_at: user.created_at,
        }
    }
}

/// OAuth 账号绑定
#[derive(Debug, Clone, FromRow, Serialize)]
pub struct OAuthAccount {
    pub id: i64,
    pub user_id: i64,
    pub provider: String,
    pub provider_user_id: String,
    pub access_token: Option<String>,
    pub refresh_token: Option<String>,
    pub expires_at: Option<DateTime<Utc>>,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

/// API Token
#[derive(Debug, Clone, FromRow, Serialize)]
pub struct ApiToken {
    pub id: i64,
    pub user_id: i64,
    pub name: String,
    pub token: String,
    pub scopes: Option<String>,
    pub last_used_at: Option<DateTime<Utc>>,
    pub expires_at: Option<DateTime<Utc>>,
    pub created_at: DateTime<Utc>,
    pub deleted_at: Option<DateTime<Utc>>,
}

impl User {
    /// 创建新用户
    pub fn new(username: String, email: String, password: String) -> Self {
        Self {
            id: 0,
            uuid: Uuid::new_v4().to_string(),
            username,
            email,
            password,
            avatar: None,
            role: "user".to_string(),
            status: 1,
            capacity_used: 0,
            capacity_max: 104857600, // 100MB 默认容量
            plan_id: None,
            plan_expires_at: None,
            email_verified_at: None,
            last_login_at: None,
            last_login_ip: None,
            created_at: Utc::now(),
            updated_at: Utc::now(),
            deleted_at: None,
        }
    }

    /// 检查用户是否为管理员
    pub fn is_admin(&self) -> bool {
        self.role == "admin" || self.role == "super_admin"
    }

    /// 检查用户是否已验证邮箱
    pub fn is_email_verified(&self) -> bool {
        self.email_verified_at.is_some()
    }

    /// 检查用户套餐是否过期
    pub fn is_plan_expired(&self) -> bool {
        match self.plan_expires_at {
            Some(expires) => expires < Utc::now(),
            None => true,
        }
    }
}
