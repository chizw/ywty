//! 认证处理器

use axum::extract::State;
use axum::Json;

use chrono::Utc;

use crate::error::{AppError, AppResult};

use crate::dto::auth::{
    AuthResponse, CaptchaResponse, CaptchaVerifyRequest, RefreshRequest, ResetPasswordRequest,
    SendVerifyCodeRequest,
};
use crate::handlers::validate_req;
use crate::handlers::CurrentUser;
use crate::utils::captcha::generate_captcha;
use crate::utils::response::ApiResponse;
use crate::AppState;

/// 注册
#[utoipa::path(
    post,
    path = "/api/v1/auth/register",
    request_body = crate::models::user::RegisterRequest,
    responses(
        (status = 200, description = "成功", body = AuthResponse),
    ),
    tag = "认证"
)]
pub async fn register(
    State(state): State<AppState>,
    Json(req): Json<crate::models::user::RegisterRequest>,
) -> AppResult<Json<ApiResponse<AuthResponse>>> {
    validate_req(&req)?;
    enforce_captcha(
        &state,
        req.captcha_id.as_deref(),
        req.captcha_code.as_deref(),
    )
    .await?;
    let resp = state
        .auth_svc
        .register(&req.username, &req.email, &req.password)
        .await?;
    Ok(Json(ApiResponse::success(resp)))
}

/// 登录（支持邮箱 / 用户名 / 手机号）
#[utoipa::path(
    post,
    path = "/api/v1/auth/login",
    request_body = crate::models::user::LoginRequest,
    responses(
        (status = 200, description = "成功", body = AuthResponse),
    ),
    tag = "认证"
)]
pub async fn login(
    State(state): State<AppState>,
    Json(req): Json<crate::models::user::LoginRequest>,
) -> AppResult<Json<ApiResponse<AuthResponse>>> {
    validate_req(&req)?;
    let resp = state.auth_svc.login(&req.account, &req.password).await?;
    Ok(Json(ApiResponse::success(resp)))
}

/// 刷新令牌
#[utoipa::path(
    post,
    path = "/api/v1/auth/refresh",
    request_body = RefreshRequest,
    responses(
        (status = 200, description = "成功", body = AuthResponse),
    ),
    tag = "认证"
)]
pub async fn refresh(
    State(state): State<AppState>,
    Json(req): Json<RefreshRequest>,
) -> AppResult<Json<ApiResponse<AuthResponse>>> {
    let resp = state.auth_svc.refresh(&req.refresh_token).await?;
    Ok(Json(ApiResponse::success(resp)))
}

/// 登出（无状态 JWT，前端清除 token 即可）
#[utoipa::path(
    post,
    path = "/api/v1/auth/logout",
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "认证"
)]
pub async fn logout() -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "message": "ok" }),
    )))
}

/// 重置密码
#[utoipa::path(
    post,
    path = "/api/v1/auth/reset-password",
    request_body = ResetPasswordRequest,
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "认证"
)]
pub async fn reset_password(
    State(state): State<AppState>,
    Json(req): Json<ResetPasswordRequest>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    validate_req(&req)?;
    enforce_captcha(
        &state,
        req.captcha_id.as_deref(),
        req.captcha_code.as_deref(),
    )
    .await?;
    state
        .auth_svc
        .reset_password(&req.email, &req.new_password, &req.verify_code)
        .await?;
    Ok(Json(ApiResponse::success_with_message(
        serde_json::json!({}),
        "密码重置成功",
    )))
}

/// 获取当前用户信息（直接返回 user 对象，不含 token 包装）
#[utoipa::path(
    get,
    path = "/api/v1/auth/me",
    responses(
        (status = 200, description = "成功", body = crate::models::user::UserPublic),
    ),
    tag = "认证"
)]
pub async fn me(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
) -> AppResult<Json<ApiResponse<crate::models::user::UserPublic>>> {
    let brief = state.auth_svc.get_me(user_id).await?;
    // 将 UserBrief 转换为 UserPublic
    Ok(Json(ApiResponse::success(
        crate::models::user::UserPublic {
            id: brief.id,
            uuid: brief.uuid,
            username: brief.username,
            avatar: brief.avatar,
            bio: None,
            role: brief.role,
            is_super_admin: brief.is_super_admin,
            created_at: brief.created_at,
        },
    )))
}

