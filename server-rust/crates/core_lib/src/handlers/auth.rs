//! 认证处理器

use axum::extract::State;
use axum::Json;

use crate::error::AppResult;

use crate::dto::auth::{
    AuthResponse, CaptchaResponse, CaptchaVerifyRequest, RefreshRequest, ResetPasswordRequest,
    SendVerifyCodeRequest,
};
use crate::handlers::validate_req;
use crate::AppState;
use crate::handlers::CurrentUser;

/// 注册
pub async fn register(
    State(state): State<AppState>,
    Json(req): Json<crate::models::user::RegisterRequest>,
) -> AppResult<Json<AuthResponse>> {
    validate_req(&req)?;
    let resp = state
        .auth_svc
        .register(&req.username, &req.email, &req.password)
        .await?;
    Ok(Json(resp))
}

/// 登录
pub async fn login(
    State(state): State<AppState>,
    Json(req): Json<crate::models::user::LoginRequest>,
) -> AppResult<Json<AuthResponse>> {
    validate_req(&req)?;
    let resp = state.auth_svc.login(&req.email, &req.password).await?;
    Ok(Json(resp))
}

/// 刷新令牌
pub async fn refresh(
    State(state): State<AppState>,
    Json(req): Json<RefreshRequest>,
) -> AppResult<Json<AuthResponse>> {
    let resp = state.auth_svc.refresh(&req.refresh_token).await?;
    Ok(Json(resp))
}

/// 登出（无状态 JWT，前端清除 token 即可）
pub async fn logout() -> AppResult<Json<serde_json::Value>> {
    Ok(Json(serde_json::json!({ "message": "ok" })))
}

/// 重置密码
pub async fn reset_password(
    State(state): State<AppState>,
    Json(req): Json<ResetPasswordRequest>,
) -> AppResult<Json<serde_json::Value>> {
    validate_req(&req)?;
    state
        .auth_svc
        .reset_password(&req.email, &req.new_password, &req.verify_code)
        .await?;
    Ok(Json(serde_json::json!({ "message": "密码重置成功" })))
}

/// 获取当前用户信息
pub async fn me(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
) -> AppResult<Json<AuthResponse>> {
    let user = state.auth_svc.get_me(user_id).await?;
    // 返回简化版（不含 token）
    Ok(Json(AuthResponse {
        access_token: String::new(),
        refresh_token: String::new(),
        token_type: "Bearer".to_string(),
        expires_in: 0,
        user,
    }))
}

/// 获取验证码（占位实现）
pub async fn get_captcha() -> AppResult<Json<CaptchaResponse>> {
    // TODO: 对接图片验证码生成
    Ok(Json(CaptchaResponse {
        captcha_id: "placeholder".to_string(),
        captcha_image: String::new(),
        expires_in: 300,
    }))
}

/// 校验验证码（占位实现）
pub async fn verify_captcha(
    Json(_req): Json<CaptchaVerifyRequest>,
) -> AppResult<Json<serde_json::Value>> {
    Ok(Json(serde_json::json!({ "valid": true })))
}

/// 发送验证码（邮件）
pub async fn send_verify_code(
    State(state): State<AppState>,
    Json(req): Json<SendVerifyCodeRequest>,
) -> AppResult<Json<serde_json::Value>> {
    validate_req(&req)?;
    let event = req.event.as_deref().unwrap_or("register");
    let code = state
        .auth_svc
        .send_verify_code(&req.email, event)
        .await?;
    // 注意：生产环境不应返回验证码，此处便于开发调试
    Ok(Json(serde_json::json!({
        "message": "验证码已发送",
        "debug_code": code
    })))
}
