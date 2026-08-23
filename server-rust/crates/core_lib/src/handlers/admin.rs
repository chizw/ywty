//! 管理员处理器

use axum::extract::{Path, Query, State};
use axum::Json;

use crate::error::{AppError, AppResult};
use crate::handlers::AdminUser;
use crate::utils::response::ApiResponse;
use crate::AppState;

#[derive(Debug, serde::Deserialize)]
pub struct UserQuery {
    pub page: Option<i32>,
    pub per_page: Option<i32>,
    pub keyword: Option<String>,
}

#[derive(Debug, serde::Deserialize)]
pub struct PhotoQuery {
    pub page: Option<i32>,
    pub per_page: Option<i32>,
    pub keyword: Option<String>,
}

/// 仪表盘统计
#[utoipa::path(
    get,
    path = "/api/v1/admin/stats",
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "管理后台"
)]
pub async fn stats(
    State(state): State<AppState>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    let stats = state.admin_svc.stats().await?;
    Ok(Json(ApiResponse::success(stats)))
}

/// 列出用户（管理端）
#[utoipa::path(
    get,
    path = "/api/v1/admin/users",
    params(
        ("page" = Option<i32>, Query, description = "页码"),
        ("per_page" = Option<i32>, Query, description = "每页数量"),
        ("keyword" = Option<String>, Query, description = "搜索关键词"),
    ),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "管理后台"
)]
pub async fn list_users(
    State(state): State<AppState>,
    Query(query): Query<UserQuery>,
) -> AppResult<Json<serde_json::Value>> {
    let (users, total) = state
        .admin_svc
        .list_users(
            query.page.unwrap_or(1),
            query.per_page.unwrap_or(20),
            query.keyword.as_deref(),
        )
        .await?;
    Ok(Json(serde_json::json!({
        "data": users,
        "meta": {
            "current_page": query.page.unwrap_or(1),
            "per_page": query.per_page.unwrap_or(20),
            "total": total,
            "last_page": (total + query.per_page.unwrap_or(20) as i64 - 1) / query.per_page.unwrap_or(20) as i64,
        }
    })))
}

/// 更新用户
#[utoipa::path(
    patch,
    path = "/api/v1/admin/users/:id",
    params(
        ("id" = i64, Path, description = "用户 ID"),
    ),
    request_body = serde_json::Value,
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "管理后台"
)]
pub async fn update_user(
    State(state): State<AppState>,
    admin: AdminUser,
    Path(id): Path<i64>,
    Json(req): Json<serde_json::Value>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    // 不允许操作自己（防自降级/自禁导致唯一超管被锁）
    if admin.user_id == id {
        return Err(AppError::Business("不能修改自己的角色或状态".to_string()));
    }
    // 目标用户元信息
    let target = state.admin_svc.get_user_meta(id).await?;
    let (target_role, target_super) =
        target.ok_or_else(|| AppError::NotFound("用户不存在".to_string()))?;
    // 超级管理员账号不可通过接口修改（仅由迁移/系统设定）
    if target_super {
        return Err(AppError::Business("超级管理员账号不可修改".to_string()));
    }

    let is_admin_req = req.get("is_admin").and_then(|v| v.as_bool());
    let status_req = req.get("status").and_then(|v| v.as_i64());
    let nickname_req = req
        .get("nickname")
        .and_then(|v| v.as_str())
        .map(|s| s.to_string());

    // 配额覆盖：字段缺省 = 不修改；null = 清除覆盖（跟随角色组）；数字 = 设置（字节）
    let quota_req: Option<Option<i64>> = if req.get("quota_override").is_some() {
        Some(req.get("quota_override").and_then(|v| v.as_i64()))
    } else {
        None
    };

    // 权限：只有超级管理员能改角色（提权/降权）；普通管理员只能改普通用户的启用状态/昵称
    let role_update = if admin.is_super_admin {
        is_admin_req.map(|ia| {
            if ia {
                "admin".to_string()
            } else {
                "user".to_string()
            }
        })
    } else {
        if is_admin_req.is_some() {
            return Err(AppError::Business("无权修改管理员角色".to_string()));
        }
        if target_role == "admin" {
            return Err(AppError::Business("无权操作其他管理员".to_string()));
        }
        None
    };

    let has_other_update = role_update.is_some() || status_req.is_some() || nickname_req.is_some();
    if has_other_update {
        state
            .admin_svc
            .update_user(id, role_update, status_req.map(|s| s as i32), nickname_req)
            .await?;
    } else if quota_req.is_none() {
        return Err(AppError::Validation("没有要更新的字段".to_string()));
    }

    // 配额覆盖单独落库（services/admin.rs 不感知该字段）
    if let Some(quota) = quota_req {
        let now = chrono::Utc::now().to_rfc3339();
        let result = sqlx::query(
            "UPDATE users SET quota_override = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL",
        )
        .bind(quota)
        .bind(&now)
        .bind(id)
        .execute(&state.db)
        .await?;
        if result.rows_affected() == 0 && !has_other_update {
            return Err(AppError::NotFound("用户不存在".to_string()));
        }
    }
    Ok(Json(ApiResponse::success(serde_json::json!({
        "data": { "id": id, "updated": true }
    }))))
}

