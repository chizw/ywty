//! 工单处理器

use axum::extract::{Path, Query, State};
use axum::Json;

use crate::error::AppResult;
use crate::handlers::{validate_req, CurrentUser};
use crate::models::ticket::CreateTicketRequest;
use crate::utils::response::ApiResponse;
use crate::AppState;

#[derive(Debug, serde::Deserialize)]
pub struct PaginationQuery {
    pub page: Option<i32>,
    pub per_page: Option<i32>,
}

#[derive(Debug, serde::Deserialize)]
pub struct AdminListQuery {
    pub page: Option<i32>,
    pub per_page: Option<i32>,
    pub status: Option<String>,
    pub level: Option<String>,
}

#[derive(Debug, serde::Deserialize, utoipa::ToSchema)]
pub struct ReplyRequest {
    pub content: String,
}

#[derive(Debug, serde::Deserialize, utoipa::ToSchema)]
pub struct UpdateStatusRequest {
    pub status: String,
}

#[utoipa::path(
    get,
    path = "/api/v1/tickets",
    params(
        ("page" = Option<i32>, Query, description = "页码"),
        ("per_page" = Option<i32>, Query, description = "每页数量"),
    ),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "工单"
)]
/// 用户侧：列出工单
pub async fn list(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Query(query): Query<PaginationQuery>,
) -> AppResult<Json<serde_json::Value>> {
    let (tickets, total) = state
        .ticket_svc
        .list(
            user_id,
            query.page.unwrap_or(1),
            query.per_page.unwrap_or(20),
        )
        .await?;
    Ok(Json(serde_json::json!({
        "data": tickets,
        "meta": {
            "current_page": query.page.unwrap_or(1),
            "per_page": query.per_page.unwrap_or(20),
            "total": total,
            "last_page": (total + query.per_page.unwrap_or(20) as i64 - 1) / query.per_page.unwrap_or(20) as i64,
        }
    })))
}

#[utoipa::path(
    post,
    path = "/api/v1/tickets",
    request_body = CreateTicketRequest,
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "工单"
)]
/// 用户侧：创建工单
pub async fn create(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Json(req): Json<CreateTicketRequest>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    validate_req(&req)?;
    let ticket = state.ticket_svc.create(user_id, &req).await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": ticket }),
    )))
}

#[utoipa::path(
    get,
    path = "/api/v1/tickets/:id",
    params(
        ("id" = i64, Path, description = "工单ID"),
    ),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "工单"
)]
/// 用户侧：获取工单详情
pub async fn get(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Path(id): Path<i64>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    let detail = state.ticket_svc.get(user_id, id).await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": detail }),
    )))
}

#[utoipa::path(
    post,
    path = "/api/v1/tickets/:id/replies",
    params(
        ("id" = i64, Path, description = "工单ID"),
    ),
    request_body = ReplyRequest,
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "工单"
)]
/// 用户侧：回复工单
pub async fn reply(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Path(id): Path<i64>,
    Json(req): Json<ReplyRequest>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    let reply = state.ticket_svc.reply(user_id, id, &req.content).await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": reply }),
    )))
}

/// 用户侧：按工单列出回复（时间正序）
#[utoipa::path(
    get,
    path = "/api/v1/tickets/:id/replies",
    params(
        ("id" = i64, Path, description = "工单ID"),
    ),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "工单"
)]
pub async fn list_replies(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Path(id): Path<i64>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    // 复用 get 做所有权校验
    let detail = state.ticket_svc.get(user_id, id).await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": detail.replies }),
    )))
}

#[utoipa::path(
    post,
    path = "/api/v1/tickets/:id/close",
    params(
        ("id" = i64, Path, description = "工单ID"),
    ),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "工单"
)]
/// 用户侧：关闭工单
pub async fn close(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Path(id): Path<i64>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    state.ticket_svc.close(user_id, id).await?;
    Ok(Json(ApiResponse::success_with_message(
        serde_json::json!({}),
        "工单已关闭",
    )))
}

#[utoipa::path(
    get,
    path = "/api/v1/admin/tickets",
    params(
        ("page" = Option<i32>, Query, description = "页码"),
        ("per_page" = Option<i32>, Query, description = "每页数量"),
        ("status" = Option<String>, Query, description = "工单状态"),
        ("level" = Option<String>, Query, description = "工单等级"),
    ),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "工单"
)]
/// 管理端：列出工单
pub async fn admin_list(
    State(state): State<AppState>,
    Query(query): Query<AdminListQuery>,
) -> AppResult<Json<serde_json::Value>> {
    let (tickets, total) = state
        .ticket_svc
        .admin_list(
            query.page.unwrap_or(1),
            query.per_page.unwrap_or(20),
            query.status.as_deref(),
            query.level.as_deref(),
        )
        .await?;
    Ok(Json(serde_json::json!({
        "data": tickets,
        "meta": {
            "current_page": query.page.unwrap_or(1),
            "per_page": query.per_page.unwrap_or(20),
            "total": total,
            "last_page": (total + query.per_page.unwrap_or(20) as i64 - 1) / query.per_page.unwrap_or(20) as i64,
        }
    })))
}

#[utoipa::path(
    get,
    path = "/api/v1/admin/tickets/:id",
    params(
        ("id" = i64, Path, description = "工单ID"),
    ),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "工单"
)]
/// 管理端：获取工单详情
pub async fn admin_get(
    State(state): State<AppState>,
    Path(id): Path<i64>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    let detail = state.ticket_svc.admin_get(id).await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": detail }),
    )))
}

/// 管理端：按工单列出回复（时间正序）
#[utoipa::path(
    get,
    path = "/api/v1/admin/tickets/:id/replies",
    params(
        ("id" = i64, Path, description = "工单ID"),
    ),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "工单"
)]
pub async fn admin_list_replies(
    State(state): State<AppState>,
    Path(id): Path<i64>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    let replies = state.ticket_svc.list_replies(id).await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": replies }),
    )))
}

#[utoipa::path(
    post,
    path = "/api/v1/admin/tickets/:id/replies",
    params(
        ("id" = i64, Path, description = "工单ID"),
    ),
    request_body = ReplyRequest,
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "工单"
)]
/// 管理端：回复工单
pub async fn admin_reply(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Path(id): Path<i64>,
    Json(req): Json<ReplyRequest>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    let reply = state
        .ticket_svc
        .admin_reply(user_id, id, &req.content)
        .await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": reply }),
    )))
}

#[utoipa::path(
    patch,
    path = "/api/v1/admin/tickets/:id/status",
    params(
        ("id" = i64, Path, description = "工单ID"),
    ),
    request_body = UpdateStatusRequest,
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "工单"
)]
/// 管理端：更新状态
pub async fn admin_update_status(
    State(state): State<AppState>,
    Path(id): Path<i64>,
    Json(req): Json<UpdateStatusRequest>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    state
        .ticket_svc
        .admin_update_status(id, &req.status)
        .await?;
    Ok(Json(ApiResponse::success_with_message(
        serde_json::json!({}),
        "状态已更新",
    )))
}
