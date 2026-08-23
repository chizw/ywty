//! 分享处理器

use axum::extract::{Path, State};
use axum::Json;

use crate::error::AppResult;
use crate::handlers::{validate_req, CurrentUser};
use crate::models::photo::CreateShareRequest;
use crate::utils::response::ApiResponse;
use crate::AppState;

/// 创建分享
#[utoipa::path(
    post,
    path = "/api/v1/shares",
    request_body = CreateShareRequest,
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "分享"
)]
pub async fn create(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Json(req): Json<CreateShareRequest>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    validate_req(&req)?;
    let share = state.share_svc.create(user_id, &req).await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": share }),
    )))
}

/// 列出我的分享
#[utoipa::path(
    get,
    path = "/api/v1/shares",
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "分享"
)]
pub async fn list(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    let shares = state.share_svc.list(user_id).await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": shares }),
    )))
}

/// 更新分享
///
/// - password: 省略 = 不修改, null/"" = 清除, 非空 = 设置
/// - expires_at: 省略 = 不修改, null = 取消过期, RFC3339 = 设置过期
#[utoipa::path(
    patch,
    path = "/api/v1/shares/:id",
    request_body = serde_json::Value,
    params(
        ("id" = i64, Path, description = "分享ID"),
    ),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "分享"
)]
pub async fn update(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Path(id): Path<i64>,
    Json(req): Json<serde_json::Value>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    // password: None = 不修改, Some(None) = 清除, Some(Some("x")) = 设置
    let password = req.get("password").map(|v| {
        if v.is_null() {
            None
        } else {
            v.as_str().map(|s| s.to_string())
        }
    });

    // expires_at: None = 不修改, Some(None) = 取消过期, Some(Some(dt)) = 设置
    let expires_at = req.get("expires_at").map(|v| {
        if v.is_null() {
            None
        } else {
            v.as_str()
                .and_then(|s| chrono::DateTime::parse_from_rfc3339(s).ok())
                .map(|dt| dt.with_timezone(&chrono::Utc))
        }
    });

    // 转换 Option<Option<String>> -> Option<Option<&str>>
    let password_ref: Option<Option<&str>> = password.as_ref().map(|opt| opt.as_deref());

    let share = state
        .share_svc
        .update(user_id, id, password_ref, expires_at)
        .await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": share }),
    )))
}

/// 删除分享
#[utoipa::path(
    delete,
    path = "/api/v1/shares/:id",
    params(
        ("id" = i64, Path, description = "分享ID"),
    ),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "分享"
)]
pub async fn delete(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Path(id): Path<i64>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    state.share_svc.delete(user_id, id).await?;
    Ok(Json(ApiResponse::success_with_message(
        serde_json::json!({}),
        "删除成功",
    )))
}