/// 删除用户（管理端）
/// - 不能删除自己
/// - 不能删除超级管理员
/// - 普通管理员只能删除普通用户；超级管理员可删除任何人（除自己/超管）
#[utoipa::path(
    delete,
    path = "/api/v1/admin/users/:id",
    params(("id" = i64, Path, description = "用户ID")),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "管理后台"
)]
pub async fn delete_user(
    State(state): State<AppState>,
    admin: AdminUser,
    Path(id): Path<i64>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    if admin.user_id == id {
        return Err(AppError::Business("不能删除自己的账号".to_string()));
    }
    let target = state.admin_svc.get_user_meta(id).await?;
    let (target_role, target_super) =
        target.ok_or_else(|| AppError::NotFound("用户不存在".to_string()))?;
    if target_super {
        return Err(AppError::Business("不能删除超级管理员账号".to_string()));
    }
    if !admin.is_super_admin && target_role == "admin" {
        return Err(AppError::Business("无权删除其他管理员".to_string()));
    }
    state.admin_svc.delete_user(id).await?;
    Ok(Json(ApiResponse::success(serde_json::json!({
        "data": { "id": id, "deleted": true }
    }))))
}

/// 列出所有图片（管理端）
#[utoipa::path(
    get,
    path = "/api/v1/admin/photos",
    params(
        ("page" = Option<i32>, Query, description = "页码"),
        ("per_page" = Option<i32>, Query, description = "每页数量"),
        ("keyword" = Option<String>, Query, description = "搜索关键词"),
    ),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "管理后台"
)]
pub async fn list_all_photos(
    State(state): State<AppState>,
    Query(query): Query<PhotoQuery>,
) -> AppResult<Json<serde_json::Value>> {
    let (photos, total) = state
        .admin_svc
        .list_all_photos(
            query.page.unwrap_or(1),
            query.per_page.unwrap_or(24),
            query.keyword.as_deref(),
        )
        .await?;
    Ok(Json(serde_json::json!({
        "data": photos,
        "meta": {
            "current_page": query.page.unwrap_or(1),
            "per_page": query.per_page.unwrap_or(24),
            "total": total,
            "last_page": (total + query.per_page.unwrap_or(24) as i64 - 1) / query.per_page.unwrap_or(24) as i64,
        }
    })))
}

/// RBAC 策略列表（占位）
#[utoipa::path(
    get,
    path = "/api/v1/admin/rbac/policies",
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "管理后台"
)]
pub async fn list_rbac_policies() -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": [] }),
    )))
}

/// 添加 RBAC 策略（占位）
#[utoipa::path(
    post,
    path = "/api/v1/admin/rbac/policies",
    request_body = serde_json::Value,
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "管理后台"
)]
pub async fn add_rbac_policy(
    Json(req): Json<serde_json::Value>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": req }),
    )))
}

/// 删除 RBAC 策略（占位）
#[utoipa::path(
    post,
    path = "/api/v1/admin/rbac/policies/delete",
    request_body = serde_json::Value,
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "管理后台"
)]
pub async fn delete_rbac_policy(
    Json(req): Json<serde_json::Value>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": req }),
    )))
}

/// RBAC 角色列表（占位）
#[utoipa::path(
    get,
    path = "/api/v1/admin/rbac/roles",
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "管理后台"
)]
pub async fn list_rbac_roles() -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": ["admin", "user"] }),
    )))
}

/// 分配 RBAC 角色（占位）
#[utoipa::path(
    post,
    path = "/api/v1/admin/rbac/roles",
    request_body = serde_json::Value,
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "管理后台"
)]
pub async fn assign_rbac_role(
    Json(req): Json<serde_json::Value>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "data": req }),
    )))
}
