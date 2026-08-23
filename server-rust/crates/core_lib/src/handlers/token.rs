//! API Token 处理器

use axum::extract::{Path, State};
use axum::Json;

use crate::error::AppResult;
use crate::handlers::CurrentUser;
use crate::utils::response::ApiResponse;
use crate::AppState;

#[derive(Debug, serde::Deserialize, utoipa::ToSchema)]
pub struct CreateTokenRequest {
    pub name: String,
    pub ttl_days: Option<i64>,
}

/// 列出 tokens
#[utoipa::path(
    get,
    path = "/api/v1/tokens",
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "管理后台"
)]
pub async fn list(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    let tokens = state.token_svc.list(user_id).await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": tokens }),
    )))
}

/// 创建 token
#[utoipa::path(
    post,
    path = "/api/v1/tokens",
    request_body = CreateTokenRequest,
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "管理后台"
)]
pub async fn create(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Json(req): Json<CreateTokenRequest>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    let (raw, info) = state
        .token_svc
        .create(user_id, &req.name, req.ttl_days)
        .await?;
    Ok(Json(ApiResponse::success(serde_json::json!({
        "data": {
            "token": raw,
            "access_token": raw,
            "info": info,
        }
    }))))
}

/// 撤销 token
#[utoipa::path(
    delete,
    path = "/api/v1/tokens/:id",
    params(
        ("id" = i64, Path, description = "Token ID"),
    ),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "管理后台"
)]
pub async fn revoke(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Path(id): Path<i64>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    state.token_svc.revoke(user_id, id).await?;
    Ok(Json(ApiResponse::success_with_message(
        serde_json::json!({}),
        "Token 已撤销",
    )))
}
