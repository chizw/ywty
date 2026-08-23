//! 点赞/举报处理器

use axum::extract::{ConnectInfo, Path, State};
use axum::Json;
use std::net::SocketAddr;

use crate::error::AppResult;
use crate::handlers::CurrentUser;
use crate::utils::{client_ip, response::ApiResponse};
use crate::AppState;

/// 点赞/取消（toggle）
#[utoipa::path(
    post,
    path = "/api/v1/likes",
    request_body = serde_json::Value,
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "点赞/举报"
)]
pub async fn toggle(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Json(req): Json<serde_json::Value>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    let target_type = req
        .get("target_type")
        .and_then(|v| v.as_str())
        .unwrap_or("");
    let target_id = req.get("target_id").and_then(|v| v.as_i64()).unwrap_or(0);

    if target_type.is_empty() || target_id == 0 {
        return Err(crate::error::AppError::Validation(
            "target_type 和 target_id 必填".to_string(),
        ));
    }

    let liked = state
        .like_svc
        .toggle(user_id, target_type, target_id)
        .await?;
    let count = state.like_svc.count(target_type, target_id).await;

    Ok(Json(ApiResponse::success(serde_json::json!({
        "data": {
            "liked": liked,
            "count": count
        }
    }))))
}

/// 查询点赞状态
#[utoipa::path(
    get,
    path = "/api/v1/likes",
    params(
        ("target_type" = String, Query, description = "目标类型"),
        ("target_id" = i64, Query, description = "目标ID"),
    ),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "点赞/举报"
)]
pub async fn status(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    axum::extract::Query(params): axum::extract::Query<std::collections::HashMap<String, String>>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    let target_type = params.get("target_type").map(|s| s.as_str()).unwrap_or("");
    let target_id: i64 = params
        .get("target_id")
        .and_then(|s| s.parse().ok())
        .unwrap_or(0);

    let count = state.like_svc.count(target_type, target_id).await;
    let liked = if user_id > 0 {
        state.like_svc.liked(user_id, target_type, target_id).await
    } else {
        false
    };

    Ok(Json(ApiResponse::success(serde_json::json!({
        "data": {
            "liked": liked,
            "count": count
        }
    }))))
}

/// 创建举报
#[utoipa::path(
    post,
    path = "/api/v1/reports",
    request_body = serde_json::Value,
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "点赞/举报"
)]
pub async fn create_report(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    ConnectInfo(addr): ConnectInfo<SocketAddr>,
    headers: axum::http::HeaderMap,
    Json(req): Json<serde_json::Value>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    let target_type = req
        .get("target_type")
        .and_then(|v| v.as_str())
        .unwrap_or("");
    let target_id = req.get("target_id").and_then(|v| v.as_i64()).unwrap_or(0);
    let content = req.get("content").and_then(|v| v.as_str()).unwrap_or("");

    if target_type.is_empty() || target_id == 0 {
        return Err(crate::error::AppError::Validation(
            "target_type 和 target_id 必填".to_string(),
        ));
    }

    // 获取客户端 IP（优先代理头，回退连接地址）
    let ip = client_ip(&headers, Some(addr));

    let report = state
        .report_svc
        .create(user_id, &ip, target_type, target_id, content)
        .await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": report }),
    )))
}

/// 后台列表举报
#[utoipa::path(
    get,
    path = "/api/v1/admin/reports",
    params(
        ("page" = u64, Query, description = "页码"),
        ("per_page" = u64, Query, description = "每页数量"),
        ("status" = Option<i32>, Query, description = "状态"),
    ),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "点赞/举报"
)]
pub async fn admin_list_reports(
    State(state): State<AppState>,
    CurrentUser { .. }: CurrentUser,
    axum::extract::Query(params): axum::extract::Query<std::collections::HashMap<String, String>>,
) -> AppResult<Json<serde_json::Value>> {
    let page: i64 = params.get("page").and_then(|s| s.parse().ok()).unwrap_or(1);
    let per_page: i64 = params
        .get("per_page")
        .and_then(|s| s.parse().ok())
        .unwrap_or(20);
    let status: Option<i32> = params.get("status").and_then(|s| s.parse().ok());

    let (rows, total) = state.report_svc.admin_list(page, per_page, status).await?;

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

/// 更新举报状态
#[utoipa::path(
    patch,
    path = "/api/v1/admin/reports/:id",
    params(
        ("id" = i64, Path, description = "ID"),
    ),
    request_body = serde_json::Value,
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "点赞/举报"
)]
pub async fn admin_update_report_status(
    State(state): State<AppState>,
    CurrentUser { .. }: CurrentUser,
    Path(id): Path<i64>,
    Json(req): Json<serde_json::Value>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    let status = req.get("status").and_then(|v| v.as_i64()).unwrap_or(0) as i32;

    state.report_svc.update_status(id, status as i64).await?;
    Ok(Json(ApiResponse::success_with_message(
        serde_json::json!({}),
        "更新成功",
    )))
}
