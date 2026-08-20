//! 认证处理器

use axum::extract::{State};
use axum::Json;

use chrono::Utc;

use crate::error::AppResult;

use crate::dto::auth::{
    AuthResponse, CaptchaResponse, CaptchaVerifyRequest, RefreshRequest, ResetPasswordRequest,
    SendVerifyCodeRequest,
};
use crate::handlers::validate_req;
use crate::utils::captcha::generate_captcha;
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

/// 获取图片验证码
///
/// 生成 4 位字符的扭曲 PNG 图片，存储到 DB，返回 base64 图片 + captcha_id。
pub async fn get_captcha(
    State(state): State<AppState>,
) -> AppResult<Json<CaptchaResponse>> {
    let captcha = generate_captcha()?;
    let captcha_id = captcha.id.clone();
    let code = captcha.code.clone();
    let image_base64 = captcha.image_base64.clone();
    let expires_in = captcha.expires_in;

    // 存储到 DB
    let now_ts = Utc::now().timestamp();
    let expired_at = now_ts + expires_in as i64;

    sqlx::query(
        r#"
        INSERT INTO captchas (captcha_id, code, expired_at, created_at)
        VALUES (?, ?, ?, datetime('now'))
        "#,
    )
    .bind(&captcha_id)
    .bind(code.to_lowercase())
    .bind(expired_at)
    .execute(&state.db)
    .await
    .map_err(|e| crate::error::AppError::Internal(format!("存储验证码失败: {}", e)))?;

    Ok(Json(CaptchaResponse {
        captcha_id,
        captcha_image: image_base64,
        expires_in,
    }))
}

/// 校验图片验证码（不区分大小写，验证后标记为已使用）
pub async fn verify_captcha(
    State(state): State<AppState>,
    Json(req): Json<CaptchaVerifyRequest>,
) -> AppResult<Json<serde_json::Value>> {
    let now_ts = Utc::now().timestamp();

    // 查询验证码
    let row: Option<(i64, String, Option<i64>)> = sqlx::query_as(
        "SELECT id, code, used_at FROM captchas WHERE captcha_id = ? AND expired_at > ? ORDER BY id DESC LIMIT 1"
    )
    .bind(&req.captcha_id)
    .bind(now_ts)
    .fetch_optional(&state.db)
    .await
    .map_err(|e| crate::error::AppError::Internal(format!("查询验证码失败: {}", e)))?;

    let (db_id, stored_code, used_at) = match row {
        Some(r) => r,
        None => {
            return Ok(Json(serde_json::json!({ "valid": false })));
        }
    };

    // 检查是否已使用
    if used_at.is_some() {
        return Ok(Json(serde_json::json!({ "valid": false })));
    }

    // 不区分大小写比对
    let valid = stored_code == req.captcha_code.to_lowercase();

    if valid {
        // 标记为已使用
        let _ = sqlx::query("UPDATE captchas SET used_at = ? WHERE id = ?")
            .bind(now_ts)
            .bind(db_id)
            .execute(&state.db)
            .await;
    }

    Ok(Json(serde_json::json!({ "valid": valid })))
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
