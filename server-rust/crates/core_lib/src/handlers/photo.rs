//! 图片处理器

use axum::extract::{Multipart, Path, Query, State};
use axum::Json;

use crate::error::AppResult;
use crate::utils::pagination::Pagination;
use crate::utils::response::ApiResponse;

use crate::dto::photo::{
    BatchIdsRequest, BatchUpdateRequest, CopyPhotoRequest, MoveToAlbumRequest, PhotoPublicResponse,
    PhotoResponse, UploadResponse,
};
use crate::dto::PaginatedData;
use crate::handlers::CurrentUser;
use crate::AppState;

/// 获取图片列表
#[utoipa::path(
    get,
    path = "/api/v1/photos",
    params(
        ("page" = u64, Query, description = "页码"),
        ("per_page" = u64, Query, description = "每页数量"),
        ("album_id" = Option<i64>, Query, description = "相册ID筛选"),
    ),
    responses(
        (status = 200, description = "成功", body = PaginatedData<PhotoResponse>),
    ),
    tag = "图片"
)]
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
#[utoipa::path(
    get,
    path = "/api/v1/photos/:id",
    params(
        ("id" = i64, Path, description = "图片ID"),
    ),
    responses(
        (status = 200, description = "成功", body = PhotoResponse),
    ),
    tag = "图片"
)]
pub async fn get(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Path(photo_id): Path<i64>,
) -> AppResult<Json<ApiResponse<PhotoResponse>>> {
    let photo = state.photo_svc.get(user_id, photo_id).await?;
    Ok(Json(ApiResponse::success(photo)))
}