/// 获取图片验证码
///
/// 生成 4 位字符的扭曲 PNG 图片，存储到 DB，返回 base64 图片 + captcha_id。
#[utoipa::path(
    get,
    path = "/api/v1/captcha",
    responses(
        (status = 200, description = "成功", body = CaptchaResponse),
    ),
    tag = "认证"
)]
pub async fn get_captcha(
    State(state): State<AppState>,
) -> AppResult<Json<ApiResponse<CaptchaResponse>>> {
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
        VALUES (?, ?, ?, ?)
        "#,
    )
    .bind(&captcha_id)
    .bind(code.to_lowercase())
    .bind(expired_at)
    .bind(crate::db::now_str())
    .execute(&state.db)
    .await
    .map_err(|e| crate::error::AppError::Internal(format!("存储验证码失败: {}", e)))?;

    Ok(Json(ApiResponse::success(CaptchaResponse {
        captcha_id,
        captcha_image: image_base64,
        expires_in,
    })))
}

/// 校验图片验证码（不区分大小写，验证后标记为已使用）
#[utoipa::path(
    post,
    path = "/api/v1/captcha/verify",
    request_body = CaptchaVerifyRequest,
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "认证"
)]
pub async fn verify_captcha(
    State(state): State<AppState>,
    Json(req): Json<CaptchaVerifyRequest>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
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
            return Ok(Json(ApiResponse::success(
                serde_json::json!({ "valid": false }),
            )));
        }
    };

    // 检查是否已使用
    if used_at.is_some() {
        return Ok(Json(ApiResponse::success(
            serde_json::json!({ "valid": false }),
        )));
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

    Ok(Json(ApiResponse::success(
        serde_json::json!({ "valid": valid }),
    )))
}

/// 发送验证码（邮件）
#[utoipa::path(
    post,
    path = "/api/v1/verify-codes",
    request_body = SendVerifyCodeRequest,
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "认证"
)]
pub async fn send_verify_code(
    State(state): State<AppState>,
    Json(req): Json<SendVerifyCodeRequest>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    validate_req(&req)?;
    let event = req.event.as_deref().unwrap_or("register");
    state.auth_svc.send_verify_code(&req.email, event).await?;
    Ok(Json(ApiResponse::success(serde_json::json!({
        "message": "验证码已发送"
    }))))
}

/// 按设置强制校验图形验证码
///
/// `security.require_captcha` 开启时，注册/找回密码必须携带有效的图形验证码；
/// 未开启时完全跳过（零配置默认关闭）。
pub(crate) async fn enforce_captcha(
    state: &AppState,
    captcha_id: Option<&str>,
    captcha_code: Option<&str>,
) -> AppResult<()> {
    let required = crate::services::settings::get_bool(
        &state.db,
        crate::services::settings::keys::SECURITY_REQUIRE_CAPTCHA,
        false,
    )
    .await?;
    if !required {
        return Ok(());
    }

    let (Some(id), Some(code)) = (captcha_id, captcha_code) else {
        return Err(AppError::Validation("请输入图形验证码".to_string()));
    };

    let now_ts = Utc::now().timestamp();
    let row: Option<(i64, String, Option<i64>)> = sqlx::query_as(
        "SELECT id, code, used_at FROM captchas WHERE captcha_id = ? AND expired_at > ? ORDER BY id DESC LIMIT 1",
    )
    .bind(id)
    .bind(now_ts)
    .fetch_optional(&state.db)
    .await
    .map_err(|e| AppError::Internal(format!("查询验证码失败: {}", e)))?;

    let Some((db_id, stored_code, used_at)) = row else {
        return Err(AppError::Validation("图形验证码无效或已过期".to_string()));
    };
    if used_at.is_some() {
        return Err(AppError::Validation(
            "图形验证码已被使用，请刷新重试".to_string(),
        ));
    }
    if stored_code != code.to_lowercase() {
        return Err(AppError::Validation("图形验证码错误".to_string()));
    }

    let _ = sqlx::query("UPDATE captchas SET used_at = ? WHERE id = ?")
        .bind(now_ts)
        .bind(db_id)
        .execute(&state.db)
        .await;
    Ok(())
}
