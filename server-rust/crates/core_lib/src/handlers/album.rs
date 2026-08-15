//! 相册处理器

use axum::extract::{Path, Query, State};
use axum::Json;

use crate::error::AppResult;
use crate::utils::pagination::Pagination;

use crate::dto::album::{
    AddPhotoToAlbumRequest, AlbumResponse, CreateAlbumRequest, UpdateAlbumRequest,
};
use crate::dto::photo::PhotoResponse;
use crate::dto::PaginatedData;
use crate::handlers::{validate_req, CurrentUser};
use crate::AppState;

/// 获取相册列表
pub async fn list(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Query(pagination): Query<Pagination>,
) -> AppResult<Json<PaginatedData<AlbumResponse>>> {
    let (rows, total) = state
        .album_svc
        .list(user_id, pagination.page, pagination.per_page)
        .await?;
    Ok(Json(PaginatedData::new(
        rows,
        total as u64,
        pagination.page,
        pagination.per_page,
    )))
}

/// 获取相册详情
pub async fn get(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Path(album_id): Path<i64>,
) -> AppResult<Json<AlbumResponse>> {
    let album = state.album_svc.get(user_id, album_id).await?;
    Ok(Json(album))
}

/// 创建相册
pub async fn create(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Json(req): Json<CreateAlbumRequest>,
) -> AppResult<Json<AlbumResponse>> {
    validate_req(&req)?;
    let album = state.album_svc.create(user_id, &req).await?;
    Ok(Json(album))
}

/// 更新相册
pub async fn update(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Path(album_id): Path<i64>,
    Json(req): Json<UpdateAlbumRequest>,
) -> AppResult<Json<AlbumResponse>> {
    validate_req(&req)?;
    let album = state.album_svc.update(user_id, album_id, &req).await?;
    Ok(Json(album))
}

/// 删除相册
pub async fn delete(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Path(album_id): Path<i64>,
) -> AppResult<Json<serde_json::Value>> {
    state.album_svc.delete(user_id, album_id).await?;
    Ok(Json(serde_json::json!({ "message": "删除成功" })))
}

/// 获取相册内的图片
pub async fn list_photos(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Path(album_id): Path<i64>,
    Query(pagination): Query<Pagination>,
) -> AppResult<Json<PaginatedData<PhotoResponse>>> {
    let (rows, total) = state
        .album_svc
        .list_photos(user_id, album_id, pagination.page, pagination.per_page)
        .await?;
    Ok(Json(PaginatedData::new(
        rows,
        total as u64,
        pagination.page,
        pagination.per_page,
    )))
}

/// 添加图片到相册
pub async fn add_photos(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Path(album_id): Path<i64>,
    Json(req): Json<AddPhotoToAlbumRequest>,
) -> AppResult<Json<serde_json::Value>> {
    let added = state
        .album_svc
        .add_photos(user_id, album_id, &req.photo_ids)
        .await?;
    Ok(Json(serde_json::json!({ "added": added })))
}

/// 从相册移除图片
/// 注意：axum 不支持多个 `Path` 参数，多参数必须用元组 `Path<(A, B)>` 按位置提取
pub async fn remove_photo(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Path((album_id, photo_id)): Path<(i64, i64)>,
) -> AppResult<Json<serde_json::Value>> {
    state
        .album_svc
        .remove_photo(user_id, album_id, photo_id)
        .await?;
    Ok(Json(serde_json::json!({ "message": "移除成功" })))
}