/// 上传图片（multipart/form-data）
///
/// 接收文件字段 `file`，可选字段 `album_id`、`is_public`。
/// 保存文件到本地存储，写入 photos 表，返回图片信息。
#[utoipa::path(
    post,
    path = "/api/v1/photos",
    request_body(content = Vec<u8>, content_type = "multipart/form-data", description = "文件上传 (field: file, album_id, is_public)"),
    responses(
        (status = 200, description = "成功", body = UploadResponse),
    ),
    tag = "图片"
)]
pub async fn upload(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    mut multipart: Multipart,
) -> AppResult<Json<ApiResponse<crate::dto::photo::UploadResponse>>> {
    let config = &state.config;
    let upload_root = config
        .storage
        .root
        .clone()
        .unwrap_or_else(|| "./uploads".to_string());
    let public_url = config.storage_public_url();

    // 当前生效的存储驱动（由 config.storage 在启动时构建）
    let driver = state.storage_driver.clone();
    let is_local_driver = driver.name() == "local";

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
                let data = field.bytes().await.map_err(|e| {
                    crate::error::AppError::Validation(format!("读取文件失败: {}", e))
                })?;

                // 限制文件大小（默认 10MB）
                let max_size = 10 * 1024 * 1024;
                if data.len() > max_size {
                    return Err(crate::error::AppError::Validation(format!(
                        "文件过大: {} 字节，最大允许 {} 字节",
                        data.len(),
                        max_size
                    )));
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
        return Err(crate::error::AppError::Validation(format!(
            "不支持的文件类型: {}，仅允许上传图片",
            mime_type
        )));
    }

    // 计算 MD5
    let md5_hash = format!("{:x}", md5::compute(&file_data));

    // 检查重复上传（同一用户、同一 MD5）
    let existing: Option<(i64,)> = sqlx::query_as(
        "SELECT id FROM photos WHERE user_id = ? AND md5 = ? AND deleted_at IS NULL LIMIT 1",
    )
    .bind(user_id)
    .bind(&md5_hash)
    .fetch_optional(&state.db)
    .await?;

    if let Some((existing_id,)) = existing {
        // 返回已存在的图片
        let photo = state.photo_svc.get(user_id, existing_id).await?;
        return Ok(Json(ApiResponse::success(
            crate::dto::photo::UploadResponse {
                id: photo.id,
                uuid: photo.uuid,
                url: photo.url,
                thumbnail_url: photo.thumbnail_url,
                size: photo.size,
                width: photo.width,
                height: photo.height,
                filename: photo.filename,
            },
        )));
    }

    // 存储配额校验（用户 override 优先，否则角色组 max_storage；无配额则不限）
    let capacity_svc = crate::services::capacity::CapacityService::new(state.db.clone());
    if let Some(limit) = capacity_svc.effective_limit_bytes(user_id).await? {
        let used = capacity_svc.used_bytes(user_id).await?;
        let remain = limit - used;
        let incoming = file_data.len() as i64;
        if incoming > remain {
            return Err(crate::error::AppError::Validation(format!(
                "存储空间不足：剩余 {} 字节，本文件 {} 字节（总限额 {} 字节，已用 {} 字节）",
                remain.max(0),
                incoming,
                limit,
                used
            )));
        }
    }

    // 生成存储路径：{yyyy/MM/dd}/{uuid}.{ext}（本地与远端驱动统一使用相对 key）
    let file_uuid = uuid::Uuid::new_v4().to_string();
    let now = chrono::Utc::now();
    let ext = std::path::PathBuf::from(&original_name)
        .extension()
        .and_then(|e| e.to_str())
        .unwrap_or("bin")
        .to_lowercase();
    let date_path = now.format("%Y/%m/%d").to_string();
    let stored_filename = format!("{}.{}", file_uuid, ext);
    let relative_path = format!("{}/{}", date_path, stored_filename);

    // 通过存储驱动写入（local 写本地磁盘；s3/oss/cos/qiniu 写远端并返回公网 URL）
    let upload_result = driver
        .upload(
            &relative_path,
            bytes::Bytes::from(file_data.clone()),
            &mime_type,
        )
        .await?;

    // 获取图片尺寸
    use std::io::Cursor;
    let (width, height) =
        match image::ImageReader::new(Cursor::new(&file_data)).with_guessed_format() {
            Ok(reader) => match reader.into_dimensions() {
                Ok((w, h)) => (Some(w as i32), Some(h as i32)),
                Err(_) => (None, None),
            },
            Err(_) => (None, None),
        };

    // URL 生成按驱动区分：远端用驱动返回的公网 URL，本地沿用 uploads 前缀
    let url = if is_local_driver {
        format!("{}/{}", public_url.trim_end_matches('/'), relative_path)
    } else {
        upload_result.url.clone()
    };

    // 创建数据库记录（缩略图初始为 None，由异步队列生成）
    let thumbnail_url: Option<String> = None;
    let photo = state
        .photo_svc
        .create_with_url(
            user_id,
            &stored_filename,
            &original_name,
            &relative_path,
            &url,
            file_data.len() as i64,
            &mime_type,
            width,
            height,
            Some(&md5_hash),
            album_id,
            thumbnail_url.as_deref(),
            is_public,
        )
        .await?;

    // 如果指定了相册，更新相册图片数
    if let Some(aid) = album_id {
        let _ = sqlx::query(
            "UPDATE albums SET photo_count = photo_count + 1, updated_at = ? WHERE id = ?",
        )
        .bind(crate::db::now_str())
        .bind(aid)
        .execute(&state.db)
        .await;
    }

    // 异步生成缩略图（投递到队列，后台 worker 处理；产物随主图走同一驱动）
    if let (Some(w), Some(h)) = (width, height) {
        if let Some(ref queue) = state.queue {
            let remote_driver = if is_local_driver {
                None
            } else {
                Some(driver.clone())
            };
            let task = crate::services::queue::ImageTask::Thumbnail {
                photo_id: photo.id,
                data: file_data.clone(),
                date_dir: date_path.clone(),
                stored_filename: stored_filename.clone(),
                upload_root: upload_root.clone(),
                orig_width: w as u32,
                orig_height: h as u32,
                public_url: public_url.clone(),
                driver: remote_driver,
            };
            if let Err(e) = queue.enqueue(task).await {
                tracing::warn!("投递缩略图任务失败: {}", e);
            }

            // 如果启用水印，同时入队水印任务
            if let Some(wm_config) = &state.config.watermark {
                if wm_config.enabled {
                    let wm_text = resolve_watermark_text(wm_config, &state.config.app.name);
                    let wm_task = crate::services::queue::ImageTask::Watermark {
                        photo_id: photo.id,
                        data: file_data.clone(),
                        date_dir: date_path.clone(),
                        stored_filename: stored_filename.clone(),
                        upload_root: upload_root.clone(),
                        public_url: public_url.clone(),
                        watermark_text: wm_text,
                        config: wm_config.clone(),
                        driver: if is_local_driver {
                            None
                        } else {
                            Some(driver.clone())
                        },
                    };
                    if let Err(e) = queue.enqueue(wm_task).await {
                        tracing::warn!("投递水印任务失败: {}", e);
                    }
                }
            }
        }
    }

    Ok(Json(ApiResponse::success(photo)))
}

