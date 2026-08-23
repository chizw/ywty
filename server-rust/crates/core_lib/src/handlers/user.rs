//! 用户处理器

use axum::extract::State;
use axum::Json;

use crate::error::AppResult;
use crate::utils::response::ApiResponse;

use crate::dto::user::{ChangeEmailRequest, UserProfile};
use crate::handlers::{validate_req, CurrentUser};
use crate::AppState;

/// 获取用户资料
#[utoipa::path(
    get,
    path = "/api/v1/user/profile",
    responses(
        (status = 200, description = "成功", body = UserProfile),
    ),
    tag = "用户"
)]
pub async fn get_profile(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
) -> AppResult<Json<ApiResponse<UserProfile>>> {
    let profile = state.user_svc.get_profile(user_id).await?;
    Ok(Json(ApiResponse::success(profile)))
}

/// 更新用户资料
#[utoipa::path(
    patch,
    path = "/api/v1/user/profile",
    request_body = crate::models::user::UpdateProfileRequest,
    responses(
        (status = 200, description = "成功", body = crate::models::user::UserPublic),
    ),
    tag = "用户"
)]
pub async fn update_profile(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Json(req): Json<crate::models::user::UpdateProfileRequest>,
) -> AppResult<Json<ApiResponse<UserProfile>>> {
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
    Ok(Json(ApiResponse::success(profile)))
}

/// 修改密码
#[utoipa::path(
    post,
    path = "/api/v1/user/change-password",
    request_body = crate::models::user::ChangePasswordRequest,
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "用户"
)]
pub async fn change_password(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Json(req): Json<crate::models::user::ChangePasswordRequest>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    validate_req(&req)?;
    state
        .user_svc
        .change_password(user_id, &req.old_password, &req.new_password)
        .await?;
    Ok(Json(ApiResponse::success_with_message(
        serde_json::json!({}),
        "密码修改成功",
    )))
}

/// 修改邮箱
#[utoipa::path(
    post,
    path = "/api/v1/user/change-email",
    request_body = ChangeEmailRequest,
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "用户"
)]
pub async fn change_email(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Json(req): Json<ChangeEmailRequest>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    validate_req(&req)?;
    state
        .user_svc
        .change_email(user_id, &req.new_email, &req.verify_code)
        .await?;
    Ok(Json(ApiResponse::success_with_message(
        serde_json::json!({}),
        "邮箱修改成功",
    )))
}
