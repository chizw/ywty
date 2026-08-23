//! 公告处理器

use axum::extract::{Path, Query, State};
use axum::Json;

use crate::error::AppResult;
use crate::handlers::{validate_req, CurrentUser};
use crate::models::notice::{CreateNoticeRequest, UpdateNoticeRequest};
use crate::utils::response::ApiResponse;
use crate::AppState;

#[derive(Debug, serde::Deserialize)]
pub struct PaginationQuery {
    pub page: Option<i32>,
    pub per_page: Option<i32>,
}

/// 公开：列出公告
#[utoipa::path(
    get,
    path = "/api/v1/notices",
    params(
        ("page" = u64, Query, description = "页码"),
        ("per_page" = u64, Query, description = "每页数量"),
    ),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "公告"
)]
pub async fn list_public(
    State(state): State<AppState>,
    Query(query): Query<PaginationQuery>,
) -> AppResult<Json<serde_json::Value>> {
    let (notices, total) = state
        .notice_svc
        .list_public(query.page.unwrap_or(1), query.per_page.unwrap_or(20))
        .await?;
    Ok(Json(serde_json::json!({
        "data": notices,
        "meta": {
            "current_page": query.page.unwrap_or(1),
            "per_page": query.per_page.unwrap_or(20),
            "total": total,
            "last_page": (total + query.per_page.unwrap_or(20) as i64 - 1) / query.per_page.unwrap_or(20) as i64,
        }
    })))
}

/// 公开：获取公告详情
#[utoipa::path(
    get,
    path = "/api/v1/notices/:id",
    params(
        ("id" = i64, Path, description = "ID"),
    ),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "公告"
)]
pub async fn get_public(
    State(state): State<AppState>,
    Path(id): Path<i64>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    let notice = state.notice_svc.get_public(id).await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": notice }),
    )))
}

/// 管理端：列出公告
#[utoipa::path(
    get,
    path = "/api/v1/admin/notices",
    params(
        ("page" = u64, Query, description = "页码"),
        ("per_page" = u64, Query, description = "每页数量"),
    ),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "公告"
)]
pub async fn admin_list(
    State(state): State<AppState>,
    Query(query): Query<PaginationQuery>,
) -> AppResult<Json<serde_json::Value>> {
    let (notices, total) = state
        .notice_svc
        .admin_list(query.page.unwrap_or(1), query.per_page.unwrap_or(20))
        .await?;
    Ok(Json(serde_json::json!({
        "data": notices,
        "meta": {
            "current_page": query.page.unwrap_or(1),
            "per_page": query.per_page.unwrap_or(20),
            "total": total,
            "last_page": (total + query.per_page.unwrap_or(20) as i64 - 1) / query.per_page.unwrap_or(20) as i64,
        }
    })))
}

/// 管理端：创建公告
#[utoipa::path(
    post,
    path = "/api/v1/admin/notices",
    request_body = CreateNoticeRequest,
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "公告"
)]
pub async fn admin_create(
    State(state): State<AppState>,
    CurrentUser { .. }: CurrentUser,
    Json(req): Json<CreateNoticeRequest>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    validate_req(&req)?;
    let notice = state.notice_svc.admin_create(&req).await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": notice }),
    )))
}

/// 管理端：更新公告
#[utoipa::path(
    patch,
    path = "/api/v1/admin/notices/:id",
    params(
        ("id" = i64, Path, description = "ID"),
    ),
    request_body = UpdateNoticeRequest,
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "公告"
)]
pub async fn admin_update(
    State(state): State<AppState>,
    Path(id): Path<i64>,
    Json(req): Json<UpdateNoticeRequest>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    validate_req(&req)?;
    let notice = state.notice_svc.admin_update(id, &req).await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": notice }),
    )))
}

/// 管理端：删除公告
#[utoipa::path(
    delete,
    path = "/api/v1/admin/notices/:id",
    params(
        ("id" = i64, Path, description = "ID"),
    ),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "公告"
)]
pub async fn admin_delete(
    State(state): State<AppState>,
    Path(id): Path<i64>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    state.notice_svc.admin_delete(id).await?;
    Ok(Json(ApiResponse::success_with_message(
        serde_json::json!({}),
        "删除成功",
    )))
}
