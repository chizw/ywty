//! 存储处理器

use axum::extract::{Query, State};
use axum::Json;
use serde::Deserialize;

use crate::error::AppResult;
use crate::handlers::CurrentUser;
use crate::utils::response::ApiResponse;
use crate::AppState;

/// 直传签名查询参数
#[derive(Debug, Deserialize, utoipa::IntoParams)]
pub struct SignQuery {
    /// 直传 object key（不传时由服务端生成）
    pub key: Option<String>,
}

/// 获取上传签名（按当前生效存储策略返回直传参数）
#[utoipa::path(
    get,
    path = "/api/v1/storage/sign",
    params(SignQuery),
    responses(
        (status = 200, description = "成功", body = crate::dto::storage::StorageSignResponse),
    ),
    tag = "存储"
)]
pub async fn sign(
    State(state): State<AppState>,
    CurrentUser { .. }: CurrentUser,
    Query(q): Query<SignQuery>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    let (base, extra) = state
        .storage_svc
        .sign(&state.config.storage, q.key.as_deref())
        .await?;

    // 合并基础信息与驱动专属直传参数（policy/signature/token/method/key 等）
    let mut merged = serde_json::to_value(&base)
        .map_err(|e| crate::error::AppError::Internal(format!("序列化签名响应失败: {}", e)))?;
    if let (Some(obj), Some(ext)) = (merged.as_object_mut(), extra.as_object()) {
        for (k, v) in ext {
            obj.insert(k.clone(), v.clone());
        }
    }

    Ok(Json(ApiResponse::success(merged)))
}
