//! 存储处理器

use axum::extract::State;
use axum::Json;

use crate::error::AppResult;

use crate::dto::storage::StorageSignResponse;
use crate::handlers::CurrentUser;
use crate::AppState;

/// 获取上传签名
pub async fn sign(
    State(state): State<AppState>,
    CurrentUser { .. }: CurrentUser,
) -> AppResult<Json<StorageSignResponse>> {
    let sign = state.storage_svc.sign().await?;
    Ok(Json(sign))
}
