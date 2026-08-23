//! 单页处理器

use axum::extract::{Path, Query, State};
use axum::Json;

use crate::error::AppResult;
use crate::handlers::{validate_req, CurrentUser};
use crate::models::page::{CreatePageRequest, UpdatePageRequest};
use crate::utils::response::ApiResponse;
use crate::AppState;

#[derive(Debug, serde::Deserialize)]
pub struct PaginationQuery {
    pub page: Option<i32>,
    pub per_page: Option<i32>,
}

/// 公开：列出页面
#[utoipa::path(
    get,
    path = "/api/v1/pages",
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "页面"
)]
pub async fn list_public(
    State(state): State<AppState>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    let pages = state.page_svc.list_public().await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": pages }),
    )))
}

/// 公开：通过 slug 获取页面
#[utoipa::path(
    get,
    path = "/api/v1/pages/:slug",
    params(
        ("slug" = String, Path, description = "页面别名"),
    ),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "页面"
)]
pub async fn get_public(
    State(state): State<AppState>,
    Path(slug): Path<String>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    let page = state.page_svc.get_by_slug(&slug).await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": page }),
    )))
}

/// 管理端：列出页面
#[utoipa::path(
    get,
    path = "/api/v1/admin/pages",
    params(
        ("page" = u64, Query, description = "页码"),
        ("per_page" = u64, Query, description = "每页数量"),
    ),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "页面"
)]
pub async fn admin_list(
    State(state): State<AppState>,
    Query(query): Query<PaginationQuery>,
) -> AppResult<Json<serde_json::Value>> {
    let (pages, total) = state
        .page_svc
        .admin_list(query.page.unwrap_or(1), query.per_page.unwrap_or(20))
        .await?;
    Ok(Json(serde_json::json!({
        "data": pages,
        "meta": {
            "current_page": query.page.unwrap_or(1),
            "per_page": query.per_page.unwrap_or(20),
            "total": total,
            "last_page": (total + query.per_page.unwrap_or(20) as i64 - 1) / query.per_page.unwrap_or(20) as i64,
        }
    })))
}

/// 管理端：获取页面详情
#[utoipa::path(
    get,
    path = "/api/v1/admin/pages/:id",
    params(
        ("id" = i64, Path, description = "ID"),
    ),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "页面"
)]
pub async fn admin_get(
    State(state): State<AppState>,
    Path(id): Path<i64>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    let page = state.page_svc.admin_get(id).await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": page }),
    )))
}

/// 管理端：创建页面
#[utoipa::path(
    post,
    path = "/api/v1/admin/pages",
    request_body = CreatePageRequest,
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "页面"
)]
pub async fn admin_create(
    State(state): State<AppState>,
    CurrentUser { .. }: CurrentUser,
    Json(req): Json<CreatePageRequest>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    validate_req(&req)?;
    let page = state.page_svc.admin_create(&req).await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": page }),
    )))
}

/// 管理端：更新页面
#[utoipa::path(
    patch,
    path = "/api/v1/admin/pages/:id",
    params(
        ("id" = i64, Path, description = "ID"),
    ),
    request_body = UpdatePageRequest,
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "页面"
)]
pub async fn admin_update(
    State(state): State<AppState>,
    Path(id): Path<i64>,
    Json(req): Json<UpdatePageRequest>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    validate_req(&req)?;
    let page = state.page_svc.admin_update(id, &req).await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": page }),
    )))
}

/// 管理端：删除页面
#[utoipa::path(
    delete,
    path = "/api/v1/admin/pages/:id",
    params(
        ("id" = i64, Path, description = "ID"),
    ),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "页面"
)]
pub async fn admin_delete(
    State(state): State<AppState>,
    Path(id): Path<i64>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    state.page_svc.admin_delete(id).await?;
    Ok(Json(ApiResponse::success_with_message(
        serde_json::json!({}),
        "删除成功",
    )))
}
