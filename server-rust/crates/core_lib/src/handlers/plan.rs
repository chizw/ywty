//! 套餐处理器

use axum::extract::{Path, State};
use axum::Json;

use crate::error::AppResult;
use crate::handlers::{validate_req, CurrentUser};
use crate::models::plan::AdminPlanRequest;
use crate::utils::response::ApiResponse;
use crate::AppState;

#[derive(Debug, serde::Deserialize)]
pub struct ListQuery {
    pub page: Option<i32>,
    pub per_page: Option<i32>,
}

#[utoipa::path(
    get,
    path = "/api/v1/plans",
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "套餐"
)]
/// 公开：列出套餐
pub async fn list_public(
    State(state): State<AppState>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    let details = state.plan_svc.list_public_with_prices().await?;
    let value = serde_json::to_value(details)
        .map_err(|e| crate::error::AppError::Internal(e.to_string()))?;
    Ok(Json(ApiResponse::success(value)))
}

#[utoipa::path(
    get,
    path = "/api/v1/plans/:id",
    params(
        ("id" = i64, Path, description = "套餐ID"),
    ),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "套餐"
)]
/// 公开：获取套餐详情
pub async fn get_public(
    State(state): State<AppState>,
    Path(id): Path<i64>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    let detail = state.plan_svc.get(id).await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": detail }),
    )))
}

#[utoipa::path(
    get,
    path = "/api/v1/admin/plans",
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "套餐"
)]
/// 管理端：列出所有套餐
pub async fn admin_list(
    State(state): State<AppState>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    let plans = state.plan_svc.list_all().await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": plans }),
    )))
}

#[utoipa::path(
    post,
    path = "/api/v1/admin/plans",
    request_body = AdminPlanRequest,
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "套餐"
)]
/// 管理端：创建套餐
pub async fn admin_create(
    State(state): State<AppState>,
    CurrentUser { .. }: CurrentUser,
    Json(req): Json<AdminPlanRequest>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    validate_req(&req)?;
    let plan = state.plan_svc.create(&req).await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": plan }),
    )))
}

#[utoipa::path(
    get,
    path = "/api/v1/admin/plans/:id",
    params(
        ("id" = i64, Path, description = "套餐ID"),
    ),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "套餐"
)]
/// 管理端：获取套餐详情
pub async fn admin_get(
    State(state): State<AppState>,
    Path(id): Path<i64>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    let detail = state.plan_svc.get(id).await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": detail }),
    )))
}

#[utoipa::path(
    patch,
    path = "/api/v1/admin/plans/:id",
    params(
        ("id" = i64, Path, description = "套餐ID"),
    ),
    request_body = AdminPlanRequest,
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "套餐"
)]
/// 管理端：更新套餐
pub async fn admin_update(
    State(state): State<AppState>,
    Path(id): Path<i64>,
    Json(req): Json<AdminPlanRequest>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    validate_req(&req)?;
    let plan = state.plan_svc.update(id, &req).await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": plan }),
    )))
}

#[utoipa::path(
    delete,
    path = "/api/v1/admin/plans/:id",
    params(
        ("id" = i64, Path, description = "套餐ID"),
    ),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "套餐"
)]
/// 管理端：删除套餐
pub async fn admin_delete(
    State(state): State<AppState>,
    Path(id): Path<i64>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    state.plan_svc.delete(id).await?;
    Ok(Json(ApiResponse::success_with_message(
        serde_json::json!({}),
        "删除成功",
    )))
}

#[utoipa::path(
    post,
    path = "/api/v1/admin/plans/:id/toggle",
    params(
        ("id" = i64, Path, description = "套餐ID"),
    ),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "套餐"
)]
/// 管理端：切换上架
pub async fn admin_toggle_up(
    State(state): State<AppState>,
    Path(id): Path<i64>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    let plan = state.plan_svc.toggle_up(id).await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": plan }),
    )))
}
