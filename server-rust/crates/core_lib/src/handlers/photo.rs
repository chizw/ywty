//! 图片处理器

use axum::extract::{Multipart, Path, Query, State};
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

/// 上传图片（multipart/form-data）
///
/// 接收文件字段 `file`，可选字段 `album_id`、`is_public`。
/// 保存文件到本地存储，写入 photos 表，返回图片信息。
pub async fn upload(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    mut multipart: Multipart,
) -> AppResult<Json<crate::dto::photo::UploadResponse>> {
    use std::path::PathBuf;
    use tokio::fs;

    let config = &state.config;
    let upload_root = config
        .storage
        .root
        .clone()
        .unwrap_or_else(|| "./uploads".to_string());
    let public_url = config
        .storage
        .url
        .clone()
        .unwrap_or_else(|| format!("http://localhost:{}", config.app.port));

    // 解析 multipart 字段
    let mut file_data: Option<Vec<u8>> = None;
    let mut filename: Option<String> = None;
    let mut album_id: Option<i64> = None;
    let mut is_public: bool = false;

    while let Some(field) = multipart
        .next_field()
        .await
        .map_err(|e| crate::error::AppError::Validation(format!("解析上传数据失败: {}", e)))?
    {
        let field_name = field.name().unwrap_or("").to_string();

        match field_name.as_str() {
            "file" => {
                // 获取原始文件名
                filename = field
                    .file_name()
                    .map(|s| s.to_string())
                    .filter(|s| !s.is_empty());

                // 读取文件内容
                let data = field
                    .bytes()
                    .await
                    .map_err(|e| crate::error::AppError::Validation(format!("读取文件失败: {}", e)))?;

                // 限制文件大小（默认 10MB）
                let max_size = 10 * 1024 * 1024;
                if data.len() > max_size {
                    return Err(crate::error::AppError::Validation(
                        format!("文件过大: {} 字节，最大允许 {} 字节", data.len(), max_size)
                    ));
                }

                if data.is_empty() {
                    return Err(crate::error::AppError::Validation("文件为空".to_string()));
                }

                file_data = Some(data.to_vec());
            }
            "album_id" => {
                let value = field.text().await.unwrap_or_default();
                album_id = value.parse().ok();
            }
            "is_public" => {
                let value = field.text().await.unwrap_or_default();
                is_public = value == "true" || value == "1";
            }
            _ => {} // 忽略未知字段
        }
    }

    let file_data = file_data
        .ok_or_else(|| crate::error::AppError::Validation("缺少文件字段 'file'".to_string()))?;
    let original_name = filename.unwrap_or_else(|| "unknown.bin".to_string());

    // 检测 MIME 类型
    let mime_type = infer::get(&file_data)
        .map(|t| t.mime_type().to_string())
        .unwrap_or_else(|| "application/octet-stream".to_string());

    // 只允许图片类型
    if !mime_type.starts_with("image/") {
        return Err(crate::error::AppError::Validation(
            format!("不支持的文件类型: {}，仅允许上传图片", mime_type)
        ));
    }

    // 计算 MD5
    let md5_hash = format!("{:x}", md5::compute(&file_data));

    // 检查重复上传（同一用户、同一 MD5）
    let existing: Option<(i64,)> = sqlx::query_as(
        "SELECT id FROM photos WHERE user_id = ? AND md5 = ? AND deleted_at IS NULL LIMIT 1"
    )
    .bind(user_id)
    .bind(&md5_hash)
    .fetch_optional(&state.db)
    .await?;

    if let Some((existing_id,)) = existing {
        // 返回已存在的图片
        let photo = state.photo_svc.get(user_id, existing_id).await?;
        return Ok(Json(crate::dto::photo::UploadResponse {
            id: photo.id,
            uuid: photo.uuid,
            url: photo.url,
            thumbnail_url: photo.thumbnail_url,
            size: photo.size,
            width: photo.width,
            height: photo.height,
            filename: photo.filename,
        }));
    }

    // 生成存储路径：uploads/2024/08/20/{uuid}.ext
    let file_uuid = uuid::Uuid::new_v4().to_string();
    let now = chrono::Utc::now();
    let ext = PathBuf::from(&original_name)
        .extension()
        .and_then(|e| e.to_str())
        .unwrap_or("bin")
        .to_lowercase();
    let date_path = now.format("%Y/%m/%d").to_string();
    let stored_filename = format!("{}.{}", file_uuid, ext);
    let relative_path = format!("{}/{}", date_path, stored_filename);
    let absolute_dir = PathBuf::from(&upload_root).join(&date_path);
    let absolute_path = absolute_dir.join(&stored_filename);

    // 创建目录
    fs::create_dir_all(&absolute_dir).await?;

    // 写入文件
    fs::write(&absolute_path, &file_data).await?;

    // 获取图片尺寸
    use std::io::Cursor;
    let (width, height) = match image::io::Reader::new(Cursor::new(&file_data)).with_guessed_format() {
        Ok(reader) => match reader.into_dimensions() {
            Ok((w, h)) => (Some(w as i32), Some(h as i32)),
            Err(_) => (None, None),
        },
        Err(_) => (None, None),
    };

    // 创建数据库记录
    let url = format!("{}/{}", public_url.trim_end_matches('/'), relative_path);
    let photo = state
        .photo_svc
        .create(
            user_id,
            &stored_filename,
            &original_name,
            &relative_path,
            file_data.len() as i64,
            &mime_type,
            width,
            height,
            Some(&md5_hash),
            album_id,
        )
        .await?;

    // 如果指定了相册，更新相册图片数
    if let Some(aid) = album_id {
        let _ = sqlx::query("UPDATE albums SET photo_count = photo_count + 1, updated_at = ? WHERE id = ?")
            .bind(now.to_rfc3339())
            .bind(aid)
            .execute(&state.db)
            .await;
    }

    Ok(Json(photo))
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