/// 解析水印文本变量
///
/// 支持变量：
/// - `{site_name}` → 站点名称
fn resolve_watermark_text(config: &crate::config::WatermarkConfig, site_name: &str) -> String {
    config.text.replace("{site_name}", site_name)
}

/// 更新图片
#[utoipa::path(
    patch,
    path = "/api/v1/photos/:id",
    params(
        ("id" = i64, Path, description = "图片ID"),
    ),
    request_body = crate::models::photo::UpdatePhotoRequest,
    responses(
        (status = 200, description = "成功", body = PhotoResponse),
    ),
    tag = "图片"
)]
pub async fn update(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Path(photo_id): Path<i64>,
    Json(req): Json<crate::models::photo::UpdatePhotoRequest>,
) -> AppResult<Json<ApiResponse<PhotoResponse>>> {
    let photo = state
        .photo_svc
        .update(user_id, photo_id, req.album_id, req.is_public)
        .await?;
    Ok(Json(ApiResponse::success(photo)))
}

/// 删除图片
#[utoipa::path(
    delete,
    path = "/api/v1/photos/:id",
    params(
        ("id" = i64, Path, description = "图片ID"),
    ),
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "图片"
)]
pub async fn delete(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Path(photo_id): Path<i64>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    state.photo_svc.delete(user_id, photo_id).await?;
    Ok(Json(ApiResponse::success_with_message(
        serde_json::json!({}),
        "删除成功",
    )))
}

/// 批量删除
#[utoipa::path(
    post,
    path = "/api/v1/photos/batch-delete",
    request_body = BatchIdsRequest,
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "图片"
)]
pub async fn batch_delete(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Json(req): Json<BatchIdsRequest>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    let affected = state.photo_svc.batch_delete(user_id, &req.ids).await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "affected": affected }),
    )))
}

/// 批量更新
#[utoipa::path(
    patch,
    path = "/api/v1/photos/batch-update",
    request_body = BatchUpdateRequest,
    responses(
        (status = 200, description = "成功", body = serde_json::Value),
    ),
    tag = "图片"
)]
pub async fn batch_update(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Json(req): Json<BatchUpdateRequest>,
) -> AppResult<Json<ApiResponse<serde_json::Value>>> {
    let affected = state.photo_svc.batch_update(user_id, &req).await?;
    Ok(Json(ApiResponse::success(
        serde_json::json!({ "affected": affected }),
    )))
}

/// 移动到相册
#[utoipa::path(
    post,
    path = "/api/v1/photos/:id/move-to-album",
    params(
        ("id" = i64, Path, description = "图片ID"),
    ),
    request_body = MoveToAlbumRequest,
    responses(
        (status = 200, description = "成功", body = PhotoResponse),
    ),
    tag = "图片"
)]
pub async fn move_to_album(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Path(photo_id): Path<i64>,
    Json(req): Json<MoveToAlbumRequest>,
) -> AppResult<Json<ApiResponse<PhotoResponse>>> {
    let photo = state
        .photo_svc
        .move_to_album(user_id, photo_id, req.album_id)
        .await?;
    Ok(Json(ApiResponse::success(photo)))
}

/// 复制图片
#[utoipa::path(
    post,
    path = "/api/v1/photos/:id/copy",
    params(
        ("id" = i64, Path, description = "图片ID"),
    ),
    request_body = CopyPhotoRequest,
    responses(
        (status = 200, description = "成功", body = UploadResponse),
    ),
    tag = "图片"
)]
pub async fn copy(
    State(state): State<AppState>,
    CurrentUser { user_id, .. }: CurrentUser,
    Path(photo_id): Path<i64>,
    Json(req): Json<CopyPhotoRequest>,
) -> AppResult<Json<ApiResponse<crate::dto::photo::UploadResponse>>> {
    let photo = state
        .photo_svc
        .copy(user_id, photo_id, req.album_id)
        .await?;
    Ok(Json(ApiResponse::success(photo)))
}

/// 公开探索列表
#[utoipa::path(
    get,
    path = "/api/v1/public/photos",
    params(
        ("page" = u64, Query, description = "页码"),
        ("per_page" = u64, Query, description = "每页数量"),
    ),
    responses(
        (status = 200, description = "成功", body = PaginatedData<PhotoPublicResponse>),
    ),
    tag = "图片"
)]
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
