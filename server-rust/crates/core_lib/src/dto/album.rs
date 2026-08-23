//! 相册域 DTO

use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use sqlx::FromRow;
use utoipa::ToSchema;
use validator::Validate;

/// 相册响应
#[derive(Debug, Clone, Serialize, FromRow, ToSchema)]
pub struct AlbumResponse {
    pub id: i64,
    pub uuid: String,
    pub user_id: i64,
    pub name: String,
    pub description: Option<String>,
    pub cover_photo_id: Option<i64>,
    pub is_public: bool,
    pub photo_count: i64,
    pub views: i64,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

/// 创建相册请求
#[derive(Debug, Clone, Deserialize, Validate, ToSchema)]
pub struct CreateAlbumRequest {
    #[validate(length(min = 1, max = 100, message = "相册名称长度必须在 1-100 之间"))]
    pub name: String,
    pub description: Option<String>,
    pub is_public: Option<bool>,
}

/// 更新相册请求
#[derive(Debug, Clone, Deserialize, Validate, Default, ToSchema)]
pub struct UpdateAlbumRequest {
    #[validate(length(min = 1, max = 100, message = "相册名称长度必须在 1-100 之间"))]
    pub name: Option<String>,
    pub description: Option<String>,
    pub cover_photo_id: Option<i64>,
    pub is_public: Option<bool>,
}

/// 添加图片到相册请求
#[derive(Debug, Clone, Deserialize, ToSchema)]
pub struct AddPhotoToAlbumRequest {
    pub photo_ids: Vec<i64>,
}
