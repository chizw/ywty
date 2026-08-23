//! 意见反馈 + 违规记录处理器

use axum::extract::{ConnectInfo, Path, State};
use axum::Json;
use std::net::SocketAddr;

use crate::error::AppResult;
use crate::handlers::{validate_req, CurrentUser};
use crate::models::feedback::{CreateFeedbackRequest, CreateViolationRequest};
use crate::utils::{client_ip, response::ApiResponse};
use crate::AppState;

// ==================== 反馈 ====================

/// 创建反馈（公开）
#[utoipa::path(
    post,
    path = "/api/v1/feedback",
    request_body = CreateFeedbackRequest,
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "反馈/违规"
)]
pub async fn create_feedback(
    State(state): State<AppState>,
    ConnectInfo(addr): ConnectInfo<SocketAddr>,
    headers: axum::http::HeaderMap,
    Json(req): Json<CreateFeedbackRequest>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    validate_req(&req)?;
    let ip = client_ip(&headers, Some(addr));
    let feedback = state.feedback_svc.create(&ip, &req).await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": feedback }),
    )))
}

/// 后台列表反馈
#[utoipa::path(
    get,
    path = "/api/v1/admin/feedbacks",
    params(
        ("page" = u64, Query, description = "页码"),
        ("per_page" = u64, Query, description = "每页数量"),
    ),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "反馈/违规"
)]
pub async fn admin_list_feedbacks(
    State(state): State<AppState>,
    CurrentUser { .. }: CurrentUser,
    axum::extract::Query(params): axum::extract::Query<std::collections::HashMap<String, String>>,
) -> AppResult<Json<serde_json::Value>> {
    let page: i64 = params.get("page").and_then(|s| s.parse().ok()).unwrap_or(1);
    let per_page: i64 = params
        .get("per_page")
        .and_then(|s| s.parse().ok())
        .unwrap_or(20);

    let (rows, total) = state.feedback_svc.list(page, per_page).await?;

    Ok(Json(serde_json::json!({
        "data": rows,
        "meta": {
            "current_page": page,
            "per_page": per_page,
            "total": total,
            "last_page": (total as f64 / per_page as f64).ceil() as i64
        }
    })))
}

/// 删除反馈
#[utoipa::path(
    delete,
    path = "/api/v1/admin/feedbacks/:id",
    params(
        ("id" = i64, Path, description = "ID"),
    ),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "反馈/违规"
)]
pub async fn admin_delete_feedback(
    State(state): State<AppState>,
    CurrentUser { .. }: CurrentUser,
    Path(id): Path<i64>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    state.feedback_svc.delete(id).await?;
    Ok(Json(ApiResponse::success_with_message(
        serde_json::json!({}),
        "删除成功",
    )))
}

// ==================== 违规记录 ====================

/// 创建违规记录
#[utoipa::path(
    post,
    path = "/api/v1/violations",
    request_body = CreateViolationRequest,
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "反馈/违规"
)]
pub async fn create_violation(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Json(req): Json<CreateViolationRequest>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    validate_req(&req)?;
    let violation = state.violation_svc.create(user_id, &req).await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": violation }),
    )))
}

/// 后台列表违规记录
#[utoipa::path(
    get,
    path = "/api/v1/admin/violations",
    params(
        ("page" = u64, Query, description = "页码"),
        ("per_page" = u64, Query, description = "每页数量"),
        ("status" = Option<String>, Query, description = "状态"),
    ),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "反馈/违规"
)]
pub async fn admin_list_violations(
    State(state): State<AppState>,
    CurrentUser { .. }: CurrentUser,
    axum::extract::Query(params): axum::extract::Query<std::collections::HashMap<String, String>>,
) -> AppResult<Json<serde_json::Value>> {
    let page: i64 = params.get("page").and_then(|s| s.parse().ok()).unwrap_or(1);
    let per_page: i64 = params
        .get("per_page")
        .and_then(|s| s.parse().ok())
        .unwrap_or(20);
    let status = params.get("status").map(|s| s.as_str());

    let (rows, total) = state.violation_svc.list(page, per_page, status).await?;

    Ok(Json(serde_json::json!({
        "data": rows,
        "meta": {
            "current_page": page,
            "per_page": per_page,
            "total": total,
            "last_page": (total as f64 / per_page as f64).ceil() as i64
        }
    })))
}

/// 更新违规记录状态
#[utoipa::path(
    patch,
    path = "/api/v1/admin/violations/:id",
    params(
        ("id" = i64, Path, description = "ID"),
    ),
    request_body = serde_json::Value,
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "反馈/违规"
)]
pub async fn admin_update_violation_status(
    State(state): State<AppState>,
    CurrentUser { .. }: CurrentUser,
    Path(id): Path<i64>,
    Json(req): Json<serde_json::Value>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    let status = req.get("status").and_then(|v| v.as_str()).unwrap_or("");

    if status.is_empty() {
        return Err(crate::error::AppError::Validation(
            "status 必填".to_string(),
        ));
    }

    state.violation_svc.update_status(id, status).await?;
    Ok(Json(ApiResponse::success_with_message(
        serde_json::json!({}),
        "更新成功",
    )))
}
