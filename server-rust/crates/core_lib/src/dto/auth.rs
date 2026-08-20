//! 认证域 DTO

use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};

/// 登录/注册 成功响应
#[derive(Debug, Clone, Serialize)]
pub struct AuthResponse {
    pub access_token: String,
    pub refresh_token: String,
    pub token_type: String,
    pub expires_in: u64,
    pub user: UserBrief,
}

/// 用户简要信息（嵌套在 AuthResponse 中）
#[derive(Debug, Clone, Serialize)]
pub struct UserBrief {
    pub id: i64,
    pub uuid: String,
    pub username: String,
    pub email: String,
    pub avatar: Option<String>,
    pub role: String,
    pub created_at: DateTime<Utc>,
}

/// 刷新令牌请求
#[derive(Debug, Clone, Deserialize)]
pub struct RefreshRequest {
    pub refresh_token: String,
}

/// 重置密码请求
#[derive(Debug, Clone, Deserialize, validator::Validate)]
pub struct ResetPasswordRequest {
    #[validate(email(message = "邮箱格式不正确"))]
    pub email: String,
    #[validate(length(min = 6, max = 64, message = "密码长度必须在 6-64 之间"))]
    pub new_password: String,
    pub verify_code: String,
}

/// 验证码响应（图片验证码，base64 PNG）
#[derive(Debug, Clone, Serialize)]
pub struct CaptchaResponse {
    pub captcha_id: String,
    pub captcha_image: String, // base64 或 URL
    pub expires_in: u64,
}

/// 验证码校验请求
#[derive(Debug, Clone, Deserialize)]
pub struct CaptchaVerifyRequest {
    pub captcha_id: String,
    pub captcha_code: String,
}

/// 发送验证码请求
#[derive(Debug, Clone, Deserialize, validator::Validate)]
pub struct SendVerifyCodeRequest {
    #[validate(email(message = "邮箱格式不正确"))]
    pub email: String,
    pub event: Option<String>, // register / login / reset_password
}
