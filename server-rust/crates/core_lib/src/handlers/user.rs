//! 用户处理器

use axum::extract::State;
use axum::Json;

use crate::error::AppResult;

use crate::dto::user::{ChangeEmailRequest, UserProfile};
use crate::handlers::{validate_req, CurrentUser};
use crate::AppState;

/// 获取用户资料
pub async fn get_profile(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
) -> AppResult<Json<UserProfile>> {
    let profile = state.user_svc.get_profile(user_id).await?;
    Ok(Json(profile))
}

/// 更新用户资料
pub async fn update_profile(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Json(req): Json<crate::models::user::UpdateProfileRequest>,
) -> AppResult<Json<UserProfile>> {
    validate_req(&req)?;
    let profile = state
        .user_svc
        .update_profile(
            user_id,
            req.username.as_deref(),
            req.avatar.as_deref(),
            req.bio.as_deref(),
        )
        .await?;
    Ok(Json(profile))
}

/// 修改密码
pub async fn change_password(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Json(req): Json<crate::models::user::ChangePasswordRequest>,
) -> AppResult<Json<serde_json::Value>> {
    validate_req(&req)?;
    state
        .user_svc
        .change_password(user_id, &req.old_password, &req.new_password)
        .await?;
    Ok(Json(serde_json::json!({ "message": "密码修改成功" })))
}

/// 修改邮箱
pub async fn change_email(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Json(req): Json<ChangeEmailRequest>,
) -> AppResult<Json<serde_json::Value>> {
    validate_req(&req)?;
    state
        .user_svc
        .change_email(user_id, &req.new_email, &req.verify_code)
        .await?;
    Ok(Json(serde_json::json!({ "message": "邮箱修改成功" })))
}
