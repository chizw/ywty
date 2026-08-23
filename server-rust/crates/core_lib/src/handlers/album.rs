//! 相册处理器

use axum::extract::{Path, Query, State};
use axum::Json;

use crate::error::AppResult;
use crate::utils::pagination::Pagination;
use crate::utils::response::ApiResponse;

use crate::dto::album::{
    AddPhotoToAlbumRequest, AlbumResponse, CreateAlbumRequest, UpdateAlbumRequest,
};
use crate::dto::photo::PhotoResponse;
use crate::dto::PaginatedData;
use crate::handlers::{validate_req, CurrentUser};
use crate::AppState;

/// 获取相册列表
#[utoipa::path(
    get,
    path = "/api/v1/albums",
    params(
        ("page" = u64, Query, description = "页码"),
        ("per_page" = u64, Query, description = "每页数量"),
    ),
    responses(
        (status = 200, description = "成功", body = PaginatedData<AlbumResponse>),
    ),
    tag = "相册"
)]
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
#[utoipa::path(
    get,
    path = "/api/v1/albums/:id",
    params(
        ("id" = i64, Path, description = "相册ID"),
    ),
    responses(
        (status = 200, description = "成功", body = ApiResponse<AlbumResponse>),
    ),
    tag = "相册"
)]
pub async fn get(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Path(album_id): Path<i64>,
) -> AppResult<Json<ApiResponse<AlbumResponse>>> {
    let album = state.album_svc.get(user_id, album_id).await?;
    Ok(Json(ApiResponse::success(album)))
}

/// 创建相册
#[utoipa::path(
    post,
    path = "/api/v1/albums",
    request_body = CreateAlbumRequest,
    responses(
        (status = 200, description = "成功", body = ApiResponse<AlbumResponse>),
    ),
    tag = "相册"
)]
pub async fn create(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Json(req): Json<CreateAlbumRequest>,
) -> AppResult<Json<ApiResponse<AlbumResponse>>> {
    validate_req(&req)?;
    let album = state.album_svc.create(user_id, &req).await?;
    Ok(Json(ApiResponse::success(album)))
}

/// 更新相册
#[utoipa::path(
    patch,
    path = "/api/v1/albums/:id",
    request_body = UpdateAlbumRequest,
    params(
        ("id" = i64, Path, description = "相册ID"),
    ),
    responses(
        (status = 200, description = "成功", body = ApiResponse<AlbumResponse>),
    ),
    tag = "相册"
)]
pub async fn update(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Path(album_id): Path<i64>,
    Json(req): Json<UpdateAlbumRequest>,
) -> AppResult<Json<ApiResponse<AlbumResponse>>> {
    validate_req(&req)?;
    let album = state.album_svc.update(user_id, album_id, &req).await?;
    Ok(Json(ApiResponse::success(album)))
}

/// 删除相册
#[utoipa::path(
    delete,
    path = "/api/v1/albums/:id",
    params(
        ("id" = i64, Path, description = "相册ID"),
    ),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "相册"
)]
pub async fn delete(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Path(album_id): Path<i64>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    state.album_svc.delete(user_id, album_id).await?;
    Ok(Json(ApiResponse::success_with_message(
        serde_json::json!({}),
        "删除成功",
    )))
}

/// 获取相册内的图片
#[utoipa::path(
    get,
    path = "/api/v1/albums/:id/photos",
    params(
        ("id" = i64, Path, description = "相册ID"),
        ("page" = u64, Query, description = "页码"),
        ("per_page" = u64, Query, description = "每页数量"),
    ),
    responses(
        (status = 200, description = "成功", body = PaginatedData<PhotoResponse>),
    ),
    tag = "相册"
)]
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
#[utoipa::path(
    post,
    path = "/api/v1/albums/:id/photos",
    request_body = AddPhotoToAlbumRequest,
    params(
        ("id" = i64, Path, description = "相册ID"),
    ),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "相册"
)]
pub async fn add_photos(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Path(album_id): Path<i64>,
    Json(req): Json<AddPhotoToAlbumRequest>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    let added = state
        .album_svc
        .add_photos(user_id, album_id, &req.photo_ids)
        .await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "added": added }),
    )))
}

/// 从相册移除图片
/// 注意：axum 不支持多个 `Path` 参数，多参数必须用元组 `Path<(A, B)>` 按位置提取
#[utoipa::path(
    delete,
    path = "/api/v1/albums/:id/photos/:photo_id",
    params(
        ("id" = i64, Path, description = "相册ID"),
        ("photo_id" = i64, Path, description = "图片ID"),
    ),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "相册"
)]
pub async fn remove_photo(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Path((album_id, photo_id)): Path<(i64, i64)>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    state
        .album_svc
        .remove_photo(user_id, album_id, photo_id)
        .await?;
    Ok(Json(ApiResponse::success_with_message(
        serde_json::json!({}),
        "移除成功",
    )))
}
