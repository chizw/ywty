//! 图片处理器

use axum::extract::{Path, Query, State};
use axum::Json;

use crate::error::AppResult;
use crate::utils::pagination::Pagination;

use crate::dto::photo::{
    BatchIdsRequest, BatchUpdateRequest, CopyPhotoRequest, MoveToAlbumRequest,
    PhotoPublicResponse, PhotoResponse,
};
use crate::dto::PaginatedData;
use crate::handlers::CurrentUser;
use crate::AppState;

/// 获取图片列表
pub async fn list(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Query(pagination): Query<Pagination>,
) -> AppResult<Json<PaginatedData<PhotoResponse>>> {
    let (rows, total) = state
        .photo_svc
        .list(
            user_id,
            pagination.page,
            pagination.per_page,
            pagination.album_id,
        )
        .await?;
    Ok(Json(PaginatedData::new(
        rows,
        total as u64,
        pagination.page,
        pagination.per_page,
    )))
}

/// 获取单张图片
pub async fn get(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Path(photo_id): Path<i64>,
) -> AppResult<Json<PhotoResponse>> {
    let photo = state.photo_svc.get(user_id, photo_id).await?;
    Ok(Json(photo))
}

/// 上传图片（简化版：接收 JSON 元数据）
pub async fn upload(
    State(_state): State<AppState>,
    CurrentUser { .. }: CurrentUser,
    Json(req): Json<crate::models::photo::UploadPhotoRequest>,
) -> AppResult<Json<serde_json::Value>> {
    // TODO: 完整 multipart 文件上传
    Ok(Json(serde_json::json!({
        "message": "请使用 /storage/sign 获取上传配置",
        "album_id": req.album_id,
        "is_public": req.is_public,
    })))
}

/// 更新图片
pub async fn update(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Path(photo_id): Path<i64>,
    Json(req): Json<crate::models::photo::UpdatePhotoRequest>,
) -> AppResult<Json<PhotoResponse>> {
    let photo = state
        .photo_svc
        .update(user_id, photo_id, req.album_id, req.is_public)
        .await?;
    Ok(Json(photo))
}

/// 删除图片
pub async fn delete(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Path(photo_id): Path<i64>,
) -> AppResult<Json<serde_json::Value>> {
    state.photo_svc.delete(user_id, photo_id).await?;
    Ok(Json(serde_json::json!({ "message": "删除成功" })))
}

/// 批量删除
pub async fn batch_delete(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Json(req): Json<BatchIdsRequest>,
) -> AppResult<Json<serde_json::Value>> {
    let affected = state.photo_svc.batch_delete(user_id, &req.ids).await?;
    Ok(Json(serde_json::json!({ "affected": affected })))
}

/// 批量更新
pub async fn batch_update(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Json(req): Json<BatchUpdateRequest>,
) -> AppResult<Json<serde_json::Value>> {
    let affected = state.photo_svc.batch_update(user_id, &req).await?;
    Ok(Json(serde_json::json!({ "affected": affected })))
}

/// 移动到相册
pub async fn move_to_album(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Path(photo_id): Path<i64>,
    Json(req): Json<MoveToAlbumRequest>,
) -> AppResult<Json<PhotoResponse>> {
    let photo = state
        .photo_svc
        .move_to_album(user_id, photo_id, req.album_id)
        .await?;
    Ok(Json(photo))
}

/// 复制图片
pub async fn copy(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Path(photo_id): Path<i64>,
    Json(req): Json<CopyPhotoRequest>,
) -> AppResult<Json<crate::dto::photo::UploadResponse>> {
    let photo = state.photo_svc.copy(user_id, photo_id, req.album_id).await?;
    Ok(Json(photo))
}

/// 公开探索列表
pub async fn list_public(
    State(state): State<AppState>,
    Query(pagination): Query<Pagination>,
) -> AppResult<Json<PaginatedData<PhotoPublicResponse>>> {
    let (rows, total) = state
        .photo_svc
        .list_public(pagination.page, pagination.per_page)
        .await?;
    Ok(Json(PaginatedData::new(
        rows,
        total as u64,
        pagination.page,
        pagination.per_page,
    )))
}
