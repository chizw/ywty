//! 标签处理器

use axum::extract::{Path, State};
use axum::Json;

use crate::error::AppResult;
use crate::handlers::CurrentUser;
use crate::utils::response::ApiResponse;
use crate::AppState;

/// 列出标签（或按图片查询标签）
#[utoipa::path(
    get,
    path = "/api/v1/tags",
    params(
        ("target_type" = String, Query, description = "目标类型"),
        ("target_id" = i64, Query, description = "目标ID"),
    ),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "标签"
)]
pub async fn list(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    axum::extract::Query(params): axum::extract::Query<std::collections::HashMap<String, String>>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    if let (Some(target_type), Some(target_id)) =
        (params.get("target_type"), params.get("target_id"))
    {
        let target_id: i64 = target_id.parse().unwrap_or(0);
        if target_type == "photo" && target_id > 0 {
            let tags = state.tag_svc.list_for_target(user_id, target_id).await?;
            return Ok(Json(ApiResponse::success(
                serde_json::json!({ "data": tags }),
            )));
        }
    }

    let tags = state.tag_svc.list().await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": tags }),
    )))
}

/// 创建标签
#[utoipa::path(
    post,
    path = "/api/v1/tags",
    request_body = serde_json::Value,
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "标签"
)]
pub async fn create(
    State(state): State<AppState>,
    CurrentUser { .. }: CurrentUser,
    Json(req): Json<serde_json::Value>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    let name = req
        .get("name")
        .and_then(|v| v.as_str())
        .unwrap_or("")
        .trim();

    if name.is_empty() {
        return Err(crate::error::AppError::Validation(
            "标签名不能为空".to_string(),
        ));
    }

    let tag = state.tag_svc.create(name).await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": tag }),
    )))
}

/// 删除标签
#[utoipa::path(
    delete,
    path = "/api/v1/tags/:id",
    params(
        ("id" = i64, Path, description = "标签ID"),
    ),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "标签"
)]
pub async fn delete(
    State(state): State<AppState>,
    CurrentUser { .. }: CurrentUser,
    Path(id): Path<i64>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    state.tag_svc.delete(id).await?;
    Ok(Json(ApiResponse::success_with_message(
        serde_json::json!({}),
        "删除成功",
    )))
}

/// 绑定标签到图片
#[utoipa::path(
    post,
    path = "/api/v1/tags/attach",
    request_body = serde_json::Value,
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "标签"
)]
pub async fn attach(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Json(req): Json<serde_json::Value>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    let _target_type = req
        .get("target_type")
        .and_then(|v| v.as_str())
        .unwrap_or("photo");
    let target_id = req.get("target_id").and_then(|v| v.as_i64()).unwrap_or(0);

    if target_id == 0 {
        return Err(crate::error::AppError::Validation(
            "target_id 无效".to_string(),
        ));
    }

    // 支持按名称批量绑定或按 tag_id 绑定
    if let Some(names) = req.get("names").and_then(|v| v.as_array()) {
        let names: Vec<String> = names
            .iter()
            .filter_map(|v| v.as_str().map(|s| s.to_string()))
            .collect();
        let tags = state
            .tag_svc
            .attach_by_names(user_id, &names, target_id)
            .await?;
        Ok(Json(ApiResponse::success(
            serde_json::json!({ "data": tags }),
        )))
    } else if let Some(tag_id) = req.get("tag_id").and_then(|v| v.as_i64()) {
        state.tag_svc.attach(user_id, tag_id, target_id).await?;
        Ok(Json(ApiResponse::success_with_message(
            serde_json::json!({}),
            "绑定成功",
        )))
    } else {
        Err(crate::error::AppError::Validation(
            "names 或 tag_id 必填".to_string(),
        ))
    }
}

/// 解绑标签
#[utoipa::path(
    post,
    path = "/api/v1/tags/detach",
    request_body = serde_json::Value,
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "标签"
)]
pub async fn detach(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Json(req): Json<serde_json::Value>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    let tag_id = req.get("tag_id").and_then(|v| v.as_i64()).unwrap_or(0);
    let _target_type = req
        .get("target_type")
        .and_then(|v| v.as_str())
        .unwrap_or("photo");
    let target_id = req.get("target_id").and_then(|v| v.as_i64()).unwrap_or(0);

    if tag_id == 0 || target_id == 0 {
        return Err(crate::error::AppError::Validation(
            "tag_id 和 target_id 必填".to_string(),
        ));
    }

    state.tag_svc.detach(user_id, tag_id, target_id).await?;
    Ok(Json(ApiResponse::success_with_message(
        serde_json::json!({}),
        "解绑成功",
    )))
}
