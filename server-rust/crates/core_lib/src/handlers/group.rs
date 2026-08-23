//! 群组处理器

use axum::extract::{Path, Query, State};
use axum::Json;

use crate::error::AppResult;
use crate::handlers::{validate_req, CurrentUser};
use crate::models::group::{CreateGroupRequest, UpdateGroupRequest};
use crate::utils::response::ApiResponse;
use crate::AppState;

#[derive(Debug, serde::Deserialize)]
pub struct PaginationQuery {
    pub page: Option<i32>,
    pub per_page: Option<i32>,
}

/// 列出群组
#[utoipa::path(
    get,
    path = "/api/v1/admin/groups",
    params(
        ("page" = Option<i32>, Query, description = "页码"),
        ("per_page" = Option<i32>, Query, description = "每页数量"),
    ),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "管理后台"
)]
pub async fn list(
    State(state): State<AppState>,
    Query(query): Query<PaginationQuery>,
) -> AppResult<Json<serde_json::Value>> {
    let (groups, total) = state
        .group_svc
        .list(query.page.unwrap_or(1), query.per_page.unwrap_or(20))
        .await?;
    Ok(Json(serde_json::json!({
        "data": groups,
        "meta": {
            "current_page": query.page.unwrap_or(1),
            "per_page": query.per_page.unwrap_or(20),
            "total": total,
            "last_page": (total + query.per_page.unwrap_or(20) as i64 - 1) / query.per_page.unwrap_or(20) as i64,
        }
    })))
}

/// 获取群组详情
#[utoipa::path(
    get,
    path = "/api/v1/admin/groups/:id",
    params(
        ("id" = i64, Path, description = "群组 ID"),
    ),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "管理后台"
)]
pub async fn get(
    State(state): State<AppState>,
    Path(id): Path<i64>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    let group = state.group_svc.get(id).await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": group }),
    )))
}

/// 创建群组
#[utoipa::path(
    post,
    path = "/api/v1/admin/groups",
    request_body = CreateGroupRequest,
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "管理后台"
)]
pub async fn create(
    State(state): State<AppState>,
    CurrentUser { .. }: CurrentUser,
    Json(req): Json<CreateGroupRequest>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    validate_req(&req)?;
    let group = state.group_svc.create(&req).await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": group }),
    )))
}

/// 更新群组
#[utoipa::path(
    put,
    path = "/api/v1/admin/groups/:id",
    params(
        ("id" = i64, Path, description = "群组 ID"),
    ),
    request_body = UpdateGroupRequest,
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "管理后台"
)]
pub async fn update(
    State(state): State<AppState>,
    Path(id): Path<i64>,
    Json(req): Json<UpdateGroupRequest>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    validate_req(&req)?;
    let group = state.group_svc.update(id, &req).await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": group }),
    )))
}

/// 删除群组
#[utoipa::path(
    delete,
    path = "/api/v1/admin/groups/:id",
    params(
        ("id" = i64, Path, description = "群组 ID"),
    ),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "管理后台"
)]
pub async fn delete(
    State(state): State<AppState>,
    Path(id): Path<i64>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    state.group_svc.delete(id).await?;
    Ok(Json(ApiResponse::success_with_message(
        serde_json::json!({}),
        "删除成功",
    )))
}
