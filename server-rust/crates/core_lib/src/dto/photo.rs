//! 图片域 DTO

use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use sqlx::FromRow;

/// 图片响应（不含敏感字段 path/md5/sha1）
#[derive(Debug, Clone, Serialize, FromRow, utoipa::ToSchema)]
pub struct PhotoResponse {
    pub id: i64,
    pub uuid: String,
    pub user_id: i64,
    pub album_id: Option<i64>,
    pub filename: String,
    pub original_name: String,
    pub url: String,
    pub thumbnail_url: Option<String>,
    pub size: i64,
    pub width: Option<i32>,
    pub height: Option<i32>,
    pub mime_type: String,
    pub exif: Option<String>,
    pub is_public: bool,
    pub views: i64,
    pub likes: i64,
    pub status: i32,
    pub expired_at: Option<DateTime<Utc>>,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

/// 图片公开信息（探索页）
#[derive(Debug, Clone, Serialize, FromRow, utoipa::ToSchema)]
pub struct PhotoPublicResponse {
    pub id: i64,
    pub uuid: String,
    pub username: String,
    pub url: String,
    pub thumbnail_url: Option<String>,
    pub width: Option<i32>,
    pub height: Option<i32>,
    pub size: i64,
    pub views: i64,
    pub likes: i64,
    pub created_at: DateTime<Utc>,
}

/// 批量操作请求
#[derive(Debug, Clone, Deserialize, utoipa::ToSchema)]
pub struct BatchIdsRequest {
    pub ids: Vec<i64>,
}

/// 批量更新请求
#[derive(Debug, Clone, Deserialize, Default, utoipa::ToSchema)]
pub struct BatchUpdateRequest {
    pub ids: Vec<i64>,
    pub album_id: Option<i64>,
    pub is_public: Option<bool>,
}

/// 移动到相册请求
#[derive(Debug, Clone, Deserialize, utoipa::ToSchema)]
pub struct MoveToAlbumRequest {
    pub album_id: i64,
}

/// 复制图片请求
#[derive(Debug, Clone, Deserialize, utoipa::ToSchema)]
pub struct CopyPhotoRequest {
    pub album_id: Option<i64>,
}

/// 上传结果响应
#[derive(Debug, Clone, Serialize, utoipa::ToSchema)]
pub struct UploadResponse {
    pub id: i64,
    pub uuid: String,
    pub url: String,
    pub thumbnail_url: Option<String>,
    pub size: i64,
    pub width: Option<i32>,
    pub height: Option<i32>,
    pub filename: String,
}
