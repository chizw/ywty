//! 容量处理器

use axum::extract::State;
use axum::Json;

use crate::error::AppResult;
use crate::handlers::CurrentUser;
use crate::utils::response::ApiResponse;
use crate::AppState;

/// 获取容量信息
#[utoipa::path(
    get,
    path = "/api/v1/capacity",
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "用户"
)]
pub async fn get(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    let info = state.capacity_svc.get_user_capacity(user_id).await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": info }),
    )))
}
