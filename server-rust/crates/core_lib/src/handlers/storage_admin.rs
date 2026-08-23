//! 存储管理处理器

use axum::extract::{Path, State};
use axum::Json;

use crate::error::AppResult;
use crate::handlers::{validate_req, CurrentUser};
use crate::utils::response::ApiResponse;
use crate::AppState;

#[derive(Debug, serde::Deserialize, validator::Validate, utoipa::ToSchema)]
pub struct CreateStorageRequest {
    pub name: String,
    pub provider: String,
    pub intro: Option<String>,
    pub prefix: Option<String>,
    pub options: Option<String>,
}

#[derive(Debug, serde::Deserialize, utoipa::ToSchema)]
pub struct UpdateStorageRequest {
    pub name: Option<String>,
    pub intro: Option<String>,
    pub prefix: Option<String>,
    pub options: Option<String>,
}

/// 列出存储驱动
#[utoipa::path(
    get,
    path = "/api/v1/admin/storage/drivers",
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "存储"
)]
pub async fn list_drivers() -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    Ok(Json(ApiResponse::success(serde_json::json!({
        "data": {
            "drivers": [{"name": "local", "id": "local"}],
            "count": 1,
        }
    }))))
}

/// 列出存储
#[utoipa::path(
    get,
    path = "/api/v1/admin/storages",
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "存储"
)]
pub async fn list_storages(
    State(state): State<AppState>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    let storages = state.storage_admin_svc.list_storages().await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": storages }),
    )))
}

/// 创建存储
#[utoipa::path(
    post,
    path = "/api/v1/admin/storages/create",
    request_body = CreateStorageRequest,
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "存储"
)]
pub async fn create_storage(
    State(state): State<AppState>,
    CurrentUser { .. }: CurrentUser,
    Json(req): Json<CreateStorageRequest>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    validate_req(&req)?;
    let storage = state
        .storage_admin_svc
        .create_storage(
            &req.name,
            &req.provider,
            req.intro.as_deref(),
            req.prefix.as_deref(),
            req.options.as_deref(),
        )
        .await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": storage }),
    )))
}

/// 更新存储
#[utoipa::path(
    patch,
    path = "/api/v1/admin/storages/update/:id",
    params(
        ("id" = i64, Path, description = "存储 ID"),
    ),
    request_body = UpdateStorageRequest,
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "存储"
)]
pub async fn update_storage(
    State(state): State<AppState>,
    Path(id): Path<i64>,
    Json(req): Json<UpdateStorageRequest>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    state
        .storage_admin_svc
        .update_storage(
            id,
            req.name.as_deref(),
            req.intro.as_deref(),
            req.prefix.as_deref(),
            req.options.as_deref(),
        )
        .await?;
    Ok(Json(ApiResponse::success_with_message(
        serde_json::json!({}),
        "更新成功",
    )))
}

/// 删除存储
#[utoipa::path(
    delete,
    path = "/api/v1/admin/storages/delete/:id",
    params(
        ("id" = i64, Path, description = "存储 ID"),
    ),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "存储"
)]
pub async fn delete_storage(
    State(state): State<AppState>,
    Path(id): Path<i64>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    state.storage_admin_svc.delete_storage(id).await?;
    Ok(Json(ApiResponse::success_with_message(
        serde_json::json!({}),
        "删除成功",
    )))
}

/// 跨存储复制（占位）
#[utoipa::path(
    post,
    path = "/api/v1/admin/storage/copy",
    request_body = serde_json::Value,
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "存储"
)]
pub async fn copy(
    State(_state): State<AppState>,
    CurrentUser { .. }: CurrentUser,
    Json(req): Json<serde_json::Value>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": req }),
    )))
}
